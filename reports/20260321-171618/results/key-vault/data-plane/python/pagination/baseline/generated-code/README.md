# Azure Key Vault Secrets Pagination Example

This example demonstrates how the Azure SDK for Python handles pagination when listing secrets in an Azure Key Vault with hundreds of secrets.

## Required Packages

```bash
pip install -r requirements.txt
```

Or manually:
```bash
pip install azure-keyvault-secrets azure-identity
```

## How Pagination Works in azure-keyvault-secrets

### ItemPaged Pattern

The `list_properties_of_secrets()` method returns an `ItemPaged[SecretProperties]` object. This is Azure's standard pagination pattern:

- **ItemPaged**: An iterator that automatically handles pagination
- **by_page()**: Returns an iterator of pages (each page is itself an iterator of items)
- **Lazy evaluation**: Pages are fetched from the server only when needed

### Key Points

1. **No secret values in list operations**: The `list_properties_of_secrets()` method only returns metadata (name, creation date, enabled status, etc.). To get actual secret values, you must call `get_secret(name)` for each secret individually.

2. **Two iteration approaches**:
   - **Explicit pagination** with `by_page()`: Gives you control over page boundaries
   - **Simple iteration**: Iterate directly over ItemPaged, pagination happens automatically

3. **Page size**: The Azure Key Vault service controls the page size. The client automatically fetches the next page when needed.

4. **Filtering**: The SDK returns all secrets; filtering (e.g., enabled vs disabled) must be done client-side.

### Properties Available on SecretProperties

From the official documentation, each `SecretProperties` object includes:

- `name` - The secret's name
- `enabled` - Whether the secret is enabled for use (bool)
- `content_type` - Arbitrary string indicating the type of the secret
- `created_on` - When the secret was created (datetime)
- `updated_on` - When the secret was last updated (datetime)
- `expires_on` - When the secret expires (datetime)
- `not_before` - The time before which the secret cannot be used (datetime)
- `id` - The secret's ID (full URI)
- `version` - The secret's version
- `vault_url` - URL of the vault containing the secret
- `tags` - Dictionary of application-specific metadata
- `managed` - Whether the secret's lifetime is managed by Key Vault
- `recoverable_days` - Days retained before permanent deletion
- `recovery_level` - The vault's deletion recovery level

## Usage

### Setup

1. Set your Key Vault URL:
```bash
export AZURE_KEYVAULT_URL='https://your-vault-name.vault.azure.net/'
```

2. Authenticate with Azure (DefaultAzureCredential supports multiple methods):
```bash
# Option 1: Azure CLI
az login

# Option 2: Set environment variables for service principal
export AZURE_CLIENT_ID='your-client-id'
export AZURE_TENANT_ID='your-tenant-id'
export AZURE_CLIENT_SECRET='your-client-secret'

# Option 3: Use Managed Identity (when running in Azure)
```

3. Ensure you have the "List" permission on secrets in the Key Vault.

### Run the Script

```bash
python list_keyvault_secrets_paginated.py
```

## Code Structure

The script includes two demonstration functions:

### 1. `list_secrets_with_pagination(vault_url)`

Shows explicit pagination using `by_page()`:

```python
secret_properties_paged = client.list_properties_of_secrets()

for page in secret_properties_paged.by_page():
    for secret_property in page:
        # Process each secret
        if secret_property.enabled:
            print(f"Secret: {secret_property.name}")
            print(f"Created: {secret_property.created_on}")
```

Benefits:
- See page boundaries
- Track number of pages
- Useful for progress reporting with large result sets

### 2. `list_secrets_simple_iteration(vault_url)`

Shows simple iteration (pagination happens automatically):

```python
for secret_property in client.list_properties_of_secrets():
    if secret_property.enabled:
        print(f"Secret: {secret_property.name}")
```

Benefits:
- Simpler code
- Still efficient (lazy loading)
- Good for most use cases

## Performance Considerations

1. **Network efficiency**: The SDK fetches secrets in batches (pages) automatically. This is much more efficient than individual requests for each secret.

2. **Memory efficiency**: Pages are loaded on-demand. The entire result set is NOT loaded into memory at once.

3. **Large vaults**: For vaults with hundreds of secrets:
   - Pagination is automatic and efficient
   - Consider using `by_page()` to show progress
   - Filter early (e.g., by enabled status) to reduce processing

4. **Secret values**: Remember that `list_properties_of_secrets()` does NOT return secret values. If you need the actual secrets, you must call `get_secret()` for each one:

```python
for secret_property in client.list_properties_of_secrets():
    if secret_property.enabled:
        # This makes an additional API call per secret
        secret = client.get_secret(secret_property.name)
        print(f"Secret value: {secret.value}")
```

## Authentication Methods

The `DefaultAzureCredential` attempts multiple authentication methods in order:

1. Environment variables (service principal)
2. Managed Identity (when running in Azure)
3. Azure CLI (`az login`)
4. Azure PowerShell
5. Interactive browser (fallback)

For production, use Managed Identity or service principal authentication.

## References

- [Azure Key Vault Secrets SDK Documentation](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)
- [SecretClient API Reference](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient)
- [ItemPaged API Reference](https://learn.microsoft.com/en-us/python/api/azure-core/azure.core.paging.itempaged)
- [DefaultAzureCredential Documentation](https://learn.microsoft.com/en-us/python/api/azure-identity/azure.identity.defaultazurecredential)
