#!/usr/bin/env python3
"""
Run doc-agent evaluate against all (or filtered) prompts in the repo.

Evaluations run in parallel by default for faster execution.

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
"""

import argparse
import datetime
import os
import shutil
import subprocess
import sys
import threading
import time
from concurrent.futures import ProcessPoolExecutor, as_completed
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


def run_single_eval_worker(prompt_id, prompt_text, tmp_output_dir, report_subdir,
                           model, timeout, verbose):
    """
    Run one eval in its own subprocess. Designed to be called from
    ProcessPoolExecutor — all arguments are picklable primitives/Paths.

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

    # Reorganize output into the final report directory
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


def _print_summary_table(results, total, elapsed_secs):
    """Print a final summary table of all evaluation results."""
    pass_count = sum(1 for r in results if r["status"] == "pass")
    fail_count = sum(1 for r in results if r["status"] == "fail")
    skip_count = sum(1 for r in results if r["status"] == "skipped")

    mins, secs = divmod(int(elapsed_secs), 60)

    print("\n" + "=" * 72)
    print("  EVALUATION SUMMARY")
    print("=" * 72)
    print(f"  Total: {total}   ✅ Passed: {pass_count}   "
          f"❌ Failed: {fail_count}   ⏭️  Skipped: {skip_count}")
    print(f"  Elapsed: {mins}m {secs}s")
    print("-" * 72)
    print(f"  {'ID':<45} {'Status':<10}")
    print("-" * 72)

    for r in results:
        status = r["status"]
        icon = {"pass": "✅", "fail": "❌", "skipped": "⏭️"}.get(status, "?")
        print(f"  {r['prompt_id']:<45} {icon} {status}")

    if fail_count:
        print("\n" + "-" * 72)
        print("  FAILURES:")
        print("-" * 72)
        for r in results:
            if r["status"] == "fail":
                err = r.get("error", "unknown error")
                print(f"  ❌ {r['prompt_id']}")
                for line in err.strip().splitlines()[:5]:
                    print(f"     {line}")

    print("=" * 72)


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
    print(f"Found {total} prompt(s) to evaluate — running {mode}")

    if args.dry_run:
        for p in prompts:
            print(f"  [{p['id']}] {p['path']}")
        return

    # Create timestamped run directory
    timestamp = datetime.datetime.utcnow().strftime("%Y-%m-%dT%H-%M-%SZ")
    run_dir = REPORTS_DIR / timestamp
    run_dir.mkdir(parents=True, exist_ok=True)

    # Thread-safe print lock for progress output
    print_lock = threading.Lock()
    completed_count = 0
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
            with print_lock:
                print(f"  ⏭️  Skipped {prompt_meta['id']} (no ## Prompt section)")
            continue

        prompt_rel = (prompt_meta["path"]
                      .replace("prompts/", "")
                      .replace(".prompt.md", ""))
        report_subdir = run_dir / prompt_rel
        # Each worker gets its own temp dir to avoid collisions
        tmp_dir = run_dir / "_tmp_eval" / prompt_meta["id"]

        work_items.append({
            "prompt_id": prompt_meta["id"],
            "prompt_text": prompt_text,
            "tmp_dir": str(tmp_dir),
            "report_subdir": str(report_subdir),
        })

    results = list(skip_results)

    if work_items:
        with ProcessPoolExecutor(max_workers=args.workers) as executor:
            future_to_id = {}
            for item in work_items:
                future = executor.submit(
                    run_single_eval_worker,
                    prompt_id=item["prompt_id"],
                    prompt_text=item["prompt_text"],
                    tmp_output_dir=item["tmp_dir"],
                    report_subdir=item["report_subdir"],
                    model=args.model,
                    timeout=args.timeout,
                    verbose=args.verbose,
                )
                future_to_id[future] = item["prompt_id"]

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
                completed_count = sum(1 for r in results
                                      if r["status"] != "skipped")
                icon = "✅" if entry["status"] == "pass" else "❌"
                with print_lock:
                    print(f"  {icon} [{completed_count}/{len(work_items)}] "
                          f"{entry['prompt_id']} — {entry['status']}")

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
    print(f"Reports: {run_dir}")


if __name__ == "__main__":
    main()
