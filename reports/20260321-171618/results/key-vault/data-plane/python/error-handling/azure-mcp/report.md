# Evaluation Report: key-vault-dp-python-error-handling

**Config:** azure-mcp | **Result:** ✅ PASSED | **Duration:** 75.4s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-error-handling` |
| Config | azure-mcp |
| Result | ✅ PASSED |
| Duration | 75.4s |
| Timestamp | 2026-03-22T00:16:18Z |
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

The user is asking for guidance on error handling with Azure Key Vault secrets in Python. This is a knowledge/example request, not a code change task. They want to understand:

1. Try/except patterns for azure-keyvault-secrets SDK
2. How to handle specific HTTP status codes (403, 404, 429)
3. How to inspect status_code and error message on HttpResponseError
4. What happens with soft-deleted secrets

This is a straightforward question that I can answer directly with code examples. I should provide:
- Import statements for the SDK
- Example code showing try/except blocks
- How to access HttpResponseError properties
- Explanation of soft-deleted secret behavior

I don't need to create a plan or explore a codebase - I can just provide a comprehensive answer with examples.

## Tool Calls

### 🔧 1. report_intent

**Input:**

```json
{"intent":"Creating error handling examples"}
```

### 🔧 2. create

**Input:**

```json
{"file_text":"\"\"\"\nAzure Key Vault Secrets - Error Handling Patterns\nDemonstrates proper exception handling for common scenarios\n\"\"\"\n\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.identity import DefaultAzureCredential\nfrom azure.core.exceptions import HttpResponseError, ResourceNotFoundError\nimport time\n\n# Initialize the client\ncredential = DefaultAzureCredential()\nvault_url = \"https://your-vault-name.vault.azure.net/\"\nclient = SecretClient(vault_url=vault_url, credential=credential)\n\n\n# Example 1: Handle specific HTTP status codes\ndef get_secret_with_status_handling(secret_name: str):\n    \"\"\"\n    Demonstrates how to handle different HTTP status codes\n    \"\"\"\n    try:\n        secret = client.get_secret(secret_name)\n        print(f\"Secret retrieved: {secret.name}\")\n        return secret.value\n        \n    except HttpResponseError as e:\n        # Access the status code\n        status_code = e.status_code\n        \n        # Access the error message\n        error_message = e.message\n        \n        # Access additional error details\n        error_code = e.error.code if e.error else None\n        \n        print(f\"Status Code: {status_code}\")\n        print(f\"Error Message: {error_message}\")\n        print(f\"Error Code: {error_code}\")\n        \n        if status_code == 403:\n            # Access Denied - RBAC permissions issue\n            print(\"ERROR: Access denied. Your identity does not have permission.\")\n            print(\"Required RBAC role: 'Key Vault Secrets User' or 'Key Vault Secrets Officer'\")\n            print(f\"Check permissions for identity accessing vault: {vault_url}\")\n            \n        elif status_code == 404:\n            # Secret not found\n            print(f\"ERROR: Secret '{secret_name}' not found in the vault.\")\n            print(\"The secret may not exist or could be soft-deleted.\")\n            \n        elif status_code == 429:\n            # Throttling - rate limit exceeded\n            print(\"ERROR: Request throttled due to rate limiting.\")\n            print(\"Too many requests in a short time period.\")\n            # Implement exponential backoff\n            retry_after = e.response.headers.get('Retry-After', 60)\n            print(f\"Retry after {retry_after} seconds\")\n            \n        else:\n            print(f\"ERROR: Unexpected HTTP error: {status_code}\")\n        \n        raise  # Re-raise if you want calling code to handle it\n\n\n# Example 2: Comprehensive error handling with retry logic for throttling\ndef get_secret_with_retry(secret_name: str, max_retries: int = 3):\n    \"\"\"\n    Get secret with exponential backoff for throttling\n    \"\"\"\n    retry_count = 0\n    base_delay = 1\n    \n    while retry_count \u003c= max_retries:\n        try:\n            secret = client.get_secret(secret_name)\n            return secret.value\n            \n        except HttpResponseError as e:\n            if e.status_code == 429:\n                # Handle throttling with exponential backoff\n                retry_count += 1\n                if retry_count \u003e max_retries:\n                    print(f\"Max retries ({max_retries}) exceeded\")\n                    raise\n                \n                # Check for Retry-After header\n                retry_after = e.response.headers.get('Retry-After')\n                if retry_after:\n                    delay = int(retry_after)\n                else:\n                    # Exponential backoff: 1s, 2s, 4s, 8s...\n                    delay = base_delay * (2 ** (retry_count - 1))\n                \n                print(f\"Throttled. Retry {retry_count}/{max_retries} after {delay}s\")\n                time.sleep(delay)\n                continue\n                \n            elif e.status_code == 403:\n                print(\"Access denied - check RBAC permissions\")\n                print(f\"Identity needs 'Key Vault Secrets User' role on {vault_url}\")\n                raise\n                \n            elif e.status_code == 404:\n                print(f\"Secret '{secret_name}' not found\")\n                raise\n                \n            else:\n                print(f\"HTTP Error {e.status_code}: {e.message}\")\n                raise\n\n\n# Example 3: Handling soft-deleted secrets\ndef handle_soft_deleted_secret(secret_name: str):\n    \"\"\"\n    What happens when you try to get a soft-deleted secret\n    \n    When soft-delete is enabled and a secret is deleted:\n    - get_secret() raises ResourceNotFoundError (404)\n    - The secret exists in a deleted state but is not accessible via get_secret()\n    - You must use get_deleted_secret() to retrieve its properties\n    - You can recover it with begin_recover_deleted_secret()\n    - Or permanently delete with begin_delete_secret() (purge)\n    \"\"\"\n    try:\n        # This will fail with 404 if the secret is soft-deleted\n        secret = client.get_secret(secret_name)\n        print(f\"Active secret found: {secret.name}\")\n        return secret.value\n        \n    except ResourceNotFoundError as e:\n        print(f\"Secret '{secret_name}' not found in active secrets (404)\")\n        print(\"Checking if it's soft-deleted...\")\n        \n        try:\n            # Try to get the deleted secret\n            deleted_secret = client.get_deleted_secret(secret_name)\n            print(f\"Secret is SOFT-DELETED:\")\n            print(f\"  - Deleted on: {deleted_secret.deleted_date}\")\n            print(f\"  - Scheduled purge: {deleted_secret.scheduled_purge_date}\")\n            print(f\"  - Recovery ID: {deleted_secret.recovery_id}\")\n            \n            # Option 1: Recover the secret\n            print(\"\\nTo recover: client.begin_recover_deleted_secret(secret_name).wait()\")\n            \n            # Option 2: Purge permanently (if you have permission)\n            print(\"To purge: client.begin_delete_secret(secret_name).wait()\")\n            \n            return None\n            \n        except HttpResponseError as del_error:\n            if del_error.status_code == 403:\n                print(\"Access denied checking deleted secrets\")\n                print(\"Need 'Key Vault Secrets Officer' role to manage deleted secrets\")\n            elif del_error.status_code == 404:\n                print(\"Secret truly does not exist (not deleted, just absent)\")\n            else:\n                print(f\"Error checking deleted secrets: {del_error.status_code}\")\n            raise\n\n\n# Example 4: Pattern for handling all common errors in one place\ndef robust_secret_operation(secret_name: str):\n    \"\"\"\n    Production-ready error handling pattern\n    \"\"\"\n    try:\n        secret = client.get_secret(secret_name)\n        return secret.value\n        \n    except ResourceNotFoundError as e:\n        # Specific exception for 404\n        print(f\"Secret not found: {secret_name}\")\n        print(\"Check if the secret name is correct or if it's been deleted\")\n        return None\n        \n    except HttpResponseError as e:\n        # General HTTP errors\n        status = e.status_code\n        \n        error_handlers = {\n            403: lambda: print(\n                \"Permission denied. Required role: 'Key Vault Secrets User'\\n\"\n                f\"Grant access: az role assignment create --role 'Key Vault Secrets User' \"\n                f\"--assignee \u003cidentity\u003e --scope \u003cvault-resource-id\u003e\"\n            ),\n            429: lambda: print(\n                \"Rate limit exceeded. Implement retry with exponential backoff.\\n\"\n                f\"Retry-After header: {e.response.headers.get('Retry-After', 'not specified')}\"\n            ),\n            503: lambda: print(\"Service unavailable. Azure Key Vault may be experiencing issues.\"),\n        }\n        \n        handler = error_handlers.get(status, lambda: print(f\"HTTP {status}: {e.message}\"))\n        handler()\n        \n        # Log additional details for debugging\n        if e.error:\n            print(f\"Error code: {e.error.code}\")\n            print(f\"Error details: {e.error.message}\")\n        \n        raise\n        \n    except Exception as e:\n        # Catch any other unexpected errors\n        print(f\"Unexpected error: {type(e).__name__}: {str(e)}\")\n        raise\n\n\n# Example 5: Inspecting all HttpResponseError properties\ndef detailed_error_inspection(secret_name: str):\n    \"\"\"\n    Shows all properties you can access on HttpResponseError\n    \"\"\"\n    try:\n        secret = client.get_secret(secret_name)\n        return secret.value\n        \n    except HttpResponseError as e:\n        print(\"=== HttpResponseError Details ===\")\n        \n        # HTTP Status Code\n        print(f\"status_code: {e.status_code}\")\n        \n        # Error message\n        print(f\"message: {e.message}\")\n        \n        # Reason phrase\n        print(f\"reason: {e.reason}\")\n        \n        # The response object\n        print(f\"response: {e.response}\")\n        \n        # Response headers\n        if e.response:\n            print(f\"headers: {dict(e.response.headers)}\")\n            print(f\"Retry-After: {e.response.headers.get('Retry-After', 'N/A')}\")\n        \n        # Error details (if available)\n        if e.error:\n            print(f\"error.code: {e.error.code}\")\n            print(f\"error.message: {e.error.message}\")\n            print(f\"error.innererror: {getattr(e.error, 'innererror', None)}\")\n        \n        # Model (deserialized error body)\n        print(f\"model: {e.model}\")\n        \n        raise\n\n\n# Example usage patterns\nif __name__ == \"__main__\":\n    secret_name = \"my-secret\"\n    \n    print(\"Example 1: Basic status code handling\")\n    try:\n        get_secret_with_status_handling(secret_name)\n    except HttpResponseError:\n        pass\n    \n    print(\"\\n\" + \"=\"*50 + \"\\n\")\n    \n    print(\"Example 2: With retry logic\")\n    try:\n        value = get_secret_with_retry(secret_name)\n        print(f\"Secret value: {value}\")\n    except HttpResponseError as e:\n        print(f\"Failed after retries: {e.status_code}\")\n    \n    print(\"\\n\" + \"=\"*50 + \"\\n\")\n    \n    print(\"Example 3: Soft-deleted secret handling\")\n    handle_soft_deleted_secret(secret_name)\n    \n    print(\"\\n\" + \"=\"*50 + \"\\n\")\n    \n    print(\"Example 4: Robust production pattern\")\n    robust_secret_operation(secret_name)\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-error-handling-azure-mcp-1472055401/key_vault_error_handling.py"}
```

## Generated Files

- `key_vault_error_handling.py`

## Copilot Response

I'll show you proper error handling patterns for Azure Key Vault secrets in Python with specific examples for each scenario.

I've created a comprehensive guide with working examples! Here are the key points:

## Key Error Handling Patterns:

**1. HttpResponseError Properties:**
```python
except HttpResponseError as e:
    status_code = e.status_code      # 403, 404, 429, etc.
    error_message = e.message         # Human-readable message
    error_code = e.error.code         # Azure error code
    retry_after = e.response.headers.get('Retry-After')  # For 429
