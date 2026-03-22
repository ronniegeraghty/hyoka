# Evaluation Report: key-vault-dp-python-error-handling

**Config:** azure-mcp | **Result:** ✅ PASSED | **Duration:** 74.1s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-error-handling` |
| Config | azure-mcp |
| Result | ✅ PASSED |
| Duration | 74.1s |
| Timestamp | 2026-03-22T00:09:00Z |
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

The user is asking for guidance on error handling with Azure Key Vault in Python. This is a specific technical question about error handling patterns with the azure-keyvault-secrets SDK. They want:

1. Try/except patterns
2. How to handle specific HTTP status codes (403, 404, 429)
3. How to inspect status_code and error messages on HttpResponseError
4. What happens with soft-deleted secrets

This is a straightforward educational/documentation request. I should create a comprehensive Python example showing best practices for error handling with the Azure Key Vault SDK.

Since this is not a codebase modification task, but rather a demonstration/example task, I should:
1. Create a well-documented example Python file showing the error handling patterns
2. Explain the key concepts
3. Show specific handling for each status code they mentioned

Let me create a comprehensive example that demonstrates all these scenarios.

## Tool Calls

### 🔧 1. report_intent

**Input:**

```json
{"intent":"Creating Key Vault error handling examples"}
```

### 🔧 2. create

**Input:**

```json
{"file_text":"\"\"\"\nAzure Key Vault Error Handling Examples\nDemonstrates proper exception handling with azure-keyvault-secrets SDK\n\"\"\"\n\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.identity import DefaultAzureCredential\nfrom azure.core.exceptions import (\n    HttpResponseError,\n    ResourceNotFoundError,\n    ServiceRequestError\n)\nimport time\n\n\ndef get_secret_with_basic_error_handling(client: SecretClient, secret_name: str):\n    \"\"\"\n    Basic error handling pattern - catches specific status codes\n    \"\"\"\n    try:\n        secret = client.get_secret(secret_name)\n        print(f\"✓ Retrieved secret: {secret.name}\")\n        return secret.value\n    \n    except HttpResponseError as e:\n        # Access the status code directly from the error\n        status_code = e.status_code\n        error_message = e.message\n        \n        if status_code == 403:\n            # Access Denied - missing RBAC role (Key Vault Secrets User or similar)\n            print(f\"❌ Access Denied (403): {error_message}\")\n            print(\"   Your identity lacks the required RBAC role.\")\n            print(\"   Needed: 'Key Vault Secrets User' or 'Key Vault Secrets Officer'\")\n        \n        elif status_code == 404:\n            # Secret Not Found\n            print(f\"❌ Secret Not Found (404): {secret_name}\")\n            print(f\"   Error: {error_message}\")\n        \n        elif status_code == 429:\n            # Throttling - too many requests\n            print(f\"❌ Throttled (429): {error_message}\")\n            print(\"   Rate limit exceeded. Implement exponential backoff.\")\n        \n        else:\n            # Other HTTP errors\n            print(f\"❌ HTTP Error {status_code}: {error_message}\")\n        \n        # Re-raise if you want calling code to handle it\n        raise\n    \n    except ServiceRequestError as e:\n        # Network connectivity issues\n        print(f\"❌ Network Error: {e}\")\n        raise\n\n\ndef get_secret_with_retry_logic(client: SecretClient, secret_name: str, max_retries: int = 3):\n    \"\"\"\n    Advanced pattern with exponential backoff for throttling (429)\n    \"\"\"\n    for attempt in range(max_retries):\n        try:\n            secret = client.get_secret(secret_name)\n            return secret.value\n        \n        except HttpResponseError as e:\n            if e.status_code == 429:\n                # Throttled - implement exponential backoff\n                if attempt \u003c max_retries - 1:\n                    wait_time = 2 ** attempt  # 1s, 2s, 4s, etc.\n                    \n                    # Check if Retry-After header is present\n                    retry_after = e.response.headers.get('Retry-After')\n                    if retry_after:\n                        wait_time = int(retry_after)\n                    \n                    print(f\"⚠️  Throttled. Retrying in {wait_time}s (attempt {attempt + 1}/{max_retries})\")\n                    time.sleep(wait_time)\n                    continue\n                else:\n                    print(f\"❌ Max retries exceeded for throttling\")\n                    raise\n            \n            elif e.status_code == 404:\n                # Don't retry on 404 - secret doesn't exist\n                print(f\"❌ Secret '{secret_name}' not found\")\n                raise\n            \n            elif e.status_code == 403:\n                # Don't retry on 403 - permission issue won't resolve\n                print(f\"❌ Access denied. Check RBAC roles.\")\n                raise\n            \n            else:\n                # Other errors - re-raise immediately\n                raise\n\n\ndef handle_soft_deleted_secret(client: SecretClient, secret_name: str):\n    \"\"\"\n    Demonstrates what happens when accessing a soft-deleted secret\n    \n    When you try to get a soft-deleted secret:\n    - You get a 404 (ResourceNotFoundError) - the secret is not \"active\"\n    - To retrieve it, you must use get_deleted_secret() instead\n    - Or recover it with begin_recover_deleted_secret()\n    \"\"\"\n    try:\n        secret = client.get_secret(secret_name)\n        print(f\"✓ Secret is active: {secret.name}\")\n        return secret.value\n    \n    except HttpResponseError as e:\n        if e.status_code == 404:\n            print(f\"⚠️  Secret not found in active secrets (404)\")\n            print(f\"   Checking if '{secret_name}' is soft-deleted...\")\n            \n            try:\n                # Try to get the deleted secret\n                deleted_secret = client.get_deleted_secret(secret_name)\n                print(f\"✓ Found soft-deleted secret: {deleted_secret.name}\")\n                print(f\"   Deleted on: {deleted_secret.deleted_date}\")\n                print(f\"   Scheduled purge: {deleted_secret.scheduled_purge_date}\")\n                print(f\"   To use it, recover with: client.begin_recover_deleted_secret('{secret_name}')\")\n                return None\n            \n            except HttpResponseError as deleted_err:\n                if deleted_err.status_code == 404:\n                    print(f\"❌ Secret '{secret_name}' does not exist (not active, not deleted)\")\n                else:\n                    print(f\"❌ Error checking deleted secrets: {deleted_err.message}\")\n                return None\n        else:\n            raise\n\n\ndef comprehensive_error_handling(client: SecretClient, secret_name: str):\n    \"\"\"\n    Production-ready error handling with detailed inspection\n    \"\"\"\n    try:\n        secret = client.get_secret(secret_name)\n        return secret.value\n    \n    except HttpResponseError as e:\n        # Detailed error inspection\n        print(f\"\\n{'='*60}\")\n        print(f\"HttpResponseError Details:\")\n        print(f\"{'='*60}\")\n        print(f\"Status Code: {e.status_code}\")\n        print(f\"Reason: {e.reason}\")\n        print(f\"Message: {e.message}\")\n        print(f\"Error Code: {e.error.code if e.error else 'N/A'}\")\n        \n        # Access response headers if needed\n        if hasattr(e, 'response') and e.response:\n            print(f\"Request ID: {e.response.headers.get('x-ms-request-id', 'N/A')}\")\n            print(f\"Retry-After: {e.response.headers.get('Retry-After', 'N/A')}\")\n        \n        # Handle specific cases\n        if e.status_code == 403:\n            print(f\"\\n🔒 RBAC Permission Required:\")\n            print(f\"   Grant your identity one of these roles on the Key Vault:\")\n            print(f\"   - Key Vault Secrets User (read-only)\")\n            print(f\"   - Key Vault Secrets Officer (read/write)\")\n            print(f\"   - Key Vault Administrator (full access)\")\n        \n        elif e.status_code == 404:\n            print(f\"\\n🔍 Secret Not Found:\")\n            print(f\"   - Verify the secret name: '{secret_name}'\")\n            print(f\"   - Check if it's been deleted (soft-delete may be enabled)\")\n            print(f\"   - Verify you're using the correct Key Vault URL\")\n        \n        elif e.status_code == 429:\n            print(f\"\\n⏱️  Rate Limit Exceeded:\")\n            print(f\"   - Implement exponential backoff\")\n            print(f\"   - Check 'Retry-After' header for wait time\")\n            print(f\"   - Consider caching secrets locally\")\n        \n        print(f\"{'='*60}\\n\")\n        raise\n\n\n# Usage Examples\nif __name__ == \"__main__\":\n    # Initialize client\n    vault_url = \"https://your-keyvault-name.vault.azure.net/\"\n    \n    try:\n        credential = DefaultAzureCredential()\n        client = SecretClient(vault_url=vault_url, credential=credential)\n        \n        # Example 1: Basic error handling\n        print(\"Example 1: Basic Error Handling\")\n        print(\"-\" * 40)\n        try:\n            value = get_secret_with_basic_error_handling(client, \"my-secret\")\n            print(f\"Secret value: {value}\")\n        except HttpResponseError as e:\n            print(f\"Failed to retrieve secret: {e.status_code}\")\n        \n        print(\"\\n\")\n        \n        # Example 2: Retry logic for throttling\n        print(\"Example 2: Retry Logic with Backoff\")\n        print(\"-\" * 40)\n        try:\n            value = get_secret_with_retry_logic(client, \"my-secret\", max_retries=3)\n            print(f\"Secret value: {value}\")\n        except HttpResponseError as e:\n            print(f\"Failed after retries: {e.status_code}\")\n        \n        print(\"\\n\")\n        \n        # Example 3: Soft-deleted secret handling\n        print(\"Example 3: Soft-Deleted Secret Handling\")\n        print(\"-\" * 40)\n        value = handle_soft_deleted_secret(client, \"deleted-secret\")\n        \n        print(\"\\n\")\n        \n        # Example 4: Comprehensive error inspection\n        print(\"Example 4: Comprehensive Error Details\")\n        print(\"-\" * 40)\n        try:\n            value = comprehensive_error_handling(client, \"test-secret\")\n        except HttpResponseError:\n            pass\n    \n    except Exception as e:\n        print(f\"Failed to initialize client: {e}\")\n\n\n# Quick Reference: Common Status Codes\n\"\"\"\nStatus Code | Meaning | Common Cause | Retry?\n------------|---------|--------------|--------\n403         | Forbidden | Missing RBAC role | No - Fix permissions\n404         | Not Found | Secret doesn't exist or is soft-deleted | No - Check name\n429         | Too Many Requests | Rate limit exceeded | Yes - With backoff\n401         | Unauthorized | Authentication failed | No - Fix credentials\n500         | Server Error | Azure service issue | Yes - Transient\n503         | Service Unavailable | Service temporarily down | Yes - Transient\n\"\"\"\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-error-handling-azure-mcp-754540071/keyvault_error_handling.py"}
```

## Generated Files

- `keyvault_error_handling.py`

## Copilot Response

I'll create a comprehensive example showing proper error handling patterns for Azure Key Vault in Python.

Perfect! I've created a comprehensive example file demonstrating Azure Key Vault error handling patterns. Here are the key points:

## Key Error Handling Patterns:

**1. Inspecting HttpResponseError:**
```python
except HttpResponseError as e:
    status_code = e.status_code  # 403, 404, 429, etc.
    error_message = e.message     # Human-readable message
    error_code = e.error.code     # Specific error code
