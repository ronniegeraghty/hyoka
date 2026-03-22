# Azure Key Vault Secrets CRUD Operations

This script demonstrates all four CRUD operations on Azure Key Vault secrets using the Azure SDK for Python.

## Prerequisites

1. **Azure Key Vault**: You need an existing Key Vault with soft-delete enabled
2. **Authentication**: Appropriate Azure credentials configured for DefaultAzureCredential
3. **Permissions**: Your identity needs the following Key Vault permissions:
   - Set (Create/Update secrets)
   - Get (Read secrets)
   - Delete (Delete secrets)
   - Purge (Permanently delete secrets)

## Installation

Install the required packages:

```bash
pip install -r requirements.txt
```

Or install individually:

```bash
pip install azure-keyvault-secrets azure-identity
```

## Configuration

Set the Key Vault URL as an environment variable:

```bash
export AZURE_KEY_VAULT_URL='https://your-vault-name.vault.azure.net/'
```

## Authentication

The script uses `DefaultAzureCredential` which attempts authentication through multiple methods in order:

1. Environment variables
2. Managed Identity
3. Visual Studio Code
4. Azure CLI
5. Azure PowerShell
6. Interactive browser

For local development, the easiest method is Azure CLI:

```bash
az login
```

## Usage

Run the script:

```bash
python key_vault_crud.py
```

## Operations Performed

1. **CREATE**: Creates a secret named "my-secret" with value "my-secret-value"
2. **READ**: Retrieves and displays the secret value
3. **UPDATE**: Updates the secret to a new value "updated-value"
4. **DELETE**: Soft-deletes the secret, then permanently purges it

## Error Handling

The script includes comprehensive error handling for:
- Missing environment variables
- Authentication failures
- Resource not found errors
- HTTP response errors
- General exceptions

## Output Example

```
Connecting to Key Vault: https://your-vault.vault.azure.net/

============================================================
1. CREATE - Setting a new secret
============================================================
✓ Secret 'my-secret' created successfully
  Version: abc123...
  Created on: 2026-03-21 06:48:41

============================================================
2. READ - Retrieving the secret
============================================================
✓ Secret 'my-secret' retrieved successfully
  Value: my-secret-value
  Version: abc123...

============================================================
3. UPDATE - Updating the secret with a new value
============================================================
✓ Secret 'my-secret' updated successfully
  New value: updated-value
  New version: def456...

============================================================
4. DELETE - Deleting and purging the secret
============================================================
✓ Secret 'my-secret' deleted successfully (soft-delete)
  Scheduled purge date: 2026-04-20 06:48:41
  Deleted date: 2026-03-21 06:48:41

  Purging secret 'my-secret'...
✓ Secret 'my-secret' purged successfully (permanent deletion)

============================================================
All CRUD operations completed successfully!
============================================================
```