```

**2. Specific Status Codes:**
- **403 (Access Denied)**: Identity needs `Key Vault Secrets User` RBAC role
- **404 (Not Found)**: Secret doesn't exist OR is soft-deleted
- **429 (Throttling)**: Implement exponential backoff with `Retry-After` header

**3. Soft-Deleted Secrets:**
When a secret is soft-deleted:
- `get_secret()` raises `ResourceNotFoundError` (404)
- Use `get_deleted_secret()` to retrieve metadata
- `begin_recover_deleted_secret()` to restore it
- Remains in deleted state until purged or recovered

**4. Best Practices:**
- Use `ResourceNotFoundError` for cleaner 404 handling
- Implement retry logic for 429 with exponential backoff
- Check `Retry-After` header before retrying
- Log error details for debugging (error.code, error.message)

The examples show production-ready patterns including retry logic, detailed error inspection, and handling both active and soft-deleted secrets!

## Verification

**Result:** ✅ PASS

**Summary:** Comprehensive implementation covering all prompt requirements: 403/404/429 handling, error inspection, soft-delete scenarios, and proper exception patterns.

**Reasoning:**

The code comprehensively addresses all requirements from the original prompt. It correctly demonstrates: (1) HttpResponseError and ResourceNotFoundError exception handling with proper imports from azure.core.exceptions; (2) Extracting status_code, error.code, and message properties with detailed examples; (3) 403 handling with specific RBAC role guidance (Key Vault Secrets User/Officer); (4) 404 handling distinguishing between non-existent and soft-deleted secrets; (5) 429 throttling with retry logic, exponential backoff, and Retry-After header inspection; (6) Comprehensive soft-delete scenario with get_deleted_secret(), begin_recover_deleted_secret(), and explanation of deleted state behavior; (7) Multiple production-ready patterns showing try/except blocks. The code is syntactically valid, uses the correct azure-keyvault-secrets SDK, and provides 5 well-structured examples covering all scenarios. Minor note: Expected coverage mentioned 'SecretClient retry configuration via kwargs' which isn't explicitly shown, but the manual retry implementation is thorough and practical. The code would run successfully (given valid credentials/vault) and serves as excellent reference material for the prompt's question.

---

[← Back to Summary](../../../../../../summary.md)
