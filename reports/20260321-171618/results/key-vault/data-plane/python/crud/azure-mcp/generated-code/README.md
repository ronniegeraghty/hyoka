# Azure Key Vault Secrets CRUD Operations

This script demonstrates all four CRUD operations on Azure Key Vault secrets using the Azure SDK for Python.

## Prerequisites

1. An Azure Key Vault with soft-delete enabled
2. Azure credentials configured (Azure CLI login, environment variables, or managed identity)
3. Appropriate permissions on the Key Vault (Secret Officer role or equivalent)

## Installation

Install required packages:

```bash
pip install -r requirements.txt
```

Or install directly:

```bash
pip install azure-keyvault-secrets azure-identity
```

## Usage

Run the script with your Key Vault URL:

```bash
python keyvault_crud.py https://<your-key-vault-name>.vault.azure.net/
```

## What the Script Does

1. **CREATE**: Creates a new secret named "my-secret" with value "my-secret-value"
2. **READ**: Retrieves the secret and displays its value
3. **UPDATE**: Updates the secret to a new value "updated-value"
4. **DELETE**: Soft-deletes the secret, then purges it permanently

## Authentication

The script uses `DefaultAzureCredential` which attempts authentication in this order:
1. Environment variables
2. Managed identity
3. Azure CLI
4. Azure PowerShell
5. Interactive browser

## Error Handling

The script includes proper error handling for:
- Resource not found errors
- HTTP response errors
- General exceptions

## Notes

- The script requires a Key Vault with soft-delete enabled for the purge operation
- Purging a secret permanently removes it and cannot be undone
- The script closes the credential properly in the finally block
