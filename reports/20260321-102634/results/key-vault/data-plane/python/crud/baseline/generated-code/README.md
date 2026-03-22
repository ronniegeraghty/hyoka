# Azure Key Vault Secrets CRUD Operations Demo

This script demonstrates all four CRUD operations on Azure Key Vault secrets using the Azure SDK for Python.

## Required Packages

Install the required packages using pip:

```bash
pip install azure-keyvault-secrets azure-identity
```

**Package Details:**
- `azure-keyvault-secrets`: Azure Key Vault Secrets client library
- `azure-identity`: Azure authentication library (includes DefaultAzureCredential)

## Prerequisites

1. **Azure Key Vault**: An existing Azure Key Vault with soft-delete enabled
2. **Python**: Python 3.9 or later
3. **Permissions**: Your credential needs the following permissions:
   - `secrets/get` - Read secrets
   - `secrets/set` - Create and update secrets
   - `secrets/delete` - Delete secrets
   - `secrets/purge` - Purge deleted secrets

## Authentication

This script uses `DefaultAzureCredential`, which attempts authentication through multiple methods in the following order:

1. Environment variables (AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET)
2. Managed Identity
3. Visual Studio Code
4. Azure CLI (`az login`)
5. Azure PowerShell
6. Interactive browser

For local development, the easiest method is to use Azure CLI:

```bash
az login
```

## Usage

Set the `VAULT_URL` environment variable to your Key Vault URL:

```bash
export VAULT_URL='https://your-vault-name.vault.azure.net/'
```

Run the script:

```bash
python keyvault_crud_demo.py
```

## Operations Performed

The script performs the following operations:

### 1. CREATE
Creates a new secret named "my-secret" with value "my-secret-value" using `set_secret()`.

### 2. READ
Retrieves the secret using `get_secret()` and displays its properties.

### 3. UPDATE
- Updates the secret value to "updated-value" using `set_secret()` (creates a new version)
- Updates metadata (content type and tags) using `update_secret_properties()`

### 4. DELETE
- Soft-deletes the secret using `begin_delete_secret()` (returns a poller)
- Permanently purges the deleted secret using `purge_deleted_secret()`

## Error Handling

The script includes comprehensive error handling for:
- Missing environment variables
- Authentication failures
- Resource not found errors
- HTTP response errors
- Permission issues

## Notes

- **Soft-delete**: The purge operation requires soft-delete to be enabled on your Key Vault
- **Versions**: Each call to `set_secret()` with an existing secret name creates a new version
- **Polling**: `begin_delete_secret()` returns a poller because deletion may take several seconds
- **Permissions**: If you lack purge permissions, the script will fail at the purge step

## References

- [Azure Key Vault Secrets Python SDK Documentation](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)
- [SecretClient API Reference](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient)
- [DefaultAzureCredential Documentation](https://learn.microsoft.com/en-us/python/api/azure-identity/azure.identity.defaultazurecredential)
