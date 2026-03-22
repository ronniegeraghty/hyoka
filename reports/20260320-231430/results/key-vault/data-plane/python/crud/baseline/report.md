# Evaluation Report: key-vault-dp-python-crud

**Config:** baseline | **Result:** ✅ PASSED | **Duration:** 110.3s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-crud` |
| Config | baseline |
| Result | ✅ PASSED |
| Duration | 110.3s |
| Timestamp | 2026-03-21T06:14:30Z |
| Files Generated | 4 |
| Event Count | 101 |

## Configuration

- **name:** baseline
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

The user wants me to create a Python script that demonstrates CRUD operations on Azure Key Vault secrets using the Azure SDK for Python. They specifically want:

1. Create a secret
2. Read the secret
3. Update the secret
4. Delete and purge the secret

They want me to use official Azure SDK documentation and use DefaultAzureCredential for authentication. They also want proper error handling and to show required pip packages.

This is a straightforward task - I need to:
1. Look up the Azure SDK for Python Key Vault documentation
2. Write a script that performs these operations
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
{"max_length":10000,"url":"https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient"}
```

### 🔧 4. create

**Input:**

```json
{"file_text":"#!/usr/bin/env python3\n\"\"\"\nAzure Key Vault Secrets CRUD Operations Demo\n\nThis script demonstrates all four CRUD operations on Azure Key Vault secrets:\n1. Create - Set a new secret\n2. Read - Retrieve a secret value\n3. Update - Update a secret to a new value\n4. Delete - Delete and purge a secret (for soft-delete enabled vaults)\n\nRequirements:\n    pip install azure-keyvault-secrets azure-identity\n\nEnvironment Variables:\n    AZURE_KEY_VAULT_URL - Your Key Vault URL (e.g., https://my-vault.vault.azure.net/)\n\nAuthentication:\n    Uses DefaultAzureCredential which supports multiple authentication methods:\n    - Azure CLI (az login)\n    - Managed Identity (in Azure)\n    - Environment variables (AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET)\n    - And more...\n\"\"\"\n\nimport os\nimport sys\nfrom azure.identity import DefaultAzureCredential\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.core.exceptions import (\n    ResourceNotFoundError,\n    HttpResponseError,\n    ClientAuthenticationError\n)\n\n\ndef main():\n    \"\"\"Main function demonstrating CRUD operations on Key Vault secrets.\"\"\"\n    \n    # Get Key Vault URL from environment variable\n    vault_url = os.environ.get(\"AZURE_KEY_VAULT_URL\")\n    if not vault_url:\n        print(\"Error: AZURE_KEY_VAULT_URL environment variable is not set\")\n        print(\"Example: export AZURE_KEY_VAULT_URL='https://my-vault.vault.azure.net/'\")\n        sys.exit(1)\n    \n    print(f\"Connecting to Key Vault: {vault_url}\")\n    print(\"-\" * 80)\n    \n    try:\n        # Initialize the credential and client\n        credential = DefaultAzureCredential()\n        client = SecretClient(vault_url=vault_url, credential=credential)\n        \n        # 1. CREATE - Set a new secret\n        print(\"\\n1. CREATE - Setting a new secret 'my-secret'\")\n        secret_name = \"my-secret\"\n        secret_value = \"my-secret-value\"\n        \n        try:\n            created_secret = client.set_secret(secret_name, secret_value)\n            print(f\"   ✓ Secret created successfully\")\n            print(f\"   Name: {created_secret.name}\")\n            print(f\"   Value: {created_secret.value}\")\n            print(f\"   Version: {created_secret.properties.version}\")\n            print(f\"   Created on: {created_secret.properties.created_on}\")\n        except HttpResponseError as e:\n            print(f\"   ✗ Failed to create secret: {e.message}\")\n            sys.exit(1)\n        \n        # 2. READ - Retrieve the secret\n        print(\"\\n2. READ - Retrieving the secret 'my-secret'\")\n        try:\n            retrieved_secret = client.get_secret(secret_name)\n            print(f\"   ✓ Secret retrieved successfully\")\n            print(f\"   Name: {retrieved_secret.name}\")\n            print(f\"   Value: {retrieved_secret.value}\")\n            print(f\"   Version: {retrieved_secret.properties.version}\")\n        except ResourceNotFoundError:\n            print(f\"   ✗ Secret '{secret_name}' not found\")\n            sys.exit(1)\n        except HttpResponseError as e:\n            print(f\"   ✗ Failed to retrieve secret: {e.message}\")\n            sys.exit(1)\n        \n        # 3. UPDATE - Update the secret to a new value\n        print(\"\\n3. UPDATE - Updating secret to a new value\")\n        new_value = \"updated-value\"\n        \n        try:\n            updated_secret = client.set_secret(secret_name, new_value)\n            print(f\"   ✓ Secret updated successfully\")\n            print(f\"   Name: {updated_secret.name}\")\n            print(f\"   New Value: {updated_secret.value}\")\n            print(f\"   New Version: {updated_secret.properties.version}\")\n            print(f\"   Updated on: {updated_secret.properties.updated_on}\")\n        except HttpResponseError as e:\n            print(f\"   ✗ Failed to update secret: {e.message}\")\n            sys.exit(1)\n        \n        # 4. DELETE - Delete and purge the secret\n        print(\"\\n4. DELETE - Deleting and purging the secret\")\n        \n        # Step 4a: Begin delete operation\n        try:\n            print(\"   4a. Starting deletion...\")\n            delete_poller = client.begin_delete_secret(secret_name)\n            deleted_secret = delete_poller.result()\n            print(f\"   ✓ Secret deleted successfully\")\n            print(f\"   Name: {deleted_secret.name}\")\n            print(f\"   Deleted on: {deleted_secret.deleted_date}\")\n            print(f\"   Scheduled purge date: {deleted_secret.scheduled_purge_date}\")\n            print(f\"   Recovery ID: {deleted_secret.recovery_id}\")\n        except ResourceNotFoundError:\n            print(f\"   ✗ Secret '{secret_name}' not found for deletion\")\n            sys.exit(1)\n        except HttpResponseError as e:\n            print(f\"   ✗ Failed to delete secret: {e.message}\")\n            sys.exit(1)\n        \n        # Step 4b: Purge the deleted secret (permanent deletion)\n        try:\n            print(\"   4b. Purging deleted secret (permanent deletion)...\")\n            client.purge_deleted_secret(secret_name)\n            print(f\"   ✓ Secret purged successfully (permanently deleted)\")\n        except ResourceNotFoundError:\n            print(f\"   ⚠ Secret '{secret_name}' not found in deleted secrets\")\n            print(\"      (May have been already purged or vault may not have soft-delete enabled)\")\n        except HttpResponseError as e:\n            if \"not enabled\" in str(e).lower() or \"soft delete\" in str(e).lower():\n                print(f\"   ⚠ Purge not needed - vault may not have soft-delete enabled\")\n                print(f\"      Secret was permanently deleted in the delete step\")\n            else:\n                print(f\"   ✗ Failed to purge secret: {e.message}\")\n                sys.exit(1)\n        \n        print(\"\\n\" + \"=\" * 80)\n        print(\"All CRUD operations completed successfully!\")\n        print(\"=\" * 80)\n        \n    except ClientAuthenticationError as e:\n        print(f\"\\nAuthentication Error: {e.message}\")\n        print(\"\\nPlease ensure you are authenticated. Try one of:\")\n        print(\"  - Run 'az login' (Azure CLI)\")\n        print(\"  - Set AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET\")\n        print(\"  - Use Managed Identity (if running in Azure)\")\n        sys.exit(1)\n    except Exception as e:\n        print(f\"\\nUnexpected error: {type(e).__name__}: {str(e)}\")\n        sys.exit(1)\n    finally:\n        # Clean up - close the credential\n        try:\n            credential.close()\n        except:\n            pass\n\n\nif __name__ == \"__main__\":\n    main()\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-2842495542/keyvault_crud_demo.py"}
```

