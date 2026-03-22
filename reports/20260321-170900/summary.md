# Evaluation Summary: 20260321-170900

## Run Statistics

| Metric | Value |
|--------|-------|
| Run ID | `20260321-170900` |
| Timestamp | 2026-03-22T00:09:00Z |
| Total Prompts | 3 |
| Total Configs | 2 |
| Total Evaluations | 6 |
| Passed | 5 |
| Failed | 1 |
| Errors | 0 |
| Duration | 181.4s |

## Comparison Matrix

| Prompt | baseline | azure-mcp |
|--------|--------|--------|
| key-vault-dp-python-crud | ✅ | ✅ |
| key-vault-dp-python-error-handling | ✅ | ✅ |
| key-vault-dp-python-pagination | ❌ | ✅ |

## Detailed Results

| Prompt | Config | Result | Score | Duration | Files |
|--------|--------|--------|-------|----------|-------|
| [key-vault-dp-python-crud](results/key-vault/data-plane/python/crud/baseline/report.md) | baseline | ✅ | — | 52.7s | 3 |
| [key-vault-dp-python-error-handling](results/key-vault/data-plane/python/error-handling/azure-mcp/report.md) | azure-mcp | ✅ | — | 74.1s | 1 |
| [key-vault-dp-python-error-handling](results/key-vault/data-plane/python/error-handling/baseline/report.md) | baseline | ✅ | — | 82.5s | 1 |
| [key-vault-dp-python-pagination](results/key-vault/data-plane/python/pagination/baseline/report.md) | baseline | ❌ | — | 80.5s | 3 |
| [key-vault-dp-python-pagination](results/key-vault/data-plane/python/pagination/azure-mcp/report.md) | azure-mcp | ✅ | — | 164.4s | 6 |
| [key-vault-dp-python-crud](results/key-vault/data-plane/python/crud/azure-mcp/report.md) | azure-mcp | ✅ | — | 128.7s | 4 |

