# Azure Key Vault Secrets Pagination Example

This script demonstrates how to efficiently list and process secrets from an Azure Key Vault containing hundreds of secrets using the Azure SDK for Python's **ItemPaged** pagination pattern.

## Key Features

- ✅ Uses `SecretClient` with `DefaultAzureCredential` for authentication
- ✅ Demonstrates the **ItemPaged** pattern for handling large result sets
- ✅ Processes secrets in pages using `by_page()` method
- ✅ Filters to show only enabled secrets
- ✅ Displays secret name, content type, and created date
- ✅ Provides page-level statistics

## Installation

Install the required packages using pip:

```bash
pip install -r requirements.txt
```

Or install packages directly:

```bash
pip install azure-keyvault-secrets azure-identity
```

### Required Packages

- **azure-keyvault-secrets** (>=4.7.0) - Azure Key Vault Secrets client library
- **azure-identity** (>=1.12.0) - Azure authentication library with DefaultAzureCredential

## Prerequisites

1. **Azure Key Vault**: You need an existing Azure Key Vault with secrets
2. **Authentication**: Configure one of the following authentication methods:
   - Azure CLI: Run `az login`
   - Environment variables: Set `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`, `AZURE_TENANT_ID`
   - Managed Identity: If running on Azure resources (VM, App Service, etc.)
   - Interactive browser authentication (fallback)

3. **Permissions**: Your identity needs the following Key Vault permissions:
   - `secrets/list` - Required to list secret properties

## Usage

Set the `VAULT_URL` environment variable to your Key Vault URL:

```bash
export VAULT_URL="https://your-vault-name.vault.azure.net/"
python azure_keyvault_pagination.py
```

Or on Windows:

```cmd
set VAULT_URL=https://your-vault-name.vault.azure.net/
python azure_keyvault_pagination.py
```

## How Pagination Works

### ItemPaged Pattern

The Azure SDK uses the **ItemPaged** pattern to handle large result sets efficiently:

```python
# Returns an ItemPaged[SecretProperties] object
secret_properties = client.list_properties_of_secrets()

# Process page by page
for page in secret_properties.by_page():
    for secret_property in page:
        # Process each secret in the page
        print(secret_property.name)
```

### Key Concepts

1. **ItemPaged Object**: `list_properties_of_secrets()` returns an `ItemPaged` object that handles pagination internally
   
2. **by_page() Method**: Converts the item iterator into a page iterator, giving you control over page-level processing

3. **Lazy Loading**: Secrets are fetched from the server only when needed, not all at once

4. **No Value Retrieval**: `list_properties_of_secrets()` returns only metadata, not secret values. Use `client.get_secret(name)` to retrieve actual values

### Benefits for Large Vaults

- **Memory Efficiency**: Only one page of results in memory at a time
- **Network Efficiency**: Fetches data in batches, not all at once
- **Progress Tracking**: Process pages individually to show progress
- **Error Recovery**: Can resume from a specific page if an error occurs

## Example Output

```
Connecting to Key Vault: https://my-vault.vault.azure.net/
================================================================================

Processing secrets by page:

--- Page 1 ---

  Secret Name: database-password
  Content Type: text/plain
  Created On: 2024-01-15 10:30:45 UTC

  Secret Name: api-key
  Content Type: application/json
  Created On: 2024-01-16 14:22:31 UTC

Secrets in this page: 50

--- Page 2 ---
...
================================================================================

Summary:
  Total pages processed: 5
  Total secrets found: 237
  Enabled secrets: 223
  Disabled secrets: 14
```

## Understanding SecretProperties

The script accesses the following properties from each `SecretProperties` object:

- **name**: The secret's name
- **content_type**: Optional field describing the secret type
- **created_on**: DateTime when the secret was created (UTC)
- **enabled**: Boolean indicating if the secret is active
- **version**: The secret's version ID
- **updated_on**: DateTime when the secret was last updated
- **expires_on**: Optional expiration date
- **tags**: Dictionary of custom metadata

## References

- [Azure Key Vault Secrets SDK Documentation](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)
- [SecretClient API Reference](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient)
- [ItemPaged Pattern Documentation](https://learn.microsoft.com/en-us/python/api/azure-core/azure.core.paging.itempaged)
- [DefaultAzureCredential Documentation](https://learn.microsoft.com/en-us/python/api/azure-identity/azure.identity.defaultazurecredential)

## Troubleshooting

### Authentication Errors

If you get authentication errors, ensure you're logged in with Azure CLI:

```bash
az login
```

### Permission Denied

If you get permission errors, ensure your identity has the `secrets/list` permission on the Key Vault. You can grant this using:

```bash
az keyvault set-policy --name your-vault-name \
  --upn your-email@domain.com \
  --secret-permissions list
```

### Connection Issues

Verify your vault URL is correct and accessible:

```bash
# Test connectivity
curl https://your-vault-name.vault.azure.net/
```
