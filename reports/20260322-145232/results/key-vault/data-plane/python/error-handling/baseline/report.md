# Evaluation Report: key-vault-dp-python-error-handling

**Config:** baseline | **Result:** ❌ FAILED | **Duration:** 31.4s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-error-handling` |
| Config | baseline |
| Result | ❌ FAILED |
| Duration | 31.4s |
| Timestamp | 2026-03-22T21:52:32Z |
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

The user is asking for guidance on error handling with Azure Key Vault secrets in Python. This is a question about best practices and code examples, not a request to modify existing code in a repository. They want to understand:

1. How to handle specific HTTP errors (403, 404, 429)
2. Try/except patterns with azure-keyvault-secrets SDK
3. How to inspect status_code and error message on HttpResponseError
4. What happens with soft-deleted secrets

This is an informational request that I should answer directly. I don't need to explore a codebase or create a plan. I should provide clear examples and explanations.

Let me provide a comprehensive answer with code examples showing:
- Proper imports
- Try/except patterns for different error scenarios
- How to inspect HttpResponseError attributes
- Explanation of soft-deleted secret behavior
- Best practices for retry logic

## Copilot Response

I'll help you understand error handling patterns for Azure Key Vault secrets in Python. Let me provide you with comprehensive examples.

## Key Error Handling Patterns

The `azure-keyvault-secrets` SDK raises `HttpResponseError` for HTTP errors. Here's how to handle common scenarios:

```python
from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient
from azure.core.exceptions import HttpResponseError, ServiceRequestError
import time

# Initialize the client
vault_url = "https://your-keyvault.vault.azure.net/"
credential = DefaultAzureCredential()
client = SecretClient(vault_url=vault_url, credential=credential)

# Example 1: Handle specific status codes
def get_secret_with_error_handling(secret_name):
    try:
        secret = client.get_secret(secret_name)
        return secret.value
    
    except HttpResponseError as e:
        # Inspect the status code
        status_code = e.status_code
        error_message = e.message
        
        if status_code == 403:
            print(f"Access Denied (403): {error_message}")
            print("Your identity lacks the required RBAC role (Key Vault Secrets User)")
            print("Or doesn't have 'Get' permission in access policies")
            # Log for security audit, alert admins
            
        elif status_code == 404:
            print(f"Secret Not Found (404): {error_message}")
            print(f"Secret '{secret_name}' does not exist or was deleted")
            # Check if it's soft-deleted (see below)
            
        elif status_code == 429:
            print(f"Throttled (429): {error_message}")
            print("Rate limit exceeded. Implement retry with exponential backoff")
            # Extract retry-after header if available
            retry_after = e.response.headers.get('Retry-After', 30)
            print(f"Retry after: {retry_after} seconds")
            
        else:
            print(f"HTTP Error {status_code}: {error_message}")
        
        # Access additional details
        print(f"Error code: {e.error.code if hasattr(e, 'error') else 'N/A'}")
        raise
    
    except ServiceRequestError as e:
        print(f"Network error: {e}")
        # Handle network connectivity issues
        raise


# Example 2: Retry logic for throttling (429)
def get_secret_with_retry(secret_name, max_retries=3):
    retries = 0
    backoff = 1  # Initial backoff in seconds
    
    while retries < max_retries:
        try:
            secret = client.get_secret(secret_name)
            return secret.value
        
        except HttpResponseError as e:
            if e.status_code == 429:
                # Extract Retry-After header
                retry_after = e.response.headers.get('Retry-After')
                if retry_after:
                    wait_time = int(retry_after)
                else:
                    # Exponential backoff
                    wait_time = backoff * (2 ** retries)
                
                print(f"Throttled. Waiting {wait_time} seconds before retry {retries + 1}/{max_retries}")
                time.sleep(wait_time)
                retries += 1
            else:
                # Non-retriable error
                raise
    
    raise Exception(f"Failed to get secret after {max_retries} retries")


# Example 3: Handling soft-deleted secrets
def handle_soft_deleted_secret(secret_name):
    """
    When a secret is deleted with soft-delete enabled:
    - get_secret() raises 404 HttpResponseError
    - The secret is in "deleted" state but can be recovered
    - You can retrieve it using get_deleted_secret()
    """
    try:
        secret = client.get_secret(secret_name)
        print(f"Secret value: {secret.value}")
        
    except HttpResponseError as e:
        if e.status_code == 404:
            print(f"Secret not found. Checking if it's soft-deleted...")
            
            try:
                # Check deleted secrets
                deleted_secret = client.get_deleted_secret(secret_name)
                print(f"Secret is soft-deleted!")
                print(f"Deleted on: {deleted_secret.deleted_date}")
                print(f"Scheduled purge date: {deleted_secret.scheduled_purge_date}")
                print(f"Recovery ID: {deleted_secret.recovery_id}")
                
                # You can recover it
                print("You can recover this secret using client.begin_recover_deleted_secret()")
                
                # Or wait for auto-purge (if soft-delete retention period expires)
                
            except HttpResponseError as deleted_error:
                if deleted_error.status_code == 404:
                    print("Secret doesn't exist (not active, not deleted)")
                else:
                    print(f"Error checking deleted secret: {deleted_error.message}")
        else:
            raise


# Example 4: Comprehensive error handling wrapper
def safe_get_secret(secret_name, default=None):
    """Production-ready wrapper with full error handling"""
    try:
        secret = client.get_secret(secret_name)
        return secret.value
    
    except HttpResponseError as e:
        error_info = {
            'status_code': e.status_code,
            'message': e.message,
            'error_code': e.error.code if hasattr(e, 'error') and e.error else None,
            'secret_name': secret_name
        }
        
        if e.status_code == 403:
            # Log security event
            print(f"ACCESS DENIED: {error_info}")
            # Alert security team
            return default
            
        elif e.status_code == 404:
            # Secret doesn't exist
            print(f"SECRET NOT FOUND: {error_info}")
            return default
            
        elif e.status_code == 429:
            # Should implement retry (see example 2)
            print(f"THROTTLED: {error_info}")
            raise  # Propagate to trigger retry logic
            
        else:
            # Unexpected error
            print(f"UNEXPECTED ERROR: {error_info}")
            raise
    
    except Exception as e:
        print(f"Unexpected exception type: {type(e).__name__}: {e}")
        raise


# Example 5: Batch operations with error handling
def get_multiple_secrets(secret_names):
    """Get multiple secrets with individual error handling"""
    results = {}
    errors = {}
    
    for secret_name in secret_names:
        try:
            results[secret_name] = client.get_secret(secret_name).value
        except HttpResponseError as e:
            errors[secret_name] = {
                'status': e.status_code,
                'message': e.message
            }
    
    return results, errors
```

## Key Points:

**HttpResponseError attributes:**
- `e.status_code` - HTTP status code (403, 404, 429, etc.)
- `e.message` - Error message
- `e.error.code` - Azure-specific error code
- `e.response.headers` - HTTP headers (including Retry-After)

**Soft-deleted secrets:**
- `get_secret()` returns **404** for soft-deleted secrets
- Use `get_deleted_secret()` to access deleted secret metadata
- Use `begin_recover_deleted_secret()` to restore
- Purge occurs automatically after retention period

**Best practices:**
- Always implement retry logic for 429 errors
- Log 403 errors for security auditing
- Use DefaultAzureCredential for production
- Consider circuit breaker pattern for repeated failures

Would you like me to create a working example or help you implement this in your existing codebase?

## Verification

**Result:** ❌ FAIL

**Summary:** FAIL — no output files found

**Reasoning:**

No files were generated.

---

[← Back to Summary](../../../../../../summary.md)