```

**2. Handling Specific Status Codes:**
- **403 (Forbidden)**: Missing RBAC role - don't retry, fix permissions
- **404 (Not Found)**: Secret doesn't exist - check if soft-deleted
- **429 (Throttled)**: Rate limit hit - retry with exponential backoff

**3. Soft-Deleted Secrets:**
When you `get_secret()` on a soft-deleted secret:
- You get **404 (Not Found)** - it's not "active"
- Use `get_deleted_secret()` to check if it's soft-deleted
- Use `begin_recover_deleted_secret()` to restore it

**4. Retry Pattern for 429:**
- Implement exponential backoff (1s, 2s, 4s...)
- Check the `Retry-After` header in the response
- Don't retry 403 or 404 - they won't resolve themselves

The file includes 4 complete examples showing basic handling, retry logic, soft-delete scenarios, and comprehensive error inspection with all the details you can extract from the exception.

## Verification

**Result:** ✅ PASS

**Summary:** Code passes: comprehensively handles Azure Key Vault errors (403/404/429), demonstrates status_code/message extraction, includes soft-delete handling, and implements retry logic with only minor omission of SecretClient retry kwargs.

**Reasoning:**

The code comprehensively addresses the prompt requirements. It correctly imports and uses HttpResponseError from azure.core.exceptions, demonstrates extraction of status_code, error.code, and message properties, and implements specific handling for 403 (RBAC), 404 (not found), and 429 (throttling) errors. The soft-delete scenario is well-explained with get_deleted_secret() and mentions begin_recover_deleted_secret(). The code includes exponential backoff for 429 errors with Retry-After header inspection, multiple error handling patterns from basic to production-ready, and is syntactically valid Python. One minor omission is not demonstrating SecretClient retry configuration via constructor kwargs (e.g., retry_total, retry_mode), though manual retry logic is implemented. The code also imports ResourceNotFoundError but doesn't explicitly use it, which is acceptable since HttpResponseError is the more common pattern. Overall, the code delivers 95%+ of expected coverage with clear examples, proper exception handling patterns, and production-ready error inspection techniques.

---

[← Back to Summary](../../../../../../summary.md)
