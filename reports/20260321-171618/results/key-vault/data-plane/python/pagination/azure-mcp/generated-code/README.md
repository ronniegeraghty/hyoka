# Azure Key Vault Secrets Pagination Demo

This script demonstrates how to properly handle pagination when listing secrets from an Azure Key Vault that contains hundreds of secrets.

## Overview

The script shows three different approaches to pagination using the Azure SDK for Python:

1. **Basic Iteration** - Automatic pagination handled by the SDK
2. **Page-by-Page Processing** - Explicit control using `by_page()`
3. **Continuation Tokens** - Resumable pagination for long-running operations

## Required Packages

Install the required packages using pip:

```bash
pip install -r requirements.txt
```

Or install individually:

```bash
pip install azure-keyvault-secrets azure-identity
```

### Package Details

- **azure-keyvault-secrets**: Azure Key Vault Secrets client library
- **azure-identity**: Azure authentication library (provides `DefaultAzureCredential`)

## Prerequisites

1. **Azure Key Vault**: You need an existing Azure Key Vault
2. **Authentication**: Configure one of the following:
   - Azure CLI: Run `az login`
   - Environment Variables: Set `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`, `AZURE_TENANT_ID`
   - Managed Identity: If running on Azure (VM, App Service, etc.)
3. **Permissions**: Your identity needs `secrets/list` permission on the Key Vault
4. **Environment Variable**: Set `VAULT_URL` to your Key Vault URL

## Usage

### Set Environment Variable

**Linux/macOS:**
```bash
export VAULT_URL="https://your-vault-name.vault.azure.net/"
```

**Windows (PowerShell):**
```powershell
$env:VAULT_URL="https://your-vault-name.vault.azure.net/"
```

**Windows (CMD):**
```cmd
set VAULT_URL=https://your-vault-name.vault.azure.net/
```

### Run the Script

```bash
python azure_keyvault_pagination.py
```

## How Azure Key Vault Handles Pagination

### ItemPaged Pattern

The Azure SDK returns an `ItemPaged[SecretProperties]` object from `list_properties_of_secrets()`. This provides:

- **Automatic iteration**: Iterate directly over the object to get all items
- **Page control**: Call `by_page()` to get pages explicitly
- **Continuation tokens**: Resume pagination from a specific point

### Key Points

1. **No Secret Values**: List operations return only metadata (name, enabled status, created date, etc.). Secret values are NOT included for performance and security reasons.

2. **Filtering**: The script filters for enabled secrets using `secret_property.enabled`

3. **Properties Available**:
   - `name`: Secret name
   - `enabled`: Whether the secret is enabled
   - `content_type`: Optional content type hint
   - `created_on`: Creation timestamp
   - `updated_on`: Last update timestamp
   - `tags`: Key-value metadata

4. **Getting Secret Values**: Use `client.get_secret(name)` to retrieve the actual secret value

### Example Code Snippet

```python
from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient

credential = DefaultAzureCredential()
client = SecretClient(vault_url="https://my-vault.vault.azure.net/", credential=credential)

# Basic iteration
for secret_property in client.list_properties_of_secrets():
    if secret_property.enabled:
        print(f"{secret_property.name}: {secret_property.created_on}")

# Page-by-page processing
pages = client.list_properties_of_secrets().by_page()
for page in pages:
    for secret_property in page:
        print(secret_property.name)
```

## Troubleshooting

### Authentication Errors

If you get authentication errors:
- Run `az login` if using Azure CLI
- Verify environment variables are set correctly
- Check that your identity has appropriate permissions

### Permission Errors

If you get "Forbidden" errors:
- Ensure your identity has `secrets/list` permission in the Key Vault
- Check the Key Vault's Access Policies or RBAC settings

### No Secrets Found

If no secrets are returned:
- Verify the Key Vault URL is correct
- Check that secrets exist in the vault
- Ensure secrets are enabled (script filters disabled secrets)

## References

- [Azure Key Vault Secrets Python SDK Documentation](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)
- [SecretClient API Reference](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient)
- [ItemPaged Documentation](https://learn.microsoft.com/en-us/python/api/azure-core/azure.core.paging.itempaged)
- [DefaultAzureCredential Documentation](https://learn.microsoft.com/en-us/python/api/azure-identity/azure.identity.defaultazurecredential)

## License

This code is provided as a demonstration based on Azure SDK documentation examples.
