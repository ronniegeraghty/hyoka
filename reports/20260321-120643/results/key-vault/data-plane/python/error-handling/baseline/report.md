# Evaluation Report: key-vault-dp-python-error-handling

**Config:** baseline | **Result:** ✅ PASSED | **Duration:** 93.8s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-error-handling` |
| Config | baseline |
| Result | ✅ PASSED |
| Duration | 93.8s |
| Timestamp | 2026-03-21T19:06:43Z |
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

The user is asking for a practical guide on error handling with Azure Key Vault secrets in Python using the `azure-keyvault-secrets` SDK. They want to understand:

1. How to handle specific HTTP errors (403, 404, 429)
2. Try/except patterns with the SDK
3. How to inspect status_code and error messages from HttpResponseError
4. What happens when trying to get a soft-deleted secret

This is a straightforward question that I can answer directly without needing to explore a codebase or make changes. I should provide clear, practical examples showing:

- The proper exception types to catch from azure-keyvault-secrets
- How to access status_code and error messages
- Examples for each error scenario they mentioned
- Explanation of soft-deleted secret behavior

I'll create a comprehensive example file demonstrating these patterns.

## Tool Calls

### 🔧 1. report_intent

**Input:**

```json
{"intent":"Creating error handling examples"}
```

### 🔧 2. create

**Input:**

```json
{"file_text":"\"\"\"\nAzure Key Vault Secrets - Error Handling Examples\n\nThis module demonstrates proper error handling patterns when working with\nAzure Key Vault secrets using the azure-keyvault-secrets SDK.\n\"\"\"\n\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.identity import DefaultAzureCredential\nfrom azure.core.exceptions import HttpResponseError, ResourceNotFoundError\nimport time\n\n\ndef handle_specific_errors(secret_client: SecretClient, secret_name: str):\n    \"\"\"\n    Example: Handle specific HTTP status codes when getting a secret.\n    \n    Common scenarios:\n    - 403 Forbidden: Missing RBAC permissions (Key Vault Secrets User role)\n    - 404 Not Found: Secret doesn't exist\n    - 429 Too Many Requests: Rate limiting/throttling\n    \"\"\"\n    try:\n        secret = secret_client.get_secret(secret_name)\n        print(f\"Successfully retrieved secret: {secret.name}\")\n        return secret.value\n    \n    except HttpResponseError as e:\n        # Inspect the status code\n        status_code = e.status_code\n        \n        # Access error message\n        error_message = e.message\n        \n        # Get additional error details from the response\n        error_code = e.error.code if hasattr(e, 'error') and e.error else None\n        \n        print(f\"HTTP Error: Status {status_code}\")\n        print(f\"Error Message: {error_message}\")\n        print(f\"Error Code: {error_code}\")\n        \n        if status_code == 403:\n            # Access Denied - Missing RBAC permissions\n            print(\"ACCESS DENIED: The application identity lacks the required RBAC role.\")\n            print(\"Required role: 'Key Vault Secrets User' or 'Key Vault Secrets Officer'\")\n            print(\"Assign permissions using: az role assignment create --role 'Key Vault Secrets User' ...\")\n            # Option: Return None, raise custom exception, or use default value\n            return None\n        \n        elif status_code == 404:\n            # Secret Not Found\n            print(f\"SECRET NOT FOUND: '{secret_name}' does not exist in the Key Vault.\")\n            print(\"The secret may have been deleted or never created.\")\n            # Option: Create the secret, return default, or raise\n            return None\n        \n        elif status_code == 429:\n            # Rate Limiting / Throttling\n            print(\"THROTTLING: Too many requests to Key Vault.\")\n            \n            # Check for Retry-After header\n            retry_after = e.response.headers.get('Retry-After')\n            if retry_after:\n                wait_time = int(retry_after)\n                print(f\"Retry after {wait_time} seconds (from Retry-After header)\")\n            else:\n                # Use exponential backoff if no Retry-After header\n                wait_time = 5\n                print(f\"Using default backoff: {wait_time} seconds\")\n            \n            time.sleep(wait_time)\n            # Retry the operation\n            return handle_specific_errors(secret_client, secret_name)\n        \n        else:\n            # Other HTTP errors\n            print(f\"Unexpected HTTP error: {status_code}\")\n            raise\n\n\ndef get_secret_with_retry(secret_client: SecretClient, secret_name: str, max_retries: int = 3):\n    \"\"\"\n    Example: Implement retry logic with exponential backoff for transient errors.\n    \"\"\"\n    for attempt in range(max_retries):\n        try:\n            secret = secret_client.get_secret(secret_name)\n            return secret.value\n        \n        except HttpResponseError as e:\n            if e.status_code == 429:\n                # Throttling - apply backoff\n                retry_after = e.response.headers.get('Retry-After')\n                wait_time = int(retry_after) if retry_after else (2 ** attempt)\n                \n                print(f\"Throttled. Retry {attempt + 1}/{max_retries} after {wait_time}s\")\n                \n                if attempt \u003c max_retries - 1:\n                    time.sleep(wait_time)\n                    continue\n                else:\n                    print(\"Max retries reached for throttling.\")\n                    raise\n            \n            elif e.status_code in [500, 502, 503, 504]:\n                # Server errors - retry with backoff\n                wait_time = 2 ** attempt\n                print(f\"Server error {e.status_code}. Retry {attempt + 1}/{max_retries} after {wait_time}s\")\n                \n                if attempt \u003c max_retries - 1:\n                    time.sleep(wait_time)\n                    continue\n                else:\n                    print(\"Max retries reached for server errors.\")\n                    raise\n            \n            else:\n                # Non-retryable errors (403, 404, etc.)\n                raise\n\n\ndef handle_soft_deleted_secret(secret_client: SecretClient, secret_name: str):\n    \"\"\"\n    Example: Handle soft-deleted secrets.\n    \n    When you try to get a soft-deleted secret:\n    - If the secret is soft-deleted (not purged), you'll get a 404 Not Found error\n    - You CANNOT retrieve a soft-deleted secret using get_secret()\n    - You must either:\n      1. Recover it first using recover_deleted_secret()\n      2. Or access it via get_deleted_secret() to view properties only (not the value)\n    \n    Soft delete is enabled by default on Key Vaults with purge protection.\n    Deleted secrets remain in soft-deleted state for the retention period (default 90 days).\n    \"\"\"\n    try:\n        # This will fail with 404 if the secret is soft-deleted\n        secret = secret_client.get_secret(secret_name)\n        print(f\"Secret retrieved: {secret.name}\")\n        return secret.value\n    \n    except HttpResponseError as e:\n        if e.status_code == 404:\n            print(f\"Secret '{secret_name}' not found (404).\")\n            print(\"Checking if it's in soft-deleted state...\")\n            \n            try:\n                # Try to get the deleted secret metadata\n                deleted_secret = secret_client.get_deleted_secret(secret_name)\n                \n                print(f\"SOFT-DELETED SECRET FOUND:\")\n                print(f\"  Name: {deleted_secret.name}\")\n                print(f\"  Deleted On: {deleted_secret.deleted_date}\")\n                print(f\"  Scheduled Purge Date: {deleted_secret.scheduled_purge_date}\")\n                print(f\"  Recovery ID: {deleted_secret.recovery_id}\")\n                \n                # NOTE: deleted_secret.value is None - you cannot get the actual secret value\n                # You must recover it first to access the value\n                \n                print(\"\\nTo recover this secret, use:\")\n                print(\"  recover_operation = secret_client.begin_recover_deleted_secret(secret_name)\")\n                print(\"  recovered_secret = recover_operation.result()\")\n                \n                return None\n            \n            except HttpResponseError as deleted_error:\n                if deleted_error.status_code == 404:\n                    print(f\"Secret '{secret_name}' is permanently purged or never existed.\")\n                else:\n                    print(f\"Error checking deleted secrets: {deleted_error.message}\")\n                return None\n        else:\n            raise\n\n\ndef recover_deleted_secret_example(secret_client: SecretClient, secret_name: str):\n    \"\"\"\n    Example: Recover a soft-deleted secret.\n    \"\"\"\n    try:\n        # Check if secret is soft-deleted\n        deleted_secret = secret_client.get_deleted_secret(secret_name)\n        print(f\"Found soft-deleted secret: {deleted_secret.name}\")\n        \n        # Begin recovery (this is a long-running operation)\n        print(\"Starting recovery operation...\")\n        recover_operation = secret_client.begin_recover_deleted_secret(secret_name)\n        \n        # Wait for recovery to complete\n        recovered_secret = recover_operation.result()\n        print(f\"Secret recovered: {recovered_secret.name}\")\n        \n        # Now you can get the secret value\n        secret = secret_client.get_secret(secret_name)\n        return secret.value\n    \n    except HttpResponseError as e:\n        if e.status_code == 404:\n            print(f\"Secret '{secret_name}' is not in deleted state or doesn't exist.\")\n        elif e.status_code == 403:\n            print(\"Access denied. Need 'Key Vault Secrets Officer' role to recover secrets.\")\n        else:\n            print(f\"Error during recovery: {e.message}\")\n        raise\n\n\ndef comprehensive_error_handling(secret_client: SecretClient, secret_name: str):\n    \"\"\"\n    Example: Comprehensive error handling covering all common scenarios.\n    \"\"\"\n    try:\n        secret = secret_client.get_secret(secret_name)\n        return secret.value\n    \n    except HttpResponseError as e:\n        # Extract error details\n        status_code = e.status_code\n        error_message = e.message\n        \n        # Log the full error details\n        print(f\"Error accessing secret '{secret_name}':\")\n        print(f\"  Status Code: {status_code}\")\n        print(f\"  Message: {error_message}\")\n        \n        # Additional details if available\n        if hasattr(e, 'error') and e.error:\n            print(f\"  Error Code: {e.error.code}\")\n            if hasattr(e.error, 'message'):\n                print(f\"  Detailed Message: {e.error.message}\")\n        \n        # Handle based on status code\n        error_handlers = {\n            403: lambda: handle_403_forbidden(secret_name),\n            404: lambda: handle_404_not_found(secret_client, secret_name),\n            429: lambda: handle_429_throttling(e),\n            500: lambda: handle_5xx_server_error(status_code),\n            502: lambda: handle_5xx_server_error(status_code),\n            503: lambda: handle_5xx_server_error(status_code),\n            504: lambda: handle_5xx_server_error(status_code),\n        }\n        \n        handler = error_handlers.get(status_code)\n        if handler:\n            return handler()\n        else:\n            print(f\"Unhandled HTTP error: {status_code}\")\n            raise\n    \n    except Exception as e:\n        # Catch other exceptions (network errors, authentication failures, etc.)\n        print(f\"Unexpected error: {type(e).__name__}: {str(e)}\")\n        raise\n\n\ndef handle_403_forbidden(secret_name: str):\n    \"\"\"Handle 403 Forbidden errors.\"\"\"\n    print(\"\\n=== ACCESS DENIED (403) ===\")\n    print(f\"Your application identity doesn't have permission to read '{secret_name}'.\")\n    print(\"\\nRequired Azure RBAC role:\")\n    print(\"  - 'Key Vault Secrets User' (read secrets)\")\n    print(\"  - 'Key Vault Secrets Officer' (read, write, delete secrets)\")\n    print(\"\\nTo grant access:\")\n    print(\"  az role assignment create \\\\\")\n    print(\"    --role 'Key Vault Secrets User' \\\\\")\n    print(\"    --assignee \u003capp-id-or-principal-id\u003e \\\\\")\n    print(\"    --scope \u003ckey-vault-resource-id\u003e\")\n    return None\n\n\ndef handle_404_not_found(secret_client: SecretClient, secret_name: str):\n    \"\"\"Handle 404 Not Found errors.\"\"\"\n    print(\"\\n=== SECRET NOT FOUND (404) ===\")\n    print(f\"Secret '{secret_name}' doesn't exist in the Key Vault.\")\n    print(\"Possible reasons:\")\n    print(\"  1. Secret was never created\")\n    print(\"  2. Secret is soft-deleted (check with get_deleted_secret)\")\n    print(\"  3. Typo in secret name\")\n    \n    # Check if soft-deleted\n    try:\n        deleted = secret_client.get_deleted_secret(secret_name)\n        print(f\"\\n✓ Secret is SOFT-DELETED (deleted on {deleted.deleted_date})\")\n        print(f\"  Purge date: {deleted.scheduled_purge_date}\")\n        print(\"  Use begin_recover_deleted_secret() to restore it.\")\n    except:\n        print(\"\\n✗ Secret is not in soft-deleted state.\")\n    \n    return None\n\n\ndef handle_429_throttling(error: HttpResponseError):\n    \"\"\"Handle 429 Too Many Requests errors.\"\"\"\n    print(\"\\n=== THROTTLING (429) ===\")\n    print(\"Too many requests to Key Vault. Rate limits exceeded.\")\n    \n    retry_after = error.response.headers.get('Retry-After')\n    if retry_after:\n        print(f\"Retry after: {retry_after} seconds\")\n    else:\n        print(\"No Retry-After header. Use exponential backoff.\")\n    \n    print(\"\\nKey Vault rate limits:\")\n    print(\"  - GET secrets: 2000 requests per 10 seconds\")\n    print(\"  - All operations: 2000 requests per 10 seconds per vault\")\n    print(\"\\nConsider:\")\n    print(\"  - Caching secret values\")\n    print(\"  - Implementing exponential backoff\")\n    print(\"  - Reducing request frequency\")\n    \n    raise\n\n\ndef handle_5xx_server_error(status_code: int):\n    \"\"\"Handle 5xx server errors.\"\"\"\n    print(f\"\\n=== SERVER ERROR ({status_code}) ===\")\n    print(\"Azure Key Vault service is experiencing issues.\")\n    print(\"These are typically transient. Implement retry with exponential backoff.\")\n    raise\n\n\n# Example usage\nif __name__ == \"__main__\":\n    # Initialize the client\n    vault_url = \"https://your-vault-name.vault.azure.net/\"\n    credential = DefaultAzureCredential()\n    client = SecretClient(vault_url=vault_url, credential=credential)\n    \n    # Example 1: Handle specific errors\n    print(\"=== Example 1: Handle Specific Errors ===\")\n    handle_specific_errors(client, \"my-secret\")\n    \n    # Example 2: Retry logic\n    print(\"\\n=== Example 2: Retry with Backoff ===\")\n    get_secret_with_retry(client, \"my-secret\", max_retries=3)\n    \n    # Example 3: Soft-deleted secrets\n    print(\"\\n=== Example 3: Soft-Deleted Secrets ===\")\n    handle_soft_deleted_secret(client, \"deleted-secret\")\n    \n    # Example 4: Comprehensive error handling\n    print(\"\\n=== Example 4: Comprehensive Handling ===\")\n    comprehensive_error_handling(client, \"my-secret\")\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-error-handling-baseline-2085732884/keyvault_error_handling.py"}
```

## Generated Files

- `keyvault_error_handling.py`

## Copilot Response

I'll create a comprehensive example showing proper error handling patterns for Azure Key Vault secrets in Python.

Perfect! I've created a comprehensive example file showing proper error handling for Azure Key Vault secrets in Python.

## Key Points:

### **HttpResponseError Structure:**
```python
except HttpResponseError as e:
    status_code = e.status_code      # HTTP status code (403, 404, 429, etc.)
    message = e.message              # Error message
    error_code = e.error.code        # Azure-specific error code
