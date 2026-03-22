# Azure Key Vault CRUD Operations

This script demonstrates all four CRUD operations on Azure Key Vault secrets using the Azure SDK for Python.

## Required Packages

Install the required packages using pip:

```bash
pip install azure-keyvault-secrets azure-identity
```

### Package Details

- **azure-keyvault-secrets**: Azure Key Vault Secrets client library for Python
- **azure-identity**: Azure Active Directory authentication library (provides DefaultAzureCredential)

## Prerequisites

1. **Python 3.9 or later** is required
2. **An Azure Key Vault** - Create one if you don't have it:
   ```bash
   az keyvault create --name <your-vault-name> --resource-group <your-resource-group> --location <location>
   ```
3. **Proper permissions** - Your identity needs the following Key Vault access policies:
   - `secrets/set` - Create and update secrets
   - `secrets/get` - Read secrets
   - `secrets/delete` - Delete secrets
   - `secrets/purge` - Purge deleted secrets (for soft-delete enabled vaults)

## Authentication Setup

The script uses `DefaultAzureCredential`, which tries multiple authentication methods in order:

### Option 1: Azure CLI (Recommended for local development)
```bash
az login
```

### Option 2: Environment Variables (For service principals)
```bash
export AZURE_CLIENT_ID="<your-client-id>"
export AZURE_TENANT_ID="<your-tenant-id>"
export AZURE_CLIENT_SECRET="<your-client-secret>"
```

### Option 3: Managed Identity
Automatically works when running on Azure resources (VMs, App Service, etc.)

## Usage

1. Set the `VAULT_URL` environment variable:
   ```bash
   export VAULT_URL="https://<your-vault-name>.vault.azure.net/"
   ```

2. Run the script:
   ```bash
   python keyvault_crud_operations.py
   ```

## What the Script Does

The script performs the following operations in sequence:

1. **CREATE** - Creates a new secret named "my-secret" with value "my-secret-value"
2. **READ** - Retrieves the secret and prints its value
3. **UPDATE** - Updates the secret to a new value "updated-value"
4. **DELETE** - Deletes the secret and purges it (permanent deletion for soft-delete enabled vaults)

## Expected Output

```
Connecting to Key Vault: https://your-vault.vault.azure.net/

==================================================
1. CREATE - Setting a new secret
==================================================
✓ Secret created successfully!
  Name: my-secret
  Value: my-secret-value
  Version: <version-id>
  Created: <timestamp>

==================================================
2. READ - Retrieving the secret
==================================================
✓ Secret retrieved successfully!
  Name: my-secret
  Value: my-secret-value
  Version: <version-id>

==================================================
3. UPDATE - Updating the secret value
==================================================
✓ Secret updated successfully!
  Name: my-secret
  New Value: updated-value
  New Version: <new-version-id>
  Updated: <timestamp>

==================================================
4. DELETE - Deleting the secret
==================================================
✓ Secret deletion initiated!
  Name: my-secret
  Deleted Date: <timestamp>
  Scheduled Purge Date: <timestamp>
  Recovery ID: <recovery-id>

Purging the deleted secret...
✓ Secret purged successfully!
  The secret 'my-secret' has been permanently deleted.

==================================================
All CRUD operations completed successfully!
==================================================
```

## Error Handling

The script includes comprehensive error handling for:

- **Authentication errors** - Invalid credentials or authentication failures
- **Permission errors** - Insufficient Key Vault access policies
- **Resource not found** - Secret doesn't exist
- **HTTP errors** - Network or service errors
- **General exceptions** - Unexpected errors

## Key Vault Soft-Delete

If your Key Vault has soft-delete enabled (default for new vaults):
- Deleted secrets are not immediately removed
- They remain in a "deleted" state for the retention period (default: 90 days)
- Use `purge_deleted_secret()` to permanently delete before the scheduled purge date
- Without purge, the secret name cannot be reused until after the retention period

## References

- [Azure Key Vault Secrets SDK Documentation](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)
- [Azure Identity SDK Documentation](https://learn.microsoft.com/en-us/python/api/overview/azure/identity-readme)
- [Azure Key Vault Documentation](https://learn.microsoft.com/en-us/azure/key-vault/)
