# Azure Key Vault Secrets Pagination Example

This example demonstrates how the `azure-keyvault-secrets` SDK handles pagination when listing secrets from Azure Key Vault vaults that contain hundreds of secrets.

## Key Concepts Demonstrated

### 1. **ItemPaged Pattern**
The `list_properties_of_secrets()` method returns an `ItemPaged[SecretProperties]` object. This is a lazy iterator that doesn't fetch all results immediately but retrieves them as needed.

### 2. **Pagination with by_page()**
The `by_page()` method returns an iterator of pages, where each page is itself an iterator of items. This provides:
- **Memory efficiency**: Process one page at a time instead of loading all secrets into memory
- **Network efficiency**: Fetch secrets in batches rather than all at once
- **Control**: Track page numbers and implement custom processing logic per page

### 3. **SecretProperties Object**
When listing secrets, you get `SecretProperties` objects (not the actual secret values). Available attributes include:
- `name` - Secret name
- `enabled` - Whether the secret is enabled
- `content_type` - Optional content type indicator
- `created_on` - Creation timestamp (UTC)
- `updated_on` - Last update timestamp (UTC)
- `expires_on` - Expiration date
- `tags` - Custom metadata dictionary

## Required Packages

Install the required packages using pip:

```bash
pip install azure-keyvault-secrets azure-identity
```

Or use the requirements file:

```bash
pip install -r requirements.txt
```

### Package Details

- **azure-keyvault-secrets** (>= 4.8.0)
  - Provides `SecretClient` for Key Vault operations
  - Returns `ItemPaged` results for list operations
  
- **azure-identity** (>= 1.16.0)
  - Provides `DefaultAzureCredential` for authentication
  - Supports multiple authentication methods (managed identity, CLI, environment variables, etc.)

- **azure-core** (installed automatically)
  - Provides the `ItemPaged` base class used for pagination

## Setup

### 1. Set Environment Variable

```bash
export AZURE_KEY_VAULT_URL='https://your-vault-name.vault.azure.net/'
```

### 2. Authenticate with Azure

The script uses `DefaultAzureCredential`, which tries these methods in order:

1. **Environment variables** - `AZURE_CLIENT_ID`, `AZURE_TENANT_ID`, `AZURE_CLIENT_SECRET`
2. **Managed Identity** - If running on Azure (VM, App Service, Functions, etc.)
3. **Azure CLI** - If you've run `az login`
4. **Azure PowerShell** - If you've run `Connect-AzAccount`
5. **Interactive browser** - Opens browser for interactive login

For local development, the easiest method is Azure CLI:

```bash
az login
```

### 3. Grant Permissions

Ensure your identity has the "Key Vault Secrets User" role or equivalent permissions:

```bash
# Using Azure CLI
az keyvault set-policy --name your-vault-name \
  --upn your-email@domain.com \
  --secret-permissions get list
```

## Usage

Run the script:

```bash
python list_secrets_paginated.py
```

### Example Output

```
Connecting to Key Vault: https://my-vault.vault.azure.net/
================================================================================

Processing secrets by page...

--- Page 1 ---
  Secret: database-password
    Content Type: text/plain
    Created On:   2024-01-15 10:30:45 UTC
    Enabled:      True

  Secret: api-key-production
    Content Type: application/x-api-key
    Created On:   2024-02-20 14:22:10 UTC
    Enabled:      True

Secrets in this page: 25

--- Page 2 ---
  Secret: storage-connection-string
    Content Type: text/plain
    Created On:   2024-03-01 09:15:30 UTC
    Enabled:      True

Secrets in this page: 25

================================================================================
Summary:
  Total pages processed:    8
  Total secrets found:      187
  Enabled secrets shown:    180
  Disabled secrets skipped: 7
```

## How Pagination Works

### Behind the Scenes

1. **First call**: `list_properties_of_secrets()` creates an `ItemPaged` object but doesn't fetch any data yet
2. **by_page()**: Returns a page iterator 
3. **First page iteration**: Makes the first HTTP request to Azure Key Vault API
4. **Continuation tokens**: If more results exist, the response includes a continuation token
5. **Subsequent pages**: When you iterate to the next page, it uses the continuation token to fetch the next batch
6. **End of results**: When no continuation token is present, iteration stops

### Key Benefits for Large Vaults

For vaults with hundreds or thousands of secrets:

- **Memory efficient**: Only one page is in memory at a time
- **Responsive**: Start processing results immediately without waiting for all secrets
- **Scalable**: Works equally well with 10 or 10,000 secrets
- **Network efficient**: Reduces total data transfer through batching

## Alternative Approach

The script also demonstrates direct iteration without explicitly calling `by_page()`:

```python
secret_properties = client.list_properties_of_secrets()

# Pagination happens automatically
for secret_property in secret_properties:
    if secret_property.enabled:
        print(secret_property.name)
```

This is simpler but gives you less control over page-level processing. Use `by_page()` when you need:
- Page-level statistics or logging
- Custom page size handling
- Progress indicators for large operations
- Batch processing logic

## Filtering

The example filters to show only enabled secrets:

```python
for secret_property in page:
    if secret_property.enabled:  # Filter disabled secrets
        print(secret_property.name)
```

You can filter on any `SecretProperties` attribute:
- `enabled` - Active vs disabled
- `content_type` - Specific secret types
- `created_on` - Date range filters
- `tags` - Custom metadata filters

## Python Version

Requires Python 3.9 or later (Azure SDK requirement).

## References

- [Azure Key Vault Secrets SDK Documentation](https://learn.microsoft.com/python/api/overview/azure/keyvault-secrets-readme)
- [SecretClient API Reference](https://learn.microsoft.com/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient)
- [ItemPaged Documentation](https://learn.microsoft.com/python/api/azure-core/azure.core.paging.itempaged)
- [SecretProperties API Reference](https://learn.microsoft.com/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretproperties)
