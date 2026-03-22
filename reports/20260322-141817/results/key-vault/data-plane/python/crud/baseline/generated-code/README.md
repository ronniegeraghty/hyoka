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

1. An Azure subscription
2. An Azure Key Vault with soft-delete enabled
3. Appropriate permissions (secrets/set, secrets/get, secrets/delete, secrets/purge)
4. Authentication configured for DefaultAzureCredential

## Authentication

The script uses `DefaultAzureCredential`, which attempts authentication using multiple methods in the following order:

1. Environment variables (AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET)
2. Managed Identity (if running on Azure)
3. Azure CLI (if logged in via `az login`)
4. Azure PowerShell
5. Interactive browser

For local development, the easiest method is to use Azure CLI:

```bash
az login
```

## Usage

Set the VAULT_URL environment variable to your Key Vault URL:

```bash
export VAULT_URL="https://your-key-vault-name.vault.azure.net/"
```

Then run the script:

```bash
python keyvault_crud.py
```

## Operations Performed

The script performs the following operations in order:

1. **CREATE** - Creates a secret named "my-secret" with value "my-secret-value"
2. **READ** - Retrieves and displays the secret value
3. **UPDATE** - Updates the secret to a new value "updated-value"
4. **DELETE** - Deletes the secret (soft-delete)
5. **PURGE** - Permanently deletes the secret (requires soft-delete enabled vault)

## Expected Output

```
Connected to Key Vault: https://your-key-vault.vault.azure.net/

1. CREATE: Creating secret 'my-secret'...
   ✓ Secret created successfully
   - Name: my-secret
   - Version: abc123...

2. READ: Reading secret 'my-secret'...
   ✓ Secret retrieved successfully
   - Name: my-secret
   - Value: my-secret-value
   - Version: abc123...

3. UPDATE: Updating secret 'my-secret' to new value...
   ✓ Secret updated successfully
   - Name: my-secret
   - New Value: updated-value
   - New Version: def456...

4. DELETE: Deleting secret 'my-secret'...
   ✓ Secret deleted successfully
   - Name: my-secret
   - Deleted Date: ...
   - Scheduled Purge Date: ...

   PURGE: Permanently purging deleted secret 'my-secret'...
   ✓ Secret purged successfully (permanently deleted)

All CRUD operations completed successfully!
```

## Error Handling

The script includes comprehensive error handling for:
- Missing VAULT_URL environment variable
- Authentication failures
- Resource not found errors
- HTTP response errors
- Permission issues

## Notes

- The UPDATE operation creates a new version of the secret (Azure Key Vault maintains version history)
- Purging is only necessary in vaults with soft-delete enabled
- After deletion, secrets can be recovered before the scheduled purge date (unless purged)
- Purge operations are permanent and cannot be undone
