# Evaluation Report: key-vault-dp-python-crud

**Config:** azure-mcp | **Result:** ✅ PASSED | **Duration:** 71.6s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-crud` |
| Config | azure-mcp |
| Result | ✅ PASSED |
| Duration | 71.6s |
| Timestamp | 2026-03-22T21:26:11Z |
| Files Generated | 3 |
| Event Count | 68 |

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

The user wants me to create a Python script that demonstrates CRUD operations on Azure Key Vault secrets using the Azure SDK for Python. They specifically want me to use only the Azure SDK for Python documentation.

This is a straightforward task - I need to:
1. Look up the Azure SDK for Python documentation for Key Vault
2. Write a script with all four CRUD operations
3. Include proper error handling
4. Show required pip packages

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
{"max_length":15000,"url":"https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme?view=azure-python"}
```

### 🔧 3. web_fetch

**Input:**

```json
{"max_length":20000,"url":"https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient?view=azure-python"}
```

### 🔧 4. create

**Input:**

```json
{"file_text":"#!/usr/bin/env python3\n\"\"\"\nAzure Key Vault CRUD Operations Demo\nDemonstrates Create, Read, Update, and Delete operations for secrets\nusing the Azure SDK for Python.\n\nRequired packages:\n- azure-keyvault-secrets\n- azure-identity\n\nInstall with: pip install azure-keyvault-secrets azure-identity\n\"\"\"\n\nimport os\nimport sys\nfrom azure.identity import DefaultAzureCredential\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.core.exceptions import ResourceNotFoundError, HttpResponseError\n\n\ndef main():\n    # Get vault URL from environment variable\n    vault_url = os.environ.get(\"AZURE_KEY_VAULT_URL\")\n    \n    if not vault_url:\n        print(\"Error: AZURE_KEY_VAULT_URL environment variable not set\")\n        print(\"Example: export AZURE_KEY_VAULT_URL='https://my-key-vault.vault.azure.net/'\")\n        sys.exit(1)\n    \n    print(f\"Connecting to Azure Key Vault: {vault_url}\")\n    \n    try:\n        # Authenticate using DefaultAzureCredential\n        # This will try multiple authentication methods in order:\n        # 1. Environment variables\n        # 2. Managed identity\n        # 3. Visual Studio Code\n        # 4. Azure CLI\n        # 5. Azure PowerShell\n        credential = DefaultAzureCredential()\n        \n        # Create the SecretClient\n        secret_client = SecretClient(vault_url=vault_url, credential=credential)\n        \n        secret_name = \"my-secret\"\n        \n        # ============================================================\n        # CREATE - Set a new secret\n        # ============================================================\n        print(\"\\n1. CREATE - Setting a new secret...\")\n        try:\n            secret = secret_client.set_secret(secret_name, \"my-secret-value\")\n            print(f\"   ✓ Secret created successfully\")\n            print(f\"   - Name: {secret.name}\")\n            print(f\"   - Value: {secret.value}\")\n            print(f\"   - Version: {secret.properties.version}\")\n        except HttpResponseError as e:\n            print(f\"   ✗ Failed to create secret: {e.message}\")\n            sys.exit(1)\n        \n        # ============================================================\n        # READ - Retrieve the secret\n        # ============================================================\n        print(\"\\n2. READ - Retrieving the secret...\")\n        try:\n            retrieved_secret = secret_client.get_secret(secret_name)\n            print(f\"   ✓ Secret retrieved successfully\")\n            print(f\"   - Name: {retrieved_secret.name}\")\n            print(f\"   - Value: {retrieved_secret.value}\")\n            print(f\"   - Version: {retrieved_secret.properties.version}\")\n        except ResourceNotFoundError:\n            print(f\"   ✗ Secret '{secret_name}' not found\")\n            sys.exit(1)\n        except HttpResponseError as e:\n            print(f\"   ✗ Failed to retrieve secret: {e.message}\")\n            sys.exit(1)\n        \n        # ============================================================\n        # UPDATE - Update the secret value\n        # ============================================================\n        print(\"\\n3. UPDATE - Updating the secret to a new value...\")\n        try:\n            # set_secret creates a new version when called on an existing secret\n            updated_secret = secret_client.set_secret(secret_name, \"updated-value\")\n            print(f\"   ✓ Secret updated successfully\")\n            print(f\"   - Name: {updated_secret.name}\")\n            print(f\"   - New Value: {updated_secret.value}\")\n            print(f\"   - New Version: {updated_secret.properties.version}\")\n        except HttpResponseError as e:\n            print(f\"   ✗ Failed to update secret: {e.message}\")\n            sys.exit(1)\n        \n        # ============================================================\n        # DELETE - Delete and purge the secret\n        # ============================================================\n        print(\"\\n4. DELETE - Deleting the secret...\")\n        try:\n            # begin_delete_secret returns a poller for long-running operations\n            delete_poller = secret_client.begin_delete_secret(secret_name)\n            deleted_secret = delete_poller.result()\n            print(f\"   ✓ Secret deleted successfully\")\n            print(f\"   - Name: {deleted_secret.name}\")\n            print(f\"   - Deleted Date: {deleted_secret.deleted_date}\")\n            print(f\"   - Scheduled Purge Date: {deleted_secret.scheduled_purge_date}\")\n        except ResourceNotFoundError:\n            print(f\"   ✗ Secret '{secret_name}' not found\")\n            sys.exit(1)\n        except HttpResponseError as e:\n            print(f\"   ✗ Failed to delete secret: {e.message}\")\n            sys.exit(1)\n        \n        # ============================================================\n        # PURGE - Permanently delete the secret (soft-delete vaults)\n        # ============================================================\n        print(\"\\n5. PURGE - Permanently deleting the secret...\")\n        try:\n            secret_client.purge_deleted_secret(secret_name)\n            print(f\"   ✓ Secret purged successfully\")\n            print(f\"   - The secret '{secret_name}' has been permanently deleted\")\n        except ResourceNotFoundError:\n            print(f\"   ✗ Deleted secret '{secret_name}' not found\")\n        except HttpResponseError as e:\n            # Purge may fail if vault doesn't have soft-delete or purge protection\n            print(f\"   ⚠ Purge operation note: {e.message}\")\n        \n        print(\"\\n\" + \"=\"*60)\n        print(\"CRUD operations completed successfully!\")\n        print(\"=\"*60)\n        \n    except Exception as e:\n        print(f\"\\n✗ Unexpected error: {type(e).__name__}: {str(e)}\")\n        sys.exit(1)\n    finally:\n        # Close the credential\n        credential.close()\n\n\nif __name__ == \"__main__\":\n    main()\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-1731692071/keyvault_crud_demo.py"}
```

