# Evaluation Report: key-vault-dp-python-error-handling

**Config:** baseline | **Result:** ✅ PASSED | **Duration:** 95.1s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-error-handling` |
| Config | baseline |
| Result | ✅ PASSED |
| Duration | 95.1s |
| Timestamp | 2026-03-21T06:48:38Z |
| Files Generated | 2 |
| Event Count | 34 |

## Configuration

- **name:** baseline
- **model:** claude-sonnet-4.5

## Prompt Sent

```
How do I properly handle errors when working with Azure Key Vault secrets in Python?
I need to handle access denied (403) when my app's identity doesn't have the right
RBAC role, secret not found (404), and throttling (429) when rate limits are hit.
Show me try/except patterns with the azure-keyvault-secrets SDK including
how to inspect the status_code and error message on HttpResponseError.
Also explain what happens when I try to get a soft-deleted secret.
```

## Copilot Reasoning

The user is asking for guidance on handling errors when working with Azure Key Vault secrets in Python. They specifically want:

1. How to handle 403 (access denied/RBAC issues)
2. How to handle 404 (secret not found)
3. How to handle 429 (throttling/rate limits)
4. Try/except patterns with azure-keyvault-secrets SDK
5. How to inspect status_code and error message on HttpResponseError
6. What happens when getting a soft-deleted secret

This is a straightforward educational question about Azure SDK error handling. I should provide a comprehensive example showing these error handling patterns. I'll create a Python file demonstrating best practices for error handling with the Azure Key Vault SDK.

Let me create a comprehensive example that shows:
- Proper imports
- Try/except blocks for different scenarios
- How to check status codes
- How to handle different error types
- Explanation of soft-deleted secrets

I'll create a practical example file with detailed comments.

## Tool Calls

### 🔧 1. report_intent

**Input:**

```json
{"intent":"Creating error handling examples"}
```

### 🔧 2. create

**Input:**

```json
{"file_text":"\"\"\"\nAzure Key Vault Secret Error Handling Examples\n\nThis module demonstrates proper error handling patterns when working with\nAzure Key Vault secrets using the azure-keyvault-secrets SDK.\n\"\"\"\n\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.identity import DefaultAzureCredential\nfrom azure.core.exceptions import (\n    HttpResponseError,\n    ResourceNotFoundError,\n    ServiceRequestError\n)\nimport time\n\n\n# Initialize the client\nvault_url = \"https://your-keyvault-name.vault.azure.net/\"\ncredential = DefaultAzureCredential()\nclient = SecretClient(vault_url=vault_url, credential=credential)\n\n\ndef handle_get_secret_basic(secret_name: str):\n    \"\"\"\n    Basic error handling pattern for getting a secret.\n    Handles common HTTP errors by checking status codes.\n    \"\"\"\n    try:\n        secret = client.get_secret(secret_name)\n        print(f\"Successfully retrieved secret: {secret.name}\")\n        return secret.value\n    \n    except HttpResponseError as e:\n        # HttpResponseError contains status_code and error details\n        status_code = e.status_code\n        error_message = e.message\n        \n        if status_code == 403:\n            # Access Denied - Missing RBAC permissions\n            print(f\"Access Denied (403): Your identity lacks permissions.\")\n            print(f\"Error: {error_message}\")\n            print(f\"Required role: 'Key Vault Secrets User' or 'Key Vault Secrets Officer'\")\n            \n        elif status_code == 404:\n            # Secret Not Found\n            print(f\"Secret Not Found (404): '{secret_name}' does not exist.\")\n            print(f\"Error: {error_message}\")\n            \n        elif status_code == 429:\n            # Rate Limit / Throttling\n            print(f\"Throttling (429): Rate limit exceeded.\")\n            print(f\"Error: {error_message}\")\n            # Check for Retry-After header if available\n            retry_after = e.response.headers.get('Retry-After')\n            if retry_after:\n                print(f\"Retry after {retry_after} seconds\")\n        \n        else:\n            # Other HTTP errors\n            print(f\"HTTP Error ({status_code}): {error_message}\")\n        \n        # Re-raise or return None based on your needs\n        raise\n\n\ndef handle_get_secret_with_specific_exceptions(secret_name: str):\n    \"\"\"\n    Using specific exception types for cleaner error handling.\n    ResourceNotFoundError is a subclass of HttpResponseError for 404s.\n    \"\"\"\n    try:\n        secret = client.get_secret(secret_name)\n        return secret.value\n    \n    except ResourceNotFoundError as e:\n        # This catches 404 specifically\n        print(f\"Secret '{secret_name}' not found (404)\")\n        print(f\"Details: {e.message}\")\n        return None\n    \n    except HttpResponseError as e:\n        if e.status_code == 403:\n            print(f\"Access denied: Check your RBAC role assignment\")\n            print(f\"Error details: {e.message}\")\n        elif e.status_code == 429:\n            print(f\"Rate limited: {e.message}\")\n        else:\n            print(f\"HTTP error {e.status_code}: {e.message}\")\n        raise\n\n\ndef handle_get_secret_with_retry(secret_name: str, max_retries: int = 3):\n    \"\"\"\n    Handling throttling (429) with exponential backoff retry logic.\n    \"\"\"\n    retry_count = 0\n    backoff = 1  # Initial backoff in seconds\n    \n    while retry_count \u003c max_retries:\n        try:\n            secret = client.get_secret(secret_name)\n            return secret.value\n        \n        except HttpResponseError as e:\n            if e.status_code == 429:\n                retry_count += 1\n                if retry_count \u003e= max_retries:\n                    print(f\"Max retries ({max_retries}) exceeded\")\n                    raise\n                \n                # Check for Retry-After header\n                retry_after = e.response.headers.get('Retry-After')\n                if retry_after:\n                    wait_time = int(retry_after)\n                else:\n                    wait_time = backoff * (2 ** retry_count)  # Exponential backoff\n                \n                print(f\"Rate limited. Retrying in {wait_time} seconds...\")\n                time.sleep(wait_time)\n            \n            elif e.status_code == 403:\n                print(f\"Access denied. Cannot retry.\")\n                raise\n            \n            elif e.status_code == 404:\n                print(f\"Secret not found. Cannot retry.\")\n                raise\n            \n            else:\n                raise\n\n\ndef handle_soft_deleted_secret(secret_name: str):\n    \"\"\"\n    Understanding soft-deleted secrets behavior.\n    \n    When a secret is deleted with soft-delete enabled (default in most vaults):\n    - The secret moves to a \"deleted\" state\n    - Attempting get_secret() raises ResourceNotFoundError (404)\n    - The name is reserved and cannot be reused until purged\n    - You must use get_deleted_secret() to access it\n    - Recovery is possible using recover_deleted_secret()\n    \"\"\"\n    try:\n        # This will fail with 404 if the secret is soft-deleted\n        secret = client.get_secret(secret_name)\n        print(f\"Secret is active: {secret.name}\")\n        return secret.value\n    \n    except ResourceNotFoundError as e:\n        print(f\"Secret not found (404). Checking if it's soft-deleted...\")\n        \n        # Try to get the deleted secret\n        try:\n            deleted_secret = client.get_deleted_secret(secret_name)\n            print(f\"Secret is soft-deleted!\")\n            print(f\"  Deleted on: {deleted_secret.deleted_date}\")\n            print(f\"  Scheduled purge: {deleted_secret.scheduled_purge_date}\")\n            print(f\"  Recovery ID: {deleted_secret.recovery_id}\")\n            print(f\"  Can be recovered using: client.begin_recover_deleted_secret()\")\n            \n            # To recover (requires 'Key Vault Secrets Officer' role):\n            # poller = client.begin_recover_deleted_secret(secret_name)\n            # recovered_secret = poller.result()\n            \n            return None\n        \n        except HttpResponseError as deleted_error:\n            if deleted_error.status_code == 404:\n                print(f\"Secret truly does not exist (not even in deleted state)\")\n            elif deleted_error.status_code == 403:\n                print(f\"No permission to view deleted secrets\")\n            else:\n                print(f\"Error checking deleted secrets: {deleted_error.message}\")\n            return None\n\n\ndef comprehensive_error_handling(secret_name: str):\n    \"\"\"\n    Comprehensive error handling covering all scenarios.\n    \"\"\"\n    try:\n        secret = client.get_secret(secret_name)\n        print(f\"✓ Successfully retrieved: {secret.name}\")\n        return secret.value\n    \n    except ResourceNotFoundError as e:\n        # HTTP 404 - Secret not found\n        print(f\"✗ Secret not found (404)\")\n        print(f\"  Message: {e.message}\")\n        print(f\"  Error code: {e.error.code if hasattr(e, 'error') else 'N/A'}\")\n        print(f\"  Tip: Check if secret is soft-deleted using get_deleted_secret()\")\n        return None\n    \n    except HttpResponseError as e:\n        # Other HTTP errors\n        status = e.status_code\n        \n        if status == 403:\n            print(f\"✗ Access Denied (403)\")\n            print(f\"  Message: {e.message}\")\n            print(f\"  Required RBAC roles:\")\n            print(f\"    - Key Vault Secrets User (read-only)\")\n            print(f\"    - Key Vault Secrets Officer (read/write)\")\n            print(f\"  Identity: Check your DefaultAzureCredential configuration\")\n        \n        elif status == 429:\n            print(f\"✗ Rate Limited (429)\")\n            print(f\"  Message: {e.message}\")\n            retry_after = e.response.headers.get('Retry-After', 'Not specified')\n            print(f\"  Retry after: {retry_after}\")\n            print(f\"  Tip: Implement exponential backoff retry logic\")\n        \n        elif status == 401:\n            print(f\"✗ Authentication Failed (401)\")\n            print(f\"  Message: {e.message}\")\n            print(f\"  Tip: Check your credential configuration and token validity\")\n        \n        else:\n            print(f\"✗ HTTP Error ({status})\")\n            print(f\"  Message: {e.message}\")\n            # Access additional error details if available\n            if hasattr(e, 'error') and e.error:\n                print(f\"  Error code: {e.error.code}\")\n                print(f\"  Error details: {e.error.message}\")\n        \n        raise\n    \n    except ServiceRequestError as e:\n        # Network-level errors (DNS, connection failures)\n        print(f\"✗ Service Request Error: {e}\")\n        print(f\"  Tip: Check network connectivity and vault URL\")\n        raise\n    \n    except Exception as e:\n        # Catch-all for unexpected errors\n        print(f\"✗ Unexpected error: {type(e).__name__}: {e}\")\n        raise\n\n\ndef inspect_error_details(secret_name: str):\n    \"\"\"\n    Demonstrates how to inspect all available error details from HttpResponseError.\n    \"\"\"\n    try:\n        secret = client.get_secret(secret_name)\n        return secret.value\n    \n    except HttpResponseError as e:\n        print(\"=== HttpResponseError Details ===\")\n        print(f\"Status Code: {e.status_code}\")\n        print(f\"Reason: {e.reason}\")\n        print(f\"Message: {e.message}\")\n        \n        # Response object details\n        if e.response:\n            print(f\"\\n=== Response Headers ===\")\n            for key, value in e.response.headers.items():\n                print(f\"{key}: {value}\")\n        \n        # Error object (if available)\n        if hasattr(e, 'error') and e.error:\n            print(f\"\\n=== Error Object ===\")\n            print(f\"Code: {e.error.code}\")\n            print(f\"Message: {e.error.message}\")\n            if hasattr(e.error, 'innererror'):\n                print(f\"Inner Error: {e.error.innererror}\")\n        \n        raise\n\n\n# Example usage patterns\nif __name__ == \"__main__\":\n    secret_name = \"my-secret\"\n    \n    print(\"Example 1: Basic error handling\")\n    print(\"-\" * 50)\n    try:\n        handle_get_secret_basic(secret_name)\n    except Exception:\n        pass\n    \n    print(\"\\n\\nExample 2: Specific exception handling\")\n    print(\"-\" * 50)\n    try:\n        handle_get_secret_with_specific_exceptions(secret_name)\n    except Exception:\n        pass\n    \n    print(\"\\n\\nExample 3: Handling with retry logic\")\n    print(\"-\" * 50)\n    try:\n        handle_get_secret_with_retry(secret_name)\n    except Exception:\n        pass\n    \n    print(\"\\n\\nExample 4: Soft-deleted secret handling\")\n    print(\"-\" * 50)\n    handle_soft_deleted_secret(secret_name)\n    \n    print(\"\\n\\nExample 5: Comprehensive error handling\")\n    print(\"-\" * 50)\n    try:\n        comprehensive_error_handling(secret_name)\n    except Exception:\n        pass\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-error-handling-baseline-2980971982/keyvault_error_handling.py"}
```

### 🔧 3. create

**Input:**

```json
{"file_text":"# Azure Key Vault Secret Error Handling Guide\n\nThis repository demonstrates proper error handling patterns for Azure Key Vault secrets using the `azure-keyvault-secrets` Python SDK.\n\n## Key Concepts\n\n### HttpResponseError Structure\n\nThe primary exception type is `HttpResponseError`, which contains:\n\n```python\nfrom azure.core.exceptions import HttpResponseError\n\ntry:\n    secret = client.get_secret(\"secret-name\")\nexcept HttpResponseError as e:\n    # Access error details:\n    status_code = e.status_code      # HTTP status code (403, 404, 429, etc.)\n    message = e.message              # Human-readable error message\n    reason = e.reason                # HTTP reason phrase\n    response = e.response            # Full HTTP response object\n    \n    # Response headers (useful for Retry-After on 429)\n    retry_after = e.response.headers.get('Retry-After')\n```\n\n### Common HTTP Status Codes\n\n| Code | Meaning | Cause | Solution |\n|------|---------|-------|----------|\n| **403** | Forbidden | Missing RBAC role | Assign \"Key Vault Secrets User\" or \"Key Vault Secrets Officer\" role |\n| **404** | Not Found | Secret doesn't exist or is soft-deleted | Check name or use `get_deleted_secret()` |\n| **429** | Too Many Requests | Rate limit exceeded | Implement retry with exponential backoff |\n| **401** | Unauthorized | Authentication failure | Check credential configuration |\n\n### Soft-Deleted Secrets\n\nWhen soft-delete is enabled (default):\n\n1. **Deleting a secret** moves it to \"deleted\" state (not permanently removed)\n2. **Getting a deleted secret** with `get_secret()` raises `ResourceNotFoundError` (404)\n3. **The name is reserved** - you cannot create a new secret with the same name\n4. **To access deleted secret**: Use `get_deleted_secret(name)`\n5. **To recover**: Use `begin_recover_deleted_secret(name)`\n6. **To permanently delete**: Use `purge_deleted_secret(name)` (requires \"Key Vault Secrets Officer\" role)\n\n```python\nfrom azure.core.exceptions import ResourceNotFoundError\n\ntry:\n    secret = client.get_secret(\"deleted-secret\")\nexcept ResourceNotFoundError:\n    # Check if it's soft-deleted\n    try:\n        deleted = client.get_deleted_secret(\"deleted-secret\")\n        print(f\"Secret is soft-deleted, scheduled purge: {deleted.scheduled_purge_date}\")\n        \n        # Recover it if needed\n        poller = client.begin_recover_deleted_secret(\"deleted-secret\")\n        recovered = poller.result()  # Wait for recovery to complete\n    except HttpResponseError as e:\n        if e.status_code == 404:\n            print(\"Secret truly doesn't exist\")\n```\n\n## Error Handling Patterns\n\n### Pattern 1: Check Status Code\n\n```python\ntry:\n    secret = client.get_secret(secret_name)\nexcept HttpResponseError as e:\n    if e.status_code == 403:\n        print(\"Access denied - check RBAC roles\")\n    elif e.status_code == 404:\n        print(\"Secret not found\")\n    elif e.status_code == 429:\n        print(\"Rate limited\")\n```\n\n### Pattern 2: Use Specific Exceptions\n\n```python\nfrom azure.core.exceptions import ResourceNotFoundError\n\ntry:\n    secret = client.get_secret(secret_name)\nexcept ResourceNotFoundError:\n    print(\"Secret not found (404)\")\nexcept HttpResponseError as e:\n    if e.status_code == 403:\n        print(\"Access denied\")\n    elif e.status_code == 429:\n        print(\"Rate limited\")\n```\n\n### Pattern 3: Retry Logic for Throttling\n\n```python\nimport time\n\ndef get_secret_with_retry(secret_name, max_retries=3):\n    for attempt in range(max_retries):\n        try:\n            return client.get_secret(secret_name)\n        except HttpResponseError as e:\n            if e.status_code == 429:\n                retry_after = e.response.headers.get('Retry-After', 5)\n                time.sleep(int(retry_after))\n            else:\n                raise\n    raise Exception(\"Max retries exceeded\")\n```\n\n## Required RBAC Roles\n\n- **Key Vault Secrets User**: Read-only access (`get_secret`)\n- **Key Vault Secrets Officer**: Full access (create, read, update, delete, recover, purge)\n\n## Installation\n\n```bash\npip install azure-keyvault-secrets azure-identity\n```\n\n## Usage\n\n```python\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.identity import DefaultAzureCredential\n\nvault_url = \"https://your-vault.vault.azure.net/\"\ncredential = DefaultAzureCredential()\nclient = SecretClient(vault_url=vault_url, credential=credential)\n\n# Use the error handling patterns from keyvault_error_handling.py\n```\n\n## See Also\n\n- [keyvault_error_handling.py](keyvault_error_handling.py) - Complete code examples\n- [Azure Key Vault Secrets SDK Documentation](https://learn.microsoft.com/python/api/azure-keyvault-secrets/)\n- [Azure RBAC for Key Vault](https://learn.microsoft.com/azure/key-vault/general/rbac-guide)\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-error-handling-baseline-2980971982/README.md"}
```

## Generated Files

- `README.md`
- `keyvault_error_handling.py`

## Copilot Response

I'll create a comprehensive example showing proper error handling patterns for Azure Key Vault secrets in Python.

## ✓ Created comprehensive error handling examples

I've created two files showing proper Azure Key Vault error handling patterns:

**Key points:**

1. **HttpResponseError inspection**:
   - `e.status_code` - HTTP status (403, 404, 429)
   - `e.message` - Error description
   - `e.response.headers.get('Retry-After')` - Retry timing for 429

2. **Common errors**:
   - **403**: Missing RBAC role - need "Key Vault Secrets User" or "Key Vault Secrets Officer"
   - **404**: Secret not found (could be soft-deleted)
   - **429**: Rate limited - implement exponential backoff

3. **Soft-deleted secrets**:
   - `get_secret()` returns 404 for deleted secrets
   - Use `get_deleted_secret()` to check deleted state
   - Use `begin_recover_deleted_secret()` to restore
   - Name is reserved until purged

The Python file includes 6 complete examples with different patterns (basic, specific exceptions, retry logic, soft-delete handling). See `keyvault_error_handling.py` for runnable code and `README.md` for a quick reference guide.

## Verification

**Result:** ✅ PASS

**Summary:** Comprehensive implementation covering all main requirements; only missing SecretClient retry kwargs configuration (manual retry shown instead).

**Reasoning:**

The generated code comprehensively addresses the original prompt and meets nearly all expected coverage criteria. ✓ Correctly imports and uses HttpResponseError and ResourceNotFoundError from azure.core.exceptions. ✓ Demonstrates extracting status_code, error.code, and message from exceptions (see inspect_error_details and comprehensive_error_handling functions). ✓ Handles 403 errors with detailed RBAC role guidance. ✓ Handles 404 errors and distinguishes between non-existent and soft-deleted secrets. ✓ Handles 429 throttling with Retry-After header inspection and retry logic. ✓ Extensively covers soft-delete behavior with get_deleted_secret() and begin_recover_deleted_secret() explanation. ✓ Shows multiple try/except patterns from basic to comprehensive. ✓ Code is syntactically valid with proper imports and structure. The only minor gap is the absence of SecretClient retry configuration via constructor kwargs (e.g., retry_policy parameter), but manual retry logic is thoroughly demonstrated. The code provides multiple well-documented functions covering all key scenarios from the prompt, making it a practical and complete solution.

---

[← Back to Summary](../../../../../../summary.md)
