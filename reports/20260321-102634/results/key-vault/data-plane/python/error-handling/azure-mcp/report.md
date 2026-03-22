# Evaluation Report: key-vault-dp-python-error-handling

**Config:** azure-mcp | **Result:** ✅ PASSED | **Duration:** 70.5s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-error-handling` |
| Config | azure-mcp |
| Result | ✅ PASSED |
| Duration | 70.5s |
| Timestamp | 2026-03-21T17:26:34Z |
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
1. Try/except patterns for azure-keyvault-secrets SDK
2. How to handle specific HTTP status codes (403, 404, 429)
3. How to inspect status_code and error message on HttpResponseError
4. What happens when trying to get a soft-deleted secret

This is a straightforward educational/example request. I should create a comprehensive Python example demonstrating proper error handling patterns for Azure Key Vault operations.

I'll create a demonstration script that shows:
- Proper exception handling with azure.core.exceptions.HttpResponseError
- How to check status codes
- How to extract error messages
- Different scenarios (403, 404, 429, soft-deleted secrets)

Let me create a practical example file.

## Tool Calls

### 🔧 1. report_intent

**Input:**

```json
{"intent":"Creating error handling examples"}
```

### 🔧 2. create

**Input:**

```json
{"file_text":"\"\"\"\nAzure Key Vault Error Handling Examples\nDemonstrates proper exception handling patterns for azure-keyvault-secrets SDK\n\"\"\"\n\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.identity import DefaultAzureCredential\nfrom azure.core.exceptions import HttpResponseError, ResourceNotFoundError\nimport time\n\n\ndef handle_get_secret_with_retry(secret_client: SecretClient, secret_name: str, max_retries: int = 3):\n    \"\"\"\n    Example: Get a secret with comprehensive error handling and retry logic for throttling\n    \"\"\"\n    retry_count = 0\n    \n    while retry_count \u003c= max_retries:\n        try:\n            secret = secret_client.get_secret(secret_name)\n            print(f\"Successfully retrieved secret: {secret.name}\")\n            print(f\"Value: {secret.value}\")\n            return secret\n            \n        except HttpResponseError as e:\n            # Check the HTTP status code\n            status_code = e.status_code\n            \n            if status_code == 403:\n                # Access Denied - RBAC permission issue\n                print(f\"❌ Access Denied (403): Your identity lacks the required RBAC role\")\n                print(f\"Error message: {e.message}\")\n                print(f\"Error code: {e.error.code if hasattr(e, 'error') and e.error else 'N/A'}\")\n                print(\"Required role: 'Key Vault Secrets User' or 'Key Vault Secrets Officer'\")\n                raise  # Don't retry on permission errors\n                \n            elif status_code == 404:\n                # Secret not found\n                print(f\"❌ Secret Not Found (404): '{secret_name}' does not exist\")\n                print(f\"Error message: {e.message}\")\n                \n                # Check if it might be a soft-deleted secret\n                try:\n                    deleted_secret = secret_client.get_deleted_secret(secret_name)\n                    print(f\"⚠️  Secret '{secret_name}' exists in soft-deleted state\")\n                    print(f\"   Deleted on: {deleted_secret.deleted_date}\")\n                    print(f\"   Scheduled purge: {deleted_secret.scheduled_purge_date}\")\n                    print(\"   Action: Recover the secret or wait for purge, then recreate\")\n                except HttpResponseError as del_error:\n                    if del_error.status_code == 404:\n                        print(\"   Secret does not exist in active or deleted state\")\n                    else:\n                        print(f\"   Could not check deleted secrets: {del_error.message}\")\n                \n                raise  # Don't retry on not found\n                \n            elif status_code == 429:\n                # Throttling - Rate limit exceeded\n                retry_count += 1\n                \n                # Try to get retry-after header\n                retry_after = e.response.headers.get('Retry-After', 5) if hasattr(e, 'response') else 5\n                try:\n                    retry_after = int(retry_after)\n                except (ValueError, TypeError):\n                    retry_after = 5\n                \n                if retry_count \u003c= max_retries:\n                    print(f\"⚠️  Throttled (429): Rate limit exceeded, attempt {retry_count}/{max_retries}\")\n                    print(f\"Error message: {e.message}\")\n                    print(f\"Waiting {retry_after} seconds before retry...\")\n                    time.sleep(retry_after)\n                    continue  # Retry the operation\n                else:\n                    print(f\"❌ Max retries exceeded after throttling\")\n                    raise\n                    \n            else:\n                # Other HTTP errors\n                print(f\"❌ HTTP Error {status_code}: {e.message}\")\n                print(f\"Response: {e.response.text() if hasattr(e, 'response') else 'N/A'}\")\n                raise\n                \n        except ResourceNotFoundError as e:\n            # Alternative exception type that may be raised for 404\n            print(f\"❌ Resource Not Found: {e.message}\")\n            raise\n            \n        except Exception as e:\n            # Catch-all for unexpected errors\n            print(f\"❌ Unexpected error: {type(e).__name__}: {str(e)}\")\n            raise\n\n\ndef handle_set_secret(secret_client: SecretClient, secret_name: str, secret_value: str):\n    \"\"\"\n    Example: Set a secret with error handling\n    \"\"\"\n    try:\n        secret = secret_client.set_secret(secret_name, secret_value)\n        print(f\"✅ Secret '{secret_name}' created/updated successfully\")\n        return secret\n        \n    except HttpResponseError as e:\n        if e.status_code == 403:\n            print(f\"❌ Access Denied (403): Cannot set secret - insufficient permissions\")\n            print(f\"Required role: 'Key Vault Secrets Officer'\")\n            print(f\"Error: {e.message}\")\n        elif e.status_code == 409:\n            print(f\"❌ Conflict (409): Secret may be in soft-deleted state\")\n            print(f\"Error: {e.message}\")\n            print(\"Action: Recover or purge the deleted secret first\")\n        else:\n            print(f\"❌ Error setting secret: {e.status_code} - {e.message}\")\n        raise\n\n\ndef handle_delete_secret(secret_client: SecretClient, secret_name: str):\n    \"\"\"\n    Example: Delete a secret with error handling\n    \"\"\"\n    try:\n        poller = secret_client.begin_delete_secret(secret_name)\n        deleted_secret = poller.result()\n        print(f\"✅ Secret '{secret_name}' deleted (soft-delete)\")\n        print(f\"   Recovery ID: {deleted_secret.recovery_id}\")\n        print(f\"   Can be recovered until: {deleted_secret.scheduled_purge_date}\")\n        return deleted_secret\n        \n    except HttpResponseError as e:\n        if e.status_code == 403:\n            print(f\"❌ Access Denied (403): Cannot delete secret\")\n            print(f\"Required role: 'Key Vault Secrets Officer'\")\n        elif e.status_code == 404:\n            print(f\"❌ Secret '{secret_name}' not found - may already be deleted\")\n        else:\n            print(f\"❌ Error deleting secret: {e.status_code} - {e.message}\")\n        raise\n\n\ndef handle_recover_deleted_secret(secret_client: SecretClient, secret_name: str):\n    \"\"\"\n    Example: Recover a soft-deleted secret\n    \"\"\"\n    try:\n        poller = secret_client.begin_recover_deleted_secret(secret_name)\n        recovered_secret = poller.result()\n        print(f\"✅ Secret '{secret_name}' recovered successfully\")\n        return recovered_secret\n        \n    except HttpResponseError as e:\n        if e.status_code == 403:\n            print(f\"❌ Access Denied (403): Cannot recover secret\")\n            print(f\"Required role: 'Key Vault Secrets Officer'\")\n        elif e.status_code == 404:\n            print(f\"❌ Secret '{secret_name}' not found in deleted state\")\n        else:\n            print(f\"❌ Error recovering secret: {e.status_code} - {e.message}\")\n        raise\n\n\ndef comprehensive_error_inspection(secret_client: SecretClient, secret_name: str):\n    \"\"\"\n    Example: Detailed error inspection showing all available properties\n    \"\"\"\n    try:\n        secret = secret_client.get_secret(secret_name)\n        return secret\n        \n    except HttpResponseError as e:\n        print(\"=== Detailed Error Information ===\")\n        print(f\"Status Code: {e.status_code}\")\n        print(f\"Reason: {e.reason}\")\n        print(f\"Message: {e.message}\")\n        \n        # Error object (if available)\n        if hasattr(e, 'error') and e.error:\n            print(f\"Error Code: {e.error.code}\")\n            print(f\"Error Message: {e.error.message}\")\n        \n        # Response object (if available)\n        if hasattr(e, 'response'):\n            print(f\"Response Headers: {dict(e.response.headers)}\")\n            print(f\"Response Text: {e.response.text()}\")\n        \n        # Additional context\n        print(f\"Exception Type: {type(e).__name__}\")\n        print(f\"Exception String: {str(e)}\")\n        \n        raise\n\n\n# Main demonstration\nif __name__ == \"__main__\":\n    # Initialize client\n    vault_url = \"https://your-keyvault-name.vault.azure.net/\"\n    \n    try:\n        credential = DefaultAzureCredential()\n        client = SecretClient(vault_url=vault_url, credential=credential)\n        \n        print(\"=== Example 1: Handling 404 (Not Found) ===\")\n        try:\n            handle_get_secret_with_retry(client, \"non-existent-secret\")\n        except HttpResponseError:\n            pass\n        \n        print(\"\\n=== Example 2: Handling 403 (Access Denied) ===\")\n        print(\"(Simulated - would occur with insufficient RBAC permissions)\")\n        \n        print(\"\\n=== Example 3: Handling 429 (Throttling) ===\")\n        print(\"(Automatic retry with exponential backoff)\")\n        \n        print(\"\\n=== Example 4: Working with Soft-Deleted Secrets ===\")\n        secret_name = \"test-secret\"\n        \n        # Set a secret\n        handle_set_secret(client, secret_name, \"test-value\")\n        \n        # Delete it (soft-delete)\n        handle_delete_secret(client, secret_name)\n        \n        # Try to get it - will get 404\n        print(\"\\nAttempting to get soft-deleted secret:\")\n        try:\n            handle_get_secret_with_retry(client, secret_name)\n        except HttpResponseError:\n            pass\n        \n        # Recover it\n        print(\"\\nRecovering the secret:\")\n        handle_recover_deleted_secret(client, secret_name)\n        \n        # Now we can get it again\n        handle_get_secret_with_retry(client, secret_name)\n        \n    except Exception as e:\n        print(f\"\\n❌ Failed to initialize: {e}\")\n\n\n\"\"\"\nKEY TAKEAWAYS:\n\n1. ALWAYS catch HttpResponseError from azure.core.exceptions\n2. Check e.status_code to handle specific HTTP errors\n3. Use e.message for human-readable error description\n4. For 429 (throttling), implement retry logic with Retry-After header\n5. For 404, check if secret is soft-deleted using get_deleted_secret()\n6. For 403, verify RBAC role assignments (Key Vault Secrets User/Officer)\n7. Soft-deleted secrets cannot be accessed until recovered or purged\n8. Use ResourceNotFoundError as alternative for 404 handling\n\nRBAC Roles Required:\n- Read secrets: 'Key Vault Secrets User'\n- Create/Update/Delete secrets: 'Key Vault Secrets Officer'\n- Purge/Recover secrets: 'Key Vault Secrets Officer'\n\"\"\"\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-error-handling-azure-mcp-370279334/keyvault_error_handling.py"}
```

## Generated Files

- `keyvault_error_handling.py`

## Copilot Response

I'll create a comprehensive example showing proper error handling patterns for Azure Key Vault secrets in Python.

Perfect! I've created a comprehensive example showing proper error handling for Azure Key Vault secrets in Python.

## Key Points:

**Error Inspection:**
```python
except HttpResponseError as e:
    status_code = e.status_code  # HTTP status code (403, 404, 429, etc.)
    message = e.message          # Human-readable error message
    error_code = e.error.code    # Specific error code (if available)
