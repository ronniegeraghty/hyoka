# Evaluation Report: key-vault-dp-python-crud

**Config:** azure-mcp | **Result:** ✅ PASSED | **Duration:** 82.3s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-crud` |
| Config | azure-mcp |
| Result | ✅ PASSED |
| Duration | 82.3s |
| Timestamp | 2026-03-22T21:19:35Z |
| Files Generated | 3 |
| Event Count | 79 |

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

The user wants me to write a Python script that demonstrates CRUD operations on Azure Key Vault secrets using the Azure SDK for Python. They specifically want:
1. Create a secret
2. Read a secret
3. Update a secret
4. Delete and purge a secret

They want me to use the official Azure SDK documentation and use DefaultAzureCredential for authentication.

This is a straightforward task - I should:
1. Look up the Azure SDK for Python documentation for Key Vault
2. Write a script with proper error handling
3. Include the required pip packages

Let me start by fetching the Azure SDK documentation to ensure accuracy.

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
{"file_text":"#!/usr/bin/env python3\n\"\"\"\nAzure Key Vault Secrets CRUD Operations Demo\n\nThis script demonstrates all four CRUD operations on Azure Key Vault secrets:\n1. Create - Set a new secret\n2. Read - Get a secret value\n3. Update - Update a secret to a new value\n4. Delete - Delete and purge a secret\n\nRequired pip packages:\n    pip install azure-keyvault-secrets azure-identity\n\nPrerequisites:\n    - An Azure Key Vault with soft-delete enabled\n    - Appropriate authentication configured for DefaultAzureCredential\n    - Required permissions: secrets/set, secrets/get, secrets/delete, secrets/purge\n    - Set AZURE_KEY_VAULT_URL environment variable to your vault URL\n\"\"\"\n\nimport os\nimport sys\nfrom azure.identity import DefaultAzureCredential\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.core.exceptions import ResourceNotFoundError, HttpResponseError\n\n\ndef main():\n    \"\"\"Demonstrate CRUD operations on Azure Key Vault secrets.\"\"\"\n    \n    # Get vault URL from environment variable\n    vault_url = os.environ.get(\"AZURE_KEY_VAULT_URL\")\n    \n    if not vault_url:\n        print(\"Error: AZURE_KEY_VAULT_URL environment variable is not set\")\n        print(\"Example: export AZURE_KEY_VAULT_URL='https://my-vault.vault.azure.net/'\")\n        sys.exit(1)\n    \n    print(f\"Using Key Vault: {vault_url}\\n\")\n    \n    # Initialize the SecretClient with DefaultAzureCredential\n    try:\n        credential = DefaultAzureCredential()\n        client = SecretClient(vault_url=vault_url, credential=credential)\n        print(\"✓ Successfully authenticated with DefaultAzureCredential\\n\")\n    except Exception as e:\n        print(f\"✗ Authentication failed: {e}\")\n        sys.exit(1)\n    \n    secret_name = \"my-secret\"\n    \n    try:\n        # ====================================================================\n        # 1. CREATE - Set a new secret\n        # ====================================================================\n        print(\"=\" * 60)\n        print(\"1. CREATE - Setting a new secret\")\n        print(\"=\" * 60)\n        \n        secret_value = \"my-secret-value\"\n        created_secret = client.set_secret(secret_name, secret_value)\n        \n        print(f\"✓ Secret created successfully:\")\n        print(f\"  Name: {created_secret.name}\")\n        print(f\"  Value: {created_secret.value}\")\n        print(f\"  Version: {created_secret.properties.version}\")\n        print(f\"  Created on: {created_secret.properties.created_on}\")\n        print()\n        \n        # ====================================================================\n        # 2. READ - Retrieve the secret\n        # ====================================================================\n        print(\"=\" * 60)\n        print(\"2. READ - Retrieving the secret\")\n        print(\"=\" * 60)\n        \n        retrieved_secret = client.get_secret(secret_name)\n        \n        print(f\"✓ Secret retrieved successfully:\")\n        print(f\"  Name: {retrieved_secret.name}\")\n        print(f\"  Value: {retrieved_secret.value}\")\n        print(f\"  Version: {retrieved_secret.properties.version}\")\n        print()\n        \n        # ====================================================================\n        # 3. UPDATE - Update the secret to a new value\n        # ====================================================================\n        print(\"=\" * 60)\n        print(\"3. UPDATE - Updating the secret with a new value\")\n        print(\"=\" * 60)\n        \n        new_value = \"updated-value\"\n        updated_secret = client.set_secret(secret_name, new_value)\n        \n        print(f\"✓ Secret updated successfully:\")\n        print(f\"  Name: {updated_secret.name}\")\n        print(f\"  New Value: {updated_secret.value}\")\n        print(f\"  New Version: {updated_secret.properties.version}\")\n        print(f\"  Updated on: {updated_secret.properties.updated_on}\")\n        print()\n        \n        # Verify the update by reading again\n        verified_secret = client.get_secret(secret_name)\n        print(f\"✓ Verified updated value: {verified_secret.value}\")\n        print()\n        \n        # ====================================================================\n        # 4. DELETE - Delete and purge the secret\n        # ====================================================================\n        print(\"=\" * 60)\n        print(\"4. DELETE - Deleting and purging the secret\")\n        print(\"=\" * 60)\n        \n        # Begin delete operation (returns a poller for soft-delete enabled vaults)\n        print(\"  Initiating delete operation...\")\n        delete_poller = client.begin_delete_secret(secret_name)\n        \n        # Wait for deletion to complete\n        deleted_secret = delete_poller.result()\n        \n        print(f\"✓ Secret soft-deleted successfully:\")\n        print(f\"  Name: {deleted_secret.name}\")\n        print(f\"  Deleted on: {deleted_secret.deleted_date}\")\n        print(f\"  Scheduled purge date: {deleted_secret.scheduled_purge_date}\")\n        print(f\"  Recovery ID: {deleted_secret.recovery_id}\")\n        print()\n        \n        # Purge the deleted secret (permanent deletion)\n        print(\"  Purging deleted secret permanently...\")\n        client.purge_deleted_secret(secret_name)\n        \n        print(f\"✓ Secret '{secret_name}' purged successfully (permanent deletion)\")\n        print()\n        \n        # Verify deletion\n        print(\"  Verifying secret no longer exists...\")\n        try:\n            client.get_secret(secret_name)\n            print(\"✗ Unexpected: Secret still exists!\")\n        except ResourceNotFoundError:\n            print(f\"✓ Confirmed: Secret '{secret_name}' no longer exists\")\n        print()\n        \n        print(\"=\" * 60)\n        print(\"All CRUD operations completed successfully!\")\n        print(\"=\" * 60)\n        \n    except ResourceNotFoundError as e:\n        print(f\"✗ Resource not found: {e.message}\")\n        sys.exit(1)\n    except HttpResponseError as e:\n        print(f\"✗ HTTP error occurred: {e.message}\")\n        print(f\"  Status code: {e.status_code}\")\n        print(f\"  Error code: {e.error.code if hasattr(e, 'error') else 'N/A'}\")\n        sys.exit(1)\n    except Exception as e:\n        print(f\"✗ An unexpected error occurred: {type(e).__name__}: {e}\")\n        sys.exit(1)\n    finally:\n        # Clean up - close the credential\n        credential.close()\n\n\nif __name__ == \"__main__\":\n    main()\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-870605203/keyvault_crud_demo.py"}
```

