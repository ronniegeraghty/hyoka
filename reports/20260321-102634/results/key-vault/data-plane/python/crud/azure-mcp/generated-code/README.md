# Azure Key Vault Secrets - CRUD Operations Demo

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

- **azure-keyvault-secrets** (>=4.8.0): Azure Key Vault Secrets client library
- **azure-identity** (>=1.15.0): Azure authentication library (includes DefaultAzureCredential)

## Prerequisites

1. **Python 3.9 or later** is required
2. **Azure Key Vault** with soft-delete enabled
3. **Azure Authentication** - DefaultAzureCredential supports multiple authentication methods (in order of precedence):
   - Environment variables (AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET)
   - Managed Identity (when running on Azure)
   - Azure CLI authentication (`az login`)
   - Azure PowerShell authentication
   - Interactive browser authentication

4. **Required Permissions** on the Key Vault:
   - secrets/set
   - secrets/get
   - secrets/delete
   - secrets/purge

## Setup

1. Set your Key Vault URL as an environment variable:

```bash
export VAULT_URL='https://your-key-vault-name.vault.azure.net/'
```

On Windows (PowerShell):
```powershell
$env:VAULT_URL='https://your-key-vault-name.vault.azure.net/'
```

On Windows (Command Prompt):
```cmd
set VAULT_URL=https://your-key-vault-name.vault.azure.net/
```

2. Authenticate with Azure (if using Azure CLI):

```bash
az login
```

## Running the Script

```bash
python keyvault_secrets_crud.py
```

## What the Script Does

The script performs four CRUD operations on a secret named "my-secret":

### 1. **CREATE** - Set a new secret
- Creates a secret with name "my-secret" and value "my-secret-value"
- Uses `set_secret()` method
- Displays the secret name, value, version, and creation date

### 2. **READ** - Retrieve the secret
- Retrieves the secret using `get_secret()` method
- Displays the secret value and metadata (version, enabled status, etc.)

### 3. **UPDATE** - Change the secret value
- Updates the secret to a new value "updated-value"
- Uses `set_secret()` again (creates a new version)
- Displays the new version and updated timestamp
- Verifies the update by reading it back

### 4. **DELETE & PURGE** - Remove the secret
- Deletes the secret using `begin_delete_secret()` (soft delete)
- Waits for deletion to complete
- Purges the secret using `purge_deleted_secret()` (permanent deletion)
- Verifies the secret is completely removed

## Error Handling

The script includes comprehensive error handling for:
- Missing environment variables
- Authentication failures
- Resource not found errors
- HTTP response errors
- Service request errors
- Network connectivity issues

## Notes

- The script uses **soft-delete** enabled vaults. If soft-delete is disabled, `begin_delete_secret()` is permanent and `purge_deleted_secret()` is not needed.
- `begin_delete_secret()` returns a poller that allows waiting for deletion completion
- `set_secret()` creates a new version if the secret already exists
- All secrets are versioned in Azure Key Vault
- The script cleans up resources properly by closing the client and credential connections

## Documentation References

- [Azure Key Vault Secrets Python SDK](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)
- [SecretClient API Reference](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient)
- [DefaultAzureCredential](https://learn.microsoft.com/en-us/python/api/azure-identity/azure.identity.defaultazurecredential)
