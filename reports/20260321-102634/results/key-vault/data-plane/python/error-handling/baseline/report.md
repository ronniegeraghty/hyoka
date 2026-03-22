# Evaluation Report: key-vault-dp-python-error-handling

**Config:** baseline | **Result:** ✅ PASSED | **Duration:** 75.6s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-error-handling` |
| Config | baseline |
| Result | ✅ PASSED |
| Duration | 75.6s |
| Timestamp | 2026-03-21T17:26:34Z |
| Files Generated | 1 |
| Event Count | 27 |

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

The user is asking for guidance on error handling with Azure Key Vault secrets in Python. They specifically want:
1. Try/except patterns for common errors (403, 404, 429)
2. How to inspect status_code and error message on HttpResponseError
3. Explanation of what happens with soft-deleted secrets
4. Use of the azure-keyvault-secrets SDK

This is a straightforward question asking for code examples and explanations. I should create a Python demonstration script showing:
- Proper error handling patterns
- How to catch and inspect HttpResponseError
- Examples of handling 403 (access denied), 404 (not found), and 429 (throttling)
- Explanation of soft-deleted secrets behavior

Since this is a demonstration/educational request and not a task to modify existing code, I should create an example file showing best practices. Let me create a comprehensive example with explanations.

## Tool Calls

### 🔧 1. report_intent

**Input:**

```json
{"intent":"Creating error handling examples"}
```

### 🔧 2. create

**Input:**

```json
{"file_text":"\"\"\"\nAzure Key Vault Secrets - Error Handling Patterns\n\nThis module demonstrates proper error handling when working with Azure Key Vault\nsecrets using the azure-keyvault-secrets SDK.\n\"\"\"\n\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.identity import DefaultAzureCredential\nfrom azure.core.exceptions import HttpResponseError, ResourceNotFoundError\nimport time\n\n\ndef handle_common_errors(vault_url: str, secret_name: str):\n    \"\"\"\n    Demonstrates handling common Key Vault errors: 403, 404, and 429.\n    \"\"\"\n    credential = DefaultAzureCredential()\n    client = SecretClient(vault_url=vault_url, credential=credential)\n    \n    try:\n        secret = client.get_secret(secret_name)\n        print(f\"Successfully retrieved secret: {secret.name}\")\n        return secret.value\n        \n    except HttpResponseError as e:\n        # Inspect the status code to handle different error scenarios\n        status_code = e.status_code\n        error_message = e.message\n        \n        if status_code == 403:\n            # Access Denied - Missing RBAC permissions\n            print(f\"Access Denied (403): {error_message}\")\n            print(\"Your app identity needs 'Key Vault Secrets User' or \"\n                  \"'Key Vault Secrets Officer' role assigned.\")\n            print(f\"Error details: {e.error}\")\n            \n        elif status_code == 404:\n            # Secret Not Found\n            print(f\"Secret Not Found (404): {error_message}\")\n            print(f\"The secret '{secret_name}' does not exist in the vault.\")\n            \n        elif status_code == 429:\n            # Throttling - Rate limit exceeded\n            print(f\"Rate Limited (429): {error_message}\")\n            print(\"Too many requests. Implementing retry with backoff...\")\n            # Check for Retry-After header\n            retry_after = e.response.headers.get('Retry-After', 60)\n            print(f\"Retry after {retry_after} seconds\")\n            \n        else:\n            # Other HTTP errors\n            print(f\"HTTP Error ({status_code}): {error_message}\")\n            print(f\"Error code: {e.error.code if e.error else 'N/A'}\")\n            \n        # Re-raise if you want calling code to handle it\n        raise\n\n\ndef handle_with_retry(vault_url: str, secret_name: str, max_retries: int = 3):\n    \"\"\"\n    Demonstrates handling 429 throttling errors with exponential backoff retry.\n    \"\"\"\n    credential = DefaultAzureCredential()\n    client = SecretClient(vault_url=vault_url, credential=credential)\n    \n    retry_count = 0\n    base_delay = 1\n    \n    while retry_count \u003c max_retries:\n        try:\n            secret = client.get_secret(secret_name)\n            return secret.value\n            \n        except HttpResponseError as e:\n            if e.status_code == 429:\n                retry_count += 1\n                if retry_count \u003e= max_retries:\n                    print(f\"Max retries ({max_retries}) exceeded\")\n                    raise\n                \n                # Use Retry-After header if available, otherwise exponential backoff\n                retry_after = e.response.headers.get('Retry-After')\n                if retry_after:\n                    delay = int(retry_after)\n                else:\n                    delay = base_delay * (2 ** (retry_count - 1))\n                \n                print(f\"Rate limited. Retrying in {delay} seconds... \"\n                      f\"(attempt {retry_count}/{max_retries})\")\n                time.sleep(delay)\n            else:\n                # For non-429 errors, don't retry\n                raise\n\n\ndef handle_soft_deleted_secret(vault_url: str, secret_name: str):\n    \"\"\"\n    Demonstrates what happens when accessing a soft-deleted secret.\n    \n    When you try to get a soft-deleted secret:\n    - You'll receive a 404 (ResourceNotFoundError/HttpResponseError)\n    - The error message will indicate the secret is in a deleted state\n    - You cannot access the secret value until you recover it\n    - You can list deleted secrets or get deleted secret properties instead\n    \"\"\"\n    credential = DefaultAzureCredential()\n    client = SecretClient(vault_url=vault_url, credential=credential)\n    \n    try:\n        # This will fail if the secret is soft-deleted\n        secret = client.get_secret(secret_name)\n        print(f\"Secret found: {secret.name}\")\n        return secret.value\n        \n    except HttpResponseError as e:\n        if e.status_code == 404:\n            print(f\"Secret not found (404): {e.message}\")\n            \n            # Check if it's a soft-deleted secret\n            try:\n                deleted_secret = client.get_deleted_secret(secret_name)\n                print(f\"\\nSecret '{secret_name}' is SOFT-DELETED:\")\n                print(f\"  Deleted on: {deleted_secret.deleted_date}\")\n                print(f\"  Scheduled purge: {deleted_secret.scheduled_purge_date}\")\n                print(f\"  Recovery ID: {deleted_secret.recovery_id}\")\n                print(\"\\nTo access this secret, you must first recover it:\")\n                print(f\"  recovery_operation = client.begin_recover_deleted_secret('{secret_name}')\")\n                print(f\"  recovered_secret = recovery_operation.result()\")\n                \n            except HttpResponseError as deleted_err:\n                if deleted_err.status_code == 404:\n                    print(f\"Secret truly does not exist (not soft-deleted)\")\n                elif deleted_err.status_code == 403:\n                    print(\"Cannot check deleted secrets - missing permissions\")\n                    print(\"Need 'Key Vault Secrets Officer' role or higher\")\n        raise\n\n\ndef inspect_error_details(vault_url: str, secret_name: str):\n    \"\"\"\n    Demonstrates how to extract detailed information from HttpResponseError.\n    \"\"\"\n    credential = DefaultAzureCredential()\n    client = SecretClient(vault_url=vault_url, credential=credential)\n    \n    try:\n        secret = client.get_secret(secret_name)\n        return secret.value\n        \n    except HttpResponseError as e:\n        print(\"=== HttpResponseError Details ===\")\n        print(f\"Status Code: {e.status_code}\")\n        print(f\"Message: {e.message}\")\n        print(f\"Reason: {e.reason}\")\n        \n        # Error object (if available)\n        if e.error:\n            print(f\"\\nError Code: {e.error.code}\")\n            print(f\"Error Message: {e.error.message}\")\n        \n        # Response headers\n        print(f\"\\nResponse Headers:\")\n        if e.response:\n            for key, value in e.response.headers.items():\n                if key.lower() in ['x-ms-request-id', 'x-ms-client-request-id', \n                                   'retry-after', 'www-authenticate']:\n                    print(f\"  {key}: {value}\")\n        \n        raise\n\n\ndef categorize_error(e: HttpResponseError) -\u003e str:\n    \"\"\"\n    Helper function to categorize Key Vault errors.\n    \"\"\"\n    status_code = e.status_code\n    \n    error_categories = {\n        403: \"PERMISSION_DENIED\",\n        404: \"NOT_FOUND\",\n        429: \"THROTTLED\",\n        401: \"AUTHENTICATION_FAILED\",\n        409: \"CONFLICT\",\n        500: \"SERVER_ERROR\",\n        503: \"SERVICE_UNAVAILABLE\"\n    }\n    \n    return error_categories.get(status_code, \"UNKNOWN_ERROR\")\n\n\n# Example usage patterns\nif __name__ == \"__main__\":\n    vault_url = \"https://your-keyvault.vault.azure.net/\"\n    \n    # Example 1: Basic error handling\n    try:\n        value = handle_common_errors(vault_url, \"my-secret\")\n    except HttpResponseError:\n        print(\"Failed to retrieve secret\")\n    \n    # Example 2: With retry logic for throttling\n    try:\n        value = handle_with_retry(vault_url, \"my-secret\", max_retries=3)\n    except HttpResponseError:\n        print(\"Failed after retries\")\n    \n    # Example 3: Handling soft-deleted secrets\n    try:\n        value = handle_soft_deleted_secret(vault_url, \"deleted-secret\")\n    except HttpResponseError:\n        print(\"Could not access secret\")\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-error-handling-baseline-3794219126/keyvault_error_handling.py"}
```

## Generated Files

- `keyvault_error_handling.py`

## Copilot Response

I'll create a comprehensive example showing proper error handling patterns for Azure Key Vault secrets in Python.

Perfect! I've created a comprehensive example showing Azure Key Vault error handling patterns. Here are the key points:

## Key Error Handling Patterns:

**1. Inspecting HttpResponseError:**
```python
except HttpResponseError as e:
    status_code = e.status_code  # 403, 404, 429, etc.
    error_message = e.message     # Human-readable message
    error_code = e.error.code     # Azure error code
