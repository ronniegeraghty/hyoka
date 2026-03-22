# Evaluation Report: key-vault-dp-python-error-handling

**Config:** azure-mcp | **Result:** ✅ PASSED | **Duration:** 86.2s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-error-handling` |
| Config | azure-mcp |
| Result | ✅ PASSED |
| Duration | 86.2s |
| Timestamp | 2026-03-22T21:52:32Z |
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
1. Try/except patterns for common errors (403, 404, 429)
2. How to inspect status_code and error message on HttpResponseError
3. What happens when getting a soft-deleted secret

This is a straightforward educational/example request. I should:
1. Create a comprehensive Python example showing error handling patterns
2. Demonstrate how to catch HttpResponseError and inspect its properties
3. Explain the soft-deleted secret scenario

I'll create an example Python file demonstrating these patterns.

## Tool Calls

### 🔧 1. report_intent

**Input:**

```json
{"intent":"Creating error handling examples"}
```

### 🔧 2. create

**Input:**

```json
{"file_text":"\"\"\"\nAzure Key Vault Secrets Error Handling Examples\n\nThis module demonstrates proper error handling patterns when working with\nAzure Key Vault secrets, including handling 403 (access denied), 404 (not found),\n429 (throttling), and soft-deleted secrets.\n\"\"\"\n\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.identity import DefaultAzureCredential\nfrom azure.core.exceptions import (\n    HttpResponseError,\n    ResourceNotFoundError,\n    ServiceRequestError,\n)\nimport time\n\n\ndef get_secret_with_basic_error_handling(client: SecretClient, secret_name: str):\n    \"\"\"\n    Basic error handling pattern using HttpResponseError.\n    \n    This demonstrates how to catch and inspect the status code and error message.\n    \"\"\"\n    try:\n        secret = client.get_secret(secret_name)\n        print(f\"✓ Successfully retrieved secret: {secret.name}\")\n        return secret.value\n    \n    except HttpResponseError as e:\n        # The status_code property contains the HTTP status code\n        status_code = e.status_code\n        \n        # The message property contains the error description\n        error_message = e.message\n        \n        # The error property contains more detailed error information\n        error_code = e.error.code if hasattr(e, 'error') and e.error else None\n        \n        print(f\"✗ HTTP Error {status_code}: {error_message}\")\n        if error_code:\n            print(f\"  Error Code: {error_code}\")\n        \n        # Handle specific status codes\n        if status_code == 403:\n            print(\"  → Access Denied: Check RBAC permissions (Key Vault Secrets User role)\")\n        elif status_code == 404:\n            print(\"  → Secret Not Found: The secret does not exist or may be soft-deleted\")\n        elif status_code == 429:\n            print(\"  → Rate Limit Exceeded: Too many requests, implement retry with backoff\")\n        \n        raise  # Re-raise after logging\n\n\ndef get_secret_with_specific_handling(client: SecretClient, secret_name: str):\n    \"\"\"\n    Handle specific error scenarios with tailored responses.\n    \"\"\"\n    try:\n        secret = client.get_secret(secret_name)\n        return secret.value\n    \n    except HttpResponseError as e:\n        if e.status_code == 403:\n            # Access denied - identity lacks required RBAC role\n            print(f\"Access denied to secret '{secret_name}'\")\n            print(\"Required RBAC role: 'Key Vault Secrets User' or 'Key Vault Secrets Officer'\")\n            print(f\"Error details: {e.message}\")\n            return None\n        \n        elif e.status_code == 404:\n            # Secret not found - may not exist or is soft-deleted\n            print(f\"Secret '{secret_name}' not found\")\n            print(\"The secret may be soft-deleted. Use get_deleted_secret() to check.\")\n            return None\n        \n        elif e.status_code == 429:\n            # Throttling - rate limit exceeded\n            retry_after = e.response.headers.get('Retry-After', 60)\n            print(f\"Rate limit exceeded. Retry after {retry_after} seconds\")\n            raise\n        \n        else:\n            # Other HTTP errors\n            print(f\"Unexpected error (HTTP {e.status_code}): {e.message}\")\n            raise\n\n\ndef get_secret_with_retry_on_throttle(client: SecretClient, secret_name: str, max_retries: int = 3):\n    \"\"\"\n    Implement retry logic for throttling (429) errors.\n    \"\"\"\n    for attempt in range(max_retries):\n        try:\n            secret = client.get_secret(secret_name)\n            return secret.value\n        \n        except HttpResponseError as e:\n            if e.status_code == 429:\n                if attempt \u003c max_retries - 1:\n                    # Get retry-after header or use exponential backoff\n                    retry_after = int(e.response.headers.get('Retry-After', 2 ** attempt))\n                    print(f\"Rate limited. Retrying in {retry_after} seconds (attempt {attempt + 1}/{max_retries})\")\n                    time.sleep(retry_after)\n                    continue\n                else:\n                    print(f\"Max retries exceeded for secret '{secret_name}'\")\n                    raise\n            else:\n                # Non-throttling errors - don't retry\n                raise\n    \n    return None\n\n\ndef handle_soft_deleted_secret(client: SecretClient, secret_name: str):\n    \"\"\"\n    Demonstrate handling of soft-deleted secrets.\n    \n    When a secret is deleted (soft delete enabled), it enters a soft-deleted state.\n    - get_secret() will return 404 (Not Found)\n    - get_deleted_secret() can retrieve information about the deleted secret\n    - You must purge or recover the secret before creating a new one with the same name\n    \"\"\"\n    try:\n        # Try to get the secret normally\n        secret = client.get_secret(secret_name)\n        print(f\"✓ Secret '{secret_name}' exists and is active\")\n        return secret.value\n    \n    except HttpResponseError as e:\n        if e.status_code == 404:\n            print(f\"Secret '{secret_name}' not found in active secrets\")\n            \n            # Check if it's soft-deleted\n            try:\n                deleted_secret = client.get_deleted_secret(secret_name)\n                print(f\"✓ Found soft-deleted secret: {deleted_secret.name}\")\n                print(f\"  Deleted on: {deleted_secret.deleted_date}\")\n                print(f\"  Scheduled purge date: {deleted_secret.scheduled_purge_date}\")\n                print(f\"  Recovery ID: {deleted_secret.recovery_id}\")\n                print(\"\\nOptions:\")\n                print(\"  1. Recover: client.begin_recover_deleted_secret(secret_name)\")\n                print(\"  2. Purge: client.purge_deleted_secret(secret_name)\")\n                return None\n            \n            except HttpResponseError as deleted_error:\n                if deleted_error.status_code == 404:\n                    print(f\"✗ Secret '{secret_name}' does not exist (not active or deleted)\")\n                else:\n                    print(f\"Error checking deleted secrets: {deleted_error.message}\")\n                return None\n        else:\n            raise\n\n\ndef comprehensive_error_handling_example(client: SecretClient, secret_name: str):\n    \"\"\"\n    Comprehensive error handling combining all scenarios.\n    \"\"\"\n    try:\n        secret = client.get_secret(secret_name)\n        print(f\"✓ Retrieved secret: {secret.name}\")\n        return secret.value\n    \n    except HttpResponseError as e:\n        # Extract error details\n        status_code = e.status_code\n        message = e.message\n        error_code = getattr(e.error, 'code', 'Unknown') if hasattr(e, 'error') and e.error else 'Unknown'\n        \n        print(f\"\\n{'='*60}\")\n        print(f\"HTTP Response Error Details:\")\n        print(f\"{'='*60}\")\n        print(f\"Status Code:  {status_code}\")\n        print(f\"Error Code:   {error_code}\")\n        print(f\"Message:      {message}\")\n        print(f\"{'='*60}\\n\")\n        \n        # Handle each scenario\n        if status_code == 403:\n            print(\"🔒 ACCESS DENIED\")\n            print(\"Cause: Your application's identity lacks the required RBAC role\")\n            print(\"Solution:\")\n            print(\"  1. Ensure managed identity is enabled (if using Azure services)\")\n            print(\"  2. Assign 'Key Vault Secrets User' role to the identity\")\n            print(\"  3. Command: az role assignment create \\\\\")\n            print(\"       --role 'Key Vault Secrets User' \\\\\")\n            print(\"       --assignee \u003cprincipal-id\u003e \\\\\")\n            print(\"       --scope \u003ckey-vault-resource-id\u003e\")\n            \n        elif status_code == 404:\n            print(\"🔍 SECRET NOT FOUND\")\n            print(\"Checking if secret is soft-deleted...\")\n            handle_soft_deleted_secret(client, secret_name)\n            \n        elif status_code == 429:\n            print(\"⏱️  RATE LIMIT EXCEEDED\")\n            retry_after = e.response.headers.get('Retry-After', 'unknown')\n            print(f\"Retry-After header: {retry_after} seconds\")\n            print(\"Solution:\")\n            print(\"  1. Implement exponential backoff retry logic\")\n            print(\"  2. Reduce request frequency\")\n            print(\"  3. Consider caching secrets if appropriate\")\n            \n        elif status_code \u003e= 500:\n            print(\"🔧 SERVER ERROR\")\n            print(\"This is a transient error from Azure Key Vault\")\n            print(\"Implement retry logic with exponential backoff\")\n        \n        return None\n    \n    except ServiceRequestError as e:\n        print(f\"Network error: {e}\")\n        print(\"Check network connectivity and firewall rules\")\n        return None\n    \n    except Exception as e:\n        print(f\"Unexpected error: {type(e).__name__}: {e}\")\n        raise\n\n\ndef main():\n    \"\"\"\n    Example usage of error handling patterns.\n    \"\"\"\n    # Initialize the client\n    # vault_url = \"https://your-keyvault.vault.azure.net/\"\n    # credential = DefaultAzureCredential()\n    # client = SecretClient(vault_url=vault_url, credential=credential)\n    \n    print(\"\"\"\nAzure Key Vault Secrets - Error Handling Patterns\n================================================\n\nThis file demonstrates how to properly handle errors when working with\nAzure Key Vault secrets in Python.\n\nKey Error Scenarios:\n-------------------\n1. **403 Forbidden (Access Denied)**\n   - Cause: Identity lacks RBAC permissions\n   - Required role: 'Key Vault Secrets User' or 'Key Vault Secrets Officer'\n   - Properties: e.status_code == 403, e.message contains details\n\n2. **404 Not Found**\n   - Cause: Secret doesn't exist or is soft-deleted\n   - Check: Use get_deleted_secret() to check soft-deleted secrets\n   - Properties: e.status_code == 404\n\n3. **429 Too Many Requests (Throttling)**\n   - Cause: Rate limit exceeded\n   - Solution: Implement retry with exponential backoff\n   - Check 'Retry-After' header: e.response.headers.get('Retry-After')\n\n4. **Soft-Deleted Secrets**\n   - get_secret() returns 404\n   - Use get_deleted_secret() to retrieve deleted secret info\n   - Options: recover (begin_recover_deleted_secret) or purge (purge_deleted_secret)\n\nHttpResponseError Properties:\n----------------------------\n- e.status_code: HTTP status code (int)\n- e.message: Error message (str)\n- e.error: Detailed error object (may have .code attribute)\n- e.response: Full HTTP response (has .headers, .status_code, etc.)\n\nExample Usage:\n-------------\n    from azure.keyvault.secrets import SecretClient\n    from azure.identity import DefaultAzureCredential\n    \n    vault_url = \"https://your-keyvault.vault.azure.net/\"\n    credential = DefaultAzureCredential()\n    client = SecretClient(vault_url=vault_url, credential=credential)\n    \n    # Basic error handling\n    secret_value = get_secret_with_basic_error_handling(client, \"my-secret\")\n    \n    # Comprehensive handling\n    secret_value = comprehensive_error_handling_example(client, \"my-secret\")\n    \n    # With retry on throttling\n    secret_value = get_secret_with_retry_on_throttle(client, \"my-secret\")\n    \n    # Check soft-deleted\n    handle_soft_deleted_secret(client, \"deleted-secret\")\n\"\"\")\n\n\nif __name__ == \"__main__\":\n    main()\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-error-handling-azure-mcp-1019751460/keyvault_error_handling.py"}
```

## Generated Files

- `keyvault_error_handling.py`

## Copilot Response

I'll create a comprehensive example demonstrating error handling patterns for Azure Key Vault secrets in Python.

Perfect! I've created a comprehensive guide to error handling with Azure Key Vault secrets in Python. Here are the key patterns:

## **Key Error Handling Patterns:**

### **1. Inspecting HttpResponseError:**
```python
except HttpResponseError as e:
    status_code = e.status_code        # HTTP status code (403, 404, 429, etc.)
    message = e.message                # Error description
    error_code = e.error.code          # Detailed error code (if available)
    retry_after = e.response.headers.get('Retry-After')  # For 429 errors
