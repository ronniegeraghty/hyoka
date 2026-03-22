# Azure Key Vault Secrets Pagination Example

This example demonstrates how to handle pagination when listing secrets from an Azure Key Vault containing hundreds of secrets using the Azure SDK for Python.

## Overview

The script showcases two approaches for handling pagination with the `ItemPaged` pattern:

1. **Explicit pagination** using `by_page()` - Provides control over page boundaries
2. **Automatic pagination** - Simpler iteration that handles pagination internally

## Required Packages

```bash
pip install -r requirements.txt
```

Or install directly:
```bash
pip install azure-keyvault-secrets azure-identity
```

**Packages:**
- `azure-keyvault-secrets` - Azure Key Vault Secrets client library
- `azure-identity` - Azure authentication library (provides DefaultAzureCredential)

## Authentication

The script uses `DefaultAzureCredential` which automatically attempts multiple authentication methods in order:

1. Environment variables
2. Managed Identity
3. Azure CLI credentials
4. Azure PowerShell credentials
5. Interactive browser authentication

For local development, the easiest method is Azure CLI:
```bash
az login
```

## Usage

Set the `VAULT_URL` environment variable to your Key Vault URL:

```bash
export VAULT_URL="https://my-key-vault.vault.azure.net/"
python list_secrets_pagination.py
```

## How Pagination Works

### ItemPaged Pattern

The `list_properties_of_secrets()` method returns an `ItemPaged[SecretProperties]` object:

```python
secret_properties = client.list_properties_of_secrets()
```

### Method 1: Explicit Pagination with by_page()

```python
page_iterator = secret_properties.by_page()

for page in page_iterator:
    for secret_property in page:
        # Process each secret
        print(secret_property.name)
```

**Benefits:**
- Control over page boundaries
- Can track page numbers and items per page
- Useful for implementing custom pagination logic

### Method 2: Automatic Pagination

```python
for secret_property in secret_properties:
    # ItemPaged handles pagination automatically
    print(secret_property.name)
```

**Benefits:**
- Simpler code
- No need to manage pages manually
- Good for straightforward iteration

## SecretProperties Attributes

The script accesses these properties (as documented in the Azure SDK):

- `name` - The secret's name
- `content_type` - Arbitrary string indicating the type of secret
- `created_on` - When the secret was created (datetime, UTC)
- `enabled` - Whether the secret is enabled for use (bool)
- `updated_on` - When the secret was last updated (datetime, UTC)
- `expires_on` - When the secret expires (datetime, UTC)
- `version` - The secret's version
- `id` - The secret's full ID
- `tags` - Application-specific metadata (dict)

**Note:** `list_properties_of_secrets()` returns metadata only. Use `client.get_secret(name)` to retrieve actual secret values.

## Key Features Demonstrated

1. **SecretClient initialization** with DefaultAzureCredential
2. **ItemPaged iteration** - Both explicit and automatic
3. **Page-by-page processing** using `by_page()`
4. **Filtering** - Shows only enabled secrets
5. **Property access** - Name, content type, created date, enabled status
6. **Resource cleanup** - Proper client and credential closure

## Permissions Required

The service principal or user must have the following Key Vault permission:
- `secrets/list` - Required for `list_properties_of_secrets()`

## Performance Considerations

- **Pagination reduces memory usage** - Secrets are fetched in batches rather than all at once
- **Azure Key Vault pages** - The service determines optimal page size
- **Network efficiency** - Reduces number of API calls for large vaults
- **No secret values** - `list_properties_of_secrets()` only retrieves metadata, not values

## Additional Resources

- [Azure SDK for Python - Key Vault Secrets Documentation](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)
- [SecretClient API Reference](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient)
- [ItemPaged Documentation](https://learn.microsoft.com/en-us/python/api/azure-core/azure.core.paging.itempaged)
- [SecretProperties API Reference](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretproperties)

## Example Output

```
================================================================================
METHOD 1: Using by_page() for explicit pagination control
================================================================================
Listing secrets from: https://my-vault.vault.azure.net/

================================================================================

--- Page 1 ---

Secret Name: database-password
  Content Type: text/plain
  Created On: 2024-01-15 10:30:00
  Enabled: True

Secret Name: api-key
  Content Type: application/json
  Created On: 2024-01-20 14:22:00
  Enabled: True

Secrets in page 1: 25

--- Page 2 ---
...

Summary:
  Total pages processed: 4
  Total secrets found: 100
  Enabled secrets: 95
  Disabled secrets: 5
```
