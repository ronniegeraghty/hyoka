# Azure Key Vault CRUD Operations Demo

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

## Prerequisites

1. **Azure Key Vault**: You need an existing Azure Key Vault with soft-delete enabled
2. **Authentication**: DefaultAzureCredential will attempt authentication in this order:
   - Environment variables (AZURE_TENANT_ID, AZURE_CLIENT_ID, AZURE_CLIENT_SECRET)
   - Managed Identity (if running on Azure)
   - Azure CLI (run `az login` first)
   - Azure PowerShell
   - Visual Studio Code

3. **Permissions**: Your identity needs the following Key Vault permissions:
   - secrets/set
   - secrets/get
   - secrets/delete
   - secrets/purge

## Setup

1. Set your Key Vault URL as an environment variable:

```bash
export AZURE_KEY_VAULT_URL="https://your-key-vault-name.vault.azure.net/"
```

2. Authenticate using Azure CLI (easiest for local development):

```bash
az login
```

## Usage

Run the script:

```bash
python keyvault_crud_demo.py
```

## What the Script Does

1. **CREATE**: Creates a new secret called "my-secret" with value "my-secret-value"
2. **READ**: Retrieves and displays the secret value
3. **UPDATE**: Updates the secret to a new value "updated-value"
4. **DELETE**: Soft-deletes the secret (for vaults with soft-delete enabled)
5. **PURGE**: Permanently deletes the secret from the vault

## Error Handling

The script includes comprehensive error handling for:
- Missing environment variables
- Authentication failures
- Resource not found errors
- HTTP response errors
- General exceptions

## Documentation Reference

This script is based on the official Azure SDK for Python documentation:
- https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme
- https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient
