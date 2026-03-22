# Azure Key Vault Pagination Demo

This script demonstrates how the Azure SDK for Python handles pagination when listing secrets from an Azure Key Vault containing hundreds of secrets.

## Prerequisites

1. **Python 3.9 or later** (Azure SDK requirement)

2. **An Azure Key Vault** with secrets
   - Create one: https://learn.microsoft.com/azure/key-vault/general/quick-create-cli

3. **Azure authentication configured**
   - For local development: Run `az login`
   - For production: Use Managed Identity or Service Principal

## Installation

Install required packages:

```bash
pip install -r requirements.txt
```

Or install individually:

```bash
pip install azure-keyvault-secrets azure-identity
```

## Required Packages

- **azure-keyvault-secrets** (>= 4.8.0): Azure Key Vault Secrets client library
- **azure-identity** (>= 1.16.0): Azure authentication library with DefaultAzureCredential

## Configuration

Set the `VAULT_URL` environment variable to your Key Vault URL:

```bash
export VAULT_URL="https://your-vault-name.vault.azure.net/"
```

## Authentication Options

The script uses `DefaultAzureCredential`, which tries multiple authentication methods in order:

1. **Environment variables** (`AZURE_CLIENT_ID`, `AZURE_TENANT_ID`, `AZURE_CLIENT_SECRET`)
2. **Managed Identity** (when running in Azure)
3. **Azure CLI** (run `az login` for local development)
4. **Azure PowerShell**
5. **Interactive browser** (fallback)

For local development, the easiest method is Azure CLI:

```bash
az login
```

## Running the Script

```bash
python azure_keyvault_pagination.py
```

## What the Script Demonstrates

### 1. **ItemPaged Pattern**
The `list_properties_of_secrets()` method returns an `ItemPaged[SecretProperties]` object that supports:
- Direct iteration (automatic pagination)
- Explicit pagination via `by_page()` method

### 2. **Automatic Pagination** (Method 1)
```python
secret_properties = client.list_properties_of_secrets()
for secret in secret_properties:
    # SDK automatically fetches next pages as needed
    print(secret.name)
```

### 3. **Explicit Page Processing** (Method 2)
```python
secret_properties = client.list_properties_of_secrets()
pages = secret_properties.by_page()

for page in pages:
    for secret in page:
        print(secret.name)
```

### 4. **Continuation Tokens** (Method 3)
```python
pages = secret_properties.by_page()
first_page = next(pages)
continuation_token = pages.continuation_token

# Resume from a specific point
resumed_pages = client.list_properties_of_secrets().by_page(
    continuation_token=continuation_token
)
```

### 5. **Filtering Enabled Secrets**
The script filters to show only enabled secrets using the `enabled` property:

```python
for secret in secret_properties:
    if secret.enabled:
        print(f"Name: {secret.name}")
        print(f"Content Type: {secret.content_type}")
        print(f"Created On: {secret.created_on}")
```

## Key Concepts

### SecretProperties vs KeyVaultSecret
- **`list_properties_of_secrets()`** returns `SecretProperties` objects (metadata only, no values)
- **`get_secret(name)`** returns `KeyVaultSecret` objects (includes the actual secret value)

This design prevents accidentally loading hundreds of secret values into memory when listing.

### ItemPaged Pagination
- Azure SDK uses the `ItemPaged` pattern for paginated results
- The SDK handles pagination transparently when iterating
- Use `by_page()` for explicit control over page processing
- Each page is itself an iterator

### Performance Considerations
For large vaults (hundreds of secrets):
- Use `by_page()` to process secrets in batches
- Filter early to reduce memory usage
- Don't call `get_secret()` for every secret unless needed (values not included in list)

## Permissions Required

The Azure identity needs the following Key Vault permissions:
- **Secrets: List** - to list secrets

Grant permissions via:
```bash
az keyvault set-policy --name YOUR_VAULT_NAME \
  --upn YOUR_EMAIL@example.com \
  --secret-permissions list
```

Or use Azure RBAC role: **Key Vault Secrets User** or **Key Vault Reader**

## References

- [Azure Key Vault Secrets SDK Documentation](https://learn.microsoft.com/python/api/overview/azure/keyvault-secrets-readme)
- [SecretClient API Reference](https://learn.microsoft.com/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient)
- [ItemPaged API Reference](https://learn.microsoft.com/python/api/azure-core/azure.core.paging.itempaged)
- [Azure Key Vault Overview](https://learn.microsoft.com/azure/key-vault/general/overview)
