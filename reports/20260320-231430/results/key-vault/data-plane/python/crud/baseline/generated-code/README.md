# Azure Key Vault Secrets CRUD Operations Demo

This script demonstrates all four CRUD operations on Azure Key Vault secrets using the Azure SDK for Python.

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

- **azure-keyvault-secrets**: Azure Key Vault Secrets client library for Python
- **azure-identity**: Azure authentication library (provides DefaultAzureCredential)

## Prerequisites

1. **Python 3.9 or later**
2. **An Azure Key Vault** - Create one using Azure CLI:
   ```bash
   az keyvault create --name <your-vault-name> --resource-group <your-rg> --location <location>
   ```
3. **Proper permissions** - Ensure you have the following Key Vault secrets permissions:
   - `secrets/set` (for create/update)
   - `secrets/get` (for read)
   - `secrets/delete` (for delete)
   - `secrets/purge` (for purge, if soft-delete is enabled)

## Authentication Setup

The script uses `DefaultAzureCredential`, which tries multiple authentication methods in order:

### Option 1: Azure CLI (Recommended for local development)
```bash
az login
```

### Option 2: Environment Variables
```bash
export AZURE_CLIENT_ID="<your-client-id>"
export AZURE_TENANT_ID="<your-tenant-id>"
export AZURE_CLIENT_SECRET="<your-client-secret>"
```

### Option 3: Managed Identity
Automatically works when running in Azure (App Service, VM, Container Instances, etc.)

## Usage

1. Set the Key Vault URL environment variable:
   ```bash
   export AZURE_KEY_VAULT_URL="https://<your-vault-name>.vault.azure.net/"
   ```

2. Run the script:
   ```bash
   python keyvault_crud_demo.py
   ```

## What the Script Does

The script performs the following operations in sequence:

1. **CREATE**: Creates a new secret named "my-secret" with value "my-secret-value"
2. **READ**: Retrieves and displays the secret value
3. **UPDATE**: Updates the secret to a new value "updated-value"
4. **DELETE**: Deletes the secret (soft-delete)
5. **PURGE**: Permanently deletes the secret (for vaults with soft-delete enabled)

## Expected Output

```
Connecting to Key Vault: https://your-vault.vault.azure.net/
--------------------------------------------------------------------------------

1. CREATE - Setting a new secret 'my-secret'
   ✓ Secret created successfully
   Name: my-secret
   Value: my-secret-value
   Version: abc123...
   Created on: 2026-03-21 06:14:33.123456

2. READ - Retrieving the secret 'my-secret'
   ✓ Secret retrieved successfully
   Name: my-secret
   Value: my-secret-value
   Version: abc123...

3. UPDATE - Updating secret to a new value
   ✓ Secret updated successfully
   Name: my-secret
   New Value: updated-value
   New Version: def456...
   Updated on: 2026-03-21 06:14:34.123456

4. DELETE - Deleting and purging the secret
   4a. Starting deletion...
   ✓ Secret deleted successfully
   Name: my-secret
   Deleted on: 2026-03-21 06:14:35.123456
   Scheduled purge date: 2026-04-20 06:14:35.123456
   Recovery ID: https://...
   4b. Purging deleted secret (permanent deletion)...
   ✓ Secret purged successfully (permanently deleted)

================================================================================
All CRUD operations completed successfully!
================================================================================
```

## Error Handling

The script includes comprehensive error handling for:

- Missing environment variables
- Authentication failures
- Resource not found errors
- Permission errors
- Soft-delete configuration issues

## Notes

- If your vault does **not** have soft-delete enabled, the delete operation is permanent, and the purge step will show a warning
- The `set_secret()` method is used for both CREATE and UPDATE operations - it creates a new version if the secret already exists
- Each secret update creates a new version; all versions are stored until deleted
- Soft-delete provides a safety net, allowing secret recovery within the retention period (default 90 days)

## References

- [Azure Key Vault Secrets Python SDK Documentation](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)
- [DefaultAzureCredential Documentation](https://learn.microsoft.com/en-us/python/api/azure-identity/azure.identity.defaultazurecredential)
- [Azure Key Vault Documentation](https://learn.microsoft.com/en-us/azure/key-vault/)
