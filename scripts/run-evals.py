#!/usr/bin/env python3
"""
Run doc-agent evaluate against all (or filtered) prompts in the repo.

Evaluations run in parallel by default with inline-updating progress bars.

Usage:
  python scripts/run-evals.py                                    # All prompts (parallel)
  python scripts/run-evals.py --workers 10                       # 10 parallel workers
  python scripts/run-evals.py --sequential                       # One at a time
  python scripts/run-evals.py --service storage                  # All Storage prompts
  python scripts/run-evals.py --language dotnet                  # All .NET prompts
  python scripts/run-evals.py --service storage --language dotnet # Storage + .NET
  python scripts/run-evals.py --category authentication          # All auth prompts
  python scripts/run-evals.py --plane data-plane                 # All data-plane prompts
  python scripts/run-evals.py --tags identity                    # Filter by tag
  python scripts/run-evals.py --prompt-id storage-dp-dotnet-auth # Single by ID
  python scripts/run-evals.py --prompt prompts/storage/.../x.prompt.md  # Single by path
  python scripts/run-evals.py --service storage --dry-run        # List without running
  python scripts/run-evals.py --debug                            # Stream subprocess output
"""

import argparse
import datetime
import os
import shutil
import subprocess
import sys
import threading
import time
from concurrent.futures import ProcessPoolExecutor, ThreadPoolExecutor, as_completed
from pathlib import Path

try:
    import yaml
except ImportError:
    print("ERROR: PyYAML is required. Install with: pip install pyyaml")
    sys.exit(1)

REPO_ROOT = Path(__file__).resolve().parent.parent
PROMPTS_DIR = REPO_ROOT / "prompts"
REPORTS_DIR = REPO_ROOT / "reports" / "runs"
MANIFEST_PATH = REPO_ROOT / "manifest.yaml"

# doc-agent evaluate output stages (ordered by file creation)
EVAL_STAGES = [
    ("task.md",        "Planning",   1),
    ("execution.log",  "Executing",  3),
    ("observations.md","Observing",  5),
    ("report.html",    "Reporting",  7),
    ("workspace",      "Complete",   8),
]
STAGE_TOTAL = 8


def load_manifest():
    """Load the central manifest."""
    if not MANIFEST_PATH.exists():
        print(f"ERROR: Manifest not found at {MANIFEST_PATH}")
        print("Run: python scripts/generate-manifest.py")
        sys.exit(1)
    with open(MANIFEST_PATH) as f:
        return yaml.safe_load(f)


def filter_prompts(manifest, args):
    """Apply all filter flags. Filters compose with AND logic."""
    prompts = manifest.get("prompts", [])

    for key in ["service", "language", "plane", "category"]:
        val = getattr(args, key, None)
        if val:
            prompts = [p for p in prompts if p.get(key) == val]

    if args.tags:
        prompts = [p for p in prompts if args.tags in p.get("tags", [])]

    if args.prompt_id:
        prompts = [p for p in prompts if p["id"] == args.prompt_id]

    if args.prompt:
        normalized = args.prompt.replace("\\", "/")
        prompts = [p for p in prompts if p["path"] == normalized]

    return prompts


def extract_prompt_text(prompt_path):
    """Extract the prompt text from the ## Prompt section of the markdown file."""
    content = Path(prompt_path).read_text(encoding="utf-8")
    in_prompt = False
    lines = []
    for line in content.split("\n"):
        if line.strip().startswith("## Prompt"):
            in_prompt = True
            continue
        if in_prompt and line.strip().startswith("## "):
            break
        if in_prompt:
            lines.append(line)
    return "\n".join(lines).strip()


def _build_eval_cmd(prompt_text, output_dir, model=None, timeout=None, verbose=False):
    """Build the doc-agent command list."""
    cmd = ["doc-agent", "evaluate", prompt_text, "-o", str(output_dir)]
    if model:
        cmd += ["-m", model]
    if timeout:
        cmd += ["-t", str(timeout)]
    if verbose:
        cmd.append("-v")
    return cmd


# ---------------------------------------------------------------------------
# Stage detection by polling output directories
# ---------------------------------------------------------------------------

def _detect_stage(tmp_output_dir):
    """
    Detect the current stage of a doc-agent evaluate run by checking
    which output files exist. Returns (stage_label, progress_filled).
    """
    tmp = Path(tmp_output_dir)
    # doc-agent writes to eval-<timestamp>-<slug>/ inside the output dir
    eval_dirs = sorted(tmp.glob("eval-*"))
    if not eval_dirs:
        return ("Starting", 0)

    d = eval_dirs[-1]
    label, filled = "Starting", 0
    for filename, stage_label, stage_filled in EVAL_STAGES:
        if (d / filename).exists():
            label, filled = stage_label, stage_filled
    return (label, filled)


def _progress_bar(filled, total=STAGE_TOTAL, width=8):
    """Render a block-style progress bar like ▓▓▓░░░░░."""
    done = int(width * filled / total) if total else 0
    return "▓" * done + "░" * (width - done)


# ---------------------------------------------------------------------------
# Worker functions
# ---------------------------------------------------------------------------

def run_single_eval_worker(prompt_id, prompt_text, tmp_output_dir, report_subdir,
                           model, timeout, verbose, eval_number=0, total=0):
    """
    Run one eval in its own subprocess (ProcessPoolExecutor compatible).
    Returns a result dict with prompt_id, status, report_path, stdout, stderr.
    """
    tmp_output = Path(tmp_output_dir)
    tmp_output.mkdir(parents=True, exist_ok=True)

    cmd = _build_eval_cmd(prompt_text, tmp_output, model, timeout, verbose)
    effective_timeout = timeout or 3600

    try:
        proc = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            timeout=effective_timeout,
        )
        exit_code = proc.returncode
        stdout = proc.stdout
        stderr = proc.stderr
    except FileNotFoundError:
        exit_code = 127
        stdout = ""
        stderr = ("ERROR: doc-agent command not found. "
                  "Install from https://github.com/coreai-microsoft/doc-review-agent")
    except subprocess.TimeoutExpired:
        exit_code = 124
        stdout = ""
        stderr = f"ERROR: Evaluation timed out after {effective_timeout}s"

    target = Path(report_subdir)
    _reorganize_eval_output(tmp_output, target)

    status = "pass" if exit_code == 0 else "fail"
    entry = {
        "prompt_id": prompt_id,
        "status": status,
        "report_path": str(target.name),
        "stdout": stdout,
        "stderr": stderr,
    }
    if exit_code != 0:
        entry["error"] = stderr[:500]
    return entry


def _run_debug_eval_worker(prompt_id, prompt_text, tmp_output_dir, report_subdir,
                           model, timeout, verbose, print_lock,
                           eval_number=0, total=0):
    """
    Run one eval using Popen and stream stdout/stderr line-by-line,
    prefixed with the eval ID.  Used in --debug mode.
    """
    if eval_number and total:
        with print_lock:
            print(f"▶️  [{eval_number}/{total}] Starting: {prompt_id}", flush=True)

    tmp_output = Path(tmp_output_dir)
    tmp_output.mkdir(parents=True, exist_ok=True)

    cmd = _build_eval_cmd(prompt_text, tmp_output, model, timeout, verbose)
    effective_timeout = timeout or 3600

    stdout_lines, stderr_lines = [], []
    exit_code = 0

    try:
        import os as _os
        env = _os.environ.copy()
        env["PYTHONUNBUFFERED"] = "1"
        proc = subprocess.Popen(
            cmd,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            bufsize=1,
            env=env,
        )

        def _stream(pipe, collector, label):
            for raw_line in pipe:
                line = raw_line.rstrip("\n")
                collector.append(line)
                with print_lock:
                    print(f"  [{prompt_id}] {line}", flush=True)

        t_out = threading.Thread(target=_stream,
                                 args=(proc.stdout, stdout_lines, "out"))
        t_err = threading.Thread(target=_stream,
                                 args=(proc.stderr, stderr_lines, "err"))
        t_out.start()
        t_err.start()

        proc.wait(timeout=effective_timeout)
        t_out.join()
        t_err.join()

        exit_code = proc.returncode
    except FileNotFoundError:
        exit_code = 127
        stderr_lines.append(
            "ERROR: doc-agent command not found. "
            "Install from https://github.com/coreai-microsoft/doc-review-agent")
    except subprocess.TimeoutExpired:
        proc.kill()
        proc.wait()
        exit_code = 124
        stderr_lines.append(f"ERROR: Evaluation timed out after {effective_timeout}s")

    stdout = "\n".join(stdout_lines)
    stderr = "\n".join(stderr_lines)

    target = Path(report_subdir)
    _reorganize_eval_output(tmp_output, target)

    status = "pass" if exit_code == 0 else "fail"
    entry = {
        "prompt_id": prompt_id,
        "status": status,
        "report_path": str(target.name),
        "stdout": stdout,
        "stderr": stderr,
    }
    if exit_code != 0:
        entry["error"] = stderr[:500]
    return entry


def _reorganize_eval_output(raw_output_dir, target_dir):
    """
    doc-agent writes to output/eval-<timestamp>-<slug>/.
    Move those files into our structured report directory.
    """
    raw_output_dir = Path(raw_output_dir)
    eval_dirs = sorted(raw_output_dir.glob("eval-*"))
    if not eval_dirs:
        return

    src = eval_dirs[-1]
    target_dir = Path(target_dir)
    target_dir.mkdir(parents=True, exist_ok=True)

    for item in src.iterdir():
        dest = target_dir / item.name
        if item.is_dir():
            shutil.copytree(str(item), str(dest), dirs_exist_ok=True)
        else:
            shutil.move(str(item), str(dest))

    shutil.rmtree(str(src), ignore_errors=True)


def _default_workers():
    """Pick a sensible default worker count."""
    cpus = os.cpu_count() or 4
    return min(cpus, 8)


# ---------------------------------------------------------------------------
# Progress display
# ---------------------------------------------------------------------------

class ProgressDisplay:
    """
    Inline-updating progress display using ANSI escape codes.

    Manages N active eval lines + 1 summary line at the bottom.
    When an eval completes, its line is updated to a final status and
    frozen; new evals are appended below.

    Falls back to simple line-by-line output when stdout is not a TTY.
    """

    def __init__(self, total_work, workers, use_ansi=True):
        self._lock = threading.Lock()
        self._total = total_work
        self._workers = workers
        self._use_ansi = use_ansi and sys.stdout.isatty()

        # Tracking state
        self._completed = 0
        self._failed = 0
        self._active = {}       # prompt_id → {num, start, tmp_dir, line_idx}
        self._line_count = 0    # how many display lines we've printed
        self._finished_ids = set()

    # -- ANSI helpers -------------------------------------------------------

    def _move_up(self, n):
        if n > 0:
            sys.stdout.write(f"\033[{n}A")

    def _clear_line(self):
        sys.stdout.write("\r\033[K")

    def _write_at(self, line_idx, text):
        """Overwrite a specific line (0-indexed from top of our display block)."""
        up = self._line_count - line_idx
        self._move_up(up)
        self._clear_line()
        sys.stdout.write(text)
        # Move back to the bottom
        down = up
        if down > 0:
            sys.stdout.write(f"\n\033[{down - 1}B" if down > 1 else "\n")
        sys.stdout.flush()

    # -- Public API ---------------------------------------------------------

    def eval_started(self, prompt_id, eval_num, tmp_dir):
        """Called when an eval is submitted to the executor."""
        with self._lock:
            line_idx = self._line_count
            self._active[prompt_id] = {
                "num": eval_num,
                "start": time.monotonic(),
                "tmp_dir": tmp_dir,
                "line_idx": line_idx,
            }
            self._line_count += 1  # the eval line
            # Print the initial line
            if self._use_ansi:
                text = self._format_active_line(prompt_id)
                sys.stdout.write(text + "\n")
                sys.stdout.flush()
                self._refresh_summary()
            else:
                num = self._active[prompt_id]["num"]
                print(f"▶️  [{num}/{self._total}] Starting: {prompt_id}",
                      flush=True)

    def eval_completed(self, prompt_id, entry):
        """Called when an eval finishes (pass or fail)."""
        with self._lock:
            self._completed += 1
            if entry["status"] == "fail":
                self._failed += 1

            info = self._active.pop(prompt_id, None)
            self._finished_ids.add(prompt_id)

            if info is None:
                return

            elapsed = int(time.monotonic() - info["start"])
            num = info["num"]

            if entry["status"] == "pass":
                icon, word = "✅", "Passed"
            else:
                icon, word = "❌", "Failed"

            if self._use_ansi:
                line = f"  [{num}/{self._total}] {prompt_id:<35} {icon} {word} ({elapsed}s)"
                self._write_at(info["line_idx"], line)
                self._refresh_summary()
            else:
                print(f"{icon} [{num}/{self._total}] {word}: {prompt_id} ({elapsed}s)",
                      flush=True)
                self._print_plain_summary()

    def refresh_stages(self):
        """Poll output dirs and update stage display for active evals."""
        if not self._use_ansi:
            return
        with self._lock:
            for pid, info in list(self._active.items()):
                text = self._format_active_line(pid)
                self._write_at(info["line_idx"], text)

    def finalize(self):
        """Move cursor below our display block so summary table prints cleanly."""
        if self._use_ansi:
            # Make sure we're past the summary line
            sys.stdout.write("\n")
            sys.stdout.flush()

    # -- Internal -----------------------------------------------------------

    def _format_active_line(self, prompt_id):
        info = self._active.get(prompt_id)
        if not info:
            return ""
        num = info["num"]
        elapsed = int(time.monotonic() - info["start"])
        stage_label, stage_filled = _detect_stage(info["tmp_dir"])
        bar = _progress_bar(stage_filled)
        return f"  [{num}/{self._total}] {prompt_id:<35} {bar} {stage_label:<12} {elapsed}s"

    def _refresh_summary(self):
        """Ensure summary line exists at the bottom and update it."""
        # The summary line is always at self._line_count
        running = len(self._active)
        queued = max(0, self._total - self._completed - running)
        text = (f"  📊 {self._completed}/{self._total} complete | "
                f"{running} running | {queued} queued | "
                f"{self._failed} failed")

        # If summary line hasn't been allocated yet, add it
        if not hasattr(self, "_summary_line"):
            self._summary_line = self._line_count
            self._line_count += 1
            sys.stdout.write(text + "\n")
            sys.stdout.flush()
        else:
            self._write_at(self._summary_line, text)

    def _print_plain_summary(self):
        running = len(self._active)
        queued = max(0, self._total - self._completed - running)
        print(f"📊 {self._completed}/{self._total} complete | "
              f"{running} running | {queued} queued | "
              f"{self._failed} failed", flush=True)


