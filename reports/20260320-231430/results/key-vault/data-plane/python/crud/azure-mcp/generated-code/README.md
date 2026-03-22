# Azure Key Vault Secrets CRUD Demo

This script demonstrates all four CRUD (Create, Read, Update, Delete) operations on Azure Key Vault secrets using the Azure SDK for Python.

## Features

- **Create**: Creates a new secret called "my-secret" with value "my-secret-value"
- **Read**: Retrieves the secret and prints its value
- **Update**: Updates the secret to a new value "updated-value"
- **Delete**: Deletes the secret and purges it (for vaults with soft-delete enabled)

## Prerequisites

1. **Python 3.9 or later**
2. **Azure Key Vault** with soft-delete enabled
3. **Azure authentication** configured (one of the following):
   - Azure CLI: `az login`
   - Managed Identity (if running on Azure resources)
   - Environment variables (AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET)
   - Visual Studio Code Azure Account extension
4. **Required permissions** on the Key Vault:
   - secrets/set
   - secrets/get
   - secrets/delete
   - secrets/purge

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

Set the VAULT_URL environment variable to your Azure Key Vault URL:

```bash
export VAULT_URL='https://your-vault-name.vault.azure.net/'
```

## Usage

Run the script:

```bash
python keyvault_crud.py
```

Or make it executable and run directly:

```bash
chmod +x keyvault_crud.py
./keyvault_crud.py
```

## Example Output

```
Connecting to Azure Key Vault: https://your-vault.vault.azure.net/

✓ Successfully created SecretClient with DefaultAzureCredential

======================================================================
1. CREATE - Setting a new secret
======================================================================
Creating secret 'my-secret' with value 'my-secret-value'...
✓ Secret created successfully!
  Name: my-secret
  Value: my-secret-value
  Version: abc123...
  Created: 2026-03-21 06:14:33

======================================================================
2. READ - Retrieving the secret
======================================================================
Reading secret 'my-secret'...
✓ Secret retrieved successfully!
  Name: my-secret
  Value: my-secret-value
  Version: abc123...

======================================================================
3. UPDATE - Updating the secret value
======================================================================
Updating secret 'my-secret' to new value 'updated-value'...
✓ Secret updated successfully!
  Name: my-secret
  New Value: updated-value
  New Version: def456...

======================================================================
4. DELETE - Deleting and purging the secret
======================================================================
Deleting secret 'my-secret'...
✓ Secret deleted successfully!
Purging deleted secret 'my-secret'...
✓ Secret purged successfully!

======================================================================
All CRUD operations completed successfully!
======================================================================
```

## Error Handling

The script includes comprehensive error handling for:
- Missing environment variables
- Authentication failures
- Resource not found errors
- HTTP response errors
- Unexpected exceptions

## Authentication Methods

`DefaultAzureCredential` attempts authentication through multiple methods in order:
1. Environment variables
2. Managed Identity
3. Visual Studio Code
4. Azure CLI
5. Azure PowerShell
6. Interactive browser

## References

- [Azure Key Vault Secrets SDK Documentation](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)
- [Azure Identity SDK Documentation](https://learn.microsoft.com/en-us/python/api/overview/azure/identity-readme)
- [Azure Key Vault Overview](https://learn.microsoft.com/en-us/azure/key-vault/general/overview)