### 🔧 5. create

**Input:**

```json
{"file_text":"azure-keyvault-secrets\nazure-identity\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-2842495542/requirements.txt"}
```

### 🔧 6. create

**Input:**

```json
{"file_text":"# Azure Key Vault Secrets CRUD Operations Demo\n\nThis script demonstrates all four CRUD operations on Azure Key Vault secrets using the Azure SDK for Python.\n\n## Required Packages\n\nInstall the required packages using pip:\n\n```bash\npip install -r requirements.txt\n```\n\nOr install individually:\n\n```bash\npip install azure-keyvault-secrets azure-identity\n```\n\n### Package Details\n\n- **azure-keyvault-secrets**: Azure Key Vault Secrets client library for Python\n- **azure-identity**: Azure authentication library (provides DefaultAzureCredential)\n\n## Prerequisites\n\n1. **Python 3.9 or later**\n2. **An Azure Key Vault** - Create one using Azure CLI:\n   ```bash\n   az keyvault create --name \u003cyour-vault-name\u003e --resource-group \u003cyour-rg\u003e --location \u003clocation\u003e\n   ```\n3. **Proper permissions** - Ensure you have the following Key Vault secrets permissions:\n   - `secrets/set` (for create/update)\n   - `secrets/get` (for read)\n   - `secrets/delete` (for delete)\n   - `secrets/purge` (for purge, if soft-delete is enabled)\n\n## Authentication Setup\n\nThe script uses `DefaultAzureCredential`, which tries multiple authentication methods in order:\n\n### Option 1: Azure CLI (Recommended for local development)\n```bash\naz login\n```\n\n### Option 2: Environment Variables\n```bash\nexport AZURE_CLIENT_ID=\"\u003cyour-client-id\u003e\"\nexport AZURE_TENANT_ID=\"\u003cyour-tenant-id\u003e\"\nexport AZURE_CLIENT_SECRET=\"\u003cyour-client-secret\u003e\"\n```\n\n### Option 3: Managed Identity\nAutomatically works when running in Azure (App Service, VM, Container Instances, etc.)\n\n## Usage\n\n1. Set the Key Vault URL environment variable:\n   ```bash\n   export AZURE_KEY_VAULT_URL=\"https://\u003cyour-vault-name\u003e.vault.azure.net/\"\n   ```\n\n2. Run the script:\n   ```bash\n   python keyvault_crud_demo.py\n   ```\n\n## What the Script Does\n\nThe script performs the following operations in sequence:\n\n1. **CREATE**: Creates a new secret named \"my-secret\" with value \"my-secret-value\"\n2. **READ**: Retrieves and displays the secret value\n3. **UPDATE**: Updates the secret to a new value \"updated-value\"\n4. **DELETE**: Deletes the secret (soft-delete)\n5. **PURGE**: Permanently deletes the secret (for vaults with soft-delete enabled)\n\n## Expected Output\n\n```\nConnecting to Key Vault: https://your-vault.vault.azure.net/\n--------------------------------------------------------------------------------\n\n1. CREATE - Setting a new secret 'my-secret'\n   ✓ Secret created successfully\n   Name: my-secret\n   Value: my-secret-value\n   Version: abc123...\n   Created on: 2026-03-21 06:14:33.123456\n\n2. READ - Retrieving the secret 'my-secret'\n   ✓ Secret retrieved successfully\n   Name: my-secret\n   Value: my-secret-value\n   Version: abc123...\n\n3. UPDATE - Updating secret to a new value\n   ✓ Secret updated successfully\n   Name: my-secret\n   New Value: updated-value\n   New Version: def456...\n   Updated on: 2026-03-21 06:14:34.123456\n\n4. DELETE - Deleting and purging the secret\n   4a. Starting deletion...\n   ✓ Secret deleted successfully\n   Name: my-secret\n   Deleted on: 2026-03-21 06:14:35.123456\n   Scheduled purge date: 2026-04-20 06:14:35.123456\n   Recovery ID: https://...\n   4b. Purging deleted secret (permanent deletion)...\n   ✓ Secret purged successfully (permanently deleted)\n\n================================================================================\nAll CRUD operations completed successfully!\n================================================================================\n```\n\n## Error Handling\n\nThe script includes comprehensive error handling for:\n\n- Missing environment variables\n- Authentication failures\n- Resource not found errors\n- Permission errors\n- Soft-delete configuration issues\n\n## Notes\n\n- If your vault does **not** have soft-delete enabled, the delete operation is permanent, and the purge step will show a warning\n- The `set_secret()` method is used for both CREATE and UPDATE operations - it creates a new version if the secret already exists\n- Each secret update creates a new version; all versions are stored until deleted\n- Soft-delete provides a safety net, allowing secret recovery within the retention period (default 90 days)\n\n## References\n\n- [Azure Key Vault Secrets Python SDK Documentation](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)\n- [DefaultAzureCredential Documentation](https://learn.microsoft.com/en-us/python/api/azure-identity/azure.identity.defaultazurecredential)\n- [Azure Key Vault Documentation](https://learn.microsoft.com/en-us/azure/key-vault/)\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-2842495542/README.md"}
```

### 🔧 7. bash

**Input:**

```json
{"command":"cd /tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-2842495542 \u0026\u0026 ls -la","description":"List files in the directory"}
```

### 🔧 8. bash

**Input:**

```json
{"command":"cd /tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-2842495542 \u0026\u0026 python3 -m py_compile keyvault_crud_demo.py \u0026\u0026 echo \"✓ Syntax check passed\"","description":"Verify Python syntax"}
```

### 🔧 9. bash

**Input:**

```json
{"command":"cd /tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-2842495542 \u0026\u0026 cat \u003c\u003c 'EOF'\n========================================\nAZURE KEY VAULT CRUD SCRIPT - SUMMARY\n========================================\n\nFiles Created:\n--------------\n1. keyvault_crud_demo.py  - Main script with all CRUD operations\n2. requirements.txt       - Required pip packages\n3. README.md              - Comprehensive usage documentation\n\nQuick Start:\n-----------\n1. Install packages:\n   pip install -r requirements.txt\n\n2. Authenticate (choose one):\n   - Azure CLI: az login\n   - Set env vars: AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET\n   - Use Managed Identity in Azure\n\n3. Set Key Vault URL:\n   export AZURE_KEY_VAULT_URL=\"https://your-vault.vault.azure.net/\"\n\n4. Run script:\n   python keyvault_crud_demo.py\n\nCRUD Operations Demonstrated:\n-----------------------------\n✓ CREATE  - set_secret(\"my-secret\", \"my-secret-value\")\n✓ READ    - get_secret(\"my-secret\")\n✓ UPDATE  - set_secret(\"my-secret\", \"updated-value\")\n✓ DELETE  - begin_delete_secret(\"my-secret\").result()\n✓ PURGE   - purge_deleted_secret(\"my-secret\")\n\nKey Features:\n------------\n✓ Based on official Azure SDK for Python documentation\n✓ Uses DefaultAzureCredential for flexible authentication\n✓ Comprehensive error handling for all operations\n✓ Supports both soft-delete enabled and disabled vaults\n✓ Clear output with success/failure indicators\n✓ Well-documented with inline comments\n\nPackages Required:\n-----------------\n- azure-keyvault-secrets (Key Vault client)\n- azure-identity (Authentication)\n\nEOF\n","description":"Display summary"}
```

## Generated Files

- `README.md`
- `__pycache__/keyvault_crud_demo.cpython-312.pyc`
- `keyvault_crud_demo.py`
- `requirements.txt`

## Copilot Response

I'll help you create a Python script for CRUD operations on Azure Key Vault secrets using the official Azure SDK documentation.

Now let me fetch information about purging secrets:

Perfect! Now I have all the information I need. Let me create a comprehensive Python script that demonstrates all CRUD operations on Azure Key Vault secrets:

Now let me create a requirements.txt file and a README with usage instructions:

Perfect! Let me now verify the script by viewing it one more time to ensure everything is correct:

Excellent! Let me also verify the script syntax:

Perfect! Let me create a quick summary document:

## Summary

✅ **Complete!** I've created a comprehensive Python script that demonstrates all four CRUD operations on Azure Key Vault secrets, based entirely on the official Azure SDK for Python documentation.

### Files Created:

1. **`keyvault_crud_demo.py`** - Main script with:
   - **CREATE**: Sets a secret named "my-secret" with value "my-secret-value"
   - **READ**: Retrieves and displays the secret value
   - **UPDATE**: Updates the secret to "updated-value"
   - **DELETE**: Deletes the secret (soft-delete)
   - **PURGE**: Permanently deletes the secret

2. **`requirements.txt`** - Required packages:
   - `azure-keyvault-secrets`
   - `azure-identity`

3. **`README.md`** - Complete documentation with setup and usage instructions

### Key Features:
- ✅ Uses `DefaultAzureCredential` for flexible authentication
- ✅ Comprehensive error handling (authentication, not found, permissions, etc.)
- ✅ Supports both soft-delete enabled and disabled vaults
- ✅ Clear, informative output with ✓/✗ indicators
- ✅ Well-documented with inline comments
- ✅ Syntax validated

### Quick Start:
```bash
pip install -r requirements.txt
export AZURE_KEY_VAULT_URL="https://your-vault.vault.azure.net/"
python keyvault_crud_demo.py
```

## Verification

**Result:** ✅ PASS

**Summary:** Code fully implements all required CRUD operations with correct Azure SDK usage, authentication, error handling, and soft-delete support.

**Reasoning:**

The generated code successfully meets all requirements from the original prompt. It correctly implements all four CRUD operations: (1) CREATE - uses set_secret() to create 'my-secret' with value 'my-secret-value', (2) READ - uses get_secret() to retrieve and print the value, (3) UPDATE - uses set_secret() again to update to 'updated-value', and (4) DELETE/PURGE - uses begin_delete_secret() with .result() to wait for completion, then purge_deleted_secret() for permanent deletion. The code uses DefaultAzureCredential for authentication and creates a SecretClient with vault URL and credential as expected. Proper error handling is implemented for ResourceNotFoundError, HttpResponseError, and ClientAuthenticationError. The requirements.txt correctly lists azure-keyvault-secrets and azure-identity packages. The code properly handles the soft-delete scenario by calling .result() on the delete poller before attempting to purge, which ensures the delete operation completes. The README.md provides comprehensive documentation covering installation, prerequisites, authentication setup, and usage. The code is syntactically valid, well-structured, follows Python best practices, and includes helpful error messages and output formatting.

---

[← Back to Summary](../../../../../../summary.md)
