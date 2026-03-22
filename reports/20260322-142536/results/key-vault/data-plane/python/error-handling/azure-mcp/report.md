# Evaluation Report: key-vault-dp-python-error-handling

**Config:** azure-mcp | **Result:** ✅ PASSED | **Duration:** 83.9s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-error-handling` |
| Config | azure-mcp |
| Result | ✅ PASSED |
| Duration | 83.9s |
| Timestamp | 2026-03-22T21:25:36Z |
| Files Generated | 1 |
| Event Count | 26 |

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

The user is asking for guidance on error handling with Azure Key Vault secrets in Python. They want to know:
1. How to handle specific HTTP errors (403, 404, 429)
2. Try/except patterns with azure-keyvault-secrets SDK
3. How to inspect status_code and error message on HttpResponseError
4. What happens when trying to get a soft-deleted secret

This is a straightforward request for code examples and explanation. I don't need to make changes to a codebase - I should create example code that demonstrates proper error handling patterns.

I'll create a comprehensive Python script that shows:
- Proper imports from azure-keyvault-secrets
- Try/except blocks for different error scenarios
- How to access status_code and error message from HttpResponseError
- Explanation of soft-deleted secret behavior

Let me create an example file that demonstrates all these patterns.

## Tool Calls

### 🔧 1. report_intent

**Input:**

```json
{"intent":"Creating error handling examples"}
```

### 🔧 2. create

**Input:**

```json
{"file_text":"\"\"\"\nAzure Key Vault Secrets - Error Handling Patterns\nDemonstrates proper exception handling for common scenarios\n\"\"\"\n\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.identity import DefaultAzureCredential\nfrom azure.core.exceptions import HttpResponseError, ResourceNotFoundError\nimport time\n\n\n# Initialize the client\ncredential = DefaultAzureCredential()\nvault_url = \"https://your-vault-name.vault.azure.net/\"\nclient = SecretClient(vault_url=vault_url, credential=credential)\n\n\n# Example 1: Handling 403 Access Denied (Missing RBAC permissions)\ndef handle_access_denied():\n    \"\"\"\n    403 Forbidden - Occurs when the identity lacks RBAC role\n    Required roles: Key Vault Secrets User (read) or Key Vault Secrets Officer (write)\n    \"\"\"\n    try:\n        secret = client.get_secret(\"my-secret\")\n        print(f\"Secret value: {secret.value}\")\n    \n    except HttpResponseError as e:\n        if e.status_code == 403:\n            print(\"Access Denied!\")\n            print(f\"Status Code: {e.status_code}\")\n            print(f\"Error Code: {e.error.code if e.error else 'N/A'}\")\n            print(f\"Error Message: {e.message}\")\n            print(\"\\nTroubleshooting:\")\n            print(\"- Verify your identity has 'Key Vault Secrets User' role\")\n            print(\"- Check RBAC assignments in Azure Portal\")\n            print(\"- Ensure you're using RBAC, not access policies\")\n        else:\n            raise\n\n\n# Example 2: Handling 404 Secret Not Found\ndef handle_secret_not_found():\n    \"\"\"\n    404 Not Found - Secret doesn't exist or has been deleted\n    \"\"\"\n    try:\n        secret = client.get_secret(\"non-existent-secret\")\n        print(f\"Secret value: {secret.value}\")\n    \n    except ResourceNotFoundError as e:\n        # ResourceNotFoundError is a subclass of HttpResponseError for 404s\n        print(\"Secret Not Found!\")\n        print(f\"Status Code: {e.status_code}\")\n        print(f\"Error Message: {e.message}\")\n        print(\"\\nPossible reasons:\")\n        print(\"- Secret name is incorrect\")\n        print(\"- Secret has been deleted\")\n        print(\"- Secret is in soft-deleted state\")\n    \n    except HttpResponseError as e:\n        if e.status_code == 404:\n            # Alternative way to handle 404\n            print(f\"Secret not found: {e.message}\")\n        else:\n            raise\n\n\n# Example 3: Handling 429 Throttling (Rate Limit)\ndef handle_throttling_with_retry():\n    \"\"\"\n    429 Too Many Requests - Azure throttles requests to protect the service\n    Default limits: ~2000 requests per 10 seconds per vault\n    \"\"\"\n    max_retries = 3\n    retry_count = 0\n    \n    while retry_count \u003c max_retries:\n        try:\n            secret = client.get_secret(\"my-secret\")\n            print(f\"Secret retrieved: {secret.name}\")\n            return secret\n        \n        except HttpResponseError as e:\n            if e.status_code == 429:\n                retry_count += 1\n                \n                # Extract retry-after header if available\n                retry_after = e.response.headers.get('Retry-After', 5)\n                retry_after = int(retry_after) if isinstance(retry_after, str) else retry_after\n                \n                print(f\"Throttled! Status Code: {e.status_code}\")\n                print(f\"Error Message: {e.message}\")\n                print(f\"Retry attempt {retry_count}/{max_retries}\")\n                print(f\"Waiting {retry_after} seconds...\")\n                \n                if retry_count \u003c max_retries:\n                    time.sleep(retry_after)\n                else:\n                    print(\"Max retries exceeded\")\n                    raise\n            else:\n                raise\n\n\n# Example 4: Handling Soft-Deleted Secrets\ndef handle_soft_deleted_secret():\n    \"\"\"\n    When you try to get_secret() on a soft-deleted secret:\n    - Returns 404 ResourceNotFoundError (secret is not \"active\")\n    - The secret exists in deleted state but is not accessible via get_secret()\n    \n    To access soft-deleted secrets, use get_deleted_secret() instead\n    \"\"\"\n    secret_name = \"deleted-secret\"\n    \n    # This will fail with 404 if the secret is soft-deleted\n    try:\n        secret = client.get_secret(secret_name)\n        print(f\"Active secret found: {secret.value}\")\n    \n    except ResourceNotFoundError as e:\n        print(f\"Secret not found in active state: {e.message}\")\n        print(\"\\nChecking if it's soft-deleted...\")\n        \n        # Try to retrieve the deleted secret\n        try:\n            deleted_secret = client.get_deleted_secret(secret_name)\n            print(f\"Found soft-deleted secret!\")\n            print(f\"Name: {deleted_secret.name}\")\n            print(f\"Deleted on: {deleted_secret.deleted_date}\")\n            print(f\"Scheduled purge: {deleted_secret.scheduled_purge_date}\")\n            print(f\"Recovery ID: {deleted_secret.recovery_id}\")\n            print(\"\\nTo recover: client.begin_recover_deleted_secret(name)\")\n            \n        except ResourceNotFoundError:\n            print(\"Secret does not exist (not active or deleted)\")\n\n\n# Example 5: Comprehensive Error Handling Pattern\ndef comprehensive_error_handling(secret_name):\n    \"\"\"\n    Complete pattern covering all common scenarios\n    \"\"\"\n    try:\n        secret = client.get_secret(secret_name)\n        return secret.value\n    \n    except ResourceNotFoundError as e:\n        # 404 - Secret doesn't exist\n        print(f\"❌ Secret '{secret_name}' not found\")\n        print(f\"   Status: {e.status_code}\")\n        print(f\"   Message: {e.message}\")\n        return None\n    \n    except HttpResponseError as e:\n        # Inspect the error details\n        status_code = e.status_code\n        error_code = e.error.code if e.error else \"Unknown\"\n        error_message = e.message\n        \n        if status_code == 403:\n            print(f\"❌ Access Denied (403)\")\n            print(f\"   Error Code: {error_code}\")\n            print(f\"   Message: {error_message}\")\n            print(f\"   Fix: Grant 'Key Vault Secrets User' RBAC role\")\n        \n        elif status_code == 429:\n            print(f\"❌ Rate Limit Exceeded (429)\")\n            print(f\"   Message: {error_message}\")\n            retry_after = e.response.headers.get('Retry-After', 'unknown')\n            print(f\"   Retry after: {retry_after} seconds\")\n        \n        elif status_code == 401:\n            print(f\"❌ Unauthorized (401)\")\n            print(f\"   Message: {error_message}\")\n            print(f\"   Fix: Check authentication credentials\")\n        \n        else:\n            print(f\"❌ HTTP Error {status_code}\")\n            print(f\"   Error Code: {error_code}\")\n            print(f\"   Message: {error_message}\")\n        \n        # Re-raise if you want the caller to handle it\n        raise\n    \n    except Exception as e:\n        # Catch-all for other exceptions (network errors, etc.)\n        print(f\"❌ Unexpected error: {type(e).__name__}\")\n        print(f\"   Details: {str(e)}\")\n        raise\n\n\n# Example 6: Inspecting HttpResponseError Details\ndef inspect_error_details():\n    \"\"\"\n    Shows all available properties on HttpResponseError\n    \"\"\"\n    try:\n        secret = client.get_secret(\"problematic-secret\")\n    \n    except HttpResponseError as e:\n        print(\"=== HttpResponseError Details ===\")\n        print(f\"status_code: {e.status_code}\")\n        print(f\"reason: {e.reason}\")\n        print(f\"message: {e.message}\")\n        \n        # Error object (may be None)\n        if e.error:\n            print(f\"error.code: {e.error.code}\")\n            print(f\"error.message: {e.error.message}\")\n        \n        # Response object\n        if e.response:\n            print(f\"response.status_code: {e.response.status_code}\")\n            print(f\"response.headers: {dict(e.response.headers)}\")\n        \n        # Additional context\n        print(f\"model: {e.model}\")\n        print(f\"exc_type: {type(e).__name__}\")\n\n\n# Example 7: Setting Secrets with Error Handling\ndef set_secret_with_error_handling(secret_name, secret_value):\n    \"\"\"\n    Error handling when creating/updating secrets\n    \"\"\"\n    try:\n        secret = client.set_secret(secret_name, secret_value)\n        print(f\"✓ Secret '{secret_name}' saved successfully\")\n        return secret\n    \n    except HttpResponseError as e:\n        if e.status_code == 403:\n            print(f\"❌ Cannot set secret - missing permissions\")\n            print(f\"   Required role: 'Key Vault Secrets Officer'\")\n        \n        elif e.status_code == 409:\n            # Conflict - secret might be in deleted state\n            print(f\"❌ Conflict (409) - Secret may be soft-deleted\")\n            print(f\"   Purge or recover the deleted secret first\")\n        \n        else:\n            print(f\"❌ Failed to set secret: {e.message}\")\n        \n        raise\n\n\nif __name__ == \"__main__\":\n    print(\"Azure Key Vault Error Handling Examples\")\n    print(\"=\" * 50)\n    \n    # Run examples (will fail without proper credentials/vault)\n    # Uncomment to test:\n    \n    # comprehensive_error_handling(\"my-secret\")\n    # handle_throttling_with_retry()\n    # handle_soft_deleted_secret()\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-error-handling-azure-mcp-1460421247/keyvault_error_handling_examples.py"}
```

## Generated Files

- `keyvault_error_handling_examples.py`

## Copilot Response

I'll create a comprehensive example showing proper error handling patterns for Azure Key Vault secrets in Python.

Perfect! I've created comprehensive examples for Azure Key Vault error handling. Here are the key patterns:

## **Key Points:**

### **1. Accessing HttpResponseError Properties:**
```python
except HttpResponseError as e:
    status_code = e.status_code          # 403, 404, 429, etc.
    error_message = e.message            # Human-readable message
    error_code = e.error.code            # Azure error code
    retry_after = e.response.headers.get('Retry-After')
