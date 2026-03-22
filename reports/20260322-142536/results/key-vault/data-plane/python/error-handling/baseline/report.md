# Evaluation Report: key-vault-dp-python-error-handling

**Config:** baseline | **Result:** ❌ FAILED | **Duration:** 33.9s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-error-handling` |
| Config | baseline |
| Result | ❌ FAILED |
| Duration | 33.9s |
| Timestamp | 2026-03-22T21:25:36Z |
| Files Generated | 0 |
| Event Count | 11 |

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

The user is asking for guidance on error handling with Azure Key Vault in Python. This is a knowledge/tutorial request rather than a code modification task. They want to understand:

1. How to handle specific HTTP errors (403, 404, 429)
2. Try/except patterns with azure-keyvault-secrets SDK
3. How to inspect status_code and error messages on HttpResponseError
4. What happens with soft-deleted secrets

This is a straightforward question that I can answer directly. I should:
- Show them concrete code examples
- Explain the error handling patterns
- Cover the specific scenarios they mentioned
- Explain soft-deleted secret behavior

I don't need to explore a codebase or make changes. I can provide a comprehensive answer with code examples.

## Copilot Response

I'll show you comprehensive error handling patterns for Azure Key Vault secrets in Python.

Here's a complete example covering all the scenarios you mentioned:

```python
from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential
from azure.core.exceptions import HttpResponseError, ResourceNotFoundError
import time

# Initialize the client
credential = DefaultAzureCredential()
client = SecretClient(vault_url="https://your-vault.vault.azure.net/", credential=credential)

def get_secret_with_error_handling(secret_name):
    """
    Demonstrates proper error handling for Azure Key Vault operations
    """
    try:
        secret = client.get_secret(secret_name)
        return secret.value
        
    except HttpResponseError as e:
        # Access the status code
        status_code = e.status_code
        
        # Access the error message
        error_message = e.message
        
        # Handle specific HTTP status codes
        if status_code == 403:
            # Access Denied - RBAC permissions issue
            print(f"Access Denied (403): {error_message}")
            print("Your identity lacks the required RBAC role.")
            print("Required: 'Key Vault Secrets User' or 'Key Vault Secrets Officer'")
            # You can also check: e.error.code, e.error.message
            
        elif status_code == 404:
            # Secret Not Found
            print(f"Secret Not Found (404): {error_message}")
            print(f"Secret '{secret_name}' does not exist in the vault")
            
        elif status_code == 429:
            # Rate Limiting / Throttling
            print(f"Throttled (429): {error_message}")
            print("Rate limit exceeded. Implementing retry with backoff...")
            
            # Check for Retry-After header
            retry_after = e.response.headers.get('Retry-After')
            if retry_after:
                wait_time = int(retry_after)
            else:
                wait_time = 5  # Default backoff
                
            print(f"Waiting {wait_time} seconds before retry...")
            time.sleep(wait_time)
            
            # Retry the operation
            return get_secret_with_retry(secret_name)
            
        else:
            # Other HTTP errors
            print(f"HTTP Error {status_code}: {error_message}")
            
        # Additional error details
        if hasattr(e, 'error'):
            print(f"Error Code: {e.error.code}")
            print(f"Error Details: {e.error.message}")
            
        raise  # Re-raise if you want calling code to handle it
        
    except ResourceNotFoundError as e:
        # This is a more specific exception for 404s
        print(f"Resource not found: {e.message}")
        raise
        
    except Exception as e:
        # Catch-all for unexpected errors
        print(f"Unexpected error: {type(e).__name__}: {str(e)}")
        raise


