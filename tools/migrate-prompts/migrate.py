#!/usr/bin/env python3
"""Migrate prompt frontmatter from flat fields to nested properties: format.

Usage:
    python3 tools/migrate-prompts/migrate.py [--dry-run] [prompts_dir]

This script:
- Finds all .prompt.md files in the prompts directory
- Moves metadata fields (service, plane, language, etc.) into a nested properties: block
- Keeps id, tags, timeout, starter_project, and other non-string fields at top level
- Preserves the Markdown body below the frontmatter
"""

import os
import sys
import yaml
from pathlib import Path
from collections import OrderedDict

# Fields that should be moved into properties:
PROPERTY_FIELDS = [
    "service", "plane", "language", "category", "difficulty",
    "description", "sdk_package", "doc_url", "created", "author",
]

# Fields that stay at top level (besides 'id')
TOP_LEVEL_FIELDS = {
    "id", "tags", "timeout", "starter_project", "project_context",
    "reference_answer", "expected_packages", "expected_tools",
}


class OrderedDumper(yaml.SafeDumper):
    """YAML dumper that preserves dict insertion order."""
    pass


def _dict_representer(dumper, data):
    return dumper.represent_mapping(
        yaml.resolver.BaseResolver.DEFAULT_MAPPING_TAG,
        data.items(),
    )


OrderedDumper.add_representer(OrderedDict, _dict_representer)
OrderedDumper.add_representer(dict, _dict_representer)


def migrate_prompt_file(filepath, dry_run=False):
    """Migrate a single prompt file to nested properties format."""
    with open(filepath, "r") as f:
        content = f.read()

    if not content.startswith("---"):
        return False, "no frontmatter"

    parts = content.split("---", 2)
    if len(parts) < 3:
        return False, "malformed frontmatter"

    frontmatter_str = parts[1].strip()
    body = parts[2]

    data = yaml.safe_load(frontmatter_str)
    if not isinstance(data, dict):
        return False, "frontmatter is not a dict"

    if "properties" in data:
        return False, "already migrated"

    new_data = OrderedDict()

    if "id" in data:
        new_data["id"] = data["id"]

    properties = OrderedDict()
    for key in PROPERTY_FIELDS:
        if key in data:
            val = data[key]
            if hasattr(val, "isoformat"):
                val = val.isoformat()
            properties[key] = val
    if properties:
        new_data["properties"] = dict(properties)

    for key in data:
        if key == "id" or key in PROPERTY_FIELDS:
            continue
        new_data[key] = data[key]

    yaml_str = yaml.dump(
        dict(new_data),
        Dumper=OrderedDumper,
        default_flow_style=False,
        sort_keys=False,
        allow_unicode=True,
        width=120,
    )

    new_content = "---\n" + yaml_str + "---" + body

    if not dry_run:
        with open(filepath, "w") as f:
            f.write(new_content)
            f.flush()
            os.fsync(f.fileno())

    return True, "migrated"


def main():
    dry_run = "--dry-run" in sys.argv
    args = [a for a in sys.argv[1:] if not a.startswith("--")]
    prompts_dir = args[0] if args else "prompts"

    files = sorted(Path(prompts_dir).rglob("*.prompt.md"))
    print(f"Found {len(files)} prompt files in {prompts_dir}/")
    if dry_run:
        print("(dry-run mode — no files will be modified)\n")

    migrated = 0
    skipped = 0
    errors = 0

    for f in files:
        try:
            ok, reason = migrate_prompt_file(str(f), dry_run=dry_run)
            if ok:
                migrated += 1
                print(f"  ✓ {f}")
            else:
                skipped += 1
                if reason != "already migrated":
                    print(f"  ⊘ {f} ({reason})")
        except Exception as e:
            errors += 1
            print(f"  ✗ {f}: {e}")

    print(f"\nResults: {migrated} migrated, {skipped} skipped, {errors} errors")
    return 0 if errors == 0 else 1


if __name__ == "__main__":
    sys.exit(main())
