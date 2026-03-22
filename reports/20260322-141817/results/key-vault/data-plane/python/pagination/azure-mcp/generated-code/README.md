# Azure Key Vault Secrets Pagination Demo

This project demonstrates how to handle pagination when listing secrets from an Azure Key Vault that contains hundreds of secrets using the Azure SDK for Python.

## Overview

The script demonstrates three different approaches to pagination using the `ItemPaged` pattern:

1. **Simple iteration**: Let `ItemPaged` handle pagination automatically (recommended for most cases)
2. **Page-by-page processing**: Use `by_page()` to process secrets in explicit pages
3. **Resumable pagination**: Use continuation tokens to pause and resume pagination

## Requirements

- Python 3.9 or later
- Azure Key Vault with secrets
- Appropriate Azure credentials configured

## Installation

Install the required packages:

```bash
pip install -r requirements.txt
```

Or install packages directly:

```bash
pip install azure-keyvault-secrets azure-identity
```

## Authentication Setup

The script uses `DefaultAzureCredential` which supports multiple authentication methods in the following order:

1. **Environment Variables** (for service principals):
   ```bash
   export AZURE_CLIENT_ID="your-client-id"
   export AZURE_CLIENT_SECRET="your-client-secret"
   export AZURE_TENANT_ID="your-tenant-id"
   ```

2. **Managed Identity**: Automatically works when running on Azure services (VM, App Service, Functions, etc.)

3. **Azure CLI**: Run `az login` to authenticate

4. **Visual Studio Code**: Use the Azure Account extension

5. **Azure PowerShell**: Use `Connect-AzAccount`

## Usage

Set the Key Vault URL environment variable:

```bash
export VAULT_URL="https://your-key-vault-name.vault.azure.net/"
```

Run the script:

```bash
python azure_keyvault_pagination_demo.py
```

## How Pagination Works

### ItemPaged Pattern

The `list_properties_of_secrets()` method returns an `ItemPaged[SecretProperties]` object:

```python
secret_properties = client.list_properties_of_secrets()
```

### Method 1: Automatic Pagination (Simple)

The simplest approach - `ItemPaged` automatically fetches additional pages as needed:

```python
for secret in secret_properties:
    if secret.enabled:
        print(f"Name: {secret.name}")
        print(f"Created: {secret.created_on}")
```

### Method 2: Explicit Page Processing

Use `by_page()` to process secrets in explicit pages:

```python
pages = secret_properties.by_page()

for page in pages:
    for secret in page:
        # Process each secret in the page
        print(secret.name)
```

### Method 3: Continuation Tokens

Save and restore pagination state using continuation tokens:

```python
# Get first page
pages = secret_properties.by_page()
first_page = next(pages)

# Save continuation token
token = pages.continuation_token

# Later, resume from saved token
resumed_pages = client.list_properties_of_secrets().by_page(continuation_token=token)
```

## Secret Properties Available

The script demonstrates accessing the following properties from `SecretProperties`:

- `name`: The secret's name
- `enabled`: Whether the secret is enabled
- `content_type`: An arbitrary string indicating the type of the secret
- `created_on`: When the secret was created (UTC datetime)
- `updated_on`: When the secret was last updated (UTC datetime)
- `version`: The secret's version
- `tags`: Application-specific metadata

**Note**: `list_properties_of_secrets()` does NOT return secret values. To get a secret's value, use `client.get_secret(name)`.

## Key Vault Permissions Required

The script requires the following Key Vault permission:

- **secrets/list**: List secret identifiers and attributes

To grant this permission using Azure CLI:

```bash
# For a user
az keyvault set-policy --name YOUR_VAULT_NAME \
    --upn user@example.com \
    --secret-permissions list

# For a service principal
az keyvault set-policy --name YOUR_VAULT_NAME \
    --spn YOUR_CLIENT_ID \
    --secret-permissions list
```

## Reference Documentation

- [Azure Key Vault Secrets README](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)
- [SecretClient API Reference](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient)
- [ItemPaged API Reference](https://learn.microsoft.com/en-us/python/api/azure-core/azure.core.paging.itempaged)
- [SecretProperties API Reference](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretproperties)
- [DefaultAzureCredential](https://learn.microsoft.com/en-us/python/api/azure-identity/azure.identity.defaultazurecredential)

## Example Output

```
Connecting to Key Vault: https://my-vault.vault.azure.net/

================================================================================
METHOD 1: Simple iteration (ItemPaged handles pagination automatically)
================================================================================

Secret #1:
  Name: database-password
  Content Type: text/plain
  Created: 2024-01-15 10:30:00
  Enabled: True

Secret #2:
  Name: api-key
  Content Type: application/json
  Created: 2024-01-16 14:20:00
  Enabled: True

...

================================================================================
METHOD 2: Process secrets page by page using by_page()
================================================================================

--- Page 1 ---
  [1] database-password
      Content Type: text/plain
      Created: 2024-01-15 10:30:00
...

Page 1 summary: 25 secrets (23 enabled)

--- Page 2 ---
...
```

## Performance Considerations

- **Page Size**: Azure Key Vault determines the page size automatically based on the number of results
- **Network Efficiency**: Using `by_page()` can be more efficient for large vaults as it allows batch processing
- **Memory**: Simple iteration is memory efficient as it streams results
- **Rate Limiting**: The SDK handles throttling automatically with built-in retry logic

## License

This demo code is provided as-is for educational purposes.
