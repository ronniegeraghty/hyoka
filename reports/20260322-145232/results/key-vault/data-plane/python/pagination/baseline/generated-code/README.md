# Azure Key Vault Secrets - Pagination Demo

This script demonstrates how the `azure-keyvault-secrets` SDK handles pagination when listing secrets in Azure Key Vault with hundreds of secrets.

## Key Concepts

### ItemPaged Pattern

The `list_properties_of_secrets()` method returns an `ItemPaged[SecretProperties]` object. This is Azure SDK's standard pagination pattern that:

- Lazily fetches pages from the service as you iterate
- Automatically handles continuation tokens
- Provides both simple iteration and explicit page-by-page control

### Pagination Methods

The script demonstrates three pagination approaches:

1. **Simple Iteration**: Let the SDK handle pagination automatically
   ```python
   for secret in client.list_properties_of_secrets():
       print(secret.name)
   ```

2. **Page-by-Page with by_page()**: Process secrets in batches
   ```python
   pages = client.list_properties_of_secrets().by_page()
   for page in pages:
       for secret in page:
           print(secret.name)
   ```

3. **Continuation Tokens**: Save position and resume later
   ```python
   pages = client.list_properties_of_secrets().by_page()
   first_page = next(pages)
   token = pages.continuation_token
   
   # Later, resume from token
   resumed_pages = client.list_properties_of_secrets().by_page(continuation_token=token)
   ```

## Installation

Install required packages:

```bash
pip install -r requirements.txt
```

Or install individually:

```bash
pip install azure-keyvault-secrets azure-identity
```

## Prerequisites

1. **Azure Key Vault**: You need an existing Azure Key Vault with secrets
2. **Authentication**: Configure one of the following for `DefaultAzureCredential`:
   - **Azure CLI**: Run `az login`
   - **Environment variables**: Set `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`, `AZURE_TENANT_ID`
   - **Managed Identity**: Available when running on Azure (VMs, App Service, etc.)
   - **Visual Studio Code**: Sign in to Azure in VS Code
   - **Other methods**: See [DefaultAzureCredential documentation](https://learn.microsoft.com/python/api/azure-identity/azure.identity.defaultazurecredential)

3. **Environment Variable**: Set your Key Vault URL
   ```bash
   export VAULT_URL="https://your-vault-name.vault.azure.net/"
   ```

## Usage

```bash
python list_secrets_paginated.py
```

## Script Features

The script demonstrates:

- ✅ Using `SecretClient` with `DefaultAzureCredential`
- ✅ Iterating through secrets using the `ItemPaged` pattern
- ✅ Processing secrets in pages using `by_page()`
- ✅ Printing name, content type, and created date of each secret
- ✅ Filtering to show only enabled secrets
- ✅ Working with continuation tokens for resumable pagination

## Understanding the Output

Each secret displays:
- **Name**: The secret's identifier in Key Vault
- **Content Type**: Optional metadata describing the secret format
- **Created**: When the secret was created (UTC)
- **Enabled**: Whether the secret is currently enabled

## Important Notes

1. **No Secret Values**: `list_properties_of_secrets()` returns only metadata, not actual secret values. To get values, call `get_secret(name)` for specific secrets.

2. **Permissions Required**: Your credential needs the `secrets/list` permission in the Key Vault's access policy or RBAC.

3. **Pagination Size**: Azure controls the page size automatically. You cannot configure it, but typically expect 25-100 items per page.

4. **Performance**: For vaults with hundreds of secrets, pagination prevents loading all data at once, improving performance and memory usage.

## References

- [Azure Key Vault Secrets SDK Documentation](https://learn.microsoft.com/python/api/overview/azure/keyvault-secrets-readme)
- [SecretClient API Reference](https://learn.microsoft.com/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient)
- [ItemPaged API Reference](https://learn.microsoft.com/python/api/azure-core/azure.core.paging.itempaged)
- [DefaultAzureCredential Documentation](https://learn.microsoft.com/python/api/azure-identity/azure.identity.defaultazurecredential)
