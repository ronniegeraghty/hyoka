# Evaluation Report: key-vault-dp-python-error-handling

**Config:** baseline | **Result:** ✅ PASSED | **Duration:** 79.3s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-error-handling` |
| Config | baseline |
| Result | ✅ PASSED |
| Duration | 79.3s |
| Timestamp | 2026-03-22T21:19:38Z |
| Files Generated | 1 |
| Event Count | 26 |

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

The user is asking for guidance on error handling with Azure Key Vault secrets in Python. This is a straightforward question about best practices and patterns, not a request to modify code in a repository. They want:

1. Try/except patterns for the azure-keyvault-secrets SDK
2. How to handle specific HTTP status codes (403, 404, 429)
3. How to inspect status_code and error message on HttpResponseError
4. Explanation of what happens with soft-deleted secrets

This is an informational/educational request. I should provide clear examples with code snippets showing the proper error handling patterns. I don't need to create a plan or modify any files - just provide helpful guidance.

Let me provide comprehensive examples showing:
- Import statements
- Basic try/except structure
- Specific error handling for 403, 404, 429
- How to access error details from HttpResponseError
- Soft-deleted secret behavior

## Tool Calls

### 🔧 1. report_intent

**Input:**

```json
{"intent":"Creating error handling examples"}
```

### 🔧 2. create

**Input:**

```json
{"file_text":"\"\"\"\nAzure Key Vault Secrets - Error Handling Examples\nDemonstrates proper error handling for common scenarios including\n403 (Forbidden), 404 (Not Found), 429 (Throttling), and soft-deleted secrets.\n\"\"\"\n\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.identity import DefaultAzureCredential\nfrom azure.core.exceptions import HttpResponseError, ResourceNotFoundError\nimport time\n\n\ndef handle_basic_errors(secret_client: SecretClient, secret_name: str):\n    \"\"\"\n    Basic error handling pattern for Key Vault operations.\n    \"\"\"\n    try:\n        secret = secret_client.get_secret(secret_name)\n        print(f\"Successfully retrieved secret: {secret.name}\")\n        return secret.value\n    \n    except HttpResponseError as e:\n        # Inspect the status code\n        status_code = e.status_code\n        error_message = e.message\n        \n        print(f\"HTTP Error: {status_code}\")\n        print(f\"Error Message: {error_message}\")\n        \n        # Access additional error details\n        if e.error:\n            print(f\"Error Code: {e.error.code}\")\n            print(f\"Error Details: {e.error.message}\")\n        \n        raise\n\n\ndef handle_specific_status_codes(secret_client: SecretClient, secret_name: str):\n    \"\"\"\n    Handle specific HTTP status codes with different retry/recovery strategies.\n    \"\"\"\n    try:\n        secret = secret_client.get_secret(secret_name)\n        return secret.value\n    \n    except HttpResponseError as e:\n        if e.status_code == 403:\n            # Access Denied - RBAC permissions missing\n            print(f\"Access Denied (403): The identity does not have permission to access secret '{secret_name}'\")\n            print(f\"Required role: 'Key Vault Secrets User' or 'Key Vault Secrets Officer'\")\n            print(f\"Error details: {e.message}\")\n            # Don't retry - this requires permission changes\n            raise PermissionError(f\"Missing RBAC role for Key Vault access\") from e\n        \n        elif e.status_code == 404:\n            # Secret not found\n            print(f\"Secret Not Found (404): '{secret_name}' does not exist\")\n            print(f\"Error details: {e.message}\")\n            # Don't retry - secret doesn't exist\n            raise\n        \n        elif e.status_code == 429:\n            # Rate limit / Throttling\n            print(f\"Rate Limited (429): Too many requests to Key Vault\")\n            print(f\"Error details: {e.message}\")\n            \n            # Check for Retry-After header\n            retry_after = e.response.headers.get('Retry-After')\n            if retry_after:\n                wait_seconds = int(retry_after)\n                print(f\"Retry after {wait_seconds} seconds\")\n                time.sleep(wait_seconds)\n            else:\n                # Default exponential backoff\n                print(\"Applying exponential backoff...\")\n                time.sleep(5)\n            \n            # Retry the operation\n            return handle_specific_status_codes(secret_client, secret_name)\n        \n        else:\n            # Other HTTP errors\n            print(f\"Unexpected HTTP Error ({e.status_code}): {e.message}\")\n            raise\n\n\ndef handle_soft_deleted_secret(secret_client: SecretClient, secret_name: str):\n    \"\"\"\n    Handle soft-deleted secrets.\n    \n    When you try to get a soft-deleted secret:\n    - You get a 404 (Not Found) error\n    - The secret is in the deleted state but can be recovered\n    - You need to use get_deleted_secret() to access its metadata\n    - You need 'recover' permission to restore it\n    \"\"\"\n    try:\n        # This will fail with 404 if the secret is soft-deleted\n        secret = secret_client.get_secret(secret_name)\n        print(f\"Retrieved active secret: {secret.name}\")\n        return secret.value\n    \n    except HttpResponseError as e:\n        if e.status_code == 404:\n            print(f\"Secret '{secret_name}' not found in active state\")\n            \n            # Check if it's soft-deleted\n            try:\n                deleted_secret = secret_client.get_deleted_secret(secret_name)\n                print(f\"Found soft-deleted secret: {deleted_secret.name}\")\n                print(f\"Deleted on: {deleted_secret.deleted_date}\")\n                print(f\"Scheduled purge date: {deleted_secret.scheduled_purge_date}\")\n                print(f\"Recovery ID: {deleted_secret.recovery_id}\")\n                \n                # To recover the secret (requires 'recover' permission):\n                print(f\"\\nTo recover this secret, use:\")\n                print(f\"  recovered_secret = secret_client.begin_recover_deleted_secret('{secret_name}')\")\n                print(f\"  recovered_secret.wait()\")\n                \n                return None\n            \n            except HttpResponseError as recover_error:\n                if recover_error.status_code == 404:\n                    print(f\"Secret '{secret_name}' does not exist (not even in deleted state)\")\n                elif recover_error.status_code == 403:\n                    print(f\"No permission to view deleted secrets (requires 'list' permission on deleted secrets)\")\n                raise\n        else:\n            raise\n\n\ndef comprehensive_error_handling(secret_client: SecretClient, secret_name: str, max_retries: int = 3):\n    \"\"\"\n    Production-ready error handling with retries and exponential backoff.\n    \"\"\"\n    retry_count = 0\n    base_delay = 2\n    \n    while retry_count \u003c max_retries:\n        try:\n            secret = secret_client.get_secret(secret_name)\n            print(f\"Successfully retrieved secret: {secret.name}\")\n            return secret.value\n        \n        except HttpResponseError as e:\n            status_code = e.status_code\n            error_code = e.error.code if e.error else \"Unknown\"\n            \n            print(f\"\\nAttempt {retry_count + 1}/{max_retries} failed\")\n            print(f\"Status Code: {status_code}\")\n            print(f\"Error Code: {error_code}\")\n            print(f\"Error Message: {e.message}\")\n            \n            # Non-retryable errors\n            if status_code in [403, 404]:\n                if status_code == 403:\n                    print(\"\\n❌ Access Denied - Check RBAC role assignments\")\n                    print(\"   Required: 'Key Vault Secrets User' or higher\")\n                elif status_code == 404:\n                    print(\"\\n❌ Secret not found - Check if it exists or is soft-deleted\")\n                raise\n            \n            # Retryable errors (429, 5xx)\n            elif status_code in [429, 500, 502, 503, 504]:\n                retry_count += 1\n                \n                if retry_count \u003e= max_retries:\n                    print(f\"\\n❌ Max retries ({max_retries}) exceeded\")\n                    raise\n                \n                # Calculate backoff with jitter\n                if status_code == 429:\n                    retry_after = e.response.headers.get('Retry-After')\n                    wait_time = int(retry_after) if retry_after else base_delay * (2 ** retry_count)\n                else:\n                    wait_time = base_delay * (2 ** retry_count)\n                \n                print(f\"⏳ Retrying in {wait_time} seconds...\")\n                time.sleep(wait_time)\n            \n            else:\n                # Unknown error\n                print(f\"\\n❌ Unexpected error: {status_code}\")\n                raise\n        \n        except Exception as e:\n            # Non-HTTP errors (network issues, etc.)\n            print(f\"\\n❌ Non-HTTP error: {type(e).__name__}: {str(e)}\")\n            raise\n\n\ndef example_usage():\n    \"\"\"\n    Example usage of error handling patterns.\n    \"\"\"\n    # Initialize the client\n    vault_url = \"https://your-keyvault-name.vault.azure.net/\"\n    credential = DefaultAzureCredential()\n    secret_client = SecretClient(vault_url=vault_url, credential=credential)\n    \n    print(\"=\" * 70)\n    print(\"Example 1: Basic Error Handling\")\n    print(\"=\" * 70)\n    try:\n        handle_basic_errors(secret_client, \"my-secret\")\n    except HttpResponseError as e:\n        print(f\"Caught error with status code: {e.status_code}\\n\")\n    \n    print(\"\\n\" + \"=\" * 70)\n    print(\"Example 2: Specific Status Code Handling\")\n    print(\"=\" * 70)\n    try:\n        handle_specific_status_codes(secret_client, \"my-secret\")\n    except Exception as e:\n        print(f\"Caught: {type(e).__name__}\\n\")\n    \n    print(\"\\n\" + \"=\" * 70)\n    print(\"Example 3: Soft-Deleted Secret Handling\")\n    print(\"=\" * 70)\n    try:\n        handle_soft_deleted_secret(secret_client, \"deleted-secret\")\n    except Exception as e:\n        print(f\"Caught: {type(e).__name__}\\n\")\n    \n    print(\"\\n\" + \"=\" * 70)\n    print(\"Example 4: Production-Ready Error Handling\")\n    print(\"=\" * 70)\n    try:\n        value = comprehensive_error_handling(secret_client, \"my-secret\", max_retries=3)\n        print(f\"Secret value: {value}\")\n    except Exception as e:\n        print(f\"Final error: {type(e).__name__}: {str(e)}\\n\")\n\n\nif __name__ == \"__main__\":\n    example_usage()\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-error-handling-baseline-2576181640/keyvault_error_handling.py"}
```

## Generated Files

- `keyvault_error_handling.py`

## Copilot Response

I'll show you comprehensive error handling patterns for Azure Key Vault secrets in Python. Let me create examples covering all the scenarios you mentioned.

Perfect! I've created a comprehensive guide for handling Azure Key Vault errors in Python. Here are the key patterns:

## **Key Error Handling Patterns:**

### **1. Inspecting HttpResponseError:**
```python
except HttpResponseError as e:
    status_code = e.status_code        # HTTP status code (403, 404, 429)
    error_message = e.message          # Human-readable message
    error_code = e.error.code          # Azure-specific error code
