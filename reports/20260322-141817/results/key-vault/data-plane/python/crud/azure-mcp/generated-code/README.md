# Azure Key Vault Secrets CRUD Demo

This script demonstrates all four CRUD operations on Azure Key Vault secrets using the Azure SDK for Python.

## Operations Demonstrated

1. **CREATE** - Set a new secret called "my-secret" with value "my-secret-value"
2. **READ** - Retrieve the secret and print its value
3. **UPDATE** - Update the secret to a new value "updated-value"
4. **DELETE** - Delete the secret and purge it (for soft-delete enabled vaults)

## Prerequisites

- Python 3.9 or later
- An Azure Key Vault with soft-delete enabled
- Appropriate permissions on the Key Vault:
  - `secrets/set` - To create and update secrets
  - `secrets/get` - To read secrets
  - `secrets/delete` - To delete secrets
  - `secrets/purge` - To permanently delete secrets

## Authentication

The script uses `DefaultAzureCredential` which automatically attempts multiple authentication methods:

1. **Environment variables** (recommended for development)
2. **Managed Identity** (for Azure-hosted applications)
3. **Azure CLI** (if logged in via `az login`)
4. **Azure PowerShell**
5. **Interactive browser**

### Quick Setup with Azure CLI

```bash
az login
```

## Installation

Install required packages:

```bash
pip install -r requirements.txt
```

Or manually:

```bash
pip install azure-keyvault-secrets azure-identity
```

## Usage

Set your Key Vault URL as an environment variable:

```bash
export AZURE_KEY_VAULT_URL='https://your-vault-name.vault.azure.net/'
```

Run the script:

```bash
python keyvault_crud_demo.py
```

## Expected Output

The script will:
- Authenticate using DefaultAzureCredential
- Create a secret named "my-secret"
- Read and display the secret value
- Update the secret to a new value
- Verify the update
- Delete the secret (soft-delete)
- Purge the secret (permanent deletion)
- Verify the secret no longer exists

## Error Handling

The script includes comprehensive error handling for:
- Authentication failures
- Missing environment variables
- Resource not found errors
- HTTP response errors
- Permission issues

## References

- [Azure Key Vault Secrets SDK Documentation](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)
- [DefaultAzureCredential Documentation](https://learn.microsoft.com/en-us/python/api/azure-identity/azure.identity.defaultazurecredential)
- [Azure Key Vault Overview](https://learn.microsoft.com/en-us/azure/key-vault/general/overview)