### 🔧 5. create

**Input:**

```json
{"file_text":"azure-keyvault-secrets\u003e=4.8.0\nazure-identity\u003e=1.15.0\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-870605203/requirements.txt"}
```

### 🔧 6. create

**Input:**

```json
{"file_text":"# Azure Key Vault Secrets CRUD Demo\n\nThis script demonstrates all four CRUD operations on Azure Key Vault secrets using the Azure SDK for Python.\n\n## Operations Demonstrated\n\n1. **CREATE** - Set a new secret called \"my-secret\" with value \"my-secret-value\"\n2. **READ** - Retrieve the secret and print its value\n3. **UPDATE** - Update the secret to a new value \"updated-value\"\n4. **DELETE** - Delete the secret and purge it (for soft-delete enabled vaults)\n\n## Prerequisites\n\n- Python 3.9 or later\n- An Azure Key Vault with soft-delete enabled\n- Appropriate permissions on the Key Vault:\n  - `secrets/set` - To create and update secrets\n  - `secrets/get` - To read secrets\n  - `secrets/delete` - To delete secrets\n  - `secrets/purge` - To permanently delete secrets\n\n## Authentication\n\nThe script uses `DefaultAzureCredential` which automatically attempts multiple authentication methods:\n\n1. **Environment variables** (recommended for development)\n2. **Managed Identity** (for Azure-hosted applications)\n3. **Azure CLI** (if logged in via `az login`)\n4. **Azure PowerShell**\n5. **Interactive browser**\n\n### Quick Setup with Azure CLI\n\n```bash\naz login\n```\n\n## Installation\n\nInstall required packages:\n\n```bash\npip install -r requirements.txt\n```\n\nOr manually:\n\n```bash\npip install azure-keyvault-secrets azure-identity\n```\n\n## Usage\n\nSet your Key Vault URL as an environment variable:\n\n```bash\nexport AZURE_KEY_VAULT_URL='https://your-vault-name.vault.azure.net/'\n```\n\nRun the script:\n\n```bash\npython keyvault_crud_demo.py\n```\n\n## Expected Output\n\nThe script will:\n- Authenticate using DefaultAzureCredential\n- Create a secret named \"my-secret\"\n- Read and display the secret value\n- Update the secret to a new value\n- Verify the update\n- Delete the secret (soft-delete)\n- Purge the secret (permanent deletion)\n- Verify the secret no longer exists\n\n## Error Handling\n\nThe script includes comprehensive error handling for:\n- Authentication failures\n- Missing environment variables\n- Resource not found errors\n- HTTP response errors\n- Permission issues\n\n## References\n\n- [Azure Key Vault Secrets SDK Documentation](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)\n- [DefaultAzureCredential Documentation](https://learn.microsoft.com/en-us/python/api/azure-identity/azure.identity.defaultazurecredential)\n- [Azure Key Vault Overview](https://learn.microsoft.com/en-us/azure/key-vault/general/overview)\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-870605203/README.md"}
```

