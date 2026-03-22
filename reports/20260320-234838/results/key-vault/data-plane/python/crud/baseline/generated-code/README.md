# Azure Key Vault Secrets CRUD Operations

This script demonstrates all four CRUD operations (Create, Read, Update, Delete) on Azure Key Vault secrets using the Azure SDK for Python.

## Prerequisites

1. **Python 3.9 or later**
2. **Azure subscription** with an Azure Key Vault
3. **Key Vault with soft-delete enabled** (default for new vaults)
4. **Proper permissions**: Your identity needs the following Key Vault access policies:
   - secrets/set
   - secrets/get
   - secrets/delete
   - secrets/purge

## Installation

Install the required packages:

```bash
pip install -r requirements.txt
```

Or install directly:

```bash
pip install azure-keyvault-secrets azure-identity
```

## Authentication Setup

The script uses `DefaultAzureCredential`, which attempts authentication through multiple methods in order:

1. **Azure CLI** (easiest for local development):
   ```bash
   az login
   ```

2. **Managed Identity** (for Azure-hosted resources)
3. **Environment variables** (for service principals)
4. **Visual Studio Code** authentication
5. **And more...**

For production, use Managed Identity. For local development, use Azure CLI login.

## Configuration

Set the `VAULT_URL` environment variable to your Key Vault URL:

```bash
export VAULT_URL="https://your-key-vault-name.vault.azure.net/"
```

## Usage

Run the script:

```bash
python keyvault_crud.py
```

## What the Script Does

1. **CREATE**: Sets a new secret named "my-secret" with value "my-secret-value"
2. **READ**: Retrieves the secret and displays its value
3. **UPDATE**: Updates the secret to a new value "updated-value" (creates a new version)
4. **DELETE**: Deletes the secret and purges it permanently (soft-delete enabled vault)

## Output Example

```
Connecting to Key Vault: https://your-vault.vault.azure.net/

============================================================
1. CREATE - Setting a new secret
============================================================
✓ Secret created successfully
  Name: my-secret
  Value: my-secret-value
  Version: abc123...
  Created: 2026-03-21 06:48:00

============================================================
2. READ - Retrieving the secret
============================================================
✓ Secret retrieved successfully
  Name: my-secret
  Value: my-secret-value
  Version: abc123...

============================================================
3. UPDATE - Updating the secret value
============================================================
✓ Secret updated successfully
  Name: my-secret
  New Value: updated-value
  New Version: def456...
  Updated: 2026-03-21 06:48:01

============================================================
4. DELETE - Deleting and purging the secret
============================================================
  Initiating deletion of 'my-secret'...
✓ Secret deleted successfully
  Name: my-secret
  Deleted Date: 2026-03-21 06:48:02
  Scheduled Purge Date: 2026-06-19 06:48:02
  Recovery ID: ...

  Purging deleted secret 'my-secret'...
✓ Secret purged successfully (permanently deleted)

============================================================
All CRUD operations completed successfully!
============================================================
```

## Error Handling

The script includes comprehensive error handling for:
- Missing environment variables
- Authentication failures
- Resource not found errors
- HTTP errors
- General exceptions

## References

- [Azure Key Vault Secrets Python SDK Documentation](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)
- [DefaultAzureCredential Documentation](https://learn.microsoft.com/en-us/python/api/azure-identity/azure.identity.defaultazurecredential)
- [Azure Key Vault Documentation](https://learn.microsoft.com/en-us/azure/key-vault/)