```

### **2. Specific Status Codes:**

- **403 (Forbidden)**: Missing RBAC role - don't retry, fix permissions
- **404 (Not Found)**: Secret doesn't exist or is soft-deleted - check state
- **429 (Throttling)**: Rate limit hit - use `Retry-After` header and exponential backoff

### **3. Soft-Deleted Secrets:**
When you call `get_secret()` on a soft-deleted secret:
- You get a **404 error** (secret not found in active state)
- Use `get_deleted_secret()` to check if it's soft-deleted
- Use `begin_recover_deleted_secret()` to restore it (requires 'recover' permission)

### **4. Best Practices:**
- ✅ Retry 429 (throttling) and 5xx errors with exponential backoff
- ❌ Don't retry 403 (permission issue) or 404 (doesn't exist)
- Check `Retry-After` header for throttling responses
- Implement max retries to avoid infinite loops

The examples in the file show progressively sophisticated patterns from basic try/except to production-ready implementations with retry logic!

## Verification

**Result:** ✅ PASS

**Summary:** Code fully meets prompt requirements with comprehensive error handling patterns for 403, 404, 429, and soft-deleted secrets using azure-keyvault-secrets SDK.

**Reasoning:**

The generated code comprehensively addresses all requirements from the original prompt. It correctly demonstrates: (1) HttpResponseError exception handling with status_code, error.code, and message extraction; (2) Specific handling for 403 (RBAC permission issues), 404 (not found), and 429 (throttling with Retry-After header); (3) Soft-deleted secret behavior explanation with get_deleted_secret() and begin_recover_deleted_secret(); (4) Multiple try/except patterns ranging from basic to production-ready; (5) Correct SDK imports from azure.keyvault.secrets and azure.core.exceptions. The code is syntactically valid Python, uses the azure-keyvault-secrets SDK correctly, includes working examples with SecretClient, and provides clear explanations in comments and print statements. It handles all key scenarios: access denied RBAC diagnosis, secret not found vs deleted-but-not-purged distinction, throttling with retry logic and exponential backoff. The only minor omission is not showing SecretClient retry configuration via constructor kwargs, but this was not explicitly requested in the original prompt—the prompt asked for 'try/except patterns' which are thoroughly demonstrated. The code goes beyond requirements with comprehensive_error_handling() showing production-ready patterns with max retries, exponential backoff, and proper error categorization.

---

[← Back to Summary](../../../../../../summary.md)
