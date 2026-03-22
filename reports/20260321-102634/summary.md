# Evaluation Summary: 20260321-102634

## Run Statistics

| Metric | Value |
|--------|-------|
| Run ID | `20260321-102634` |
| Timestamp | 2026-03-21T17:26:34Z |
| Total Prompts | 3 |
| Total Configs | 2 |
| Total Evaluations | 6 |
| Passed | 5 |
| Failed | 1 |
| Errors | 0 |
| Duration | 168.3s |

## Comparison Matrix

| Prompt | azure-mcp | baseline |
|--------|--------|--------|
| key-vault-dp-python-error-handling | ✅ | ✅ |
| key-vault-dp-python-pagination | ✅ | ❌ |
| key-vault-dp-python-crud | ✅ | ✅ |

## Detailed Results

| Prompt | Config | Result | Score | Duration | Files |
|--------|--------|--------|-------|----------|-------|
| [key-vault-dp-python-error-handling](results/key-vault/data-plane/python/error-handling/azure-mcp/report.md) | azure-mcp | ✅ | — | 70.5s | 1 |
| [key-vault-dp-python-error-handling](results/key-vault/data-plane/python/error-handling/baseline/report.md) | baseline | ✅ | — | 75.6s | 1 |
| [key-vault-dp-python-pagination](results/key-vault/data-plane/python/pagination/baseline/report.md) | baseline | ❌ | — | 140.4s | 4 |
| [key-vault-dp-python-crud](results/key-vault/data-plane/python/crud/baseline/report.md) | baseline | ✅ | — | 74.4s | 2 |
| [key-vault-dp-python-crud](results/key-vault/data-plane/python/crud/azure-mcp/report.md) | azure-mcp | ✅ | — | 88.2s | 3 |
| [key-vault-dp-python-pagination](results/key-vault/data-plane/python/pagination/azure-mcp/report.md) | azure-mcp | ✅ | — | 168.3s | 5 |

