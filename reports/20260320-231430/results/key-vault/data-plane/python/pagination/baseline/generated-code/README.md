# Azure Key Vault Secrets Pagination Demo

This script demonstrates how to handle pagination when listing secrets from an Azure Key Vault that contains hundreds of secrets using the Azure SDK for Python.

## Key Concepts from Azure SDK Documentation

### ItemPaged Pattern

The `list_properties_of_secrets()` method returns an `ItemPaged[SecretProperties]` object. This is a special iterator that supports pagination for efficiently processing large result sets.

### by_page() Method

The `by_page()` method transforms the item iterator into a page iterator, where each page contains multiple items. This is useful for:
- Processing secrets in manageable chunks
- Implementing rate limiting or throttling
- Better memory management with large vaults
- Using continuation tokens to resume interrupted operations

### SecretProperties Attributes

When listing secrets, you get `SecretProperties` objects (not full secrets with values). Available properties include:
- `name`: The secret's name
- `content_type`: Application-specific content type indicator
- `created_on`: When the secret was created (datetime, UTC)
- `enabled`: Whether the secret is enabled for use (bool)
- `updated_on`: When the secret was last updated
- `expires_on`: When the secret expires
- `tags`: Custom metadata dictionary
- `version`: The secret's version

**Note**: List operations do NOT return secret values. Use `get_secret(name)` to retrieve the actual secret value.

## Installation

```bash
pip install -r requirements.txt
```

Or install packages individually:

```bash
pip install azure-keyvault-secrets azure-identity
```

## Prerequisites

1. An Azure Key Vault with secrets
2. Python 3.9 or later
3. Proper authentication configured (see below)

## Authentication Setup

The script uses `DefaultAzureCredential`, which tries multiple authentication methods in order:

1. **Environment Variables**: Set these for service principal authentication:
   ```bash
   export AZURE_CLIENT_ID="your-client-id"
   export AZURE_CLIENT_SECRET="your-client-secret"
   export AZURE_TENANT_ID="your-tenant-id"
   ```

2. **Azure CLI**: Login using Azure CLI:
   ```bash
   az login
   ```

3. **Managed Identity**: Automatically used when running on Azure resources (VMs, App Service, Functions, etc.)

4. **Interactive Browser**: Falls back to browser-based authentication

## Required Permissions

Your Azure identity must have the following permission on the Key Vault:
- `secrets/list` - to list secret properties

You can grant this using Azure RBAC or Key Vault access policies.

## Usage

Set the Key Vault URL and run the script:

```bash
export VAULT_URL="https://your-vault-name.vault.azure.net/"
python list_secrets_paginated.py
```

## How Pagination Works

The script demonstrates the recommended pagination pattern from the Azure SDK documentation:

```python
# Get the ItemPaged iterator
secret_properties_paged = client.list_properties_of_secrets()

# Use by_page() to iterate through pages
for page in secret_properties_paged.by_page():
    # Each page is itself an iterator of SecretProperties
    for secret_properties in page:
        # Process each secret
        print(secret_properties.name)
```

### Why Use by_page()?

1. **Memory Efficiency**: Process secrets in chunks rather than loading all at once
2. **Performance**: Better for large vaults with hundreds or thousands of secrets
3. **Control**: You can track progress and implement your own batching logic
4. **Continuation Tokens**: Can save state and resume processing later

### Alternative: Direct Iteration

For simpler use cases, you can iterate directly:

```python
for secret_properties in client.list_properties_of_secrets():
    print(secret_properties.name)
```

This is simpler but gives you less control over pagination.

## Output Example

```
Connecting to Key Vault: https://your-vault-name.vault.azure.net/

================================================================================
Listing all enabled secrets with pagination
================================================================================

--- Page 1 ---
  Secret: database-password
    Content Type: text/plain
    Created: 2024-01-15 10:30:45 UTC
    Enabled: True

  Secret: api-key
    Content Type: application/json
    Created: 2024-01-16 14:22:10 UTC
    Enabled: True

Processed 25 secrets in this page

--- Page 2 ---
  ...

================================================================================
Summary
================================================================================
Total pages processed: 8
Total secrets found: 200
Enabled secrets: 185
Disabled secrets: 15

Done!
```

## References

- [Azure Key Vault Secrets SDK Documentation](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)
- [SecretClient API Reference](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient)
- [ItemPaged API Reference](https://learn.microsoft.com/en-us/python/api/azure-core/azure.core.paging.itempaged)
- [DefaultAzureCredential Documentation](https://learn.microsoft.com/en-us/python/api/azure-identity/azure.identity.defaultazurecredential)