def _print_summary_table(results, total, elapsed_secs):
    """Print a final summary table of all evaluation results."""
    pass_count = sum(1 for r in results if r["status"] == "pass")
    fail_count = sum(1 for r in results if r["status"] == "fail")
    skip_count = sum(1 for r in results if r["status"] == "skipped")

    mins, secs = divmod(int(elapsed_secs), 60)

    print("\n" + "=" * 72, flush=True)
    print("  EVALUATION SUMMARY", flush=True)
    print("=" * 72, flush=True)
    print(f"  Total: {total}   ✅ Passed: {pass_count}   "
          f"❌ Failed: {fail_count}   ⏭️  Skipped: {skip_count}", flush=True)
    print(f"  Elapsed: {mins}m {secs}s", flush=True)
    print("-" * 72, flush=True)
    print(f"  {'ID':<45} {'Status':<10}", flush=True)
    print("-" * 72, flush=True)

    for r in results:
        status = r["status"]
        icon = {"pass": "✅", "fail": "❌", "skipped": "⏭️"}.get(status, "?")
        print(f"  {r['prompt_id']:<45} {icon} {status}", flush=True)

    if fail_count:
        print("\n" + "-" * 72, flush=True)
        print("  FAILURES:", flush=True)
        print("-" * 72, flush=True)
        for r in results:
            if r["status"] == "fail":
                err = r.get("error", "unknown error")
                print(f"  ❌ {r['prompt_id']}", flush=True)
                for line in err.strip().splitlines()[:5]:
                    print(f"     {line}", flush=True)

    print("=" * 72, flush=True)


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main():
    parser = argparse.ArgumentParser(
        description="Run doc-agent evaluate against prompts.",
        epilog="No arguments = run ALL prompts in parallel. Flags compose with AND logic.",
    )

    # Filter flags
    parser.add_argument("--service", help="Filter by service (e.g., storage, key-vault)")
    parser.add_argument("--language", help="Filter by language (e.g., dotnet, python)")
    parser.add_argument("--plane", help="Filter by plane (data-plane, management-plane)")
    parser.add_argument("--category", help="Filter by category (e.g., authentication, crud)")
    parser.add_argument("--tags", help="Filter by tag")
    parser.add_argument("--prompt-id", help="Run a single prompt by its ID")
    parser.add_argument("--prompt", help="Run a single prompt by file path")

    # Execution options
    parser.add_argument("--dry-run", action="store_true",
                        help="List matching prompts without running evaluations")
    parser.add_argument("--model", "-m", help="Override doc-agent model")
    parser.add_argument("--timeout", "-t", type=int,
                        help="Timeout per evaluation in seconds (default: 3600)")
    parser.add_argument("--verbose", "-v", action="store_true",
                        help="Verbose doc-agent output")
    parser.add_argument("--debug", action="store_true",
                        help="Stream doc-agent stdout/stderr with eval-ID prefixes "
                             "(disables progress bars)")

    # Parallelism options
    parser.add_argument("--workers", "-w", type=int, default=_default_workers(),
                        help=f"Number of parallel workers (default: {_default_workers()})")
    parser.add_argument("--sequential", action="store_true",
                        help="Run evaluations one at a time (equivalent to --workers 1)")

    args = parser.parse_args()

    if args.sequential:
        args.workers = 1

    manifest = load_manifest()
    prompts = filter_prompts(manifest, args)

    if not prompts:
        print("No prompts matched the given filters.")
        sys.exit(1)

    total = len(prompts)
    mode = "sequentially" if args.workers == 1 else f"in parallel ({args.workers} workers)"
    print(f"Found {total} prompt(s) to evaluate — running {mode}", flush=True)

    if args.dry_run:
        for p in prompts:
            print(f"  [{p['id']}] {p['path']}")
        return

    # Create timestamped run directory
    timestamp = datetime.datetime.utcnow().strftime("%Y-%m-%dT%H-%M-%SZ")
    run_dir = REPORTS_DIR / timestamp
    run_dir.mkdir(parents=True, exist_ok=True)

    print_lock = threading.Lock()
    start_time = time.monotonic()

    # Pre-extract prompt texts and build work items
    work_items = []
    skip_results = []
    for prompt_meta in prompts:
        prompt_path = REPO_ROOT / prompt_meta["path"]
        prompt_text = extract_prompt_text(prompt_path)

        if not prompt_text:
            skip_results.append({
                "prompt_id": prompt_meta["id"],
                "status": "skipped",
                "report_path": "",
                "error": "No ## Prompt section found",
            })
            print(f"  ⏭️  Skipped {prompt_meta['id']} (no ## Prompt section)", flush=True)
            continue

        prompt_rel = (prompt_meta["path"]
                      .replace("prompts/", "")
                      .replace(".prompt.md", ""))
        report_subdir = run_dir / prompt_rel
        tmp_dir = run_dir / "_tmp_eval" / prompt_meta["id"]

        work_items.append({
            "prompt_id": prompt_meta["id"],
            "prompt_text": prompt_text,
            "tmp_dir": str(tmp_dir),
            "report_subdir": str(report_subdir),
        })

    results = list(skip_results)
    total_work = len(work_items)

    if work_items:
        print(f"\n🚀 Starting evaluation: {total_work} prompts with {args.workers} workers\n",
              flush=True)

        if args.debug:
            _run_debug_mode(args, work_items, total_work, results, print_lock)
        else:
            _run_progress_mode(args, work_items, total_work, results)

    # Clean up all temp dirs
    tmp_root = run_dir / "_tmp_eval"
    if tmp_root.exists():
        shutil.rmtree(str(tmp_root), ignore_errors=True)

    # Gather active filters for metadata
    active_filters = {}
    for key in ["service", "language", "plane", "category", "tags",
                "prompt_id", "prompt"]:
        val = getattr(args, key.replace("-", "_"), None)
        if val:
            active_filters[key] = val

    pass_count = sum(1 for r in results if r["status"] == "pass")
    fail_count = sum(1 for r in results if r["status"] == "fail")
    skip_count = sum(1 for r in results if r["status"] == "skipped")

    # Strip stdout/stderr from persisted metadata (keep reports lean)
    clean_results = []
    for r in results:
        clean = {k: v for k, v in r.items() if k not in ("stdout", "stderr")}
        clean_results.append(clean)

    elapsed = time.monotonic() - start_time
    run_meta = {
        "timestamp": timestamp,
        "prompt_count": total,
        "pass_count": pass_count,
        "fail_count": fail_count,
        "skip_count": skip_count,
        "workers": args.workers,
        "elapsed_seconds": round(elapsed, 1),
        "filters": active_filters if active_filters else "none (all prompts)",
        "results": clean_results,
    }

    with open(run_dir / "run-metadata.yaml", "w") as f:
        yaml.dump(run_meta, f, default_flow_style=False, sort_keys=False)

    # Update latest symlink
    latest = REPORTS_DIR / "latest"
    if latest.is_symlink() or latest.exists():
        latest.unlink()
    os.symlink(timestamp, str(latest))

    _print_summary_table(results, total, elapsed)
    print(f"Reports: {run_dir}", flush=True)


