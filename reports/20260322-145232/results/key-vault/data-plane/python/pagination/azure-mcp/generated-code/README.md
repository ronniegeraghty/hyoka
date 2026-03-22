# Azure Key Vault Secrets Pagination Guide

This example demonstrates how the `azure-keyvault-secrets` SDK handles pagination for large vaults using the **ItemPaged** pattern.

## Required Packages

Install the required packages using pip:

```bash
pip install -r requirements.txt
```

Or install individually:

```bash
pip install azure-keyvault-secrets azure-identity
```

### Package Details

- **azure-keyvault-secrets** (>=4.8.0): SDK for managing secrets in Azure Key Vault
- **azure-identity** (>=1.15.0): Authentication library supporting multiple credential types
- **azure-core** (>=1.29.0): Contains the ItemPaged base class

## How ItemPaged Pagination Works

### What is ItemPaged?

`ItemPaged` is Azure SDK's standard pagination pattern. When you call `list_properties_of_secrets()`, you get an `ItemPaged[SecretProperties]` object that:

1. **Lazy loads data**: Secrets are fetched from Azure only as you iterate
2. **Handles pagination automatically**: Makes multiple API calls behind the scenes
3. **Supports two iteration modes**:
   - **Item-by-item**: Iterate directly over secrets (automatic pagination)
   - **Page-by-page**: Use `by_page()` to process secrets in batches

### Pagination Flow

```
Azure Key Vault
    ↓
SecretClient.list_properties_of_secrets()
    ↓
Returns: ItemPaged[SecretProperties]
    ↓
┌─────────────────────────────────────┐
│ Iteration Mode 1: Item-by-item     │
│ for secret in secrets:              │
│   # Automatically fetches new pages │
└─────────────────────────────────────┘
    OR
┌─────────────────────────────────────┐
│ Iteration Mode 2: Page-by-page     │
│ for page in secrets.by_page():      │
│   for secret in page:               │
│     # Process each secret           │
└─────────────────────────────────────┘
```

### Key Concepts from the Documentation

1. **ItemPaged** (azure.core.paging.ItemPaged):
   - Iterator that returns items one at a time
   - Automatically handles pagination
   - Method: `by_page(continuation_token=None)` - Get pages instead of items

2. **SecretProperties**:
   - Returned by list operations (doesn't include secret values)
   - Properties: `name`, `enabled`, `content_type`, `created_on`, `updated_on`, etc.
   - Use `get_secret(name)` to retrieve the actual secret value

3. **Continuation Tokens**:
   - Opaque strings that mark a position in the result set
   - Can be used to resume pagination from a specific point
   - Useful for distributed processing or resuming after failures

## Running the Example

### Prerequisites

1. **Azure Key Vault**: You need an existing Key Vault with secrets
2. **Authentication**: Set up one of the following:

   **Option A: Azure CLI (easiest for local development)**
   ```bash
   az login
   ```

   **Option B: Service Principal with environment variables**
   ```bash
   export AZURE_CLIENT_ID="your-client-id"
   export AZURE_CLIENT_SECRET="your-client-secret"
   export AZURE_TENANT_ID="your-tenant-id"
   ```

   **Option C: Managed Identity** (when running on Azure VMs, App Service, etc.)

3. **Permissions**: Your identity needs the `secrets/list` permission on the Key Vault

### Set Environment Variable

```bash
export VAULT_URL="https://your-vault-name.vault.azure.net/"
```

### Run the Script

```bash
python azure_keyvault_pagination.py
```

## Script Features

The script demonstrates three pagination methods:

### Method 1: Simple Iteration (Automatic Pagination)
```python
secrets = client.list_properties_of_secrets()
for secret in secrets:
    if secret.enabled:
        print(secret.name)
```

**When to use**: 
- Default approach for most scenarios
- Simplest code
- Pagination handled automatically

### Method 2: Page-by-Page Processing
```python
secrets = client.list_properties_of_secrets()
pages = secrets.by_page()
for page in pages:
    for secret in page:
        if secret.enabled:
            print(secret.name)
```

**When to use**:
- Need to track progress (e.g., "Processing page 5 of 10")
- Want to measure API call count
- Batch processing with per-page operations
- Better memory control for very large vaults

### Method 3: Continuation Tokens
```python
# Get first page
pages = secrets.by_page()
first_page = next(pages)
continuation_token = first_page.continuation_token

# Resume from token
remaining = client.list_properties_of_secrets()
remaining_pages = remaining.by_page(continuation_token=continuation_token)
```

**When to use**:
- Resuming after interruption
- Distributing work across multiple processes/workers
- Implementing custom pagination UI (e.g., "Load more" buttons)
- Checkpointing long-running operations

## Understanding the Output

The script filters and displays only **enabled** secrets, showing:

- **Name**: The secret identifier
- **Content Type**: Optional metadata describing the secret format (e.g., "text/plain", "application/json")
- **Created Date**: When the secret was first created
- **Enabled Status**: Whether the secret is currently active

## Performance Considerations

### Page Size
- Azure Key Vault determines page size automatically (typically 25 items)
- Cannot be configured by the client
- May vary based on server load and throttling

### Best Practices for Large Vaults

1. **Use `by_page()` for better control**:
   ```python
   for page in secrets.by_page():
       # Process batch
       # Optional: Add logging, progress tracking, error handling per page
   ```

2. **Don't retrieve secret values unless needed**:
   ```python
   # Good: Only list properties
   secrets = client.list_properties_of_secrets()
   
   # Avoid: Calling get_secret() for every secret in a large vault
   for secret in secrets:
       value = client.get_secret(secret.name)  # Additional API call per secret!
   ```

3. **Filter server-side when possible**:
   - The SDK doesn't support server-side filtering
   - Filter on the client side (as shown in the example)

4. **Handle throttling**:
   ```python
   from azure.core.exceptions import HttpResponseError
   
   try:
       secrets = client.list_properties_of_secrets()
       for secret in secrets:
           process(secret)
   except HttpResponseError as e:
       if e.status_code == 429:  # Too Many Requests
           # Implement retry logic with exponential backoff
           pass
   ```

## Troubleshooting

### "DefaultAzureCredential failed to retrieve a token"
- Run `az login` if using Azure CLI authentication
- Verify environment variables if using service principal
- Check that managed identity is configured if running on Azure

### "Forbidden" or "Access Denied"
- Verify you have `secrets/list` permission on the Key Vault
- Check your Key Vault's access policies or RBAC settings
- Ensure you're using the correct VAULT_URL

### "No secrets found in the vault"
- Verify the vault contains secrets
- Check that the secrets are not all disabled
- Ensure you have permission to list secrets

## Additional Resources

- [Azure Key Vault Secrets SDK Documentation](https://learn.microsoft.com/python/api/overview/azure/keyvault-secrets-readme)
- [ItemPaged API Reference](https://learn.microsoft.com/python/api/azure-core/azure.core.paging.itempaged)
- [DefaultAzureCredential Documentation](https://learn.microsoft.com/python/api/azure-identity/azure.identity.defaultazurecredential)
- [Azure SDK Pagination Guidelines](https://azure.github.io/azure-sdk/python_design.html#pagination)

## License

This example is provided for educational purposes to demonstrate Azure SDK pagination patterns.
