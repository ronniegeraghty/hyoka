# Evaluation Summary: 20260322-141817

## Run Statistics

| Metric | Value |
|--------|-------|
| Run ID | `20260322-141817` |
| Timestamp | 2026-03-22T21:18:17Z |
| Total Prompts | 3 |
| Total Configs | 2 |
| Total Evaluations | 6 |
| Passed | 5 |
| Failed | 1 |
| Errors | 0 |
| Duration | 158.1s |

## Comparison Matrix

| Prompt | baseline | azure-mcp |
|--------|--------|--------|
| key-vault-dp-python-crud | ✅ | ✅ |
| key-vault-dp-python-error-handling | ✅ | ✅ |
| key-vault-dp-python-pagination | ❌ | ✅ |

## Detailed Results

| Prompt | Config | Result | Score | Duration | Files |
|--------|--------|--------|-------|----------|-------|
| [key-vault-dp-python-crud](results/key-vault/data-plane/python/crud/baseline/report.md) | baseline | ✅ | — | 75.6s | 3 |
| [key-vault-dp-python-error-handling](results/key-vault/data-plane/python/error-handling/azure-mcp/report.md) | azure-mcp | ✅ | — | 78.8s | 1 |
| [key-vault-dp-python-pagination](results/key-vault/data-plane/python/pagination/azure-mcp/report.md) | azure-mcp | ✅ | — | 146.4s | 5 |
| [key-vault-dp-python-pagination](results/key-vault/data-plane/python/pagination/baseline/report.md) | baseline | ❌ | — | 147.0s | 5 |
| [key-vault-dp-python-crud](results/key-vault/data-plane/python/crud/azure-mcp/report.md) | azure-mcp | ✅ | — | 82.3s | 3 |
| [key-vault-dp-python-error-handling](results/key-vault/data-plane/python/error-handling/baseline/report.md) | baseline | ✅ | — | 79.3s | 1 |

