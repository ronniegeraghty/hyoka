# Azure Key Vault Secrets - Error Handling Guide

## Overview

This guide demonstrates proper error handling patterns when working with Azure Key Vault secrets using the `azure-keyvault-secrets` Python SDK.

## Common Error Scenarios

### 1. **403 Forbidden - Access Denied**

Occurs when your identity lacks the required RBAC role.

```python
from azure.core.exceptions import HttpResponseError

try:
    secret = secret_client.get_secret("my-secret")
except HttpResponseError as e:
    if e.status_code == 403:
        print(f"Access denied: {e.message}")
        print("Required role: 'Key Vault Secrets User' or 'Key Vault Secrets Officer'")
```

**Required Azure RBAC roles:**
- `Key Vault Secrets User` - Read-only access to secrets
- `Key Vault Secrets Officer` - Full access to secrets

**How to grant access:**
```bash
az role assignment create \
  --role "Key Vault Secrets User" \
  --assignee <user-principal-id-or-email> \
  --scope /subscriptions/<subscription-id>/resourceGroups/<resource-group>/providers/Microsoft.KeyVault/vaults/<keyvault-name>
```

### 2. **404 Not Found - Secret Doesn't Exist**

Occurs when the secret doesn't exist or is soft-deleted.

```python
try:
    secret = secret_client.get_secret("non-existent")
except HttpResponseError as e:
    if e.status_code == 404:
        print(f"Secret not found: {e.message}")
```

### 3. **429 Too Many Requests - Throttling**

Occurs when you exceed Key Vault's rate limits.

```python
try:
    secret = secret_client.get_secret("my-secret")
except HttpResponseError as e:
    if e.status_code == 429:
        # Get retry delay from response header
        retry_after = e.response.headers.get('Retry-After')
        print(f"Throttled. Retry after {retry_after} seconds")
        time.sleep(int(retry_after))
```

## Inspecting Error Details

The `HttpResponseError` exception provides several useful attributes:

```python
try:
    secret = secret_client.get_secret("my-secret")
except HttpResponseError as e:
    # HTTP status code (403, 404, 429, etc.)
    status_code = e.status_code
    
    # Human-readable error message
    error_message = e.message
    
    # Azure-specific error code (if available)
    if hasattr(e, 'error') and e.error:
        error_code = e.error.code
        detailed_message = e.error.message
    
    # Response headers (for Retry-After, etc.)
    if hasattr(e, 'response') and e.response:
        retry_after = e.response.headers.get('Retry-After')
```

## Soft-Deleted Secrets

When you delete a secret with soft-delete enabled (default):

1. **`get_secret()` returns 404** - The secret appears to not exist
2. **The secret is recoverable** for the retention period (90 days by default)
3. **Use `get_deleted_secret()`** to check if it's soft-deleted

```python
try:
    secret = secret_client.get_secret("my-secret")
except HttpResponseError as e:
    if e.status_code == 404:
        # Check if soft-deleted
        try:
            deleted = secret_client.get_deleted_secret("my-secret")
            print(f"Secret is soft-deleted")
            print(f"Deleted on: {deleted.deleted_date}")
            print(f"Scheduled purge: {deleted.scheduled_purge_date}")
            
            # Recover it
            recover_operation = secret_client.begin_recover_deleted_secret("my-secret")
            recovered_secret = recover_operation.result()
            
        except HttpResponseError:
            print("Secret does not exist and is not soft-deleted")
```

## Production-Ready Pattern with Retry Logic

```python
def get_secret_with_retry(secret_client, secret_name, max_retries=3):
    """Get secret with exponential backoff for throttling."""
    retry_count = 0
    base_delay = 1
    
    while retry_count < max_retries:
        try:
            return secret_client.get_secret(secret_name)
        
        except HttpResponseError as e:
            if e.status_code == 429:
                # Throttling - retry with backoff
                retry_count += 1
                retry_after = e.response.headers.get('Retry-After')
                delay = int(retry_after) if retry_after else base_delay * (2 ** (retry_count - 1))
                
                if retry_count < max_retries:
                    time.sleep(delay)
                else:
                    raise
                    
            elif e.status_code in [403, 404]:
                # Don't retry permission or not found errors
                raise
            else:
                # Other errors - retry with backoff
                retry_count += 1
                if retry_count < max_retries:
                    time.sleep(base_delay * (2 ** (retry_count - 1)))
                else:
                    raise
    
    raise Exception(f"Failed after {max_retries} retries")
```

## Key Vault Service Limits

Be aware of Azure Key Vault throttling limits:

- **GET requests**: 2,000 per 10 seconds per vault
- **All requests**: 2,000 per 10 seconds per vault (standard tier)
- **Premium tier**: Higher limits available

When you exceed these limits, you'll receive a 429 response with a `Retry-After` header indicating when to retry.

## Best Practices

1. ✅ **Always handle specific status codes** (403, 404, 429)
2. ✅ **Implement exponential backoff** for 429 errors
3. ✅ **Check for soft-deleted secrets** when you get 404
4. ✅ **Don't retry 403 errors** - they require permission changes
5. ✅ **Use the `Retry-After` header** when throttled
6. ✅ **Cache secret values** to reduce API calls
7. ✅ **Log errors with context** for debugging

## Dependencies

```bash
pip install azure-keyvault-secrets azure-identity
```

## Example Usage

```python
from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient
from azure.core.exceptions import HttpResponseError

# Initialize client
vault_url = "https://your-keyvault.vault.azure.net/"
credential = DefaultAzureCredential()
client = SecretClient(vault_url=vault_url, credential=credential)

# Get secret with error handling
try:
    secret = client.get_secret("database-password")
    print(f"Secret value: {secret.value}")
    
except HttpResponseError as e:
    if e.status_code == 403:
        print("Access denied - check RBAC permissions")
    elif e.status_code == 404:
        print("Secret not found")
    elif e.status_code == 429:
        print("Rate limited - implement retry logic")
    else:
        print(f"Error {e.status_code}: {e.message}")
```

## Additional Resources

- [Azure Key Vault secrets client library for Python](https://learn.microsoft.com/python/api/overview/azure/keyvault-secrets-readme)
- [Azure Key Vault service limits](https://learn.microsoft.com/azure/key-vault/general/service-limits)
- [Azure RBAC for Key Vault](https://learn.microsoft.com/azure/key-vault/general/rbac-guide)