def get_secret_with_retry(secret_name, max_retries=3):
    """
    Implements exponential backoff for retries
    """
    for attempt in range(max_retries):
        try:
            secret = client.get_secret(secret_name)
            return secret.value
            
        except HttpResponseError as e:
            if e.status_code == 429:
                if attempt < max_retries - 1:
                    wait_time = 2 ** attempt  # Exponential backoff: 1, 2, 4 seconds
                    print(f"Retry attempt {attempt + 1}/{max_retries} after {wait_time}s")
                    time.sleep(wait_time)
                else:
                    print(f"Max retries ({max_retries}) exceeded")
                    raise
            else:
                raise  # Re-raise non-429 errors immediately


def handle_soft_deleted_secret(secret_name):
    """
    Demonstrates what happens with soft-deleted secrets
    """
    try:
        # Trying to get a soft-deleted secret returns 404
        secret = client.get_secret(secret_name)
        return secret.value
        
    except HttpResponseError as e:
        if e.status_code == 404:
            print(f"Secret '{secret_name}' not found (404)")
            print("This could mean:")
            print("1. The secret never existed")
            print("2. The secret was deleted and is in soft-delete state")
            
            # Check if it's soft-deleted
            try:
                deleted_secret = client.get_deleted_secret(secret_name)
                print(f"\nSecret IS soft-deleted!")
                print(f"Deleted on: {deleted_secret.deleted_date}")
                print(f"Scheduled purge: {deleted_secret.scheduled_purge_date}")
                print(f"Recovery ID: {deleted_secret.recovery_id}")
                
                # You can recover it if needed:
                # recovered = client.begin_recover_deleted_secret(secret_name).result()
                
            except HttpResponseError as recover_err:
                if recover_err.status_code == 404:
                    print("\nSecret is NOT soft-deleted - it simply doesn't exist")
                elif recover_err.status_code == 403:
                    print("\nNo permission to check deleted secrets")
                    print("Required: 'Key Vault Secrets Officer' role")


# Example: Complete error handling workflow
def production_secret_retrieval(secret_name):
    """
    Production-ready pattern with logging and fallback
    """
    try:
        secret = client.get_secret(secret_name)
        return secret.value
        
    except HttpResponseError as e:
        # Log the full error details
        error_info = {
            'status_code': e.status_code,
            'message': e.message,
            'secret_name': secret_name,
        }
        
        if hasattr(e, 'error'):
            error_info['error_code'] = e.error.code
            
        # Handle based on status
        if e.status_code == 403:
            # Log security event and alert
            print(f"SECURITY: Unauthorized access attempt: {error_info}")
            return None  # Or use a default value
            
        elif e.status_code == 404:
            # Could use a fallback or default
            print(f"WARNING: Secret missing: {error_info}")
            return None
            
        elif e.status_code == 429:
            # Implement retry with backoff
            return get_secret_with_retry(secret_name)
            
        else:
            # Unknown error - log and fail
            print(f"ERROR: Key Vault operation failed: {error_info}")
            raise


# Usage examples
if __name__ == "__main__":
    # Example 1: Basic error handling
    try:
        value = get_secret_with_error_handling("my-secret")
        print(f"Secret value: {value}")
    except Exception as e:
        print(f"Failed to retrieve secret: {e}")
    
    # Example 2: Check soft-deleted secret
    handle_soft_deleted_secret("deleted-secret")
```

**Key Points:**

1. **HttpResponseError** is the main exception - check `e.status_code` for HTTP codes
2. **Soft-deleted secrets** return 404 on `get_secret()` - use `get_deleted_secret()` to check if they're recoverable
3. **For 429 throttling**: Check `Retry-After` header and implement exponential backoff
4. **For 403 errors**: User needs "Key Vault Secrets User" (read) or "Key Vault Secrets Officer" (read/write/delete) RBAC role
5. **Error object structure**: `e.error.code` and `e.error.message` provide additional details

The SDK also supports automatic retries via `RetryPolicy` if you want to configure it at the client level.

## Verification

**Result:** ❌ FAIL

**Summary:** FAIL — no output files found

**Reasoning:**

No files were generated.

---

[← Back to Summary](../../../../../../summary.md)
