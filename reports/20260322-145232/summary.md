# Evaluation Summary: 20260322-145232

## Run Statistics

| Metric | Value |
|--------|-------|
| Run ID | `20260322-145232` |
| Timestamp | 2026-03-22T21:52:32Z |
| Total Prompts | 3 |
| Total Configs | 2 |
| Total Evaluations | 6 |
| Passed | 4 |
| Failed | 2 |
| Errors | 0 |
| Duration | 169.9s |

## Comparison Matrix

| Prompt | baseline | azure-mcp |
|--------|--------|--------|
| key-vault-dp-python-error-handling | ❌ | ✅ |
| key-vault-dp-python-pagination | ❌ | ✅ |
| key-vault-dp-python-crud | ✅ | ✅ |

## Detailed Results

| Prompt | Config | Result | Score | Duration | Files |
|--------|--------|--------|-------|----------|-------|
| [key-vault-dp-python-error-handling](results/key-vault/data-plane/python/error-handling/baseline/report.md) | baseline | ❌ | — | 31.4s | 0 |
| [key-vault-dp-python-error-handling](results/key-vault/data-plane/python/error-handling/azure-mcp/report.md) | azure-mcp | ✅ | — | 86.2s | 1 |
| [key-vault-dp-python-pagination](results/key-vault/data-plane/python/pagination/baseline/report.md) | baseline | ❌ | — | 118.6s | 4 |
| [key-vault-dp-python-crud](results/key-vault/data-plane/python/crud/baseline/report.md) | baseline | ✅ | — | 94.1s | 4 |
| [key-vault-dp-python-pagination](results/key-vault/data-plane/python/pagination/azure-mcp/report.md) | azure-mcp | ✅ | — | 145.6s | 3 |
| [key-vault-dp-python-crud](results/key-vault/data-plane/python/crud/azure-mcp/report.md) | azure-mcp | ✅ | — | 83.7s | 3 |

