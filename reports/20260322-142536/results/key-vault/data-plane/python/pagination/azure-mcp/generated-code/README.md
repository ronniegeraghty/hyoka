# Azure Key Vault Secrets Pagination Demo

This script demonstrates how the Azure SDK for Python handles pagination when listing secrets from an Azure Key Vault with hundreds of secrets.

## Features Demonstrated

1. **SecretClient with DefaultAzureCredential** - Secure authentication using Azure's recommended credential chain
2. **ItemPaged Pattern** - Understanding how Azure SDK returns paginated results
3. **by_page() Method** - Explicit control over page-by-page processing
4. **Continuation Tokens** - Resumable pagination for large result sets
5. **Filtering** - Show only enabled secrets
6. **Secret Properties** - Access name, content type, and created date without retrieving secret values

## Installation

Install the required packages:

```bash
pip install -r requirements.txt
```

Or install directly:

```bash
pip install azure-keyvault-secrets azure-identity
```

## Prerequisites

1. **Azure Key Vault**: You need an existing Azure Key Vault with secrets
2. **Authentication**: Configure one of these authentication methods:
   - **Azure CLI**: Run `az login`
   - **Managed Identity**: If running on Azure (VM, App Service, etc.)
   - **Environment Variables**: Set `AZURE_TENANT_ID`, `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`
   - **Service Principal**: Configured through Azure CLI or environment variables

3. **Permissions**: Your identity needs the `secrets/list` permission on the Key Vault

4. **Environment Variable**: Set the Key Vault URL:
   ```bash
   export AZURE_KEY_VAULT_URL="https://your-vault-name.vault.azure.net/"
   ```

## Usage

```bash
python list_secrets_paginated.py
```

## How Pagination Works

### ItemPaged Pattern

The `list_properties_of_secrets()` method returns an `ItemPaged[SecretProperties]` object:

```python
secret_properties = client.list_properties_of_secrets()
```

This is an iterator that:
- Automatically fetches additional pages from the server as you iterate
- Handles all HTTP requests and pagination logic internally
- Returns `SecretProperties` objects (not the secret values)

### Three Pagination Methods

#### Method 1: Automatic Pagination (Simple)
```python
for secret in client.list_properties_of_secrets():
    if secret.enabled:
        print(f"{secret.name}: {secret.created_on}")
```
The SDK automatically fetches new pages as needed.

#### Method 2: Explicit Page Processing
```python
pages = client.list_properties_of_secrets().by_page()
for page in pages:
    for secret in page:
        print(secret.name)
```
Process one page at a time with full control.

#### Method 3: Resumable Pagination
```python
pages = client.list_properties_of_secrets().by_page(continuation_token=None)
for page in pages:
    token = page.continuation_token
    # Save token to resume later
    break

# Resume later
pages = client.list_properties_of_secrets().by_page(continuation_token=token)
```
Use continuation tokens to resume from where you left off.

## Secret Properties vs Secret Values

**Important**: The `list_properties_of_secrets()` method returns metadata only, not the actual secret values. This is for security and performance:

- ✅ Returns: name, enabled status, created_on, updated_on, content_type, tags, etc.
- ❌ Does NOT return: the actual secret value

To get the secret value, use:
```python
secret = client.get_secret("secret-name")
print(secret.value)  # The actual secret value
```

## Performance Considerations

- **Page Size**: The server determines page size (typically 25-100 items per page)
- **Network Calls**: Each page requires one HTTP request to Azure
- **Large Vaults**: For vaults with hundreds/thousands of secrets, use `by_page()` for better memory efficiency
- **Filtering**: Apply filters in your code; the API doesn't support server-side filtering

## Example Output

```
Connecting to Key Vault: https://my-vault.vault.azure.net/
================================================================================

=== Method 1: Simple iteration (automatic pagination) ===

Name: database-password
  Content Type: text/plain
  Created: 2024-01-15 10:30:00
  Enabled: True

Name: api-key
  Content Type: Not set
  Created: 2024-01-16 14:22:00
  Enabled: True

Total secrets: 150
Enabled secrets: 142

================================================================================
=== Method 2: Explicit page-by-page processing using by_page() ===

--- Page 1 ---
  [1] database-password
      Content Type: text/plain
      Created: 2024-01-15 10:30:00
      Enabled: True
...
```

## References

- [Azure Key Vault Secrets Python SDK Documentation](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme?view=azure-python)
- [SecretClient API Reference](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient?view=azure-python)
- [ItemPaged Pattern Documentation](https://learn.microsoft.com/en-us/python/api/azure-core/azure.core.paging.itempaged?view=azure-python)
- [DefaultAzureCredential Documentation](https://learn.microsoft.com/en-us/python/api/azure-identity/azure.identity.defaultazurecredential?view=azure-python)

## License

This example code is provided for educational purposes.
