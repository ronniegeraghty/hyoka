# Azure Key Vault Secrets Pagination Demo

This script demonstrates how the `azure-keyvault-secrets` SDK handles pagination for Azure Key Vaults containing hundreds of secrets.

## Features

- ✅ Uses `SecretClient` with `DefaultAzureCredential` for authentication
- ✅ Leverages the `ItemPaged` pattern for efficient iteration
- ✅ Processes secrets in pages using the `by_page()` method
- ✅ Displays name, content type, and created date for each secret
- ✅ Filters to show only enabled secrets
- ✅ Demonstrates continuation tokens for resumable pagination

## Prerequisites

1. **Python 3.9 or later**
2. **Azure Key Vault** with secrets configured
3. **Authentication** configured for `DefaultAzureCredential`:
   - Service Principal (environment variables)
   - Azure CLI (`az login`)
   - Managed Identity (when running in Azure)

## Installation

Install the required packages:

```bash
pip install -r requirements.txt
```

Or install individually:

```bash
pip install azure-keyvault-secrets azure-identity
```

## Configuration

Set the `VAULT_URL` environment variable to your Key Vault URL:

```bash
export VAULT_URL="https://your-vault-name.vault.azure.net/"
```

### Authentication Options

**Option 1: Service Principal (recommended for automation)**
```bash
export AZURE_CLIENT_ID="your-client-id"
export AZURE_CLIENT_SECRET="your-client-secret"
export AZURE_TENANT_ID="your-tenant-id"
```

**Option 2: Azure CLI (recommended for local development)**
```bash
az login
```

**Option 3: Managed Identity (automatic in Azure)**
No configuration needed when running in Azure with Managed Identity enabled.

## Permissions Required

Ensure your identity has the following Key Vault permissions:
- `secrets/list` - to list secret properties

You can grant this via Azure RBAC role:
- **Key Vault Secrets User** or **Key Vault Reader**

Or via Access Policies (legacy):
- List permission for secrets

## Usage

Run the script:

```bash
python keyvault_pagination_demo.py
```

## How Pagination Works

### ItemPaged Pattern

The `list_properties_of_secrets()` method returns an `ItemPaged[SecretProperties]` object:

```python
secret_properties_paged = client.list_properties_of_secrets()
```

### Two Ways to Iterate

**1. Direct iteration (automatic pagination):**
```python
for secret_property in secret_properties_paged:
    print(secret_property.name)
```

**2. Page-by-page iteration (explicit control):**
```python
for page in secret_properties_paged.by_page():
    for secret_property in page:
        print(secret_property.name)
```

### Why Use by_page()?

- **Better control** over pagination flow
- **Monitor progress** by tracking page numbers
- **Handle rate limits** by pausing between pages
- **Resume processing** using continuation tokens
- **Batch processing** - process page-size chunks of data

### Continuation Tokens

You can pause and resume pagination:

```python
pages = secret_properties_paged.by_page()
first_page = next(pages)
# Get continuation token from first_page (implementation detail)
# Resume later: by_page(continuation_token=token)
```

## Expected Output

```
Connecting to Key Vault: https://my-vault.vault.azure.net/
================================================================================

Processing secrets page by page...

--- Page 1 ---
  Secret Name: database-password
    Content Type: text/plain
    Created On: 2024-01-15 10:30:45 UTC
    Enabled: Yes

  Secret Name: api-key
    Content Type: application/json
    Created On: 2024-02-20 14:22:10 UTC
    Enabled: Yes

Secrets in this page: 25

--- Page 2 ---
  ...

================================================================================

Summary:
  Total pages processed: 8
  Total secrets found: 187
  Enabled secrets: 182
  Disabled secrets: 5

✓ Script completed successfully!
```

## Key Concepts

### SecretProperties vs KeyVaultSecret

- **SecretProperties**: Metadata only (returned by `list_properties_of_secrets()`)
  - Name, enabled status, created date, content type
  - Does NOT include the actual secret value
  
- **KeyVaultSecret**: Full secret including value (returned by `get_secret()`)
  - Contains both properties and the secret value
  - Requires additional API call per secret

### Performance Considerations

- **Listing is cheap**: Only metadata is transferred
- **Getting values is expensive**: Each `get_secret()` is a separate API call
- **Pagination helps**: Process large datasets without loading everything into memory
- **Page size**: Controlled by Azure (typically 25 items per page)

## Troubleshooting

**Error: VAULT_URL environment variable is not set**
- Set the environment variable with your vault URL

**Authentication errors**
- Verify `DefaultAzureCredential` is properly configured
- Check Azure CLI is logged in: `az account show`
- Verify service principal credentials if using environment variables

**Permission denied**
- Ensure your identity has `secrets/list` permission on the Key Vault
- Check Azure RBAC roles or Access Policies

**No secrets found**
- Verify the vault contains secrets
- Check if secrets are disabled (script filters to enabled only)

## References

- [Azure Key Vault Secrets Python SDK](https://learn.microsoft.com/python/api/overview/azure/keyvault-secrets-readme)
- [SecretClient Documentation](https://learn.microsoft.com/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient)
- [ItemPaged Documentation](https://learn.microsoft.com/python/api/azure-core/azure.core.paging.itempaged)
- [DefaultAzureCredential](https://learn.microsoft.com/python/api/azure-identity/azure.identity.defaultazurecredential)
