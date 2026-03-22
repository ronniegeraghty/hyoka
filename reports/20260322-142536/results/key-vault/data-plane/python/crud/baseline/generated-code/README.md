# Azure Key Vault Secrets CRUD Operations

This script demonstrates all four CRUD operations on Azure Key Vault secrets using the Azure SDK for Python.

## Prerequisites

1. An Azure subscription
2. An Azure Key Vault with soft-delete enabled
3. Appropriate permissions to create, read, update, delete, and purge secrets
4. Python 3.7 or later

## Installation

Install the required packages:

```bash
pip install -r requirements.txt
```

Or install them individually:

```bash
pip install azure-keyvault-secrets azure-identity
```

## Authentication

This script uses `DefaultAzureCredential`, which attempts authentication through multiple methods in order:

1. Environment variables
2. Managed Identity
3. Visual Studio Code
4. Azure CLI
5. Azure PowerShell
6. Interactive browser

The easiest way to authenticate locally is using Azure CLI:

```bash
az login
```

## Configuration

Set the Key Vault URL as an environment variable:

```bash
export AZURE_KEY_VAULT_URL='https://your-vault-name.vault.azure.net/'
```

On Windows (PowerShell):

```powershell
$env:AZURE_KEY_VAULT_URL='https://your-vault-name.vault.azure.net/'
```

## Usage

Run the script:

```bash
python key_vault_crud.py
```

## Operations Performed

1. **CREATE**: Creates a new secret called "my-secret" with value "my-secret-value"
2. **READ**: Reads the secret back and prints its value
3. **UPDATE**: Updates the secret to a new value "updated-value"
4. **DELETE**: Deletes the secret (soft-delete)
5. **PURGE**: Permanently purges the deleted secret

## Required Permissions

Your Azure identity needs the following Key Vault permissions:

- Get (secrets)
- List (secrets)
- Set (secrets)
- Delete (secrets)
- Purge (secrets)

## Notes

- The Key Vault must have soft-delete enabled for the purge operation
- Purging a secret permanently deletes it and cannot be undone
- Each update creates a new version of the secret