def _run_progress_mode(args, work_items, total_work, results):
    """
    Run evals with inline-updating progress bars.
    Uses ProcessPoolExecutor + a background monitor thread for stage polling.
    """
    display = ProgressDisplay(total_work, args.workers, use_ansi=True)
    monitor_stop = threading.Event()

    def _monitor_loop():
        while not monitor_stop.is_set():
            display.refresh_stages()
            monitor_stop.wait(2.0)

    monitor = threading.Thread(target=_monitor_loop, daemon=True)
    monitor.start()

    eval_start_times = {}

    with ProcessPoolExecutor(max_workers=args.workers) as executor:
        future_to_id = {}
        for idx, item in enumerate(work_items, 1):
            pid = item["prompt_id"]
            eval_start_times[pid] = time.monotonic()
            display.eval_started(pid, idx, item["tmp_dir"])

            future = executor.submit(
                run_single_eval_worker,
                prompt_id=pid,
                prompt_text=item["prompt_text"],
                tmp_output_dir=item["tmp_dir"],
                report_subdir=item["report_subdir"],
                model=args.model,
                timeout=args.timeout,
                verbose=args.verbose,
            )
            future_to_id[future] = pid

        for future in as_completed(future_to_id):
            prompt_id = future_to_id[future]
            try:
                entry = future.result()
            except Exception as exc:
                entry = {
                    "prompt_id": prompt_id,
                    "status": "fail",
                    "report_path": "",
                    "error": f"Worker exception: {exc}",
                    "stdout": "",
                    "stderr": str(exc),
                }

            results.append(entry)
            display.eval_completed(prompt_id, entry)

    monitor_stop.set()
    monitor.join(timeout=3)
    display.finalize()


