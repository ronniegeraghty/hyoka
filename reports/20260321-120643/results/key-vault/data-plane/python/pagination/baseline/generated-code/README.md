# Azure Key Vault Secrets - Pagination Example

This script demonstrates how to handle pagination when listing secrets in an Azure Key Vault that contains hundreds of secrets.

## Key Concepts

### ItemPaged Pattern

The Azure SDK for Python uses the `ItemPaged[T]` pattern for list operations. When you call `list_properties_of_secrets()`, it returns an `ItemPaged[SecretProperties]` object that:

1. **Automatically handles pagination** when you iterate over it directly
2. **Supports manual pagination** via the `by_page()` method
3. **Lazily fetches data** - pages are only retrieved as needed

### Pagination Methods

#### Method 1: Automatic Pagination (Simple Iteration)
```python
secret_properties = client.list_properties_of_secrets()
for secret in secret_properties:
    print(secret.name)
```
The SDK automatically fetches additional pages as you iterate.

#### Method 2: Manual Pagination (by_page())
```python
secret_pages = client.list_properties_of_secrets().by_page()
for page in secret_pages:
    for secret in page:
        print(secret.name)
```
This gives you control over page-by-page processing.

## Installation

Install the required packages:

```bash
pip install -r requirements.txt
```

## Authentication

The script uses `DefaultAzureCredential` which attempts authentication via:
1. Environment variables (AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET)
2. Managed Identity (when running in Azure)
3. Azure CLI credentials
4. Azure PowerShell credentials
5. Interactive browser authentication

For local development, the easiest method is Azure CLI:
```bash
az login
```

## Usage

Set the vault URL environment variable:
```bash
export VAULT_URL="https://your-vault-name.vault.azure.net/"
```

Run the script:
```bash
python list_secrets_pagination.py
```

## What the Script Demonstrates

1. **Simple iteration**: Automatic pagination handling
2. **Manual pagination**: Using `by_page()` for page-by-page processing
3. **Filtering**: Processing only enabled secrets with pagination
4. **Secret metadata**: Displaying name, content type, and created date

## Important Notes

- `list_properties_of_secrets()` returns **metadata only**, not secret values
- To get secret values, use `get_secret(name)` for each secret
- The default page size is controlled by the Azure Key Vault service
- Pagination is handled efficiently - pages are fetched on demand

## References

- [SecretClient API Documentation](https://learn.microsoft.com/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient)
- [ItemPaged API Documentation](https://learn.microsoft.com/python/api/azure-core/azure.core.paging.itempaged)
- [Azure Key Vault Secrets README](https://learn.microsoft.com/python/api/overview/azure/keyvault-secrets-readme)
- [DefaultAzureCredential Documentation](https://learn.microsoft.com/python/api/azure-identity/azure.identity.defaultazurecredential)
