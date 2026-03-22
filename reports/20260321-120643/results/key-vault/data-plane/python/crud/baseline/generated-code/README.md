# Azure Key Vault Secrets CRUD Operations Demo

This Python script demonstrates all four CRUD operations on Azure Key Vault secrets using the official Azure SDK for Python.

## Operations Demonstrated

1. **CREATE**: Set a new secret called "my-secret" with value "my-secret-value"
2. **READ**: Retrieve the secret and print its value
3. **UPDATE**: Update the secret to a new value "updated-value"
4. **DELETE**: Delete the secret and purge it (for soft-delete enabled vaults)

## Prerequisites

- Python 3.9 or later
- An Azure Key Vault instance with soft-delete enabled
- Azure credentials configured for authentication
- Required Key Vault permissions:
  - `secrets/set`
  - `secrets/get`
  - `secrets/delete`
  - `secrets/purge`

## Installation

Install the required packages:

```bash
pip install -r requirements.txt
```

Or install directly:

```bash
pip install azure-keyvault-secrets azure-identity
```

## Configuration

Set the `VAULT_URL` environment variable to your Azure Key Vault URL:

```bash
export VAULT_URL='https://your-key-vault-name.vault.azure.net/'
```

## Authentication

The script uses `DefaultAzureCredential` which automatically tries multiple authentication methods in order:

1. Environment variables
2. Managed Identity
3. Azure CLI credentials
4. Azure PowerShell credentials
5. Interactive browser authentication

For local development, the easiest method is to authenticate using Azure CLI:

```bash
az login
```

## Usage

Run the script:

```bash
python keyvault_crud_demo.py
```

## Expected Output

```
Connected to Key Vault: https://your-key-vault-name.vault.azure.net/

======================================================================
1. CREATE: Setting a new secret
======================================================================
✓ Secret created successfully!
  Name: my-secret
  Value: my-secret-value
  Version: abc123...

======================================================================
2. READ: Retrieving the secret
======================================================================
✓ Secret retrieved successfully!
  Name: my-secret
  Value: my-secret-value
  Version: abc123...

======================================================================
3. UPDATE: Updating the secret to a new value
======================================================================
✓ Secret updated successfully!
  Name: my-secret
  Value: updated-value
  New Version: def456...

======================================================================
4. DELETE: Deleting and purging the secret
======================================================================
Deleting secret 'my-secret'...
✓ Secret soft-deleted successfully!
  Name: my-secret
  Deleted Date: 2026-03-21 19:06:46
  Scheduled Purge Date: 2026-06-19 19:06:46

Purging secret 'my-secret' permanently...
✓ Secret purged successfully!
  The secret 'my-secret' has been permanently deleted.

======================================================================
All CRUD operations completed successfully!
======================================================================
```

## Error Handling

The script includes comprehensive error handling for:

- Missing environment variables
- Authentication failures
- Permission issues
- Resource not found errors
- HTTP response errors

## Notes

- The `set_secret()` method creates a new secret if it doesn't exist, or creates a new version if it does exist
- For vaults without soft-delete enabled, the `begin_delete_secret()` operation is permanent and `purge_deleted_secret()` is not needed
- Each secret update creates a new version while preserving the history
- Purging a secret is permanent and cannot be recovered

## References

- [Azure Key Vault Secrets Documentation](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)
- [SecretClient API Reference](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient)
- [DefaultAzureCredential Documentation](https://learn.microsoft.com/en-us/python/api/azure-identity/azure.identity.defaultazurecredential)
