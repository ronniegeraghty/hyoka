# Evaluation Report: key-vault-dp-python-error-handling

**Config:** azure-mcp | **Result:** ✅ PASSED | **Duration:** 82.5s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-error-handling` |
| Config | azure-mcp |
| Result | ✅ PASSED |
| Duration | 82.5s |
| Timestamp | 2026-03-21T19:08:02Z |
| Files Generated | 1 |
| Event Count | 27 |

## Configuration

- **name:** azure-mcp
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

The user is asking for guidance on error handling with Azure Key Vault secrets in Python. They want:
1. Try/except patterns for handling specific HTTP errors (403, 404, 429)
2. How to inspect status_code and error message on HttpResponseError
3. What happens with soft-deleted secrets

This is a straightforward educational request. I should create a comprehensive example showing:
- Proper imports from azure-keyvault-secrets
- Exception handling with HttpResponseError
- How to check status codes
- Different error scenarios (403, 404, 429)
- Soft-deleted secret behavior

I'll create a Python example file that demonstrates all these patterns clearly.

## Tool Calls

### 🔧 1. report_intent

**Input:**

```json
{"intent":"Creating Azure Key Vault error handling examples"}
```

### 🔧 2. create

**Input:**

```json
{"file_text":"\"\"\"\nAzure Key Vault Error Handling Examples\nDemonstrates proper error handling patterns for azure-keyvault-secrets SDK\n\"\"\"\n\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.core.exceptions import (\n    HttpResponseError,\n    ResourceNotFoundError,\n    ServiceRequestError\n)\nfrom azure.identity import DefaultAzureCredential\nimport time\n\n\ndef handle_specific_status_codes(secret_client: SecretClient, secret_name: str):\n    \"\"\"\n    Example: Handling specific HTTP status codes (403, 404, 429)\n    \"\"\"\n    try:\n        secret = secret_client.get_secret(secret_name)\n        print(f\"Successfully retrieved secret: {secret.name}\")\n        return secret.value\n    \n    except HttpResponseError as e:\n        # Inspect the status code and error details\n        status_code = e.status_code\n        error_code = e.error.code if e.error else \"Unknown\"\n        error_message = e.message\n        \n        print(f\"HTTP Status Code: {status_code}\")\n        print(f\"Error Code: {error_code}\")\n        print(f\"Error Message: {error_message}\")\n        \n        # Handle specific status codes\n        if status_code == 403:\n            # Access Denied - RBAC permissions issue\n            print(\"ERROR: Access Denied (403)\")\n            print(\"Your identity doesn't have the required RBAC role.\")\n            print(\"Needed: 'Key Vault Secrets User' or 'Key Vault Secrets Officer'\")\n            print(f\"Make sure your identity has GET permission on secret: {secret_name}\")\n            # Could trigger an alert, log to monitoring system, etc.\n            raise\n        \n        elif status_code == 404:\n            # Secret Not Found\n            print(f\"ERROR: Secret '{secret_name}' not found (404)\")\n            print(\"The secret may not exist or could be soft-deleted.\")\n            # Could fall back to default value, create the secret, etc.\n            return None\n        \n        elif status_code == 429:\n            # Throttling - Rate limit exceeded\n            print(\"ERROR: Throttling (429) - Rate limit exceeded\")\n            print(\"Too many requests to Key Vault. Implementing exponential backoff...\")\n            \n            # Get retry-after header if available\n            retry_after = e.response.headers.get('Retry-After', 5)\n            retry_after = int(retry_after) if isinstance(retry_after, str) else retry_after\n            \n            print(f\"Waiting {retry_after} seconds before retry...\")\n            time.sleep(retry_after)\n            \n            # Retry the request (in production, use exponential backoff)\n            return handle_specific_status_codes(secret_client, secret_name)\n        \n        else:\n            # Other HTTP errors\n            print(f\"ERROR: Unexpected HTTP error (Status: {status_code})\")\n            raise\n\n\ndef handle_soft_deleted_secret(secret_client: SecretClient, secret_name: str):\n    \"\"\"\n    Example: What happens when trying to get a soft-deleted secret\n    \n    When soft-delete is enabled:\n    - get_secret() will raise 404 (ResourceNotFoundError)\n    - The secret still exists in the deleted state\n    - Use get_deleted_secret() to retrieve it\n    - Use recover_deleted_secret() to restore it\n    \"\"\"\n    try:\n        secret = secret_client.get_secret(secret_name)\n        print(f\"Secret found: {secret.name}\")\n        return secret.value\n    \n    except ResourceNotFoundError as e:\n        print(f\"Secret '{secret_name}' not found with get_secret()\")\n        print(\"Checking if it's soft-deleted...\")\n        \n        try:\n            # Try to get the deleted secret\n            deleted_secret = secret_client.get_deleted_secret(secret_name)\n            print(f\"Found soft-deleted secret: {deleted_secret.name}\")\n            print(f\"Deleted on: {deleted_secret.deleted_date}\")\n            print(f\"Scheduled purge date: {deleted_secret.scheduled_purge_date}\")\n            print(f\"Recovery ID: {deleted_secret.recovery_id}\")\n            \n            # Option 1: Recover the secret\n            print(\"Recovering the soft-deleted secret...\")\n            recovered_secret = secret_client.begin_recover_deleted_secret(secret_name).result()\n            print(f\"Secret recovered: {recovered_secret.name}\")\n            \n            # Now we can get it normally\n            secret = secret_client.get_secret(secret_name)\n            return secret.value\n        \n        except ResourceNotFoundError:\n            print(f\"Secret '{secret_name}' is not soft-deleted either.\")\n            print(\"It truly doesn't exist or was purged.\")\n            return None\n        \n        except HttpResponseError as recover_error:\n            if recover_error.status_code == 403:\n                print(\"ERROR: Cannot recover - missing permissions\")\n                print(\"Needed: 'Key Vault Secrets Officer' role for recovery\")\n            raise\n\n\ndef comprehensive_error_handling(secret_client: SecretClient, secret_name: str):\n    \"\"\"\n    Example: Comprehensive error handling with multiple exception types\n    \"\"\"\n    max_retries = 3\n    retry_count = 0\n    base_delay = 2  # seconds\n    \n    while retry_count \u003c max_retries:\n        try:\n            secret = secret_client.get_secret(secret_name)\n            print(f\"✓ Successfully retrieved secret: {secret.name}\")\n            return secret.value\n        \n        except ResourceNotFoundError as e:\n            # 404 - Secret not found\n            print(f\"✗ Secret '{secret_name}' not found (404)\")\n            # Check if soft-deleted\n            try:\n                deleted = secret_client.get_deleted_secret(secret_name)\n                print(f\"  → Found in soft-deleted state. Deleted: {deleted.deleted_date}\")\n                return None  # Or trigger recovery workflow\n            except:\n                print(f\"  → Secret doesn't exist (not even in deleted state)\")\n                return None\n        \n        except HttpResponseError as e:\n            status_code = e.status_code\n            error_code = e.error.code if e.error else \"Unknown\"\n            \n            print(f\"✗ HTTP Error: {status_code} - {error_code}\")\n            print(f\"  Message: {e.message}\")\n            \n            if status_code == 403:\n                # Permission denied - don't retry\n                print(f\"  → Access denied. Check RBAC permissions.\")\n                print(f\"  → Required role: 'Key Vault Secrets User'\")\n                raise  # Don't retry permission errors\n            \n            elif status_code == 429:\n                # Throttling - retry with backoff\n                retry_after = int(e.response.headers.get('Retry-After', base_delay * (2 ** retry_count)))\n                print(f\"  → Throttled. Retry after {retry_after}s (attempt {retry_count + 1}/{max_retries})\")\n                time.sleep(retry_after)\n                retry_count += 1\n                continue\n            \n            elif status_code == 503:\n                # Service unavailable - retry with backoff\n                delay = base_delay * (2 ** retry_count)\n                print(f\"  → Service unavailable. Retry in {delay}s (attempt {retry_count + 1}/{max_retries})\")\n                time.sleep(delay)\n                retry_count += 1\n                continue\n            \n            else:\n                # Other errors - don't retry\n                print(f\"  → Unexpected error. Not retrying.\")\n                raise\n        \n        except ServiceRequestError as e:\n            # Network errors, DNS resolution failures, etc.\n            print(f\"✗ Service request error: {e}\")\n            print(f\"  → Network or connection issue. Retry {retry_count + 1}/{max_retries}\")\n            delay = base_delay * (2 ** retry_count)\n            time.sleep(delay)\n            retry_count += 1\n            continue\n        \n        except Exception as e:\n            # Catch-all for unexpected errors\n            print(f\"✗ Unexpected error: {type(e).__name__}: {e}\")\n            raise\n    \n    print(f\"✗ Failed after {max_retries} retries\")\n    return None\n\n\ndef batch_get_with_error_handling(secret_client: SecretClient, secret_names: list):\n    \"\"\"\n    Example: Batch operations with per-secret error handling\n    \"\"\"\n    results = {}\n    errors = {}\n    \n    for secret_name in secret_names:\n        try:\n            secret = secret_client.get_secret(secret_name)\n            results[secret_name] = secret.value\n            print(f\"✓ {secret_name}: retrieved\")\n        \n        except HttpResponseError as e:\n            error_info = {\n                'status_code': e.status_code,\n                'error_code': e.error.code if e.error else None,\n                'message': e.message\n            }\n            errors[secret_name] = error_info\n            print(f\"✗ {secret_name}: {e.status_code} - {error_info['error_code']}\")\n        \n        except Exception as e:\n            errors[secret_name] = {'error': str(e)}\n            print(f\"✗ {secret_name}: {type(e).__name__}: {e}\")\n    \n    return results, errors\n\n\ndef inspect_error_details(secret_client: SecretClient, secret_name: str):\n    \"\"\"\n    Example: Detailed error inspection showing all available properties\n    \"\"\"\n    try:\n        secret = secret_client.get_secret(secret_name)\n        return secret.value\n    \n    except HttpResponseError as e:\n        print(\"=\" * 60)\n        print(\"HttpResponseError Details:\")\n        print(\"=\" * 60)\n        \n        # Status code\n        print(f\"status_code: {e.status_code}\")\n        \n        # Error object (if available)\n        if e.error:\n            print(f\"error.code: {e.error.code}\")\n            print(f\"error.message: {e.error.message}\")\n        \n        # Message\n        print(f\"message: {e.message}\")\n        \n        # Response object (if available)\n        if e.response:\n            print(f\"response.status_code: {e.response.status_code}\")\n            print(f\"response.reason: {e.response.reason}\")\n            print(f\"response.headers: {dict(e.response.headers)}\")\n        \n        # Additional attributes\n        print(f\"reason: {e.reason if hasattr(e, 'reason') else 'N/A'}\")\n        \n        print(\"=\" * 60)\n        raise\n\n\n# Example usage\nif __name__ == \"__main__\":\n    # Initialize client\n    vault_url = \"https://your-vault-name.vault.azure.net/\"\n    credential = DefaultAzureCredential()\n    secret_client = SecretClient(vault_url=vault_url, credential=credential)\n    \n    # Example 1: Handle specific status codes\n    print(\"\\n--- Example 1: Specific Status Code Handling ---\")\n    try:\n        value = handle_specific_status_codes(secret_client, \"my-secret\")\n    except Exception as e:\n        print(f\"Failed: {e}\")\n    \n    # Example 2: Soft-deleted secret handling\n    print(\"\\n--- Example 2: Soft-Deleted Secret Handling ---\")\n    try:\n        value = handle_soft_deleted_secret(secret_client, \"deleted-secret\")\n    except Exception as e:\n        print(f\"Failed: {e}\")\n    \n    # Example 3: Comprehensive error handling with retries\n    print(\"\\n--- Example 3: Comprehensive Error Handling ---\")\n    value = comprehensive_error_handling(secret_client, \"my-secret\")\n    \n    # Example 4: Batch operations\n    print(\"\\n--- Example 4: Batch Operations ---\")\n    secret_names = [\"secret1\", \"secret2\", \"secret3\"]\n    results, errors = batch_get_with_error_handling(secret_client, secret_names)\n    print(f\"Retrieved: {len(results)}, Failed: {len(errors)}\")\n    \n    # Example 5: Detailed error inspection\n    print(\"\\n--- Example 5: Detailed Error Inspection ---\")\n    try:\n        inspect_error_details(secret_client, \"nonexistent-secret\")\n    except Exception:\n        pass\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-error-handling-azure-mcp-3153508792/keyvault_error_handling.py"}
```

## Generated Files

- `keyvault_error_handling.py`

## Copilot Response

I'll create a comprehensive example showing Azure Key Vault error handling patterns in Python.

Perfect! I've created a comprehensive guide showing Azure Key Vault error handling patterns. Here are the key points:

## Key Patterns:

**1. Handling Specific Status Codes (403, 404, 429):**
- Catch `HttpResponseError` and inspect `e.status_code`
- **403 (Forbidden)**: RBAC permission issue - needs "Key Vault Secrets User" or "Secrets Officer" role
- **404 (Not Found)**: Secret doesn't exist (could be soft-deleted)
- **429 (Throttling)**: Implement exponential backoff, check `Retry-After` header

**2. Inspecting HttpResponseError:**
```python
except HttpResponseError as e:
    status_code = e.status_code
    error_code = e.error.code if e.error else None
    error_message = e.message
    retry_after = e.response.headers.get('Retry-After')
