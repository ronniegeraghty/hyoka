# Evaluation Summary: 20260320-231430

## Run Statistics

| Metric | Value |
|--------|-------|
| Run ID | `20260320-231430` |
| Timestamp | 2026-03-21T06:14:30Z |
| Total Prompts | 3 |
| Total Configs | 2 |
| Total Evaluations | 6 |
| Passed | 4 |
| Failed | 2 |
| Errors | 0 |
| Duration | 276.7s |

## Comparison Matrix

| Prompt | baseline | azure-mcp |
|--------|--------|--------|
| key-vault-dp-python-error-handling | ✅ | ✅ |
| key-vault-dp-python-crud | ✅ | ✅ |
| key-vault-dp-python-pagination | ❌ | ❌ |

## Detailed Results

| Prompt | Config | Result | Score | Duration | Files |
|--------|--------|--------|-------|----------|-------|
| [key-vault-dp-python-error-handling](results/key-vault/data-plane/python/error-handling/baseline/report.md) | baseline | ✅ | — | 80.5s | 1 |
| [key-vault-dp-python-crud](results/key-vault/data-plane/python/crud/azure-mcp/report.md) | azure-mcp | ✅ | — | 98.5s | 3 |
| [key-vault-dp-python-crud](results/key-vault/data-plane/python/crud/baseline/report.md) | baseline | ✅ | — | 110.3s | 4 |
| [key-vault-dp-python-pagination](results/key-vault/data-plane/python/pagination/azure-mcp/report.md) | azure-mcp | ❌ | — | 152.4s | 4 |
| [key-vault-dp-python-error-handling](results/key-vault/data-plane/python/error-handling/azure-mcp/report.md) | azure-mcp | ✅ | — | 78.3s | 1 |
| [key-vault-dp-python-pagination](results/key-vault/data-plane/python/pagination/baseline/report.md) | baseline | ❌ | — | 178.1s | 6 |

