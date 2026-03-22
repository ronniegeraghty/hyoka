# Azure Key Vault Secrets Pagination Example

This Python script demonstrates how to efficiently list and process secrets from an Azure Key Vault that contains hundreds of secrets using the Azure SDK for Python's ItemPaged pagination pattern.

## Features

- ✅ Uses `SecretClient` with `DefaultAzureCredential` for authentication
- ✅ Iterates through secrets using the `ItemPaged` pattern
- ✅ Processes secrets in pages using `by_page()` method
- ✅ Displays name, content type, and created date for each secret
- ✅ Filters to show only enabled secrets

## Prerequisites

- Python 3.9 or later
- An Azure subscription
- An existing Azure Key Vault with secrets
- Appropriate permissions (secrets/list and secrets/get)

## Installation

Install the required packages:

```bash
pip install -r requirements.txt
```

This installs:
- `azure-keyvault-secrets` - Azure Key Vault Secrets client library
- `azure-identity` - Azure authentication library (provides DefaultAzureCredential)

## Authentication

This script uses `DefaultAzureCredential`, which attempts authentication through multiple methods in order:

1. Environment variables (AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET)
2. Managed Identity (when running on Azure)
3. Azure CLI (`az login`)
4. Azure PowerShell
5. Interactive browser

For local development, the easiest method is Azure CLI:

```bash
az login
```

## Usage

Set the Key Vault URL environment variable:

```bash
export AZURE_KEY_VAULT_URL='https://your-vault-name.vault.azure.net/'
```

Run the script:

```bash
python list_secrets_paginated.py
```

## How Pagination Works

The script demonstrates the Azure SDK's pagination pattern:

1. **ItemPaged Object**: `list_properties_of_secrets()` returns an `ItemPaged[SecretProperties]` object
2. **by_page() Method**: Call `by_page()` to get an iterator of pages instead of individual items
3. **Page Iteration**: Each page is itself an iterator of `SecretProperties` objects
4. **Filtering**: Secrets are filtered to show only enabled ones (`enabled=True`)

### Code Structure

```python
# Get ItemPaged object
secret_properties_paged = client.list_properties_of_secrets()

# Get page iterator
page_iterator = secret_properties_paged.by_page()

# Iterate through pages
for page in page_iterator:
    # Iterate through items in each page
    for secret_property in page:
        if secret_property.enabled:
            # Process enabled secret
            print(secret_property.name)
            print(secret_property.content_type)
            print(secret_property.created_on)
```

## Output

The script outputs:
- Each page of secrets with details (name, content type, created date, enabled status)
- Summary statistics (total pages, total secrets, enabled/disabled counts)

Example output:
```
Listing secrets from vault: https://my-vault.vault.azure.net/

================================================================================

--- Page 1 ---

Secret Name:    database-password
Content Type:   text/plain
Created Date:   2024-01-15 10:30:45 UTC
Enabled:        True
--------------------------------------------------------------------------------
Secret Name:    api-key
Content Type:   Not specified
Created Date:   2024-01-16 14:22:10 UTC
Enabled:        True
--------------------------------------------------------------------------------

Secrets in page 1: 25

--- Page 2 ---

...

================================================================================

Summary:
  Total pages processed:     4
  Total secrets found:       100
  Enabled secrets (shown):   95
  Disabled secrets (hidden): 5
```

## Important Notes

1. **list_properties_of_secrets()** does NOT retrieve secret values - only metadata
2. To get secret values, use `client.get_secret(name)` for specific secrets
3. The script filters at the client side - all secrets are retrieved from Key Vault
4. Pagination is automatic - the SDK handles continuation tokens internally

## References

- [Azure Key Vault Secrets Python SDK](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)
- [SecretClient API Reference](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient)
- [ItemPaged API Reference](https://learn.microsoft.com/en-us/python/api/azure-core/azure.core.paging.itempaged)
- [SecretProperties API Reference](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretproperties)
