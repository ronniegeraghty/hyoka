# Evaluation Summary: 20260322-142536

## Run Statistics

| Metric | Value |
|--------|-------|
| Run ID | `20260322-142536` |
| Timestamp | 2026-03-22T21:25:36Z |
| Total Prompts | 3 |
| Total Configs | 2 |
| Total Evaluations | 6 |
| Passed | 5 |
| Failed | 1 |
| Errors | 0 |
| Duration | 159.5s |

## Comparison Matrix

| Prompt | baseline | azure-mcp |
|--------|--------|--------|
| key-vault-dp-python-error-handling | ❌ | ✅ |
| key-vault-dp-python-crud | ✅ | ✅ |
| key-vault-dp-python-pagination | ✅ | ✅ |

## Detailed Results

| Prompt | Config | Result | Score | Duration | Files |
|--------|--------|--------|-------|----------|-------|
| [key-vault-dp-python-error-handling](results/key-vault/data-plane/python/error-handling/baseline/report.md) | baseline | ❌ | — | 33.9s | 0 |
| [key-vault-dp-python-crud](results/key-vault/data-plane/python/crud/baseline/report.md) | baseline | ✅ | — | 60.9s | 3 |
| [key-vault-dp-python-error-handling](results/key-vault/data-plane/python/error-handling/azure-mcp/report.md) | azure-mcp | ✅ | — | 83.9s | 1 |
| [key-vault-dp-python-pagination](results/key-vault/data-plane/python/pagination/azure-mcp/report.md) | azure-mcp | ✅ | — | 103.4s | 3 |
| [key-vault-dp-python-crud](results/key-vault/data-plane/python/crud/azure-mcp/report.md) | azure-mcp | ✅ | — | 71.6s | 3 |
| [key-vault-dp-python-pagination](results/key-vault/data-plane/python/pagination/baseline/report.md) | baseline | ✅ | — | 98.5s | 3 |

