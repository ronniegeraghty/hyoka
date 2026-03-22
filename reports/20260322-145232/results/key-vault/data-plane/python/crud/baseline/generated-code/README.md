# Azure Key Vault Secrets CRUD Operations

This script demonstrates all four CRUD operations on Azure Key Vault secrets using the Azure SDK for Python.

## Operations Performed

1. **CREATE** - Creates a new secret named "my-secret" with value "my-secret-value"
2. **READ** - Retrieves the secret and prints its value
3. **UPDATE** - Updates the secret to a new value "updated-value"
4. **DELETE** - Deletes the secret and purges it (for soft-delete enabled vaults)

## Prerequisites

- Python 3.9 or later
- An Azure subscription
- An Azure Key Vault (with soft-delete enabled for purge operation)
- Appropriate permissions on the Key Vault:
  - `secrets/set` - for creating and updating secrets
  - `secrets/get` - for reading secrets
  - `secrets/delete` - for deleting secrets
  - `secrets/purge` - for purging deleted secrets

## Installation

Install the required packages:

```bash
pip install -r requirements.txt
```

Or install individually:

```bash
pip install azure-keyvault-secrets azure-identity
```

## Authentication

The script uses `DefaultAzureCredential` which supports multiple authentication methods in the following order:

1. **Environment variables** - Set these for service principal authentication:
   - `AZURE_TENANT_ID`
   - `AZURE_CLIENT_ID`
   - `AZURE_CLIENT_SECRET`

2. **Managed Identity** - Works automatically when running on Azure resources

3. **Azure CLI** - Uses credentials from `az login`

4. **Azure PowerShell** - Uses credentials from `Connect-AzAccount`

5. **Interactive browser** - Opens a browser for user authentication

For local development, the easiest method is to use Azure CLI:

```bash
az login
```

## Usage

1. Set the Key Vault URL environment variable:

```bash
export AZURE_KEY_VAULT_URL="https://your-key-vault-name.vault.azure.net/"
```

On Windows (PowerShell):
```powershell
$env:AZURE_KEY_VAULT_URL="https://your-key-vault-name.vault.azure.net/"
```

2. Run the script:

```bash
python keyvault_crud.py
```

## Expected Output

```
Using Key Vault: https://your-key-vault-name.vault.azure.net/

✓ Successfully initialized SecretClient with DefaultAzureCredential

============================================================
1. CREATE - Setting a new secret
============================================================
✓ Secret created successfully
  Name: my-secret
  Value: my-secret-value
  Version: abc123...
  Created: 2026-03-22 21:53:00

============================================================
2. READ - Retrieving the secret
============================================================
✓ Secret retrieved successfully
  Name: my-secret
  Value: my-secret-value
  Version: abc123...
  Enabled: True

============================================================
3. UPDATE - Updating the secret to a new value
============================================================
✓ Secret updated successfully
  Name: my-secret
  New Value: updated-value
  New Version: def456...
  Updated: 2026-03-22 21:53:01

============================================================
4. DELETE - Deleting and purging the secret
============================================================
Step 1: Deleting secret (soft delete)...
✓ Secret deleted successfully
  Name: my-secret
  Deleted Date: 2026-03-22 21:53:02
  Scheduled Purge Date: 2026-06-20 21:53:02
  Recovery ID: https://...

Step 2: Purging secret (permanent deletion)...
✓ Secret purged successfully
  The secret 'my-secret' has been permanently deleted.

============================================================
All CRUD operations completed successfully!
============================================================
```

## Error Handling

The script includes comprehensive error handling for:

- Missing environment variables
- Authentication failures
- HTTP errors (with status codes)
- Resource not found errors
- Permission errors

## References

- [Azure Key Vault Secrets Python SDK Documentation](https://learn.microsoft.com/python/api/overview/azure/keyvault-secrets-readme)
- [Azure Identity Python SDK Documentation](https://learn.microsoft.com/python/api/overview/azure/identity-readme)
- [DefaultAzureCredential Documentation](https://learn.microsoft.com/python/api/azure-identity/azure.identity.defaultazurecredential)