```

### **Common Scenarios:**

1. **403 Forbidden** - Missing RBAC permissions. Need "Key Vault Secrets User" role assigned to your app's identity.

2. **404 Not Found** - Secret doesn't exist OR is soft-deleted (can't get value directly).

3. **429 Throttling** - Check `e.response.headers.get('Retry-After')` and implement exponential backoff.

### **Soft-Deleted Secrets:**
When you try `get_secret()` on a soft-deleted secret:
- ❌ Returns **404 Not Found** 
- ❌ You **cannot** get the secret value
- ✅ Use `get_deleted_secret()` to check deletion status (metadata only)
- ✅ Use `begin_recover_deleted_secret()` to restore it first
- Then call `get_secret()` to retrieve the value

The example file includes retry logic, comprehensive error handling patterns, and detailed explanations for each scenario you mentioned!

## Verification

**Result:** ✅ PASS

**Summary:** Code fully meets all requirements with comprehensive error handling examples for Azure Key Vault operations

**Reasoning:**

The generated code comprehensively meets all requirements from the original prompt. It correctly imports and uses HttpResponseError and ResourceNotFoundError from azure.core.exceptions, and SecretClient from azure.keyvault.secrets. The code demonstrates proper error handling for all requested scenarios: 403 (access denied with RBAC role explanations), 404 (secret not found with soft-delete detection), and 429 (throttling with Retry-After header inspection). It shows how to extract status_code, error.code, and message from HttpResponseError exceptions. The soft-delete scenario is thoroughly covered with get_deleted_secret() and begin_recover_deleted_secret() examples, explaining that soft-deleted secrets return 404 and must be recovered before accessing their values. The code includes retry logic with exponential backoff, explains Key Vault rate limits, and provides multiple complete working examples. While it doesn't show SecretClient retry configuration via constructor kwargs, it implements comprehensive manual retry logic which is a valid and common approach. The code is syntactically correct, well-documented, and would execute properly given appropriate Azure credentials and vault setup.

---

[← Back to Summary](../../../../../../summary.md)
