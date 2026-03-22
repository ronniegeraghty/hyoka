# Azure Key Vault Secret Error Handling Guide

This repository demonstrates proper error handling patterns for Azure Key Vault secrets using the `azure-keyvault-secrets` Python SDK.

## Key Concepts

### HttpResponseError Structure

The primary exception type is `HttpResponseError`, which contains:

```python
from azure.core.exceptions import HttpResponseError

try:
    secret = client.get_secret("secret-name")
except HttpResponseError as e:
    # Access error details:
    status_code = e.status_code      # HTTP status code (403, 404, 429, etc.)
    message = e.message              # Human-readable error message
    reason = e.reason                # HTTP reason phrase
    response = e.response            # Full HTTP response object
    
    # Response headers (useful for Retry-After on 429)
    retry_after = e.response.headers.get('Retry-After')
```

### Common HTTP Status Codes

| Code | Meaning | Cause | Solution |
|------|---------|-------|----------|
| **403** | Forbidden | Missing RBAC role | Assign "Key Vault Secrets User" or "Key Vault Secrets Officer" role |
| **404** | Not Found | Secret doesn't exist or is soft-deleted | Check name or use `get_deleted_secret()` |
| **429** | Too Many Requests | Rate limit exceeded | Implement retry with exponential backoff |
| **401** | Unauthorized | Authentication failure | Check credential configuration |

### Soft-Deleted Secrets

When soft-delete is enabled (default):

1. **Deleting a secret** moves it to "deleted" state (not permanently removed)
2. **Getting a deleted secret** with `get_secret()` raises `ResourceNotFoundError` (404)
3. **The name is reserved** - you cannot create a new secret with the same name
4. **To access deleted secret**: Use `get_deleted_secret(name)`
5. **To recover**: Use `begin_recover_deleted_secret(name)`
6. **To permanently delete**: Use `purge_deleted_secret(name)` (requires "Key Vault Secrets Officer" role)

```python
from azure.core.exceptions import ResourceNotFoundError

try:
    secret = client.get_secret("deleted-secret")
except ResourceNotFoundError:
    # Check if it's soft-deleted
    try:
        deleted = client.get_deleted_secret("deleted-secret")
        print(f"Secret is soft-deleted, scheduled purge: {deleted.scheduled_purge_date}")
        
        # Recover it if needed
        poller = client.begin_recover_deleted_secret("deleted-secret")
        recovered = poller.result()  # Wait for recovery to complete
    except HttpResponseError as e:
        if e.status_code == 404:
            print("Secret truly doesn't exist")
```

## Error Handling Patterns

### Pattern 1: Check Status Code

```python
try:
    secret = client.get_secret(secret_name)
except HttpResponseError as e:
    if e.status_code == 403:
        print("Access denied - check RBAC roles")
    elif e.status_code == 404:
        print("Secret not found")
    elif e.status_code == 429:
        print("Rate limited")
```

### Pattern 2: Use Specific Exceptions

```python
from azure.core.exceptions import ResourceNotFoundError

try:
    secret = client.get_secret(secret_name)
except ResourceNotFoundError:
    print("Secret not found (404)")
except HttpResponseError as e:
    if e.status_code == 403:
        print("Access denied")
    elif e.status_code == 429:
        print("Rate limited")
```

### Pattern 3: Retry Logic for Throttling

```python
import time

def get_secret_with_retry(secret_name, max_retries=3):
    for attempt in range(max_retries):
        try:
            return client.get_secret(secret_name)
        except HttpResponseError as e:
            if e.status_code == 429:
                retry_after = e.response.headers.get('Retry-After', 5)
                time.sleep(int(retry_after))
            else:
                raise
    raise Exception("Max retries exceeded")
```

## Required RBAC Roles

- **Key Vault Secrets User**: Read-only access (`get_secret`)
- **Key Vault Secrets Officer**: Full access (create, read, update, delete, recover, purge)

## Installation

```bash
pip install azure-keyvault-secrets azure-identity
```

## Usage

```python
from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential

vault_url = "https://your-vault.vault.azure.net/"
credential = DefaultAzureCredential()
client = SecretClient(vault_url=vault_url, credential=credential)

# Use the error handling patterns from keyvault_error_handling.py
```

## See Also

- [keyvault_error_handling.py](keyvault_error_handling.py) - Complete code examples
- [Azure Key Vault Secrets SDK Documentation](https://learn.microsoft.com/python/api/azure-keyvault-secrets/)
- [Azure RBAC for Key Vault](https://learn.microsoft.com/azure/key-vault/general/rbac-guide)