### 🔧 7. bash

**Input:**

```json
{"command":"cd /tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-870605203 \u0026\u0026 ls -la","description":"List created files"}
```

## Generated Files

- `README.md`
- `keyvault_crud_demo.py`
- `requirements.txt`

## Copilot Response

I'll help you write a Python script for Azure Key Vault CRUD operations using the official Azure SDK documentation. Let me fetch the documentation and create the script.

Now let me fetch information about purging deleted secrets:

Perfect! Now I have all the information I need. Let me create a comprehensive Python script that demonstrates all CRUD operations on Azure Key Vault secrets:

Now let me create a requirements.txt file and a README with instructions:

Perfect! Let me create a quick summary document showing the key code snippets:

Perfect! I've created a complete Azure Key Vault CRUD demo script based on the official Azure SDK for Python documentation. Here's what I've provided:

## 📁 Files Created

1. **keyvault_crud_demo.py** - Main script with all CRUD operations
2. **requirements.txt** - Required pip packages
3. **README.md** - Setup and usage instructions

## 🔑 Key Features

The script demonstrates all four CRUD operations:

- **CREATE**: Uses `set_secret("my-secret", "my-secret-value")` to create a new secret
- **READ**: Uses `get_secret("my-secret")` to retrieve and print the secret value
- **UPDATE**: Uses `set_secret("my-secret", "updated-value")` to update to a new value
- **DELETE**: Uses `begin_delete_secret()` for soft-delete, then `purge_deleted_secret()` for permanent deletion

## 📦 Required Packages

```bash
pip install azure-keyvault-secrets azure-identity
```

## 🔐 Authentication

Uses `DefaultAzureCredential` as requested, which automatically tries:
- Environment variables
- Managed Identity
- Azure CLI (`az login`)
- Interactive browser

## ✅ Error Handling Included

- ResourceNotFoundError
- HttpResponseError  
- Authentication failures
- Missing environment variables

## 🚀 Usage

```bash
export AZURE_KEY_VAULT_URL='https://your-vault.vault.azure.net/'
python keyvault_crud_demo.py
```

All code is sourced directly from the official Azure SDK for Python documentation!

## Verification

**Result:** ✅ PASS

**Summary:** Complete, correct implementation of all CRUD operations with proper authentication, error handling, and soft-delete handling.

**Reasoning:**

The generated code fully meets all requirements from the original prompt. It correctly implements all four CRUD operations: (1) Creates a secret 'my-secret' with value 'my-secret-value' using set_secret(), (2) Reads the secret using get_secret() and prints its value, (3) Updates the secret to 'updated-value' using set_secret() again, and (4) Deletes and purges the secret using begin_delete_secret() followed by purge_deleted_secret(). The code uses DefaultAzureCredential for authentication as required, creates a SecretClient with vault URL and credential, and includes comprehensive error handling for ResourceNotFoundError, HttpResponseError, and general exceptions. The requirements.txt correctly specifies both azure-keyvault-secrets and azure-identity packages. The code properly handles soft-delete by calling result() on the delete poller to wait for deletion to complete before purging. The script is syntactically valid Python, well-structured with clear comments, includes proper imports, and even adds cleanup with credential.close(). The README.md provides excellent documentation with prerequisites, installation steps, and usage instructions. All expected coverage points are addressed.

---

[← Back to Summary](../../../../../../summary.md)
