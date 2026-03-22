# Azure Key Vault Secrets Pagination Example

This example demonstrates how the Azure SDK for Python handles pagination when listing secrets from a Key Vault containing hundreds of secrets.

## Installation

Install the required packages:

```bash
pip install -r requirements.txt
```

## Required Packages

- **azure-keyvault-secrets** (>=4.7.0): Core SDK for Key Vault secrets operations
- **azure-identity** (>=1.15.0): Provides DefaultAzureCredential for authentication
- **azure-core** (>=1.29.0): Provides the ItemPaged pagination pattern

## Authentication

The script uses `DefaultAzureCredential`, which automatically tries multiple authentication methods in order:

1. Environment variables (AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET)
2. Managed Identity (when running in Azure)
3. Azure CLI credentials (`az login`)
4. Azure PowerShell credentials
5. Interactive browser authentication

Ensure you're authenticated via one of these methods before running the script.

## Usage

```bash
python list_key_vault_secrets.py https://your-vault-name.vault.azure.net/
```

## How Azure SDK Pagination Works

### ItemPaged Pattern

The Azure SDK uses the **ItemPaged** pattern (from `azure.core.paging`) for paginated results. When you call `list_properties_of_secrets()`, it returns an `ItemPaged[SecretProperties]` object.

### Two Ways to Iterate

#### 1. Page-by-Page Iteration (Recommended for large datasets)

```python
secret_properties = client.list_properties_of_secrets()

for page in secret_properties.by_page():
    # Each page is an iterator of SecretProperties
    for secret_prop in page:
        print(secret_prop.name)
```

**Benefits:**
- Full control over pagination
- Can track page count
- Can process pages differently
- Better for monitoring progress with large datasets

#### 2. Direct Iteration (Simpler)

```python
secret_properties = client.list_properties_of_secrets()

for secret_prop in secret_properties:
    # Pages fetched automatically as needed
    print(secret_prop.name)
```

**Benefits:**
- Simpler code
- Automatic page management
- Good for straightforward iteration

### Key Concepts

1. **Lazy Loading**: Pages are fetched from the server on-demand, not all at once
2. **No Secret Values**: `list_properties_of_secrets()` returns metadata only, not actual secret values (requires secrets/list permission)
3. **SecretProperties Object**: Contains attributes like name, content_type, created_on, enabled, etc.
4. **Filtering**: You can filter results in your code (e.g., only enabled secrets)

### SecretProperties Attributes

The script demonstrates accessing these attributes:
- `name`: The secret's name
- `content_type`: An arbitrary string indicating the type
- `created_on`: When the secret was created (UTC datetime)
- `enabled`: Whether the secret is enabled for use (bool)
- Other available attributes: `expires_on`, `not_before`, `tags`, `version`, `updated_on`

## Permissions Required

- **secrets/list**: Required to list secret properties

## Performance Considerations

- The SDK handles pagination automatically
- Pages are fetched incrementally, reducing memory usage
- For vaults with hundreds of secrets, page-by-page iteration provides better observability
- Each page typically contains multiple secrets (server-determined page size)

## References

- [SecretClient Documentation](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient)
- [ItemPaged Documentation](https://learn.microsoft.com/en-us/python/api/azure-core/azure.core.paging.itempaged)
- [SecretProperties Documentation](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretproperties)