```

### **2. Specific Error Handling:**

- **403 (Access Denied)**: Missing RBAC role - needs "Key Vault Secrets User" for read or "Key Vault Secrets Officer" for write
- **404 (Not Found)**: Use `ResourceNotFoundError` (subclass of HttpResponseError) - secret doesn't exist or is deleted
- **429 (Throttling)**: Extract `Retry-After` header and implement exponential backoff

### **3. Soft-Deleted Secrets:**
**Important**: `get_secret()` returns **404** for soft-deleted secrets! They're not accessible via normal operations. Use:
```python
deleted_secret = client.get_deleted_secret(secret_name)
```

The soft-deleted secret contains `deleted_date`, `scheduled_purge_date`, and `recovery_id` for recovery operations.

### **4. Best Practice Pattern:**
Catch `ResourceNotFoundError` first (for 404s), then broader `HttpResponseError` for other status codes, using a final catch-all for unexpected errors.

Run the examples with your vault URL and credentials to see the patterns in action!

## Verification

**Result:** ✅ PASS

**Summary:** Code fully addresses all prompt requirements with comprehensive error handling patterns; minor omission of SecretClient retry kwargs configuration.

**Reasoning:**

The generated code successfully addresses all core requirements from the prompt. It correctly imports and uses azure.keyvault.secrets.SecretClient and azure.core.exceptions (HttpResponseError, ResourceNotFoundError). The code demonstrates: (1) 403 handling with status_code inspection, error.code/message extraction, and RBAC troubleshooting guidance; (2) 404 handling using both ResourceNotFoundError and HttpResponseError patterns, explaining secret not found vs. soft-deleted states; (3) 429 throttling with retry logic, Retry-After header extraction, and rate limit documentation; (4) comprehensive soft-delete handling showing that get_secret() returns 404 for deleted secrets, with get_deleted_secret() as the solution and begin_recover_deleted_secret() mentioned for recovery; (5) detailed error inspection examples showing how to access status_code, error.code, error.message, response.headers, and other HttpResponseError properties. The code is syntactically valid, well-organized with 7 focused examples, and includes practical troubleshooting guidance. Minor omission: it doesn't show SecretClient retry configuration via constructor kwargs (e.g., retry_policy parameter), though it does demonstrate manual retry handling for 429 errors. Overall, this is a comprehensive, production-ready example that exceeds basic requirements.

---

[← Back to Summary](../../../../../../summary.md)