```

### **2. Common Scenarios:**

- **403 (Access Denied)**: Your app's managed identity needs the "Key Vault Secrets User" or "Key Vault Secrets Officer" RBAC role
- **404 (Not Found)**: Secret doesn't exist OR is soft-deleted (check with `get_deleted_secret()`)
- **429 (Throttling)**: Implement retry with exponential backoff, respect the `Retry-After` header

### **3. Soft-Deleted Secrets:**
When you delete a secret (with soft-delete enabled):
- `get_secret()` returns **404 Not Found**
- Use `get_deleted_secret()` to retrieve info about the deleted secret
- You can **recover** it with `begin_recover_deleted_secret()` or **purge** it with `purge_deleted_secret()`
- You **cannot** create a new secret with the same name until it's recovered or purged

The file includes 5 different functions demonstrating progressively more sophisticated error handling patterns. Run `python keyvault_error_handling.py` to see the documentation, or use the individual functions as templates for your code!

## Verification

**Result:** ✅ PASS

**Summary:** Code passes - comprehensively handles all required error scenarios with proper exception inspection and soft-delete awareness

**Reasoning:**

The generated code comprehensively addresses the prompt requirements. It correctly imports and uses HttpResponseError from azure.core.exceptions, demonstrates extracting status_code, error.code, and message properties, and provides detailed handling for all three error scenarios: 403 (access denied with RBAC role guidance), 404 (with soft-delete awareness), and 429 (throttling with retry-after header extraction). The code includes multiple well-documented functions showing basic error handling, specific scenario handling, retry logic for throttling, and comprehensive soft-delete handling using get_deleted_secret() and mentioning begin_recover_deleted_secret(). The syntax is valid Python and would execute correctly with proper Azure credentials. The only minor omission is not demonstrating SecretClient retry configuration via constructor kwargs (e.g., retry_total, retry_backoff_factor), but this is offset by the manual retry implementation shown in get_secret_with_retry_on_throttle(). The code provides educational value with clear examples, comments, and a detailed main() docstring explaining all scenarios.

---

[← Back to Summary](../../../../../../summary.md)