```

**Common Scenarios:**

1. **403 (Access Denied)** - Missing RBAC role. Requires `Key Vault Secrets User` (read) or `Key Vault Secrets Officer` (write/delete)

2. **404 (Not Found)** - Secret doesn't exist. Check if it's soft-deleted using `get_deleted_secret()`

3. **429 (Throttling)** - Rate limit hit. Implement retry logic with `Retry-After` header

**Soft-Deleted Secrets:**
- When you delete a secret, it goes into soft-deleted state (if soft-delete is enabled)
- Trying to `get_secret()` on a soft-deleted secret returns **404**
- Use `get_deleted_secret()` to check if it exists in deleted state
- Use `begin_recover_deleted_secret()` to restore it
- You cannot create a new secret with the same name until the deleted one is purged

The example includes retry logic for throttling, detailed error inspection, and practical patterns for all scenarios you mentioned!

## Verification

**Result:** ✅ PASS

**Summary:** Code fully meets requirements with comprehensive error handling for Azure Key Vault including 403/404/429 scenarios, soft-delete handling, and proper exception inspection.

**Reasoning:**

The generated code comprehensively addresses all requirements from the original prompt. It correctly imports and uses HttpResponseError and ResourceNotFoundError from azure.core.exceptions. It demonstrates extracting status_code (e.status_code), error.code (e.error.code), and message (e.message) from exceptions. The code handles 403 errors with clear RBAC role guidance ('Key Vault Secrets User' or 'Key Vault Secrets Officer'), handles 404 errors by checking for soft-deleted secrets using get_deleted_secret(), and implements retry logic for 429 throttling with Retry-After header parsing. It shows soft-delete aware error handling with begin_recover_deleted_secret() and includes a complete example demonstrating what happens when trying to get a soft-deleted secret (returns 404 until recovered). The code is syntactically valid Python using the correct azure-keyvault-secrets SDK. The only minor omission is not showing SecretClient retry configuration via constructor kwargs, but it provides comprehensive manual retry logic instead. The code includes excellent documentation, multiple practical examples, and clear explanations of RBAC requirements.

---

[← Back to Summary](../../../../../../summary.md)