```

**3. Soft-Deleted Secrets:**
When you try to `get_secret()` on a soft-deleted secret:
- Raises **404/ResourceNotFoundError** (same as non-existent secret)
- Use `get_deleted_secret()` to check if it's soft-deleted
- Use `begin_recover_deleted_secret()` to restore it (needs "Secrets Officer" role)

**4. Best Practices:**
- Don't retry 403 errors (permission issues won't resolve automatically)
- Implement exponential backoff for 429/503 errors
- Use `ResourceNotFoundError` subclass for cleaner 404 handling
- Check for soft-deleted secrets when getting 404s

The file includes 5 complete examples showing different scenarios and patterns!

## Verification

**Result:** ✅ PASS

**Summary:** Comprehensive implementation covering all error handling scenarios with proper exception types, status code inspection, and soft-delete handling.

**Reasoning:**

The code successfully addresses all requirements from the original prompt. It correctly imports and uses HttpResponseError and ResourceNotFoundError from azure.core.exceptions. It demonstrates proper try/except patterns for handling 403 (access denied with RBAC role guidance), 404 (secret not found), and 429 (throttling with Retry-After header inspection). The code shows how to extract status_code, error.code, and message properties from exceptions. It includes a dedicated function (handle_soft_deleted_secret) that explains what happens with soft-deleted secrets and demonstrates using begin_recover_deleted_secret. The inspect_error_details function comprehensively shows all available error properties. The code is syntactically valid Python, uses the correct azure-keyvault-secrets SDK, and provides multiple practical examples including batch operations and exponential backoff retry logic. The only minor gap is that it doesn't show SecretClient retry configuration via constructor kwargs (retry_total, retry_backoff_factor), but this wasn't explicitly requested in the original prompt, only in the expected coverage checklist. The code implements manual retry logic instead, which is also valid. Overall, this is a comprehensive, production-ready solution that meets all the user's stated needs.

---

[← Back to Summary](../../../../../../summary.md)
