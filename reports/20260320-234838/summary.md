# Evaluation Summary: 20260320-234838

## Run Statistics

| Metric | Value |
|--------|-------|
| Run ID | `20260320-234838` |
| Timestamp | 2026-03-21T06:48:38Z |
| Total Prompts | 3 |
| Total Configs | 2 |
| Total Evaluations | 6 |
| Passed | 5 |
| Failed | 1 |
| Errors | 0 |
| Duration | 191.3s |

## Comparison Matrix

| Prompt | azure-mcp | baseline |
|--------|--------|--------|
| key-vault-dp-python-crud | ✅ | ✅ |
| key-vault-dp-python-error-handling | ✅ | ✅ |
| key-vault-dp-python-pagination | ✅ | ❌ |

## Detailed Results

| Prompt | Config | Result | Score | Duration | Files |
|--------|--------|--------|-------|----------|-------|
| [key-vault-dp-python-crud](results/key-vault/data-plane/python/crud/azure-mcp/report.md) | azure-mcp | ✅ | — | 65.5s | 3 |
| [key-vault-dp-python-error-handling](results/key-vault/data-plane/python/error-handling/baseline/report.md) | baseline | ✅ | — | 95.1s | 2 |
| [key-vault-dp-python-crud](results/key-vault/data-plane/python/crud/baseline/report.md) | baseline | ✅ | — | 100.6s | 4 |
| [key-vault-dp-python-pagination](results/key-vault/data-plane/python/pagination/azure-mcp/report.md) | azure-mcp | ✅ | — | 118.2s | 4 |
| [key-vault-dp-python-error-handling](results/key-vault/data-plane/python/error-handling/azure-mcp/report.md) | azure-mcp | ✅ | — | 79.9s | 1 |
| [key-vault-dp-python-pagination](results/key-vault/data-plane/python/pagination/baseline/report.md) | baseline | ❌ | — | 96.1s | 3 |

