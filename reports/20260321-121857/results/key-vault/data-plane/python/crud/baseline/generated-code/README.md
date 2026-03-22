# Azure Key Vault Secrets CRUD Operations

This script demonstrates all four CRUD operations on Azure Key Vault secrets using the Azure SDK for Python.

## Operations Demonstrated

1. **CREATE** - Create a new secret called "my-secret" with value "my-secret-value"
2. **READ** - Read the secret back and print its value
3. **UPDATE** - Update the secret to a new value "updated-value"
4. **DELETE** - Delete the secret and purge it (for soft-delete enabled vaults)

## Installation

Install the required packages using pip:

```bash
pip install -r requirements.txt
```

Or install packages individually:

```bash
pip install azure-keyvault-secrets azure-identity
```

## Prerequisites

1. **Azure Key Vault**: You need an existing Azure Key Vault with:
   - Soft-delete enabled (required for purge operation)
   - The vault URL (e.g., `https://your-vault-name.vault.azure.net/`)

2. **Authentication**: Configure one of the following authentication methods:
   
   - **Azure CLI** (easiest for local development):
     ```bash
     az login
     ```
   
   - **Environment Variables** (for service principal):
     ```bash
     export AZURE_CLIENT_ID="your-client-id"
     export AZURE_TENANT_ID="your-tenant-id"
     export AZURE_CLIENT_SECRET="your-client-secret"
     ```
   
   - **Managed Identity** (when running on Azure resources)

3. **Permissions**: Your identity needs the following Key Vault permissions:
   - `secrets/set` - To create and update secrets
   - `secrets/get` - To read secrets
   - `secrets/delete` - To delete secrets
   - `secrets/purge` - To permanently delete secrets

## Usage

Set the VAULT_URL environment variable and run the script:

```bash
export VAULT_URL="https://your-vault-name.vault.azure.net/"
python key_vault_crud.py
```

## Expected Output

The script will:

1. Connect to your Key Vault
2. Create a secret named "my-secret" with value "my-secret-value"
3. Retrieve and display the secret
4. Update the secret to value "updated-value"
5. Update secret metadata (content type)
6. Delete the secret (soft-delete)
7. Purge the secret (permanent deletion)
8. Verify the secret no longer exists

## Error Handling

The script includes comprehensive error handling for:

- Missing VAULT_URL environment variable
- Authentication failures
- Insufficient permissions (403 errors)
- Resource not found errors
- General HTTP errors

## Python Version

Requires Python 3.9 or later.

## Documentation

Based on official Azure SDK for Python documentation:
- [Azure Key Vault Secrets Client Library](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)
- [SecretClient API Reference](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient)
- [DefaultAzureCredential](https://learn.microsoft.com/en-us/python/api/azure-identity/azure.identity.defaultazurecredential)