```

**2. Common Errors:**

- **403 (Access Denied)**: Your app identity needs `Key Vault Secrets User` or `Key Vault Secrets Officer` RBAC role
- **404 (Not Found)**: Secret doesn't exist (or is soft-deleted)
- **429 (Throttling)**: Rate limit hit - check `Retry-After` header and implement exponential backoff

**3. Soft-Deleted Secrets:**
When you try to `get_secret()` on a soft-deleted secret:
- You get a **404 error**
- Must use `get_deleted_secret()` to check if it's soft-deleted
- Cannot access the value until you call `begin_recover_deleted_secret()`
- The secret stays in deleted state until recovered or purged

**4. Best Practices:**
- Always catch `HttpResponseError` for HTTP-level errors
- Check `status_code` to determine error type
- Implement retry logic with exponential backoff for 429 errors
- Use the `Retry-After` header when provided
- Check deleted secrets when encountering 404s

The example file includes ready-to-use functions with proper error handling, retry logic, and detailed error inspection!

## Verification

**Result:** ✅ PASS

**Summary:** Code fully meets all requirements: proper SDK usage, comprehensive error handling for 403/404/429, status code inspection, soft-delete handling, and retry logic.

**Reasoning:**

The generated code comprehensively addresses all requirements from the original prompt. It correctly imports HttpResponseError and ResourceNotFoundError from azure.core.exceptions, demonstrates handling of 403 (access denied/RBAC), 404 (not found), and 429 (throttling) errors with proper status_code inspection. The code shows how to extract error.code and message from HttpResponseError in the inspect_error_details function. It includes a dedicated function (handle_soft_deleted_secret) that explains what happens when accessing a soft-deleted secret, demonstrates checking with get_deleted_secret, and references begin_recover_deleted_secret for recovery. The handle_with_retry function implements retry logic for 429 errors with Retry-After header inspection and exponential backoff. The code uses the correct azure-keyvault-secrets SDK with SecretClient and DefaultAzureCredential. All syntax is valid Python and the code would execute properly with appropriate Azure credentials. The only minor omission is not showing SecretClient initialization with custom retry policy kwargs, but the manual retry implementation is equally practical and meets the requirement. The code includes helpful comments, multiple example functions, and a main block demonstrating usage patterns.

---

[← Back to Summary](../../../../../../summary.md)
