# Generator Skills

This directory holds skills loaded by the **generator agent** during evaluations.

## Adding Skills

Use [`npx skills add`](https://github.com/microsoft/skills) to install skills from a registry:

```bash
npx skills add microsoft/skills --directory skills/generator
```

This launches a wizard where you select which skills to install. Each skill gets its own subdirectory with a `SKILL.md` and optional `references/` folder.

## Example

To install the Java Key Vault Secrets skill:

```bash
npx skills add microsoft/skills --directory skills/generator
# Select "keyvault-secrets-java" from the wizard
```

Then reference this directory in your config:

```yaml
generator_skill_directories:
  - "./skills/generator"
```

See `configs/baseline-opus-skills.yaml` for a working example.
