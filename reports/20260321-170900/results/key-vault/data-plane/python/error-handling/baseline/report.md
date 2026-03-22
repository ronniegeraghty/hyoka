# Evaluation Report: key-vault-dp-python-error-handling

**Config:** baseline | **Result:** ✅ PASSED | **Duration:** 82.5s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-error-handling` |
| Config | baseline |
| Result | ✅ PASSED |
| Duration | 82.5s |
| Timestamp | 2026-03-22T00:09:00Z |
| Files Generated | 1 |
| Event Count | 26 |

## Configuration

- **model:** claude-sonnet-4.5
- **name:** baseline

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
1. Try/except patterns for common errors (403, 404, 429)
2. How to inspect status_code and error message on HttpResponseError
3. Explanation of what happens with soft-deleted secrets
4. Examples using azure-keyvault-secrets SDK

This is a straightforward educational/documentation request. I should:
1. Create a comprehensive example script showing error handling patterns
2. Include examples for each error scenario
3. Explain the soft-deleted secret behavior
4. Show how to access HttpResponseError properties

I don't need to create a plan for this - it's a single educational script/example, not a multi-step implementation task. I'll create a well-documented Python example file showing the patterns.

## Tool Calls

### 🔧 1. report_intent

**Input:**

```json
{"intent":"Creating error handling examples"}
```

### 🔧 2. create

**Input:**

```json
{"file_text":"\"\"\"\nAzure Key Vault Secrets - Error Handling Examples\n\nThis module demonstrates proper error handling patterns when working with\nAzure Key Vault secrets using the azure-keyvault-secrets SDK.\n\"\"\"\n\nfrom azure.identity import DefaultAzureCredential\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.core.exceptions import (\n    HttpResponseError,\n    ResourceNotFoundError,\n    ServiceRequestError,\n)\nimport time\n\n\ndef create_secret_client(vault_url: str) -\u003e SecretClient:\n    \"\"\"Create and return a SecretClient instance.\"\"\"\n    credential = DefaultAzureCredential()\n    return SecretClient(vault_url=vault_url, credential=credential)\n\n\n# ==============================================================================\n# EXAMPLE 1: Basic Error Handling - Inspecting HttpResponseError\n# ==============================================================================\n\ndef get_secret_with_basic_error_handling(client: SecretClient, secret_name: str):\n    \"\"\"\n    Demonstrates basic error handling with HttpResponseError inspection.\n    Shows how to access status_code, error message, and other properties.\n    \"\"\"\n    try:\n        secret = client.get_secret(secret_name)\n        print(f\"Successfully retrieved secret: {secret.name}\")\n        return secret.value\n    \n    except HttpResponseError as e:\n        # HttpResponseError has several useful properties:\n        # - status_code: HTTP status code (403, 404, 429, etc.)\n        # - message: Human-readable error message\n        # - error: Detailed error information\n        # - reason: HTTP reason phrase\n        \n        print(f\"HTTP Error occurred:\")\n        print(f\"  Status Code: {e.status_code}\")\n        print(f\"  Message: {e.message}\")\n        print(f\"  Reason: {e.reason}\")\n        \n        # The error attribute contains structured error details\n        if hasattr(e, 'error') and e.error:\n            print(f\"  Error Code: {e.error.code if hasattr(e.error, 'code') else 'N/A'}\")\n            print(f\"  Error Message: {e.error.message if hasattr(e.error, 'message') else 'N/A'}\")\n        \n        raise\n\n\n# ==============================================================================\n# EXAMPLE 2: Handling Specific Status Codes (403, 404, 429)\n# ==============================================================================\n\ndef get_secret_with_specific_error_handling(\n    client: SecretClient, \n    secret_name: str,\n    max_retries: int = 3\n):\n    \"\"\"\n    Demonstrates handling specific HTTP status codes:\n    - 403 Forbidden: Access denied (insufficient RBAC permissions)\n    - 404 Not Found: Secret doesn't exist\n    - 429 Too Many Requests: Rate limiting/throttling\n    \"\"\"\n    retry_count = 0\n    base_delay = 1  # seconds\n    \n    while retry_count \u003c= max_retries:\n        try:\n            secret = client.get_secret(secret_name)\n            print(f\"Successfully retrieved secret: {secret.name}\")\n            return secret.value\n        \n        except HttpResponseError as e:\n            \n            # 403 Forbidden - Access Denied\n            if e.status_code == 403:\n                print(f\"ERROR: Access denied to secret '{secret_name}'\")\n                print(\"Your application identity does not have the required permissions.\")\n                print(\"Required RBAC role: 'Key Vault Secrets User' or 'Key Vault Secrets Officer'\")\n                print(f\"Error details: {e.message}\")\n                # Don't retry - this requires permission changes\n                raise\n            \n            # 404 Not Found - Secret doesn't exist\n            elif e.status_code == 404:\n                print(f\"ERROR: Secret '{secret_name}' not found in the Key Vault\")\n                print(f\"Error details: {e.message}\")\n                # Don't retry - secret doesn't exist\n                raise\n            \n            # 429 Too Many Requests - Throttling\n            elif e.status_code == 429:\n                retry_count += 1\n                \n                if retry_count \u003e max_retries:\n                    print(f\"ERROR: Max retries ({max_retries}) exceeded due to throttling\")\n                    raise\n                \n                # Check for Retry-After header\n                retry_after = None\n                if hasattr(e, 'response') and e.response:\n                    retry_after = e.response.headers.get('Retry-After')\n                \n                if retry_after:\n                    # Retry-After can be in seconds or a date\n                    try:\n                        delay = int(retry_after)\n                    except ValueError:\n                        # If it's a date, use exponential backoff\n                        delay = base_delay * (2 ** (retry_count - 1))\n                else:\n                    # Exponential backoff: 1s, 2s, 4s, 8s, etc.\n                    delay = base_delay * (2 ** (retry_count - 1))\n                \n                print(f\"Throttled (429). Retry {retry_count}/{max_retries} after {delay}s...\")\n                time.sleep(delay)\n                continue\n            \n            # Other HTTP errors\n            else:\n                print(f\"ERROR: Unexpected HTTP error (status {e.status_code})\")\n                print(f\"Message: {e.message}\")\n                raise\n        \n        except ServiceRequestError as e:\n            # Network errors, DNS failures, connection timeouts\n            print(f\"ERROR: Service request failed - {str(e)}\")\n            print(\"This typically indicates network connectivity issues\")\n            raise\n        \n        except Exception as e:\n            # Catch any other unexpected errors\n            print(f\"ERROR: Unexpected error occurred - {type(e).__name__}: {str(e)}\")\n            raise\n\n\n# ==============================================================================\n# EXAMPLE 3: Handling Soft-Deleted Secrets\n# ==============================================================================\n\ndef get_secret_handling_soft_delete(client: SecretClient, secret_name: str):\n    \"\"\"\n    Demonstrates what happens when trying to get a soft-deleted secret.\n    \n    IMPORTANT: When you try to get a soft-deleted secret using get_secret(),\n    you will receive a 404 (Not Found) error - the same as if the secret\n    never existed. You CANNOT retrieve the value of a soft-deleted secret.\n    \n    To work with soft-deleted secrets, you must use:\n    - get_deleted_secret() - to see deleted secret metadata\n    - recover_deleted_secret() - to restore it\n    - purge_deleted_secret() - to permanently delete it\n    \"\"\"\n    try:\n        # This will fail with 404 if the secret is soft-deleted\n        secret = client.get_secret(secret_name)\n        print(f\"Secret found: {secret.name}\")\n        return secret.value\n    \n    except HttpResponseError as e:\n        if e.status_code == 404:\n            print(f\"Secret '{secret_name}' not found.\")\n            print(\"Checking if it's soft-deleted...\")\n            \n            try:\n                # Try to get the deleted secret metadata\n                deleted_secret = client.get_deleted_secret(secret_name)\n                \n                print(f\"\\nFound soft-deleted secret:\")\n                print(f\"  Name: {deleted_secret.name}\")\n                print(f\"  Deleted On: {deleted_secret.deleted_on}\")\n                print(f\"  Scheduled Purge Date: {deleted_secret.scheduled_purge_date}\")\n                print(f\"  Recovery ID: {deleted_secret.recovery_id}\")\n                print(\"\\nThe secret is soft-deleted. You can:\")\n                print(\"  1. Recover it using recover_deleted_secret()\")\n                print(\"  2. Wait for automatic purge\")\n                print(\"  3. Manually purge it using purge_deleted_secret() (if you have permissions)\")\n                \n                return None\n            \n            except HttpResponseError as deleted_check_error:\n                if deleted_check_error.status_code == 404:\n                    print(\"Secret does not exist (not found in active or deleted state)\")\n                elif deleted_check_error.status_code == 403:\n                    print(\"Cannot check deleted secrets - insufficient permissions\")\n                    print(\"Required: 'Key Vault Secrets Officer' role or 'List' + 'Get' on deleted secrets\")\n                else:\n                    print(f\"Error checking deleted secrets: {deleted_check_error.message}\")\n                \n                raise deleted_check_error\n        else:\n            raise\n\n\n# ==============================================================================\n# EXAMPLE 4: Complete Error Handling with Recovery\n# ==============================================================================\n\ndef get_secret_with_recovery(client: SecretClient, secret_name: str):\n    \"\"\"\n    Complete example that handles soft-deleted secrets by recovering them.\n    \"\"\"\n    try:\n        secret = client.get_secret(secret_name)\n        return secret.value\n    \n    except HttpResponseError as e:\n        if e.status_code == 404:\n            # Check if soft-deleted and attempt recovery\n            try:\n                deleted_secret = client.get_deleted_secret(secret_name)\n                print(f\"Secret '{secret_name}' is soft-deleted. Attempting recovery...\")\n                \n                # Recover the secret\n                recover_operation = client.begin_recover_deleted_secret(secret_name)\n                recovered_secret = recover_operation.result()\n                \n                print(f\"Secret '{secret_name}' recovered successfully!\")\n                \n                # Now retrieve it\n                secret = client.get_secret(secret_name)\n                return secret.value\n            \n            except HttpResponseError as recovery_error:\n                if recovery_error.status_code == 403:\n                    print(\"Cannot recover secret - insufficient permissions\")\n                    print(\"Required: 'Key Vault Secrets Officer' role or 'Recover' permission\")\n                else:\n                    print(f\"Recovery failed: {recovery_error.message}\")\n                raise\n        \n        elif e.status_code == 403:\n            print(f\"Access denied. Check RBAC role assignments for your identity.\")\n            print(\"Required role: 'Key Vault Secrets User' (for get) or 'Key Vault Secrets Officer' (for all operations)\")\n            raise\n        \n        elif e.status_code == 429:\n            print(\"Rate limited. Implement retry logic with exponential backoff.\")\n            raise\n        \n        else:\n            print(f\"Unexpected error (status {e.status_code}): {e.message}\")\n            raise\n\n\n# ==============================================================================\n# EXAMPLE 5: Using ResourceNotFoundError (Convenience Exception)\n# ==============================================================================\n\ndef get_secret_using_resource_not_found(client: SecretClient, secret_name: str):\n    \"\"\"\n    The SDK also provides ResourceNotFoundError as a convenience exception\n    that you can catch specifically for 404 errors.\n    \"\"\"\n    try:\n        secret = client.get_secret(secret_name)\n        return secret.value\n    \n    except ResourceNotFoundError as e:\n        # This is raised for 404 errors - more specific than HttpResponseError\n        print(f\"Secret '{secret_name}' not found\")\n        print(f\"Status code: {e.status_code}\")  # Will be 404\n        print(f\"Message: {e.message}\")\n        raise\n    \n    except HttpResponseError as e:\n        # Handle other HTTP errors (403, 429, etc.)\n        if e.status_code == 403:\n            print(\"Access denied - check RBAC permissions\")\n        elif e.status_code == 429:\n            print(\"Throttled - implement retry logic\")\n        else:\n            print(f\"HTTP error {e.status_code}: {e.message}\")\n        raise\n\n\n# ==============================================================================\n# USAGE EXAMPLES\n# ==============================================================================\n\nif __name__ == \"__main__\":\n    # Replace with your Key Vault URL\n    vault_url = \"https://your-keyvault-name.vault.azure.net/\"\n    \n    try:\n        client = create_secret_client(vault_url)\n        \n        # Example 1: Basic error handling\n        print(\"=\" * 70)\n        print(\"Example 1: Basic Error Handling\")\n        print(\"=\" * 70)\n        get_secret_with_basic_error_handling(client, \"my-secret\")\n        \n        # Example 2: Specific status code handling\n        print(\"\\n\" + \"=\" * 70)\n        print(\"Example 2: Handling Specific Status Codes\")\n        print(\"=\" * 70)\n        get_secret_with_specific_error_handling(client, \"my-secret\", max_retries=3)\n        \n        # Example 3: Soft-deleted secret handling\n        print(\"\\n\" + \"=\" * 70)\n        print(\"Example 3: Handling Soft-Deleted Secrets\")\n        print(\"=\" * 70)\n        get_secret_handling_soft_delete(client, \"deleted-secret\")\n        \n        # Example 4: With automatic recovery\n        print(\"\\n\" + \"=\" * 70)\n        print(\"Example 4: Error Handling with Recovery\")\n        print(\"=\" * 70)\n        get_secret_with_recovery(client, \"my-secret\")\n        \n        # Example 5: Using ResourceNotFoundError\n        print(\"\\n\" + \"=\" * 70)\n        print(\"Example 5: Using ResourceNotFoundError\")\n        print(\"=\" * 70)\n        get_secret_using_resource_not_found(client, \"my-secret\")\n    \n    except Exception as e:\n        print(f\"\\nFatal error: {type(e).__name__}: {str(e)}\")\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-error-handling-baseline-593784303/keyvault_error_handling.py"}
```

## Generated Files

- `keyvault_error_handling.py`

## Copilot Response

I'll create a comprehensive example showing proper error handling patterns for Azure Key Vault secrets in Python.

I've created a comprehensive example file showing proper error handling patterns for Azure Key Vault secrets in Python! Here's what it covers:

## Key Points:

**1. HttpResponseError Properties:**
- `status_code` - HTTP status (403, 404, 429, etc.)
- `message` - Human-readable error message
- `reason` - HTTP reason phrase
- `error.code` and `error.message` - Structured error details

**2. Common Status Codes:**
- **403 (Forbidden)**: Access denied - need "Key Vault Secrets User" RBAC role
- **404 (Not Found)**: Secret doesn't exist (or is soft-deleted)
- **429 (Too Many Requests)**: Throttling - implement exponential backoff with retry

**3. Soft-Deleted Secrets:**
When you try to `get_secret()` on a soft-deleted secret, you get **404 - same as if it never existed**. You cannot retrieve the value. Instead:
- Use `get_deleted_secret()` to see metadata
- Use `begin_recover_deleted_secret()` to restore it
- Use `purge_deleted_secret()` to permanently delete

**4. Best Practices:**
- Use `ResourceNotFoundError` for cleaner 404 handling
- Implement exponential backoff for 429 errors
- Check `Retry-After` header when throttled
- Don't retry 403/404 errors
- Handle `ServiceRequestError` for network issues

The file includes 5 complete examples showing different patterns. Run them with your Key Vault URL to see the error handling in action!

## Verification

**Result:** ✅ PASS

**Summary:** Code fully meets requirements with comprehensive error handling for 403/404/429, proper exception inspection, and soft-delete scenarios.

**Reasoning:**

The generated code comprehensively addresses all requirements from the original prompt. It correctly uses the azure-keyvault-secrets SDK with proper imports from azure.core.exceptions (HttpResponseError, ResourceNotFoundError). The code demonstrates: (1) Inspecting status_code, error.code, and message properties on HttpResponseError with clear examples in Example 1; (2) Proper 403 handling with specific guidance about RBAC roles ('Key Vault Secrets User' and 'Key Vault Secrets Officer') and explaining not to retry permission errors; (3) Complete 404 handling that distinguishes between non-existent secrets and soft-deleted secrets using get_deleted_secret(); (4) Robust 429 throttling handling with exponential backoff, Retry-After header inspection, and configurable max_retries; (5) Comprehensive soft-delete handling explaining that get_secret() returns 404 for soft-deleted secrets, with examples of get_deleted_secret() and begin_recover_deleted_secret(); (6) Multiple clear try/except patterns across 5 well-structured examples. The code is syntactically valid, well-commented, and demonstrates both HttpResponseError and ResourceNotFoundError usage. Minor note: while the expected coverage mentions 'SecretClient retry configuration via kwargs', the code demonstrates manual retry logic instead, which is practical and acceptable. The code fully answers the original question about properly handling errors when working with Azure Key Vault secrets.

---

[← Back to Summary](../../../../../../summary.md)
