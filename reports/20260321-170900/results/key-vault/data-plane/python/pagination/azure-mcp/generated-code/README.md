# Azure Key Vault Secrets Pagination Demo

This project demonstrates how the Azure SDK for Python handles pagination when listing secrets from an Azure Key Vault with hundreds of secrets.

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

- **azure-keyvault-secrets** (>=4.8.0): Provides `SecretClient` for Key Vault operations
- **azure-identity** (>=1.15.0): Provides `DefaultAzureCredential` for authentication
- **azure-core**: (automatically installed) Provides `ItemPaged` pagination support

## How Azure SDK Handles Pagination

### ItemPaged Pattern

The Azure SDK uses the `ItemPaged` pattern for listing operations:

1. **`list_properties_of_secrets()`** returns an `ItemPaged[SecretProperties]` object
2. `ItemPaged` is an iterator that automatically handles pagination behind the scenes
3. Two ways to iterate:
   - **Simple iteration**: Iterate directly over items (pagination is automatic and transparent)
   - **Page-by-page**: Call `by_page()` to get explicit control over pages

### Key Characteristics

- **Lazy evaluation**: Results are fetched on-demand as you iterate
- **Automatic continuation**: The SDK handles continuation tokens internally
- **Memory efficient**: Only one page is loaded in memory at a time
- **No manual page size control**: Azure Key Vault determines the page size server-side

### Method Details

```python
# Returns ItemPaged[SecretProperties]
secret_properties = client.list_properties_of_secrets()

# Option 1: Direct iteration (simple, automatic pagination)
for secret in secret_properties:
    print(secret.name)

# Option 2: Page-by-page iteration (explicit control)
for page in secret_properties.by_page():
    for secret in page:
        print(secret.name)
```

## SecretProperties Attributes

When listing secrets, you get `SecretProperties` objects (NOT the secret values). Available attributes:

- **name**: Secret name
- **content_type**: Optional content type string
- **created_on**: datetime when created (UTC)
- **updated_on**: datetime when last updated (UTC)
- **enabled**: bool indicating if secret is enabled
- **expires_on**: Optional expiration datetime
- **not_before**: Optional datetime before which secret can't be used
- **tags**: Dictionary of custom tags
- **version**: Secret version ID
- **vault_url**: URL of the containing vault

**Note**: To get the actual secret value, you must call `client.get_secret(name)`

## Authentication

The script uses `DefaultAzureCredential`, which tries multiple authentication methods in order:

1. **Environment variables**: `AZURE_CLIENT_ID`, `AZURE_TENANT_ID`, `AZURE_CLIENT_SECRET`
2. **Managed Identity**: For Azure-hosted applications
3. **Azure CLI**: If logged in via `az login`
4. **Azure PowerShell**: If logged in via PowerShell
5. **Visual Studio Code**: If signed into Azure extension
6. **And more...**

### Setup Authentication

Easiest method for local development:

```bash
# Install Azure CLI
# https://docs.microsoft.com/en-us/cli/azure/install-azure-cli

# Login
az login

# Set your subscription (if you have multiple)
az account set --subscription "your-subscription-id"
```

## Usage

```bash
python list_keyvault_secrets_paginated.py https://your-vault-name.vault.azure.net/
```

Replace `your-vault-name` with your actual Key Vault name.

## Required Permissions

The script requires the following Key Vault permission:
- **secrets/list**: To list secret properties

### Grant Access

Using Azure CLI:

```bash
# Get your user's object ID
USER_ID=$(az ad signed-in-user show --query id -o tsv)

# Grant Secret List permission
az keyvault set-policy \
  --name your-vault-name \
  --object-id $USER_ID \
  --secret-permissions list
```

## Script Output

The script demonstrates page-by-page processing and shows:

- Page number being processed
- For each enabled secret:
  - Secret name
  - Content type
  - Created date
  - Enabled status
- Secrets per page count
- Summary statistics (total secrets, enabled/disabled counts, page count)

## Performance Considerations

### For Hundreds of Secrets

- **Page size**: Server-determined (typically 25-100 items per page)
- **Memory usage**: Only one page in memory at a time
- **Network calls**: One HTTP request per page
- **Total time**: Depends on number of pages and network latency

### Best Practices

1. **Use `by_page()` for progress tracking**: When processing hundreds of secrets, `by_page()` lets you track progress and implement checkpointing
2. **Filter early**: Apply filters (like `enabled` check) immediately to reduce processing
3. **Don't fetch secret values unnecessarily**: `list_properties_of_secrets()` is fast; `get_secret()` is slower
4. **Use connection pooling**: The SDK handles this automatically
5. **Close the client**: Always call `client.close()` or use context manager

## Example with Context Manager

```python
from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential

vault_url = "https://your-vault.vault.azure.net/"

with SecretClient(vault_url=vault_url, credential=DefaultAzureCredential()) as client:
    for page_num, page in enumerate(client.list_properties_of_secrets().by_page(), start=1):
        print(f"Processing page {page_num}")
        for secret in page:
            if secret.enabled:
                print(f"  - {secret.name}")
```

## Troubleshooting

### Common Issues

1. **Authentication failed**: Run `az login` or set environment variables
2. **Forbidden (403)**: Check Key Vault access policies - you need `secrets/list` permission
3. **Vault not found (404)**: Verify the vault URL is correct
4. **Timeout**: Large vaults may take time; this is normal for page-by-page processing

## References

- [Azure Key Vault Secrets SDK Documentation](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/)
- [SecretClient Class](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient)
- [ItemPaged Class](https://learn.microsoft.com/en-us/python/api/azure-core/azure.core.paging.itempaged)
- [SecretProperties Class](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretproperties)
- [DefaultAzureCredential](https://learn.microsoft.com/en-us/python/api/azure-identity/azure.identity.defaultazurecredential)
