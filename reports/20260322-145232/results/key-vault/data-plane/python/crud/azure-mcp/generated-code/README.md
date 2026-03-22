# Azure Key Vault CRUD Operations

This script demonstrates all four CRUD operations on Azure Key Vault secrets using the Azure SDK for Python.

## Operations Performed

1. **CREATE** - Set a new secret called "my-secret" with value "my-secret-value"
2. **READ** - Retrieve the secret and print its value
3. **UPDATE** - Update the secret to a new value "updated-value"
4. **DELETE** - Delete the secret (soft-delete)
5. **PURGE** - Permanently delete the secret (requires soft-delete enabled vault)

## Prerequisites

- Python 3.9 or later
- An Azure subscription
- An Azure Key Vault with soft-delete enabled
- Appropriate Azure credentials configured (see Authentication below)

## Installation

Install the required packages:

```bash
pip install -r requirements.txt
```

Or install packages directly:

```bash
pip install azure-keyvault-secrets azure-identity
```

## Required Packages

- **azure-keyvault-secrets** - Azure Key Vault Secrets client library
- **azure-identity** - Azure authentication library providing DefaultAzureCredential

## Authentication

This script uses `DefaultAzureCredential` which attempts to authenticate via multiple methods in order:

1. Environment variables (AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET)
2. Managed Identity (if deployed to Azure)
3. Azure CLI (if logged in via `az login`)
4. Azure PowerShell
5. Interactive browser

For local development, the easiest method is Azure CLI:

```bash
az login
```

## Configuration

Set the `VAULT_URL` environment variable to your Azure Key Vault URL:

```bash
export VAULT_URL='https://your-vault-name.vault.azure.net/'
```

## Usage

Run the script:

```bash
python keyvault_crud.py
```

## Required Permissions

Ensure your Azure identity has the following permissions on the Key Vault:

- `secrets/set` - For creating and updating secrets
- `secrets/get` - For reading secrets
- `secrets/delete` - For deleting secrets
- `secrets/purge` - For purging deleted secrets

You can assign these using Azure RBAC role "Key Vault Secrets Officer" or via Key Vault access policies.

## Error Handling

The script includes comprehensive error handling for:

- Missing environment variables
- Authentication failures
- Resource not found errors
- HTTP response errors
- General exceptions

## Sample Output

```
Connecting to Key Vault: https://my-vault.vault.azure.net/
✓ Successfully authenticated

==================================================
1. CREATE - Setting secret
==================================================
✓ Secret created successfully
  Name: my-secret
  Value: my-secret-value
  Version: abc123...

==================================================
2. READ - Retrieving secret
==================================================
✓ Secret retrieved successfully
  Name: my-secret
  Value: my-secret-value
  Content Type: None
  Enabled: True

==================================================
3. UPDATE - Updating secret value
==================================================
✓ Secret updated successfully
  Name: my-secret
  New Value: updated-value
  New Version: def456...

==================================================
4. DELETE - Deleting secret
==================================================
✓ Secret deleted successfully
  Name: my-secret
  Deleted Date: 2026-03-22 21:54:00+00:00
  Scheduled Purge Date: 2026-06-20 21:54:00+00:00
  Recovery ID: https://...

==================================================
5. PURGE - Permanently deleting secret
==================================================
✓ Secret purged successfully
  The secret 'my-secret' has been permanently deleted

==================================================
✓ All CRUD operations completed successfully!
==================================================
```

## Notes

- The vault must have soft-delete enabled for the purge operation to work
- Deleted secrets can be recovered before purging (not demonstrated in this script)
- Secret values are stored as strings in Azure Key Vault
- Each update creates a new version of the secret

## References

- [Azure Key Vault Secrets Python SDK Documentation](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)
- [Azure Identity Python SDK Documentation](https://learn.microsoft.com/en-us/python/api/overview/azure/identity-readme)
- [Azure Key Vault Overview](https://learn.microsoft.com/en-us/azure/key-vault/general/overview)
