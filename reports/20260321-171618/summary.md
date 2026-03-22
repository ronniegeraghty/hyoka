# Evaluation Summary: 20260321-171618

## Run Statistics

| Metric | Value |
|--------|-------|
| Run ID | `20260321-171618` |
| Timestamp | 2026-03-22T00:16:18Z |
| Total Prompts | 3 |
| Total Configs | 2 |
| Total Evaluations | 6 |
| Passed | 5 |
| Failed | 1 |
| Errors | 0 |
| Duration | 204.0s |

## Comparison Matrix

| Prompt | azure-mcp | baseline |
|--------|--------|--------|
| key-vault-dp-python-error-handling | ✅ | ✅ |
| key-vault-dp-python-pagination | ❌ | ✅ |
| key-vault-dp-python-crud | ✅ | ✅ |

## Detailed Results

| Prompt | Config | Result | Score | Duration | Files |
|--------|--------|--------|-------|----------|-------|
| [key-vault-dp-python-error-handling](results/key-vault/data-plane/python/error-handling/azure-mcp/report.md) | azure-mcp | ✅ | — | 75.4s | 1 |
| [key-vault-dp-python-error-handling](results/key-vault/data-plane/python/error-handling/baseline/report.md) | baseline | ✅ | — | 76.9s | 1 |
| [key-vault-dp-python-pagination](results/key-vault/data-plane/python/pagination/baseline/report.md) | baseline | ✅ | — | 115.1s | 3 |
| [key-vault-dp-python-crud](results/key-vault/data-plane/python/crud/azure-mcp/report.md) | azure-mcp | ✅ | — | 56.4s | 3 |
| [key-vault-dp-python-crud](results/key-vault/data-plane/python/crud/baseline/report.md) | baseline | ✅ | — | 91.0s | 3 |
| [key-vault-dp-python-pagination](results/key-vault/data-plane/python/pagination/azure-mcp/report.md) | azure-mcp | ❌ | — | 204.0s | 7 |

