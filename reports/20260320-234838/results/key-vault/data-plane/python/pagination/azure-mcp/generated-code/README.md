# Azure Key Vault Secrets - Pagination Example

This example demonstrates how the Azure SDK for Python handles pagination when listing hundreds of secrets in an Azure Key Vault.

## Required Packages

Install the required packages using pip:

```bash
pip install -r requirements.txt
```

Or install individually:

```bash
pip install azure-keyvault-secrets azure-identity azure-core
```

### Package Breakdown:

- **azure-keyvault-secrets**: Provides `SecretClient` for managing Key Vault secrets
- **azure-identity**: Provides `DefaultAzureCredential` for authentication
- **azure-core**: Provides the `ItemPaged` pagination pattern (installed as a dependency)

## How Azure SDK Pagination Works

### ItemPaged Pattern

The Azure SDK uses the `ItemPaged` class from `azure-core` for paginated results. When you call `list_properties_of_secrets()`, it returns an `ItemPaged[SecretProperties]` object.

### Two Ways to Use ItemPaged:

#### 1. Automatic Pagination (Simple Iterator)
```python
for secret in client.list_properties_of_secrets():
    print(secret.name)
```
- Simplest approach
- Pagination happens automatically behind the scenes
- Good for small to medium vaults

#### 2. Explicit Page-by-Page Iteration (by_page())
```python
for page in client.list_properties_of_secrets().by_page():
    for secret in page:
        print(secret.name)
```
- More control over pagination
- Process secrets in chunks/batches
- Better for large vaults with hundreds or thousands of secrets
- Allows tracking progress per page
- More memory efficient for very large result sets

### Key Points:

1. **No Secret Values**: `list_properties_of_secrets()` returns only metadata (SecretProperties), not actual secret values. Use `get_secret(name)` to retrieve values.

2. **Lazy Evaluation**: Pages are fetched on-demand as you iterate. The API doesn't fetch all results upfront.

3. **Page Size**: The Azure service determines page size automatically. You can't explicitly set it in the SDK.

4. **Continuation Tokens**: The `by_page()` method accepts an optional `continuation_token` parameter to resume from a specific point.

## Authentication

The script uses `DefaultAzureCredential`, which attempts authentication in this order:

1. **Environment variables** (AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET)
2. **Managed Identity** (if running in Azure)
3. **Azure CLI** (`az login`)
4. **Azure PowerShell**
5. **Interactive browser** (as fallback)

### Quick Setup:

```bash
# Login using Azure CLI
az login

# Set your Key Vault URL
export AZURE_KEYVAULT_URL='https://your-vault-name.vault.azure.net/'
```

## Usage

### Option 1: Environment Variable
```bash
export AZURE_KEYVAULT_URL='https://your-vault-name.vault.azure.net/'
python list_keyvault_secrets_paginated.py
```

### Option 2: Command Line Argument
```bash
python list_keyvault_secrets_paginated.py https://your-vault-name.vault.azure.net/
```

## Permissions Required

Your authenticated identity needs the following Key Vault permission:
- **Secrets: List** - to enumerate secrets

You can grant this via:
- Key Vault Access Policies (legacy)
- Azure RBAC roles: "Key Vault Secrets User" or "Key Vault Reader"

## Script Output

The script will:
1. List all secrets page by page
2. Show page number for each batch
3. Filter to only show **enabled** secrets
4. Display for each secret:
   - Name
   - Content Type
   - Created Date (UTC)
   - Enabled status
5. Print summary statistics

### Example Output:
```
Listing secrets from: https://myvault.vault.azure.net/

================================================================================

--- Page 1 ---
  Name:         database-password
  Content Type: text/plain
  Created:      2024-01-15 10:30:45 UTC
  Enabled:      True

  Name:         api-key
  Content Type: application/json
  Created:      2024-02-20 14:22:10 UTC
  Enabled:      True

Secrets in this page: 2

--- Page 2 ---
  Name:         connection-string
  Content Type: Not set
  Created:      2024-03-01 08:15:30 UTC
  Enabled:      True

Secrets in this page: 1

================================================================================

Summary:
  Total pages processed: 2
  Total secrets found:   3
  Enabled secrets:       3
  Disabled secrets:      0
```

## Code Structure

The script provides two functions:

### `list_secrets_paginated(vault_url)`
Demonstrates explicit page-by-page iteration using `by_page()`. This is the recommended approach for vaults with hundreds of secrets.

### `list_secrets_simple(vault_url)`
Shows the simpler automatic pagination approach. Good for understanding how ItemPaged works as a basic iterator.

## Performance Considerations

For vaults with hundreds of secrets:
- Use `by_page()` to process in chunks
- Apply filters early (e.g., check `enabled` status before processing)
- Only call `get_secret()` when you need the actual secret value
- Consider implementing retry logic for transient failures
- Monitor API rate limits for very large operations

## References

- [Azure Key Vault Secrets Python SDK](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/)
- [SecretClient Documentation](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient)
- [ItemPaged Documentation](https://learn.microsoft.com/en-us/python/api/azure-core/azure.core.paging.itempaged)
- [DefaultAzureCredential](https://learn.microsoft.com/en-us/python/api/azure-identity/azure.identity.defaultazurecredential)