def _run_debug_mode(args, work_items, total_work, results, print_lock):
    """
    Run evals in --debug mode: stream subprocess output with eval-ID prefixes.
    Uses ThreadPoolExecutor so Popen streaming works in-process.
    """
    eval_start_times = {}

    with ThreadPoolExecutor(max_workers=args.workers) as executor:
        future_to_id = {}
        for idx, item in enumerate(work_items, 1):
            pid = item["prompt_id"]
            eval_start_times[pid] = time.monotonic()
            future = executor.submit(
                _run_debug_eval_worker,
                prompt_id=pid,
                eval_number=idx,
                total=total_work,
                prompt_text=item["prompt_text"],
                tmp_output_dir=item["tmp_dir"],
                report_subdir=item["report_subdir"],
                model=args.model,
                timeout=args.timeout,
                verbose=args.verbose,
                print_lock=print_lock,
            )
            future_to_id[future] = pid

        completed = 0
        for future in as_completed(future_to_id):
            prompt_id = future_to_id[future]
            try:
                entry = future.result()
            except Exception as exc:
                entry = {
                    "prompt_id": prompt_id,
                    "status": "fail",
                    "report_path": "",
                    "error": f"Worker exception: {exc}",
                    "stdout": "",
                    "stderr": str(exc),
                }

            results.append(entry)
            completed += 1
            elapsed_item = int(time.monotonic() - eval_start_times[prompt_id])
            icon = "✅" if entry["status"] == "pass" else "❌"
            word = "Passed" if entry["status"] == "pass" else "Failed"
            running = len(future_to_id) - completed
            with print_lock:
                print(f"{icon} {word}: {prompt_id} ({elapsed_item}s)  "
                      f"[{completed}/{total_work}]", flush=True)


if __name__ == "__main__":
    main()
