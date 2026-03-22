# Evaluation Report: key-vault-dp-python-error-handling

**Config:** azure-mcp | **Result:** ❌ FAILED | **Duration:** 30.3s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-error-handling` |
| Config | azure-mcp |
| Result | ❌ FAILED |
| Duration | 30.3s |
| Timestamp | 2026-03-21T06:38:17Z |
| Files Generated | 0 |
| Event Count | 11 |

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

The user is asking for guidance on error handling with Azure Key Vault secrets in Python. This is a knowledge/documentation question rather than a task requiring code changes in a repository. They want to learn about:

1. Handling specific HTTP errors (403, 404, 429)
2. Try/except patterns with azure-keyvault-secrets SDK
3. How to inspect status_code and error message on HttpResponseError
4. What happens with soft-deleted secrets

Since this is a direct question asking for information/examples rather than a task to implement in a codebase, I should provide a concise, helpful response with code examples. I don't need to use the plan mode or explore the codebase - there's no indication we're working in an existing project that needs modification.

Let me provide a comprehensive but concise answer with practical code examples.

## Copilot Response

I'll show you comprehensive error handling patterns for Azure Key Vault secrets in Python with the `azure-keyvault-secrets` SDK.

Here are the key error handling patterns:

## Basic Error Handling

```python
from azure.keyvault.secrets import SecretClient
from azure.core.exceptions import HttpResponseError, ResourceNotFoundError
from azure.identity import DefaultAzureCredential

credential = DefaultAzureCredential()
client = SecretClient(vault_url="https://your-vault.vault.azure.net/", credential=credential)

# Pattern 1: Specific exception handling
try:
    secret = client.get_secret("my-secret")
    print(f"Secret value: {secret.value}")
except ResourceNotFoundError:
    print("Secret not found (404)")
except HttpResponseError as e:
    if e.status_code == 403:
        print(f"Access denied: {e.message}")
        print("Check RBAC role assignments (need 'Key Vault Secrets User' or similar)")
    elif e.status_code == 429:
        print(f"Rate limited: {e.message}")
        print(f"Retry after: {e.response.headers.get('Retry-After', 'unknown')} seconds")
    else:
        print(f"HTTP error {e.status_code}: {e.message}")
```

## Inspecting HttpResponseError Details

```python
try:
    secret = client.get_secret("secret-name")
except HttpResponseError as e:
    print(f"Status Code: {e.status_code}")
    print(f"Error Message: {e.message}")
    print(f"Error Code: {e.error.code if e.error else 'N/A'}")
    print(f"Full Error: {e.error}")
    
    # Access response headers (useful for retry logic)
    if e.response:
        print(f"Retry-After header: {e.response.headers.get('Retry-After')}")
        print(f"Request ID: {e.response.headers.get('x-ms-request-id')}")
```

## Handling Soft-Deleted Secrets

```python
# When you try to get a soft-deleted secret, you get a 404
try:
    secret = client.get_secret("deleted-secret")
except ResourceNotFoundError:
    print("Secret not found - may be soft-deleted")
    
    # Check if it's in the deleted secrets
    try:
        deleted_secret = client.get_deleted_secret("deleted-secret")
        print(f"Found in deleted secrets, deleted on: {deleted_secret.deleted_date}")
        print(f"Scheduled purge: {deleted_secret.scheduled_purge_date}")
        
        # Recover if needed
        recovered = client.begin_recover_deleted_secret("deleted-secret").result()
        print(f"Recovered secret")
    except ResourceNotFoundError:
        print("Secret truly doesn't exist (not even soft-deleted)")
    except HttpResponseError as e:
        if e.status_code == 403:
            print("No permission to view/recover deleted secrets")
```

## Comprehensive Pattern with Retry Logic

```python
import time
from azure.core.exceptions import HttpResponseError, ResourceNotFoundError

def get_secret_with_retry(client, secret_name, max_retries=3):
    """Get secret with exponential backoff for rate limiting"""
    for attempt in range(max_retries):
        try:
            return client.get_secret(secret_name)
        
        except ResourceNotFoundError:
            raise  # Don't retry 404s
        
        except HttpResponseError as e:
            if e.status_code == 403:
                # Access denied - no point retrying
                raise ValueError(
                    f"Access denied to secret '{secret_name}'. "
                    f"Ensure the identity has 'Key Vault Secrets User' role."
                ) from e
            
            elif e.status_code == 429:
                # Rate limited - retry with backoff
                retry_after = int(e.response.headers.get('Retry-After', 2 ** attempt))
                if attempt < max_retries - 1:
                    print(f"Rate limited. Retrying after {retry_after}s...")
                    time.sleep(retry_after)
                else:
                    raise ValueError(f"Rate limit exceeded after {max_retries} attempts") from e
            
            else:
                # Other HTTP errors
                raise

# Usage
try:
    secret = get_secret_with_retry(client, "my-secret")
    print(f"Got secret: {secret.name}")
except Exception as e:
    print(f"Failed to get secret: {e}")
```

## Common Status Codes

- **403 Forbidden**: Missing RBAC role (e.g., "Key Vault Secrets User")
- **404 Not Found**: Secret doesn't exist or is soft-deleted
- **429 Too Many Requests**: Rate limit hit (check `Retry-After` header)
- **401 Unauthorized**: Authentication failure (credential issue)

## Key Points

1. **Soft-deleted secrets**: Return 404 on normal `get_secret()` - use `get_deleted_secret()` to access them
2. **RBAC roles needed**: Minimum "Key Vault Secrets User" for read, "Key Vault Secrets Officer" for write/delete
3. **Rate limiting**: Always check the `Retry-After` header and implement exponential backoff
4. **ResourceNotFoundError** is a convenience exception for 404s, but you can also catch it as `HttpResponseError`

Would you like me to create a complete working example script with all these patterns?

## Verification

**Result:** ❌ FAIL

**Summary:** FAIL — no output files found

**Reasoning:**

No files were generated.

---

[← Back to Summary](../../../../../../summary.md)
