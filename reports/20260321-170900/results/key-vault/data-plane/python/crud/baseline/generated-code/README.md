# Azure Key Vault Secret CRUD Operations

This script demonstrates all four CRUD operations on Azure Key Vault secrets using the Azure SDK for Python.

## Prerequisites

1. An Azure subscription
2. An Azure Key Vault with soft-delete enabled
3. Appropriate permissions (Secret Get, Set, Delete, Purge)
4. Python 3.7 or higher

## Installation

Install the required packages:

```bash
pip install -r requirements.txt
```

Or install individually:

```bash
pip install azure-keyvault-secrets azure-identity
```

## Authentication

This script uses `DefaultAzureCredential`, which attempts multiple authentication methods in order:
1. Environment variables
2. Managed Identity
3. Azure CLI
4. Azure PowerShell
5. Interactive browser

For local development, the easiest method is Azure CLI:

```bash
az login
```

## Usage

Run the script:

```bash
python key_vault_crud.py
```

When prompted, enter your Key Vault URL in the format:
```
https://<your-key-vault-name>.vault.azure.net/
```

## Operations Performed

1. **CREATE**: Creates a secret named "my-secret" with value "my-secret-value"
2. **READ**: Retrieves the secret and prints its value
3. **UPDATE**: Updates the secret to a new value "updated-value"
4. **DELETE**: Soft-deletes the secret and then purges it permanently

## Error Handling

The script includes error handling for:
- Resource not found errors
- HTTP errors (401 Unauthorized, 403 Forbidden, etc.)
- General exceptions

## Required Permissions

Ensure your identity has the following Key Vault permissions:
- `Secret Get`
- `Secret Set`
- `Secret Delete`
- `Secret Purge`

You can assign these via Azure RBAC role "Key Vault Secrets Officer" or through access policies.
