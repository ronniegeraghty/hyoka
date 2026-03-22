# Evaluation Report: key-vault-dp-python-crud

**Config:** azure-mcp | **Result:** ✅ PASSED | **Duration:** 88.2s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-crud` |
| Config | azure-mcp |
| Result | ✅ PASSED |
| Duration | 88.2s |
| Timestamp | 2026-03-21T17:27:45Z |
| Files Generated | 3 |
| Event Count | 67 |

## Configuration

- **name:** azure-mcp
- **model:** claude-sonnet-4.5

## Prompt Sent

```
Using only the Azure SDK for Python documentation, write a script that performs
all four CRUD operations on Azure Key Vault secrets:
1. Create a new secret called "my-secret" with value "my-secret-value"
2. Read the secret back and print its value
3. Update the secret to a new value "updated-value"
4. Delete the secret and purge it (soft-delete enabled vault)

Use DefaultAzureCredential for authentication. Include proper error handling
and show required pip packages.
```

## Copilot Reasoning

The user wants me to write a Python script that performs CRUD operations on Azure Key Vault secrets using the Azure SDK for Python. They specifically want:
1. Create a secret
2. Read the secret
3. Update the secret
4. Delete and purge the secret

They want me to use only the Azure SDK for Python documentation, use DefaultAzureCredential, include error handling, and show required pip packages.

This is a straightforward task - I need to:
1. Look up the Azure SDK for Python documentation for Key Vault
2. Write a script that demonstrates all CRUD operations
3. Include error handling
4. Document the required packages

Let me start by fetching the Azure SDK documentation for Key Vault secrets.

## Tool Calls

### 🔧 1. report_intent

**Input:**

```json
{"intent":"Fetching Azure SDK documentation"}
```

### 🔧 2. web_fetch

**Input:**

```json
{"max_length":15000,"url":"https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme"}
```

### 🔧 3. web_fetch

**Input:**

```json
{"max_length":15000,"url":"https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient"}
```

### 🔧 4. create

**Input:**

```json
{"file_text":"#!/usr/bin/env python3\n\"\"\"\nAzure Key Vault Secrets CRUD Operations Demo\n\nThis script demonstrates all four CRUD operations on Azure Key Vault secrets:\n1. Create - Set a new secret\n2. Read - Get the secret value\n3. Update - Change the secret value\n4. Delete - Delete and purge the secret\n\nPrerequisites:\n- Azure Key Vault with soft-delete enabled\n- Proper authentication configured (Azure CLI login, managed identity, etc.)\n- Required permissions: secrets/set, secrets/get, secrets/delete, secrets/purge\n\nEnvironment Variables Required:\n- VAULT_URL: Your Key Vault URL (e.g., https://my-key-vault.vault.azure.net/)\n\"\"\"\n\nimport os\nimport sys\nimport time\nfrom azure.identity import DefaultAzureCredential\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.core.exceptions import (\n    ResourceNotFoundError,\n    HttpResponseError,\n    ServiceRequestError\n)\n\n\ndef main():\n    \"\"\"Main function demonstrating CRUD operations on Azure Key Vault secrets.\"\"\"\n    \n    # Get vault URL from environment variable\n    vault_url = os.environ.get(\"VAULT_URL\")\n    if not vault_url:\n        print(\"Error: VAULT_URL environment variable is not set.\")\n        print(\"Example: export VAULT_URL='https://my-key-vault.vault.azure.net/'\")\n        sys.exit(1)\n    \n    print(f\"Using Key Vault: {vault_url}\\n\")\n    \n    # Initialize credential and client\n    try:\n        credential = DefaultAzureCredential()\n        client = SecretClient(vault_url=vault_url, credential=credential)\n        print(\"✓ Successfully initialized SecretClient with DefaultAzureCredential\\n\")\n    except Exception as e:\n        print(f\"Error initializing client: {e}\")\n        sys.exit(1)\n    \n    secret_name = \"my-secret\"\n    \n    try:\n        # ==================== CREATE ====================\n        print(\"=\" * 60)\n        print(\"1. CREATE - Setting a new secret\")\n        print(\"=\" * 60)\n        \n        initial_value = \"my-secret-value\"\n        try:\n            secret = client.set_secret(secret_name, initial_value)\n            print(f\"✓ Secret created successfully!\")\n            print(f\"  Name: {secret.name}\")\n            print(f\"  Value: {secret.value}\")\n            print(f\"  Version: {secret.properties.version}\")\n            print(f\"  Created on: {secret.properties.created_on}\")\n            print()\n        except HttpResponseError as e:\n            print(f\"✗ Failed to create secret: {e.message}\")\n            raise\n        \n        # ==================== READ ====================\n        print(\"=\" * 60)\n        print(\"2. READ - Retrieving the secret\")\n        print(\"=\" * 60)\n        \n        try:\n            retrieved_secret = client.get_secret(secret_name)\n            print(f\"✓ Secret retrieved successfully!\")\n            print(f\"  Name: {retrieved_secret.name}\")\n            print(f\"  Value: {retrieved_secret.value}\")\n            print(f\"  Version: {retrieved_secret.properties.version}\")\n            print(f\"  Content Type: {retrieved_secret.properties.content_type}\")\n            print(f\"  Enabled: {retrieved_secret.properties.enabled}\")\n            print()\n        except ResourceNotFoundError:\n            print(f\"✗ Secret '{secret_name}' not found in the vault\")\n            raise\n        except HttpResponseError as e:\n            print(f\"✗ Failed to retrieve secret: {e.message}\")\n            raise\n        \n        # ==================== UPDATE ====================\n        print(\"=\" * 60)\n        print(\"3. UPDATE - Updating the secret value\")\n        print(\"=\" * 60)\n        \n        new_value = \"updated-value\"\n        try:\n            # Setting a secret with an existing name creates a new version\n            updated_secret = client.set_secret(secret_name, new_value)\n            print(f\"✓ Secret updated successfully!\")\n            print(f\"  Name: {updated_secret.name}\")\n            print(f\"  New Value: {updated_secret.value}\")\n            print(f\"  New Version: {updated_secret.properties.version}\")\n            print(f\"  Updated on: {updated_secret.properties.updated_on}\")\n            print()\n            \n            # Verify the update by reading it back\n            verify_secret = client.get_secret(secret_name)\n            print(f\"✓ Verified updated value: {verify_secret.value}\")\n            print()\n        except HttpResponseError as e:\n            print(f\"✗ Failed to update secret: {e.message}\")\n            raise\n        \n        # ==================== DELETE \u0026 PURGE ====================\n        print(\"=\" * 60)\n        print(\"4. DELETE - Deleting and purging the secret\")\n        print(\"=\" * 60)\n        \n        try:\n            # Begin delete operation (returns a poller)\n            print(f\"Initiating delete operation for '{secret_name}'...\")\n            delete_poller = client.begin_delete_secret(secret_name)\n            \n            # Wait for deletion to complete\n            deleted_secret = delete_poller.result()\n            print(f\"✓ Secret deleted successfully!\")\n            print(f\"  Name: {deleted_secret.name}\")\n            print(f\"  Deleted on: {deleted_secret.deleted_date}\")\n            print(f\"  Scheduled purge date: {deleted_secret.scheduled_purge_date}\")\n            print(f\"  Recovery ID: {deleted_secret.recovery_id}\")\n            print()\n            \n            # Purge the deleted secret (permanent deletion)\n            print(f\"Purging deleted secret '{secret_name}'...\")\n            client.purge_deleted_secret(secret_name)\n            print(f\"✓ Secret purged successfully!\")\n            print(f\"  The secret has been permanently deleted and cannot be recovered.\")\n            print()\n            \n            # Wait a moment for purge to complete\n            time.sleep(2)\n            \n            # Verify the secret is gone\n            print(\"Verifying secret deletion...\")\n            try:\n                client.get_secret(secret_name)\n                print(\"✗ Warning: Secret still exists (unexpected)\")\n            except ResourceNotFoundError:\n                print(\"✓ Confirmed: Secret no longer exists in the vault\")\n            \n        except HttpResponseError as e:\n            if \"not currently in a deleted state\" in str(e):\n                print(f\"Note: Secret may already be deleted. Error: {e.message}\")\n            else:\n                print(f\"✗ Failed to delete/purge secret: {e.message}\")\n                raise\n        \n        print()\n        print(\"=\" * 60)\n        print(\"✓ All CRUD operations completed successfully!\")\n        print(\"=\" * 60)\n        \n    except ServiceRequestError as e:\n        print(f\"\\n✗ Service request error: {e}\")\n        print(\"Check your network connection and vault URL.\")\n        sys.exit(1)\n    except Exception as e:\n        print(f\"\\n✗ Unexpected error: {e}\")\n        sys.exit(1)\n    finally:\n        # Clean up resources\n        client.close()\n        credential.close()\n\n\nif __name__ == \"__main__\":\n    main()\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-2960641506/keyvault_secrets_crud.py"}
```

### 🔧 5. create

**Input:**

```json
{"file_text":"# Azure Key Vault Secrets - Required Packages\n# Install with: pip install -r requirements.txt\n\nazure-keyvault-secrets\u003e=4.8.0\nazure-identity\u003e=1.15.0\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-2960641506/requirements.txt"}
```

### 🔧 6. create

**Input:**

```json
{"file_text":"# Azure Key Vault Secrets - CRUD Operations Demo\n\nThis script demonstrates all four CRUD operations on Azure Key Vault secrets using the Azure SDK for Python.\n\n## Required Packages\n\nInstall the required packages using pip:\n\n```bash\npip install -r requirements.txt\n```\n\nOr install individually:\n\n```bash\npip install azure-keyvault-secrets azure-identity\n```\n\n### Package Details\n\n- **azure-keyvault-secrets** (\u003e=4.8.0): Azure Key Vault Secrets client library\n- **azure-identity** (\u003e=1.15.0): Azure authentication library (includes DefaultAzureCredential)\n\n## Prerequisites\n\n1. **Python 3.9 or later** is required\n2. **Azure Key Vault** with soft-delete enabled\n3. **Azure Authentication** - DefaultAzureCredential supports multiple authentication methods (in order of precedence):\n   - Environment variables (AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET)\n   - Managed Identity (when running on Azure)\n   - Azure CLI authentication (`az login`)\n   - Azure PowerShell authentication\n   - Interactive browser authentication\n\n4. **Required Permissions** on the Key Vault:\n   - secrets/set\n   - secrets/get\n   - secrets/delete\n   - secrets/purge\n\n## Setup\n\n1. Set your Key Vault URL as an environment variable:\n\n```bash\nexport VAULT_URL='https://your-key-vault-name.vault.azure.net/'\n```\n\nOn Windows (PowerShell):\n```powershell\n$env:VAULT_URL='https://your-key-vault-name.vault.azure.net/'\n```\n\nOn Windows (Command Prompt):\n```cmd\nset VAULT_URL=https://your-key-vault-name.vault.azure.net/\n```\n\n2. Authenticate with Azure (if using Azure CLI):\n\n```bash\naz login\n```\n\n## Running the Script\n\n```bash\npython keyvault_secrets_crud.py\n```\n\n## What the Script Does\n\nThe script performs four CRUD operations on a secret named \"my-secret\":\n\n### 1. **CREATE** - Set a new secret\n- Creates a secret with name \"my-secret\" and value \"my-secret-value\"\n- Uses `set_secret()` method\n- Displays the secret name, value, version, and creation date\n\n### 2. **READ** - Retrieve the secret\n- Retrieves the secret using `get_secret()` method\n- Displays the secret value and metadata (version, enabled status, etc.)\n\n### 3. **UPDATE** - Change the secret value\n- Updates the secret to a new value \"updated-value\"\n- Uses `set_secret()` again (creates a new version)\n- Displays the new version and updated timestamp\n- Verifies the update by reading it back\n\n### 4. **DELETE \u0026 PURGE** - Remove the secret\n- Deletes the secret using `begin_delete_secret()` (soft delete)\n- Waits for deletion to complete\n- Purges the secret using `purge_deleted_secret()` (permanent deletion)\n- Verifies the secret is completely removed\n\n## Error Handling\n\nThe script includes comprehensive error handling for:\n- Missing environment variables\n- Authentication failures\n- Resource not found errors\n- HTTP response errors\n- Service request errors\n- Network connectivity issues\n\n## Notes\n\n- The script uses **soft-delete** enabled vaults. If soft-delete is disabled, `begin_delete_secret()` is permanent and `purge_deleted_secret()` is not needed.\n- `begin_delete_secret()` returns a poller that allows waiting for deletion completion\n- `set_secret()` creates a new version if the secret already exists\n- All secrets are versioned in Azure Key Vault\n- The script cleans up resources properly by closing the client and credential connections\n\n## Documentation References\n\n- [Azure Key Vault Secrets Python SDK](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)\n- [SecretClient API Reference](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient)\n- [DefaultAzureCredential](https://learn.microsoft.com/en-us/python/api/azure-identity/azure.identity.defaultazurecredential)\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-2960641506/README.md"}
```

## Generated Files

- `README.md`
- `keyvault_secrets_crud.py`
- `requirements.txt`

## Copilot Response

I'll help you create a Python script that performs CRUD operations on Azure Key Vault secrets using the official Azure SDK documentation. Let me first check the Azure SDK documentation and then write the script.

Let me also fetch information about purging secrets:

Perfect! Now I have all the information needed. Let me create a comprehensive Python script that demonstrates all CRUD operations on Azure Key Vault secrets.

Now let me create a requirements.txt file and a README with instructions:

Perfect! I've created a complete solution for you. Here's what I've provided:

## Summary

✅ **keyvault_secrets_crud.py** - A comprehensive Python script that demonstrates all four CRUD operations:

1. **CREATE** - Uses `set_secret()` to create "my-secret" with value "my-secret-value"
2. **READ** - Uses `get_secret()` to retrieve and display the secret value
3. **UPDATE** - Uses `set_secret()` again to update the value to "updated-value" (creates new version)
4. **DELETE & PURGE** - Uses `begin_delete_secret()` for soft delete, then `purge_deleted_secret()` for permanent deletion

## Key Features:

✅ **DefaultAzureCredential** for authentication (supports multiple auth methods)  
✅ **Comprehensive error handling** for all common exceptions  
✅ **Clear output** with visual indicators (✓/✗) for each operation  
✅ **Proper resource cleanup** with client.close() and credential.close()  
✅ **Verification steps** to confirm each operation succeeded  

## Required Packages:

```
pip install azure-keyvault-secrets azure-identity
```

Or use the provided `requirements.txt`:
```
pip install -r requirements.txt
```

## To Run:

```bash
export VAULT_URL='https://your-key-vault-name.vault.azure.net/'
python keyvault_secrets_crud.py
```

The script includes detailed documentation, follows Azure SDK best practices from the official documentation, and handles both soft-delete enabled and disabled vaults appropriately!

## Verification

**Result:** ✅ PASS

**Summary:** Code fully implements all CRUD operations with correct SDK usage, authentication, error handling, and soft-delete support as specified.

**Reasoning:**

The generated code fully meets all requirements from the original prompt. (1) CRUD Operations: Correctly implements Create using set_secret() with 'my-secret' and 'my-secret-value', Read using get_secret() with printed output, Update using set_secret() with 'updated-value', and Delete/Purge using begin_delete_secret() followed by purge_deleted_secret(). (2) Authentication: Uses DefaultAzureCredential as specified. (3) Error Handling: Includes comprehensive exception handling for ResourceNotFoundError, HttpResponseError, and ServiceRequestError with proper try-except blocks and cleanup in finally. (4) Packages: requirements.txt correctly lists azure-keyvault-secrets>=4.8.0 and azure-identity>=1.15.0 with installation instructions. (5) Soft-Delete: Properly handles soft-delete by calling begin_delete_secret(), waiting for completion with result(), then purging. (6) Syntax: Code is syntactically valid Python with correct imports and SDK method usage. (7) Documentation: Comprehensive README covers setup, prerequisites, authentication methods, and usage. The code addresses all expected coverage points including SecretClient initialization, all required methods, soft-delete waiting mechanism, and ResourceNotFoundError handling.

---

[← Back to Summary](../../../../../../summary.md)
