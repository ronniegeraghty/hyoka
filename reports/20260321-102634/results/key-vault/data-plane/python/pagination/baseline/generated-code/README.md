# Azure Key Vault Secrets - Pagination Demo

This script demonstrates how to handle pagination when working with Azure Key Vault vaults containing hundreds of secrets using the Azure SDK for Python.

## Required Packages

Install the required packages using pip:

```bash
pip install azure-keyvault-secrets azure-identity
```

Or install from the requirements.txt:

```bash
pip install -r requirements.txt
```

### Package Details

- **azure-keyvault-secrets** (>=4.8.0): Provides `SecretClient` for interacting with Azure Key Vault secrets
- **azure-identity** (>=1.15.0): Provides `DefaultAzureCredential` for authentication

## Prerequisites

1. **Azure Key Vault**: You need an existing Azure Key Vault with secrets
2. **Authentication**: Configure authentication for `DefaultAzureCredential` (see below)
3. **Permissions**: Ensure you have `secrets/list` permission on the vault
4. **Environment Variable**: Set `VAULT_URL` to your Key Vault URL

```bash
export VAULT_URL="https://your-vault-name.vault.azure.net/"
```

## Understanding ItemPaged Pagination

The Azure SDK for Python uses the `ItemPaged` pattern for pagination. When you call `list_properties_of_secrets()`, it returns an `ItemPaged[SecretProperties]` object that supports two iteration modes:

### 1. Automatic Iteration (Simple)

The simplest approach - iterate directly over the `ItemPaged` object:

```python
secret_properties = client.list_properties_of_secrets()

for secret_property in secret_properties:
    print(secret_property.name)
```

**How it works:**
- The SDK automatically fetches pages from the server as needed
- You don't need to worry about page boundaries
- Ideal for processing all secrets sequentially

### 2. Page-by-Page Iteration (Explicit Control)

Use the `by_page()` method for explicit control over pagination:

```python
secret_properties = client.list_properties_of_secrets()
page_iterator = secret_properties.by_page()

for page in page_iterator:
    # Process all secrets in this page
    for secret_property in page:
        print(secret_property.name)
```

**How it works:**
- `by_page()` returns an iterator of pages
- Each page is itself an iterator of `SecretProperties` objects
- Useful when you need to process secrets in batches
- Allows for progress tracking and batch operations

### 3. Continuation Tokens (Resumable Iteration)

For advanced scenarios, you can use continuation tokens:

```python
secret_properties = client.list_properties_of_secrets()
page_iterator = secret_properties.by_page(continuation_token=saved_token)

for page in page_iterator:
    # Process page
    pass
```

**Use cases:**
- Resume iteration after interruption
- Implement custom pagination in web applications
- Save state between runs

## Script Demonstration Methods

The script includes four demonstration methods:

### Method 1: Basic Iteration
Shows automatic pagination with filtering for enabled secrets only.

### Method 2: Page-by-Page Processing
Demonstrates explicit page control with per-page statistics.

### Method 3: Continuation Tokens
Shows how to work with continuation tokens for resumable pagination.

### Method 4: Detailed Properties
Displays all available properties of `SecretProperties` objects.

## Running the Script

```bash
# Set your vault URL
export VAULT_URL="https://your-vault-name.vault.azure.net/"

# Run the script
python azure_keyvault_pagination.py
```

## Authentication with DefaultAzureCredential

`DefaultAzureCredential` tries multiple authentication methods in order:

1. **Environment variables** (service principal):
   ```bash
   export AZURE_TENANT_ID="your-tenant-id"
   export AZURE_CLIENT_ID="your-client-id"
   export AZURE_CLIENT_SECRET="your-client-secret"
   ```

2. **Managed Identity**: If running on Azure (VM, App Service, etc.)

3. **Azure CLI**: If you're logged in via `az login`

4. **Azure PowerShell**: If you're logged in via PowerShell

5. **Interactive browser**: As a fallback

## SecretProperties Attributes

When listing secrets, you receive `SecretProperties` objects (not the secret values). Available attributes:

- **name**: The secret's name
- **enabled**: Whether the secret is enabled for use
- **content_type**: Optional content type indicator
- **created_on**: When the secret was created (UTC datetime)
- **updated_on**: When the secret was last updated (UTC datetime)
- **expires_on**: When the secret expires (optional)
- **not_before**: Time before which secret cannot be used (optional)
- **version**: The secret's version identifier
- **vault_url**: URL of the containing vault
- **managed**: Whether lifetime is managed by Key Vault
- **recovery_level**: Deletion recovery level
- **recoverable_days**: Days retained before permanent deletion
- **tags**: Application-specific metadata dictionary

**Note**: `list_properties_of_secrets()` does NOT return secret values. Use `client.get_secret(name)` to retrieve the actual secret value.

## Performance Considerations

- **Page Size**: The Azure service determines page size (typically 25 items)
- **Network Calls**: Each page requires a network request to Azure
- **Filtering**: Filtering is done client-side after fetching pages
- **Large Vaults**: For vaults with hundreds of secrets, use page-by-page iteration for better control

## Error Handling

Common exceptions from `azure.core.exceptions`:

- **ResourceNotFoundError**: Secret doesn't exist
- **HttpResponseError**: Network or service errors
- **ClientAuthenticationError**: Authentication failed

Example:
```python
from azure.core.exceptions import ResourceNotFoundError

try:
    secret_properties = client.list_properties_of_secrets()
    for prop in secret_properties:
        print(prop.name)
except ResourceNotFoundError as e:
    print(f"Resource not found: {e}")
```

## Additional Resources

- [Azure Key Vault Secrets SDK Documentation](https://learn.microsoft.com/python/api/overview/azure/keyvault-secrets-readme)
- [SecretClient API Reference](https://learn.microsoft.com/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient)
- [ItemPaged Documentation](https://learn.microsoft.com/python/api/azure-core/azure.core.paging.itempaged)
- [DefaultAzureCredential Documentation](https://learn.microsoft.com/python/api/azure-identity/azure.identity.defaultazurecredential)
