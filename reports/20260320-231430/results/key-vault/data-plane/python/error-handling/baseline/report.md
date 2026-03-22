# Evaluation Report: key-vault-dp-python-error-handling

**Config:** baseline | **Result:** ✅ PASSED | **Duration:** 80.5s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-error-handling` |
| Config | baseline |
| Result | ✅ PASSED |
| Duration | 80.5s |
| Timestamp | 2026-03-21T06:14:30Z |
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

The user is asking for guidance on error handling with Azure Key Vault secrets in Python, specifically:
1. Handling 403 (access denied)
2. Handling 404 (secret not found)
3. Handling 429 (throttling/rate limits)
4. How to inspect status_code and error message on HttpResponseError
5. What happens with soft-deleted secrets

This is a straightforward educational/example request. I should create a comprehensive Python example demonstrating these error handling patterns. I'll create a clear, working example that shows:
- Proper imports
- Try/except patterns for different error codes
- How to inspect HttpResponseError properties
- Explanation of soft-deleted secret behavior

Since this is a quick educational task (creating example code), I don't need to create a plan.md file. I'll just create a well-documented Python example file.

## Tool Calls

### 🔧 1. report_intent

**Input:**

```json
{"intent":"Creating error handling examples"}
```

### 🔧 2. create

**Input:**

```json
{"file_text":"\"\"\"\nAzure Key Vault Secrets - Error Handling Patterns\n\nThis module demonstrates proper error handling when working with Azure Key Vault\nsecrets, including handling 403 (access denied), 404 (not found), 429 (throttling),\nand soft-deleted secrets.\n\"\"\"\n\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.identity import DefaultAzureCredential\nfrom azure.core.exceptions import HttpResponseError, ResourceNotFoundError\nimport time\n\n\ndef handle_keyvault_errors_basic(vault_url: str, secret_name: str):\n    \"\"\"\n    Basic error handling pattern with status code inspection.\n    \n    Demonstrates how to:\n    - Catch HttpResponseError\n    - Inspect status_code property\n    - Access error message details\n    \"\"\"\n    credential = DefaultAzureCredential()\n    client = SecretClient(vault_url=vault_url, credential=credential)\n    \n    try:\n        secret = client.get_secret(secret_name)\n        print(f\"Successfully retrieved secret: {secret.name}\")\n        return secret.value\n        \n    except HttpResponseError as e:\n        # Inspect the status code to determine the error type\n        status_code = e.status_code\n        error_message = e.message\n        \n        if status_code == 403:\n            # Access Denied - RBAC permissions issue\n            print(f\"Access Denied (403): {error_message}\")\n            print(\"Your identity lacks the required RBAC role.\")\n            print(\"Required role: 'Key Vault Secrets User' or 'Key Vault Secrets Officer'\")\n            \n        elif status_code == 404:\n            # Secret Not Found\n            print(f\"Secret Not Found (404): {error_message}\")\n            print(f\"The secret '{secret_name}' does not exist in the vault.\")\n            \n        elif status_code == 429:\n            # Rate Limiting / Throttling\n            print(f\"Throttling (429): {error_message}\")\n            print(\"Too many requests. Service is rate limiting your requests.\")\n            # Check for Retry-After header\n            if hasattr(e, 'response') and e.response:\n                retry_after = e.response.headers.get('Retry-After')\n                if retry_after:\n                    print(f\"Retry after {retry_after} seconds\")\n                    \n        else:\n            # Other HTTP errors\n            print(f\"HTTP Error ({status_code}): {error_message}\")\n            \n        # Re-raise if you want calling code to handle it\n        raise\n\n\ndef handle_specific_errors_separately(vault_url: str, secret_name: str):\n    \"\"\"\n    Pattern with separate exception handlers for different scenarios.\n    \n    Uses ResourceNotFoundError for 404s (more specific than HttpResponseError).\n    \"\"\"\n    credential = DefaultAzureCredential()\n    client = SecretClient(vault_url=vault_url, credential=credential)\n    \n    try:\n        secret = client.get_secret(secret_name)\n        return secret.value\n        \n    except ResourceNotFoundError as e:\n        # More specific exception for 404 errors\n        print(f\"Secret '{secret_name}' not found: {e.message}\")\n        return None\n        \n    except HttpResponseError as e:\n        if e.status_code == 403:\n            print(f\"Access denied. Check RBAC permissions: {e.message}\")\n            # Log additional details\n            if hasattr(e, 'error'):\n                print(f\"Error code: {e.error.code}\")\n            raise\n            \n        elif e.status_code == 429:\n            print(f\"Rate limited: {e.message}\")\n            # Implement retry logic with exponential backoff\n            raise\n            \n        else:\n            print(f\"Unexpected error ({e.status_code}): {e.message}\")\n            raise\n\n\ndef handle_with_retry_on_throttling(vault_url: str, secret_name: str, max_retries: int = 3):\n    \"\"\"\n    Error handling with automatic retry for throttling (429).\n    \n    Implements exponential backoff when encountering rate limits.\n    \"\"\"\n    credential = DefaultAzureCredential()\n    client = SecretClient(vault_url=vault_url, credential=credential)\n    \n    retry_count = 0\n    base_delay = 1  # seconds\n    \n    while retry_count \u003c= max_retries:\n        try:\n            secret = client.get_secret(secret_name)\n            return secret.value\n            \n        except HttpResponseError as e:\n            if e.status_code == 429 and retry_count \u003c max_retries:\n                # Calculate retry delay (exponential backoff)\n                retry_after = base_delay * (2 ** retry_count)\n                \n                # Check if server provided Retry-After header\n                if hasattr(e, 'response') and e.response:\n                    server_retry_after = e.response.headers.get('Retry-After')\n                    if server_retry_after:\n                        retry_after = int(server_retry_after)\n                \n                print(f\"Throttled (429). Retrying in {retry_after} seconds... (attempt {retry_count + 1}/{max_retries})\")\n                time.sleep(retry_after)\n                retry_count += 1\n                \n            elif e.status_code == 403:\n                print(f\"Access Denied (403): {e.message}\")\n                print(\"RBAC role required: Key Vault Secrets User or Key Vault Secrets Officer\")\n                raise\n                \n            elif e.status_code == 404:\n                print(f\"Secret not found (404): {e.message}\")\n                return None\n                \n            else:\n                # For other errors or exceeded retries, raise\n                print(f\"Error ({e.status_code}): {e.message}\")\n                raise\n        \n        except Exception as e:\n            print(f\"Unexpected error: {type(e).__name__}: {str(e)}\")\n            raise\n    \n    # If we exhausted all retries\n    raise Exception(f\"Failed to retrieve secret after {max_retries} retries due to throttling\")\n\n\ndef handle_soft_deleted_secrets(vault_url: str, secret_name: str):\n    \"\"\"\n    Demonstrates behavior when attempting to access soft-deleted secrets.\n    \n    Key Points:\n    1. get_secret() on a soft-deleted secret raises ResourceNotFoundError (404)\n    2. The secret exists in the deleted state but cannot be retrieved via get_secret()\n    3. You must use get_deleted_secret() to access properties of deleted secrets\n    4. You cannot retrieve the VALUE of a deleted secret\n    5. To use the secret name again, either purge or recover it first\n    \"\"\"\n    credential = DefaultAzureCredential()\n    client = SecretClient(vault_url=vault_url, credential=credential)\n    \n    print(f\"\\n=== Attempting to get soft-deleted secret '{secret_name}' ===\")\n    \n    try:\n        # This will fail with 404 if the secret is soft-deleted\n        secret = client.get_secret(secret_name)\n        print(f\"Secret retrieved successfully: {secret.name}\")\n        return secret.value\n        \n    except ResourceNotFoundError as e:\n        print(f\"Secret not found (404): {e.message}\")\n        print(\"\\nChecking if secret is in deleted state...\")\n        \n        try:\n            # Try to get the deleted secret\n            deleted_secret = client.get_deleted_secret(secret_name)\n            \n            print(f\"\\n✓ Secret '{secret_name}' exists but is SOFT-DELETED\")\n            print(f\"  Deleted on: {deleted_secret.deleted_date}\")\n            print(f\"  Scheduled purge date: {deleted_secret.scheduled_purge_date}\")\n            print(f\"  Recovery ID: {deleted_secret.recovery_id}\")\n            \n            print(\"\\nTo use this secret:\")\n            print(\"  Option 1: Recover it using client.begin_recover_deleted_secret()\")\n            print(\"  Option 2: Purge it using client.purge_deleted_secret() (requires purge permission)\")\n            print(\"\\nNote: You CANNOT retrieve the value of a deleted secret.\")\n            \n            return None\n            \n        except ResourceNotFoundError:\n            print(f\"\\n✗ Secret '{secret_name}' does not exist (not active or deleted)\")\n            return None\n            \n        except HttpResponseError as deleted_error:\n            if deleted_error.status_code == 403:\n                print(\"\\n✗ Cannot check deleted secrets - missing 'List' permission\")\n                print(\"  Required: 'Key Vault Secrets User' role or 'List' data action\")\n            else:\n                print(f\"\\nError checking deleted secrets: {deleted_error.message}\")\n            raise\n    \n    except HttpResponseError as e:\n        if e.status_code == 403:\n            print(f\"Access Denied (403): {e.message}\")\n            print(\"Check RBAC role assignment on the Key Vault\")\n        elif e.status_code == 429:\n            print(f\"Throttled (429): {e.message}\")\n        else:\n            print(f\"HTTP Error ({e.status_code}): {e.message}\")\n        raise\n\n\ndef inspect_error_details(vault_url: str, secret_name: str):\n    \"\"\"\n    Demonstrates how to inspect all available properties on HttpResponseError.\n    \n    Useful properties:\n    - status_code: HTTP status code (int)\n    - message: Error message (str)\n    - error: Structured error object (if available)\n    - response: Raw HTTP response (if available)\n    \"\"\"\n    credential = DefaultAzureCredential()\n    client = SecretClient(vault_url=vault_url, credential=credential)\n    \n    try:\n        secret = client.get_secret(secret_name)\n        return secret.value\n        \n    except HttpResponseError as e:\n        print(\"=== HttpResponseError Details ===\")\n        print(f\"Status Code: {e.status_code}\")\n        print(f\"Message: {e.message}\")\n        \n        # Check for structured error object\n        if hasattr(e, 'error') and e.error:\n            print(f\"Error Code: {e.error.code if hasattr(e.error, 'code') else 'N/A'}\")\n            print(f\"Error Message: {e.error.message if hasattr(e.error, 'message') else 'N/A'}\")\n        \n        # Check response headers\n        if hasattr(e, 'response') and e.response:\n            print(\"\\nResponse Headers:\")\n            for header, value in e.response.headers.items():\n                if header.lower() in ['retry-after', 'x-ms-request-id', 'x-ms-keyvault-service-version']:\n                    print(f\"  {header}: {value}\")\n        \n        # Additional context\n        print(f\"\\nException Type: {type(e).__name__}\")\n        print(f\"String Representation: {str(e)}\")\n        \n        raise\n\n\ndef main():\n    \"\"\"\n    Example usage demonstrating different error handling patterns.\n    \"\"\"\n    vault_url = \"https://your-keyvault-name.vault.azure.net/\"\n    \n    print(\"Example 1: Basic error handling with status code inspection\")\n    print(\"-\" * 60)\n    try:\n        handle_keyvault_errors_basic(vault_url, \"my-secret\")\n    except HttpResponseError:\n        pass\n    \n    print(\"\\n\\nExample 2: Handling throttling with retry logic\")\n    print(\"-\" * 60)\n    try:\n        handle_with_retry_on_throttling(vault_url, \"my-secret\", max_retries=3)\n    except Exception:\n        pass\n    \n    print(\"\\n\\nExample 3: Soft-deleted secret handling\")\n    print(\"-\" * 60)\n    try:\n        handle_soft_deleted_secrets(vault_url, \"deleted-secret\")\n    except HttpResponseError:\n        pass\n    \n    print(\"\\n\\nExample 4: Detailed error inspection\")\n    print(\"-\" * 60)\n    try:\n        inspect_error_details(vault_url, \"my-secret\")\n    except HttpResponseError:\n        pass\n\n\nif __name__ == \"__main__\":\n    main()\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-error-handling-baseline-2181013305/keyvault_error_handling.py"}
```

## Generated Files

- `keyvault_error_handling.py`

## Copilot Response

I'll create a comprehensive example showing proper error handling patterns for Azure Key Vault secrets in Python.

Perfect! I've created a comprehensive guide for handling Azure Key Vault errors in Python. Here's what the example covers:

## Key Error Handling Patterns:

**1. Status Code Inspection**
- Access `e.status_code` to get the HTTP status code
- Access `e.message` to get the error message
- Handle 403 (access denied), 404 (not found), 429 (throttling) specifically

**2. Exception Types**
- `HttpResponseError` - base class for HTTP errors
- `ResourceNotFoundError` - more specific for 404s

**3. Retry Logic for Throttling (429)**
- Implements exponential backoff
- Checks `Retry-After` header from response
- Configurable max retries

**4. Soft-Deleted Secrets Behavior**
- `get_secret()` on soft-deleted secret → **404 ResourceNotFoundError**
- Must use `get_deleted_secret()` to check deleted state
- **Cannot retrieve the value** of deleted secrets
- Must recover or purge before reusing the name

**5. Error Details Available**
- `status_code` - HTTP status code (403, 404, 429, etc.)
- `message` - Human-readable error message
- `error.code` - Structured error code
- `response.headers` - HTTP headers including Retry-After

The examples show production-ready patterns with proper logging and retry mechanisms!

## Verification

**Result:** ✅ PASS

**Summary:** Code fully meets requirements with comprehensive error handling patterns for 403/404/429, proper exception inspection, and thorough soft-delete explanation.

**Reasoning:**

The code comprehensively addresses all requirements from the original prompt. It demonstrates: (1) Proper try/except patterns with azure-keyvault-secrets SDK using HttpResponseError and ResourceNotFoundError; (2) Handling 403 access denied with RBAC role diagnostics; (3) Handling 404 secret not found scenarios; (4) Handling 429 throttling with retry-after header inspection and exponential backoff; (5) Extracting status_code, error.code, and message properties from exceptions; (6) Detailed explanation of soft-deleted secret behavior showing get_secret() raises 404, get_deleted_secret() can retrieve metadata, and mentions recovery via begin_recover_deleted_secret(). The code imports from azure.core.exceptions correctly, provides multiple practical patterns (basic, separate handlers, retry logic, soft-delete handling, detailed inspection), and is syntactically valid Python that would run correctly. While it doesn't demonstrate SecretClient retry configuration kwargs (mentioned in expected coverage but not the original prompt), it exceeds requirements by providing comprehensive examples with good documentation. The soft-delete handling function thoroughly explains the behavior with working code examples.

---

[← Back to Summary](../../../../../../summary.md)
