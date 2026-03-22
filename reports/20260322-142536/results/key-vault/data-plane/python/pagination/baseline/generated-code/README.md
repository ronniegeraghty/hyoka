# Azure Key Vault Secrets Pagination Example

This example demonstrates how to list all secrets in an Azure Key Vault using the ItemPaged pagination pattern from the Azure SDK for Python.

## Required Packages

```bash
pip install azure-keyvault-secrets azure-identity
```

Or using the requirements file:

```bash
pip install -r requirements.txt
```

## Authentication

The script uses `DefaultAzureCredential`, which supports multiple authentication methods:

1. **Environment variables** - Set these variables:
   - `AZURE_TENANT_ID`
   - `AZURE_CLIENT_ID`
   - `AZURE_CLIENT_SECRET`

2. **Managed Identity** - When running on Azure (App Service, VM, etc.)

3. **Azure CLI** - Run `az login` first

4. **Azure PowerShell** - Run `Connect-AzAccount` first

5. **Interactive browser** - Falls back to interactive login if needed

## Usage

Set your Key Vault URL:

```bash
export AZURE_KEYVAULT_URL='https://your-vault-name.vault.azure.net/'
```

Run the script:

```bash
python list_key_vault_secrets_paginated.py
```

## Key Concepts

### ItemPaged Pattern

The `list_properties_of_secrets()` method returns an `ItemPaged[SecretProperties]` object:

- **ItemPaged** is an iterable that automatically handles pagination
- Each item is a `SecretProperties` object with metadata (not the secret value)
- The SDK makes multiple API calls as needed when iterating

### Three Pagination Methods

#### 1. Simple Iteration (Recommended)
```python
secret_properties = client.list_properties_of_secrets()
for secret_property in secret_properties:
    print(secret_property.name)
```
- Easiest to use
- Pagination handled automatically
- Best for most use cases

#### 2. Explicit Page Processing
```python
pages = client.list_properties_of_secrets().by_page()
for page in pages:
    for secret_property in page:
        print(secret_property.name)
```
- Process secrets page by page
- Useful for progress tracking
- Better control over API requests

#### 3. Continuation Tokens
```python
pages = client.list_properties_of_secrets().by_page(continuation_token=saved_token)
```
- Resume pagination from a specific point
- Useful for long-running operations
- Enable resumable processing

## SecretProperties Attributes

When listing secrets, you get `SecretProperties` objects with these attributes:

- `name` - Secret name
- `enabled` - Whether the secret is enabled
- `content_type` - User-defined content type (optional)
- `created_on` - Creation timestamp (datetime)
- `updated_on` - Last update timestamp (datetime)
- `expires_on` - Expiration timestamp (datetime, optional)
- `not_before` - Valid-from timestamp (datetime, optional)
- `tags` - Dictionary of tags
- `vault_url` - Key Vault URL
- `version` - Secret version ID

**Note:** `list_properties_of_secrets()` does NOT return secret values. Use `get_secret(name)` to retrieve the actual secret value.

## Permissions Required

Your Azure identity needs the following Key Vault permission:
- **Secrets: List** - To list secret properties

## Pagination Behavior

- Azure Key Vault returns secrets in pages (default: 25 secrets per page)
- The `ItemPaged` object automatically fetches additional pages as you iterate
- For very large vaults (hundreds of secrets), this is more efficient than loading all at once
- Each page requires a separate API call to Key Vault

## Filtering

The script demonstrates filtering for enabled secrets:

```python
for secret_property in secret_properties:
    if secret_property.enabled:
        # Process only enabled secrets
        print(secret_property.name)
```

## References

- [Azure Key Vault Secrets Documentation](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)
- [SecretClient API Reference](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient)
- [ItemPaged API Reference](https://learn.microsoft.com/en-us/python/api/azure-core/azure.core.paging.itempaged)
- [SecretProperties API Reference](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretproperties)