### 🔧 5. create

**Input:**

```json
{"file_text":"azure-keyvault-secrets\nazure-identity\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-1731692071/requirements.txt"}
```

### 🔧 6. create

**Input:**

```json
{"file_text":"# Azure Key Vault CRUD Operations Demo\n\nThis script demonstrates all four CRUD operations on Azure Key Vault secrets using the Azure SDK for Python.\n\n## Required Packages\n\nInstall the required packages using pip:\n\n```bash\npip install -r requirements.txt\n```\n\nOr install individually:\n\n```bash\npip install azure-keyvault-secrets azure-identity\n```\n\n## Prerequisites\n\n1. **Azure Key Vault**: You need an existing Azure Key Vault with soft-delete enabled\n2. **Authentication**: DefaultAzureCredential will attempt authentication in this order:\n   - Environment variables (AZURE_TENANT_ID, AZURE_CLIENT_ID, AZURE_CLIENT_SECRET)\n   - Managed Identity (if running on Azure)\n   - Azure CLI (run `az login` first)\n   - Azure PowerShell\n   - Visual Studio Code\n\n3. **Permissions**: Your identity needs the following Key Vault permissions:\n   - secrets/set\n   - secrets/get\n   - secrets/delete\n   - secrets/purge\n\n## Setup\n\n1. Set your Key Vault URL as an environment variable:\n\n```bash\nexport AZURE_KEY_VAULT_URL=\"https://your-key-vault-name.vault.azure.net/\"\n```\n\n2. Authenticate using Azure CLI (easiest for local development):\n\n```bash\naz login\n```\n\n## Usage\n\nRun the script:\n\n```bash\npython keyvault_crud_demo.py\n```\n\n## What the Script Does\n\n1. **CREATE**: Creates a new secret called \"my-secret\" with value \"my-secret-value\"\n2. **READ**: Retrieves and displays the secret value\n3. **UPDATE**: Updates the secret to a new value \"updated-value\"\n4. **DELETE**: Soft-deletes the secret (for vaults with soft-delete enabled)\n5. **PURGE**: Permanently deletes the secret from the vault\n\n## Error Handling\n\nThe script includes comprehensive error handling for:\n- Missing environment variables\n- Authentication failures\n- Resource not found errors\n- HTTP response errors\n- General exceptions\n\n## Documentation Reference\n\nThis script is based on the official Azure SDK for Python documentation:\n- https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme\n- https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-1731692071/README.md"}
```

## Generated Files

- `README.md`
- `keyvault_crud_demo.py`
- `requirements.txt`

## Copilot Response

I'll help you create a Python script for Azure Key Vault CRUD operations using the official Azure SDK documentation.

Perfect! I have the documentation I need. Let me also check for the purge operation documentation:

Great! Now I have all the information I need. Let me create a comprehensive Python script that demonstrates all four CRUD operations:

Now let me create a requirements.txt file and a README with setup instructions:

Perfect! I've created a complete Azure Key Vault CRUD operations script based on the official Azure SDK for Python documentation. Here's what I've provided:

## 📄 Files Created

1. **keyvault_crud_demo.py** - Main script with all CRUD operations
2. **requirements.txt** - Required Python packages
3. **README.md** - Setup and usage instructions

## ✨ Key Features

The script demonstrates:

1. **CREATE**: Uses `set_secret()` to create "my-secret" with value "my-secret-value"
2. **READ**: Uses `get_secret()` to retrieve and display the secret
3. **UPDATE**: Uses `set_secret()` again to update to "updated-value" (creates a new version)
4. **DELETE**: Uses `begin_delete_secret()` for soft-delete
5. **PURGE**: Uses `purge_deleted_secret()` to permanently remove the secret

## 🔐 Authentication

Uses `DefaultAzureCredential` which automatically tries multiple authentication methods:
- Environment variables
- Managed Identity
- Azure CLI (`az login`)
- Azure PowerShell
- VS Code

## 🛡️ Error Handling

Includes proper exception handling for:
- `ResourceNotFoundError` - when secrets don't exist
- `HttpResponseError` - for API errors
- Missing environment variables
- General exceptions

## 📦 Required Packages

```
azure-keyvault-secrets
azure-identity
```

The script is production-ready with clear output, proper error handling, and follows Azure SDK best practices from the official documentation!

## Verification

**Result:** ✅ PASS

**Summary:** All CRUD operations correctly implemented with proper authentication, error handling, and soft-delete handling

**Reasoning:**

The generated code fully meets all requirements from the original prompt. It correctly implements all four CRUD operations: (1) CREATE using set_secret() with 'my-secret' and 'my-secret-value', (2) READ using get_secret() and printing the value, (3) UPDATE using set_secret() with 'updated-value', (4) DELETE and PURGE using begin_delete_secret() followed by purge_deleted_secret(). The code properly uses DefaultAzureCredential for authentication, creates a SecretClient with vault URL and credential, and includes comprehensive error handling for ResourceNotFoundError and HttpResponseError. The requirements.txt correctly lists azure-keyvault-secrets and azure-identity packages. Critically, the code handles soft-delete correctly by calling .result() on the delete_poller to wait for the delete operation to complete before purging. The README provides clear installation instructions and prerequisites. The code is syntactically valid Python, follows Azure SDK best practices, and would execute successfully given proper Azure credentials and Key Vault setup.

---

[← Back to Summary](../../../../../../summary.md)
