# Azure Key Vault Secrets CRUD Demo

This script demonstrates all four CRUD operations (Create, Read, Update, Delete) on Azure Key Vault secrets using the Azure SDK for Python.

## Prerequisites

1. **Python 3.9 or later**
2. **An Azure Key Vault** with soft-delete enabled
3. **Appropriate permissions**: Your Azure identity needs the following Key Vault secret permissions:
   - `secrets/set` - to create/update secrets
   - `secrets/get` - to read secrets
   - `secrets/delete` - to delete secrets
   - `secrets/purge` - to permanently delete secrets

## Installation

Install the required Python packages:

```bash
pip install -r requirements.txt
```

Or install directly:

```bash
pip install azure-keyvault-secrets azure-identity
```

## Authentication

The script uses `DefaultAzureCredential` which attempts authentication in the following order:

1. **Environment variables** - Set these for service principal authentication:
   - `AZURE_CLIENT_ID`
   - `AZURE_TENANT_ID`
   - `AZURE_CLIENT_SECRET`

2. **Managed Identity** - Automatically used when running on Azure services

3. **Azure CLI** - Run `az login` first:
   ```bash
   az login
   ```

4. **Azure PowerShell** - Run `Connect-AzAccount` first

5. **Interactive browser** - Opens a browser window for authentication

## Usage

1. Set the Key Vault URL environment variable:
   ```bash
   export VAULT_URL="https://your-vault-name.vault.azure.net/"
   ```

2. Run the script:
   ```bash
   python keyvault_crud_demo.py
   ```

## What the Script Does

### 1. CREATE
Creates a new secret named "my-secret" with value "my-secret-value"

### 2. READ
Retrieves the secret and displays its value and properties

### 3. UPDATE
- Updates the secret to a new value "updated-value" (creates a new version)
- Updates secret metadata (content type, enabled status)

### 4. DELETE
- Soft-deletes the secret (can be recovered within retention period)
- Purges the secret permanently (irreversible deletion)

## Example Output

```
Connecting to Key Vault: https://my-vault.vault.azure.net/
--------------------------------------------------------------------------------

1. CREATE - Setting a new secret
--------------------------------------------------------------------------------
✓ Secret created successfully!
  Name: my-secret
  Value: my-secret-value
  Version: abc123...
  Created: 2026-03-21 19:00:00

2. READ - Retrieving the secret
--------------------------------------------------------------------------------
✓ Secret retrieved successfully!
  Name: my-secret
  Value: my-secret-value
  Version: abc123...

3. UPDATE - Updating the secret value
--------------------------------------------------------------------------------
✓ Secret value updated successfully!
  Name: my-secret
  New Value: updated-value
  New Version: def456...

4. DELETE - Deleting and purging the secret
--------------------------------------------------------------------------------
✓ Secret deleted successfully!
✓ Secret purged successfully!
  The secret has been permanently deleted and cannot be recovered.

================================================================================
All CRUD operations completed successfully!
================================================================================
```

## Error Handling

The script includes comprehensive error handling for:
- Missing environment variables
- Authentication failures
- Permission issues
- Resource not found errors
- Azure service errors

## References

- [Azure Key Vault Secrets SDK Documentation](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)
- [Azure Identity Documentation](https://learn.microsoft.com/en-us/python/api/overview/azure/identity-readme)
- [DefaultAzureCredential Documentation](https://learn.microsoft.com/en-us/python/api/azure-identity/azure.identity.defaultazurecredential)
