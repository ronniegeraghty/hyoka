# Azure Key Vault Secrets Pagination Example

This example demonstrates how to handle pagination when listing secrets from an Azure Key Vault that contains hundreds of secrets.

## Key Concepts

### ItemPaged Pattern
The Azure SDK for Python uses the `ItemPaged` pattern for paginated operations. The `list_properties_of_secrets()` method returns an `ItemPaged[SecretProperties]` object.

### Two Ways to Iterate

1. **Page-by-page iteration** (using `by_page()`):
   - Gives you control over each page of results
   - Useful for batch processing or progress tracking
   - Each page is an iterator of items

2. **Direct iteration**:
   - Iterate directly over the ItemPaged object
   - Pagination happens automatically behind the scenes
   - Simpler code when you don't need page-level control

## Installation

```bash
pip install -r requirements.txt
```

Or install packages individually:
```bash
pip install azure-keyvault-secrets azure-identity
```

## Required Packages

- **azure-keyvault-secrets**: Azure Key Vault Secrets client library
- **azure-identity**: Azure authentication library (provides DefaultAzureCredential)

## Usage

1. Set the Key Vault URL:
```bash
export AZURE_KEYVAULT_URL='https://your-vault-name.vault.azure.net/'
```

2. Ensure you're authenticated:
```bash
az login
```

3. Run the script:
```bash
python list_keyvault_secrets_paginated.py
```

## Authentication

The script uses `DefaultAzureCredential` which attempts authentication through multiple methods in order:
1. Environment variables
2. Managed Identity (if running in Azure)
3. Azure CLI credentials
4. Interactive browser (if available)

## Required Permissions

Your Azure identity needs the following Key Vault permission:
- **secrets/list**: To list secret properties

## What the Script Does

1. Creates a `SecretClient` with `DefaultAzureCredential`
2. Calls `list_properties_of_secrets()` to get an `ItemPaged[SecretProperties]` iterator
3. Uses `by_page()` to process secrets page by page
4. Filters to show only enabled secrets
5. Prints name, content type, and created date for each enabled secret
6. Displays pagination statistics (pages, total secrets, enabled/disabled counts)

## Key Vault Pagination Behavior

- The Azure Key Vault service returns results in pages
- Page size is determined by the service (typically 25 items per page)
- The `ItemPaged` object handles continuation tokens automatically
- No need to manually manage continuation tokens unless you want to resume from a specific point

## Notes

- `list_properties_of_secrets()` returns only metadata, not secret values
- Use `client.get_secret(name)` to retrieve actual secret values
- The script filters for enabled secrets using the `enabled` property
- SecretProperties includes: name, content_type, created_on, enabled, expires_on, tags, and more
