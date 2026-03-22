# Azure Key Vault Secrets CRUD Operations

This script demonstrates all four CRUD (Create, Read, Update, Delete) operations on Azure Key Vault secrets using the Azure SDK for Python.

## Features

The script performs the following operations in sequence:

1. **CREATE**: Creates a new secret called "my-secret" with value "my-secret-value"
2. **READ**: Retrieves the secret and displays its value
3. **UPDATE**: Updates the secret to a new value "updated-value"
4. **DELETE**: Deletes the secret (soft-delete)
5. **PURGE**: Permanently purges the deleted secret (for soft-delete enabled vaults)

## Prerequisites

### Azure Resources
- An Azure subscription
- An Azure Key Vault with soft-delete enabled
- Appropriate permissions on the Key Vault:
  - `secrets/set` - To create and update secrets
  - `secrets/get` - To read secrets
  - `secrets/delete` - To delete secrets
  - `secrets/purge` - To permanently delete secrets

### Python Requirements
- Python 3.9 or later

## Installation

Install the required packages using pip:

```bash
pip install -r requirements.txt
```

Or install them directly:

```bash
pip install azure-keyvault-secrets azure-identity
```

## Required Packages

- **azure-keyvault-secrets**: Azure Key Vault Secrets client library
- **azure-identity**: Azure authentication library providing DefaultAzureCredential

## Authentication

The script uses `DefaultAzureCredential` from the Azure Identity library, which attempts multiple authentication methods in the following order:

1. **Environment variables** - `AZURE_CLIENT_ID`, `AZURE_TENANT_ID`, `AZURE_CLIENT_SECRET`
2. **Managed Identity** - If deployed to Azure with managed identity enabled
3. **Azure CLI** - If logged in via `az login`
4. **Azure PowerShell** - If logged in via `Connect-AzAccount`
5. **Interactive browser** - As a fallback

### Quick Setup with Azure CLI

```bash
# Login to Azure
az login

# Set your subscription (if you have multiple)
az account set --subscription "your-subscription-id"
```

## Configuration

Set the Key Vault URL as an environment variable:

### Linux/macOS:
```bash
export AZURE_KEY_VAULT_URL="https://your-vault-name.vault.azure.net/"
```

### Windows (PowerShell):
```powershell
$env:AZURE_KEY_VAULT_URL="https://your-vault-name.vault.azure.net/"
```

### Windows (Command Prompt):
```cmd
set AZURE_KEY_VAULT_URL=https://your-vault-name.vault.azure.net/
```

## Usage

Run the script:

```bash
python keyvault_crud.py
```

## Example Output

```
Connecting to Key Vault: https://my-vault.vault.azure.net/

============================================================
1. CREATE OPERATION
============================================================
Creating secret 'my-secret' with value 'my-secret-value'...
✓ Secret created successfully!
  Name: my-secret
  Value: my-secret-value
  Version: abc123def456
  Created: 2026-03-22 00:00:00+00:00

============================================================
2. READ OPERATION
============================================================
Reading secret 'my-secret'...
✓ Secret retrieved successfully!
  Name: my-secret
  Value: my-secret-value
  Version: abc123def456
  Content Type: None
  Enabled: True

============================================================
3. UPDATE OPERATION
============================================================
Updating secret 'my-secret' to new value 'updated-value'...
✓ Secret updated successfully!
  Name: my-secret
  New Value: updated-value
  New Version: def789ghi012
  Updated: 2026-03-22 00:01:00+00:00

Verifying update by reading secret again...
✓ Verified value: updated-value

============================================================
4. DELETE OPERATION
============================================================
Deleting secret 'my-secret'...
✓ Secret deleted successfully!
  Name: my-secret
  Deleted Date: 2026-03-22 00:02:00+00:00
  Scheduled Purge Date: 2026-04-21 00:02:00+00:00
  Recovery ID: https://my-vault.vault.azure.net/deletedsecrets/my-secret

============================================================
5. PURGE OPERATION (Permanent Deletion)
============================================================
Purging deleted secret 'my-secret' permanently...
✓ Secret purged successfully!
  The secret 'my-secret' has been permanently deleted.
  It cannot be recovered.

============================================================
All CRUD operations completed successfully!
============================================================
```

## Error Handling

The script includes comprehensive error handling for:

- **Authentication errors**: When credentials are invalid or missing
- **Missing Key Vault URL**: When the environment variable is not set
- **Resource not found**: When trying to access a non-existent secret
- **Permission errors**: When the user lacks necessary permissions
- **General HTTP errors**: For other API-related issues

## Troubleshooting

### Authentication Failed
```
✗ Authentication failed
```

**Solutions**:
- Ensure you're logged in: `az login`
- Verify your account has access to the Key Vault
- Check that the Key Vault's access policies or RBAC includes your account

### Permission Denied
```
✗ Failed to create secret: Forbidden
```

**Solutions**:
- Add appropriate access policies in the Azure Portal
- Or assign the "Key Vault Secrets Officer" role if using RBAC

### Purge Operation Failed
```
Note: Purge operation failed or not needed
```

**Reason**: This typically means soft-delete is not enabled on the vault. In this case, the delete operation is already permanent.

## Key Vault Setup

To create a Key Vault with soft-delete enabled:

```bash
# Create a resource group
az group create --name myResourceGroup --location eastus

# Create a Key Vault with soft-delete enabled (default since 2020)
az keyvault create \
  --name myKeyVault \
  --resource-group myResourceGroup \
  --location eastus

# Grant yourself permissions
az keyvault set-policy \
  --name myKeyVault \
  --upn user@example.com \
  --secret-permissions get set delete purge list
```

## API Reference

Based on the Azure SDK for Python documentation:

- [SecretClient](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient)
- [DefaultAzureCredential](https://learn.microsoft.com/en-us/python/api/azure-identity/azure.identity.defaultazurecredential)
- [Azure Key Vault Overview](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)

## Notes

- **Versioning**: Each time you update a secret using `set_secret()`, a new version is created
- **Soft-delete**: When soft-delete is enabled, deleted secrets can be recovered within the retention period
- **Purge**: The purge operation permanently deletes the secret and cannot be undone
- **Connection cleanup**: The script properly closes the client and credential connections

## License

This is a demonstration script for educational purposes.
