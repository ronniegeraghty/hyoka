# Azure Key Vault Secrets CRUD Operations

This script demonstrates all four CRUD operations on Azure Key Vault secrets using the official Azure SDK for Python.

## Features

✅ **Create** - Set a new secret with name "my-secret" and value "my-secret-value"  
✅ **Read** - Retrieve the secret and print its value  
✅ **Update** - Update the secret to a new value "updated-value"  
✅ **Delete** - Delete the secret and purge it (for soft-delete enabled vaults)

## Installation

Install the required packages:

```bash
pip install -r requirements.txt
```

Or install manually:

```bash
pip install azure-keyvault-secrets azure-identity
```

## Prerequisites

1. **Azure subscription**
2. **Azure Key Vault** with soft-delete enabled
3. **Authentication** - Configure one of the following for DefaultAzureCredential:
   - Azure CLI: Run `az login`
   - Environment variables (Service Principal)
   - Managed Identity (when running in Azure)
   - Interactive browser authentication

4. **Permissions** - Your account/identity needs the following Key Vault permissions:
   - `secrets/get`
   - `secrets/set`
   - `secrets/delete`
   - `secrets/purge`

## Usage

Set the VAULT_URL environment variable to your Key Vault URL:

```bash
export VAULT_URL="https://your-key-vault-name.vault.azure.net/"
```

Run the script:

```bash
python azure_keyvault_crud.py
```

## Example Output

```
Connecting to Key Vault: https://my-key-vault.vault.azure.net/
------------------------------------------------------------
✓ Successfully authenticated with DefaultAzureCredential

1. CREATE - Setting a new secret
------------------------------------------------------------
✓ Created secret: my-secret
  Value: my-secret-value
  Version: abc123...
  Created on: 2026-03-21 06:38:20.123456+00:00

2. READ - Retrieving the secret
------------------------------------------------------------
✓ Retrieved secret: my-secret
  Value: my-secret-value
  Version: abc123...

3. UPDATE - Updating the secret to a new value
------------------------------------------------------------
✓ Updated secret: my-secret
  New value: updated-value
  New version: def456...
  Updated on: 2026-03-21 06:38:21.234567+00:00

4. DELETE - Deleting and purging the secret
------------------------------------------------------------
✓ Deleted secret: my-secret
  Deleted on: 2026-03-21 06:38:22.345678+00:00
  Scheduled purge date: 2026-04-20 06:38:22.345678+00:00
✓ Purged secret 'my-secret' permanently

============================================================
All CRUD operations completed successfully!
============================================================
```

## Error Handling

The script includes comprehensive error handling for:
- Missing environment variables
- Authentication failures
- Resource not found errors
- HTTP response errors
- Permission issues

## Documentation References

- [Azure Key Vault Secrets SDK for Python](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)
- [SecretClient API Reference](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient)
- [DefaultAzureCredential](https://learn.microsoft.com/en-us/python/api/azure-identity/azure.identity.defaultazurecredential)
