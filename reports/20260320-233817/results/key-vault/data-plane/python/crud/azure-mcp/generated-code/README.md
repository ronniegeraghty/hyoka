# Azure Key Vault Secrets CRUD Demo

This script demonstrates all four CRUD operations on Azure Key Vault secrets using the Azure SDK for Python.

## Prerequisites

1. **Python 3.9 or later**
2. **An Azure Key Vault** with soft-delete enabled (recommended)
3. **Proper authentication** configured (see below)

## Installation

Install the required packages:

```bash
pip install -r requirements.txt
```

Or install directly:

```bash
pip install azure-keyvault-secrets azure-identity
```

## Required Packages

- `azure-keyvault-secrets`: Azure Key Vault Secrets client library
- `azure-identity`: Azure authentication library (provides DefaultAzureCredential)

## Configuration

Set the Key Vault URL as an environment variable:

```bash
export AZURE_KEY_VAULT_URL="https://your-key-vault-name.vault.azure.net/"
```

## Authentication

The script uses `DefaultAzureCredential`, which automatically tries multiple authentication methods in order:

1. **Environment Variables**:
   ```bash
   export AZURE_CLIENT_ID="your-client-id"
   export AZURE_TENANT_ID="your-tenant-id"
   export AZURE_CLIENT_SECRET="your-client-secret"
   ```

2. **Azure CLI** (easiest for local development):
   ```bash
   az login
   ```

3. **Managed Identity** (when running in Azure)

4. **Interactive Browser** (fallback)

## Permissions Required

Ensure your Azure identity has the following Key Vault permissions:

- `secrets/set` - To create and update secrets
- `secrets/get` - To read secrets
- `secrets/delete` - To delete secrets
- `secrets/purge` - To permanently delete secrets (only needed for soft-delete enabled vaults)

You can assign these permissions using Azure RBAC role "Key Vault Secrets Officer" or via Access Policies.

## Usage

Run the script:

```bash
python keyvault_crud_demo.py
```

## What the Script Does

The script performs these operations in sequence:

1. **CREATE**: Creates a secret named "my-secret" with value "my-secret-value"
2. **READ**: Retrieves the secret and prints its value
3. **UPDATE**: Updates the secret to a new value "updated-value" (creates a new version)
4. **DELETE**: Soft-deletes the secret
5. **PURGE**: Permanently deletes the secret (for soft-delete enabled vaults)

## Output Example

```
Using Key Vault: https://my-key-vault.vault.azure.net/

✓ Successfully initialized SecretClient with DefaultAzureCredential

============================================================
1. CREATE - Creating a new secret
============================================================
✓ Secret created successfully!
  Name:    my-secret
  Value:   my-secret-value
  Version: abc123...
  Created: 2026-03-21 06:30:00.123456+00:00

...
```

## Error Handling

The script includes comprehensive error handling for:

- Missing environment variables
- Authentication failures
- Resource not found errors
- HTTP response errors
- Permission errors
- Purge operation restrictions

## Notes

- If your vault has **purge protection** enabled, the purge operation will fail. This is a safety feature.
- If your vault does **not** have soft-delete enabled, the delete operation is permanent and purge is not needed.
- The script uses `begin_delete_secret()` which returns a poller to wait for deletion to complete.
- Secret values are stored as strings in Azure Key Vault.

## Reference

Based on official Azure SDK for Python documentation:
- [Azure Key Vault Secrets README](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)
- [SecretClient API Reference](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient)
