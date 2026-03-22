# Evaluation Report: key-vault-dp-python-crud

**Config:** baseline | **Result:** ✅ PASSED | **Duration:** 94.1s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-crud` |
| Config | baseline |
| Result | ✅ PASSED |
| Duration | 94.1s |
| Timestamp | 2026-03-22T21:53:05Z |
| Files Generated | 4 |
| Event Count | 105 |

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

The user wants me to write a Python script using the Azure SDK for Python that demonstrates CRUD operations on Azure Key Vault secrets. They want me to use only the Azure SDK documentation as a reference.

Let me break down what I need to do:
1. Look up the Azure SDK for Python Key Vault documentation
2. Write a script that:
   - Creates a secret
   - Reads the secret
   - Updates the secret
   - Deletes and purges the secret
3. Use DefaultAzureCredential for authentication
4. Include error handling
5. Show required pip packages

I should first search for the Azure SDK documentation to ensure I'm using the correct API methods and patterns.

## Tool Calls

### 🔧 1. report_intent

**Input:**

```json
{"intent":"Writing Azure Key Vault script"}
```

### 🔧 2. web_fetch

**Input:**

```json
{"max_length":10000,"url":"https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme"}
```

### 🔧 3. web_fetch

**Input:**

```json
{"max_length":5000,"start_index":10000,"url":"https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme"}
```

### 🔧 4. web_fetch

**Input:**

```json
{"max_length":5000,"url":"https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient"}
```

### 🔧 5. create

**Input:**

```json
{"file_text":"#!/usr/bin/env python3\n\"\"\"\nAzure Key Vault Secrets CRUD Operations Script\n\nThis script demonstrates all four CRUD operations on Azure Key Vault secrets:\n1. Create - Set a new secret\n2. Read - Get a secret value\n3. Update - Update secret to a new value\n4. Delete - Delete and purge a secret (soft-delete enabled vault)\n\nRequired packages:\n    pip install azure-keyvault-secrets azure-identity\n\nPrerequisites:\n    - Azure subscription with a Key Vault created\n    - Set environment variable AZURE_KEY_VAULT_URL with your vault URL\n    - Appropriate authentication configured for DefaultAzureCredential\n\"\"\"\n\nimport os\nimport sys\nfrom azure.identity import DefaultAzureCredential\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.core.exceptions import (\n    ResourceNotFoundError,\n    HttpResponseError,\n    ClientAuthenticationError\n)\n\n\ndef main():\n    \"\"\"Main function to demonstrate CRUD operations on Azure Key Vault secrets.\"\"\"\n    \n    # Get the Key Vault URL from environment variable\n    vault_url = os.environ.get(\"AZURE_KEY_VAULT_URL\")\n    if not vault_url:\n        print(\"Error: AZURE_KEY_VAULT_URL environment variable is not set.\")\n        print(\"Please set it to your Key Vault URL (e.g., https://my-key-vault.vault.azure.net/)\")\n        sys.exit(1)\n    \n    print(f\"Using Key Vault: {vault_url}\\n\")\n    \n    # Initialize credential and client\n    try:\n        credential = DefaultAzureCredential()\n        secret_client = SecretClient(vault_url=vault_url, credential=credential)\n        print(\"✓ Successfully initialized SecretClient with DefaultAzureCredential\\n\")\n    except ClientAuthenticationError as e:\n        print(f\"Authentication failed: {e.message}\")\n        sys.exit(1)\n    except Exception as e:\n        print(f\"Error initializing client: {e}\")\n        sys.exit(1)\n    \n    secret_name = \"my-secret\"\n    \n    # ========================================\n    # 1. CREATE - Set a new secret\n    # ========================================\n    print(\"=\" * 60)\n    print(\"1. CREATE - Setting a new secret\")\n    print(\"=\" * 60)\n    try:\n        initial_value = \"my-secret-value\"\n        secret = secret_client.set_secret(secret_name, initial_value)\n        print(f\"✓ Secret created successfully\")\n        print(f\"  Name: {secret.name}\")\n        print(f\"  Value: {secret.value}\")\n        print(f\"  Version: {secret.properties.version}\")\n        print(f\"  Created: {secret.properties.created_on}\\n\")\n    except HttpResponseError as e:\n        print(f\"✗ HTTP error creating secret: {e.message}\")\n        print(f\"  Status code: {e.status_code}\")\n        sys.exit(1)\n    except Exception as e:\n        print(f\"✗ Error creating secret: {e}\")\n        sys.exit(1)\n    \n    # ========================================\n    # 2. READ - Get the secret value\n    # ========================================\n    print(\"=\" * 60)\n    print(\"2. READ - Retrieving the secret\")\n    print(\"=\" * 60)\n    try:\n        retrieved_secret = secret_client.get_secret(secret_name)\n        print(f\"✓ Secret retrieved successfully\")\n        print(f\"  Name: {retrieved_secret.name}\")\n        print(f\"  Value: {retrieved_secret.value}\")\n        print(f\"  Version: {retrieved_secret.properties.version}\")\n        print(f\"  Enabled: {retrieved_secret.properties.enabled}\\n\")\n    except ResourceNotFoundError:\n        print(f\"✗ Secret '{secret_name}' not found in the vault\")\n        sys.exit(1)\n    except HttpResponseError as e:\n        print(f\"✗ HTTP error retrieving secret: {e.message}\")\n        sys.exit(1)\n    except Exception as e:\n        print(f\"✗ Error retrieving secret: {e}\")\n        sys.exit(1)\n    \n    # ========================================\n    # 3. UPDATE - Update the secret value\n    # ========================================\n    print(\"=\" * 60)\n    print(\"3. UPDATE - Updating the secret to a new value\")\n    print(\"=\" * 60)\n    try:\n        new_value = \"updated-value\"\n        updated_secret = secret_client.set_secret(secret_name, new_value)\n        print(f\"✓ Secret updated successfully\")\n        print(f\"  Name: {updated_secret.name}\")\n        print(f\"  New Value: {updated_secret.value}\")\n        print(f\"  New Version: {updated_secret.properties.version}\")\n        print(f\"  Updated: {updated_secret.properties.updated_on}\\n\")\n    except HttpResponseError as e:\n        print(f\"✗ HTTP error updating secret: {e.message}\")\n        sys.exit(1)\n    except Exception as e:\n        print(f\"✗ Error updating secret: {e}\")\n        sys.exit(1)\n    \n    # ========================================\n    # 4. DELETE - Delete and purge the secret\n    # ========================================\n    print(\"=\" * 60)\n    print(\"4. DELETE - Deleting and purging the secret\")\n    print(\"=\" * 60)\n    try:\n        # Delete the secret (soft delete)\n        print(\"Step 1: Deleting secret (soft delete)...\")\n        delete_poller = secret_client.begin_delete_secret(secret_name)\n        deleted_secret = delete_poller.result()\n        print(f\"✓ Secret deleted successfully\")\n        print(f\"  Name: {deleted_secret.name}\")\n        print(f\"  Deleted Date: {deleted_secret.deleted_date}\")\n        print(f\"  Scheduled Purge Date: {deleted_secret.scheduled_purge_date}\")\n        print(f\"  Recovery ID: {deleted_secret.recovery_id}\\n\")\n        \n        # Purge the secret (permanent deletion)\n        print(\"Step 2: Purging secret (permanent deletion)...\")\n        secret_client.purge_deleted_secret(secret_name)\n        print(f\"✓ Secret purged successfully\")\n        print(f\"  The secret '{secret_name}' has been permanently deleted.\\n\")\n        \n    except ResourceNotFoundError:\n        print(f\"✗ Secret '{secret_name}' not found for deletion\")\n        sys.exit(1)\n    except HttpResponseError as e:\n        print(f\"✗ HTTP error during delete/purge: {e.message}\")\n        print(f\"  Status code: {e.status_code}\")\n        if e.status_code == 403:\n            print(\"  Note: Ensure you have 'secrets/delete' and 'secrets/purge' permissions\")\n        sys.exit(1)\n    except Exception as e:\n        print(f\"✗ Error deleting/purging secret: {e}\")\n        sys.exit(1)\n    \n    print(\"=\" * 60)\n    print(\"All CRUD operations completed successfully!\")\n    print(\"=\" * 60)\n\n\nif __name__ == \"__main__\":\n    main()\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-1340918120/keyvault_crud.py"}
```

### 🔧 6. create

**Input:**

```json
{"file_text":"azure-keyvault-secrets\nazure-identity\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-1340918120/requirements.txt"}
```

### 🔧 7. create

**Input:**

```json
{"file_text":"# Azure Key Vault Secrets CRUD Operations\n\nThis script demonstrates all four CRUD operations on Azure Key Vault secrets using the Azure SDK for Python.\n\n## Operations Performed\n\n1. **CREATE** - Creates a new secret named \"my-secret\" with value \"my-secret-value\"\n2. **READ** - Retrieves the secret and prints its value\n3. **UPDATE** - Updates the secret to a new value \"updated-value\"\n4. **DELETE** - Deletes the secret and purges it (for soft-delete enabled vaults)\n\n## Prerequisites\n\n- Python 3.9 or later\n- An Azure subscription\n- An Azure Key Vault (with soft-delete enabled for purge operation)\n- Appropriate permissions on the Key Vault:\n  - `secrets/set` - for creating and updating secrets\n  - `secrets/get` - for reading secrets\n  - `secrets/delete` - for deleting secrets\n  - `secrets/purge` - for purging deleted secrets\n\n## Installation\n\nInstall the required packages:\n\n```bash\npip install -r requirements.txt\n```\n\nOr install individually:\n\n```bash\npip install azure-keyvault-secrets azure-identity\n```\n\n## Authentication\n\nThe script uses `DefaultAzureCredential` which supports multiple authentication methods in the following order:\n\n1. **Environment variables** - Set these for service principal authentication:\n   - `AZURE_TENANT_ID`\n   - `AZURE_CLIENT_ID`\n   - `AZURE_CLIENT_SECRET`\n\n2. **Managed Identity** - Works automatically when running on Azure resources\n\n3. **Azure CLI** - Uses credentials from `az login`\n\n4. **Azure PowerShell** - Uses credentials from `Connect-AzAccount`\n\n5. **Interactive browser** - Opens a browser for user authentication\n\nFor local development, the easiest method is to use Azure CLI:\n\n```bash\naz login\n```\n\n## Usage\n\n1. Set the Key Vault URL environment variable:\n\n```bash\nexport AZURE_KEY_VAULT_URL=\"https://your-key-vault-name.vault.azure.net/\"\n```\n\nOn Windows (PowerShell):\n```powershell\n$env:AZURE_KEY_VAULT_URL=\"https://your-key-vault-name.vault.azure.net/\"\n```\n\n2. Run the script:\n\n```bash\npython keyvault_crud.py\n```\n\n## Expected Output\n\n```\nUsing Key Vault: https://your-key-vault-name.vault.azure.net/\n\n✓ Successfully initialized SecretClient with DefaultAzureCredential\n\n============================================================\n1. CREATE - Setting a new secret\n============================================================\n✓ Secret created successfully\n  Name: my-secret\n  Value: my-secret-value\n  Version: abc123...\n  Created: 2026-03-22 21:53:00\n\n============================================================\n2. READ - Retrieving the secret\n============================================================\n✓ Secret retrieved successfully\n  Name: my-secret\n  Value: my-secret-value\n  Version: abc123...\n  Enabled: True\n\n============================================================\n3. UPDATE - Updating the secret to a new value\n============================================================\n✓ Secret updated successfully\n  Name: my-secret\n  New Value: updated-value\n  New Version: def456...\n  Updated: 2026-03-22 21:53:01\n\n============================================================\n4. DELETE - Deleting and purging the secret\n============================================================\nStep 1: Deleting secret (soft delete)...\n✓ Secret deleted successfully\n  Name: my-secret\n  Deleted Date: 2026-03-22 21:53:02\n  Scheduled Purge Date: 2026-06-20 21:53:02\n  Recovery ID: https://...\n\nStep 2: Purging secret (permanent deletion)...\n✓ Secret purged successfully\n  The secret 'my-secret' has been permanently deleted.\n\n============================================================\nAll CRUD operations completed successfully!\n============================================================\n```\n\n## Error Handling\n\nThe script includes comprehensive error handling for:\n\n- Missing environment variables\n- Authentication failures\n- HTTP errors (with status codes)\n- Resource not found errors\n- Permission errors\n\n## References\n\n- [Azure Key Vault Secrets Python SDK Documentation](https://learn.microsoft.com/python/api/overview/azure/keyvault-secrets-readme)\n- [Azure Identity Python SDK Documentation](https://learn.microsoft.com/python/api/overview/azure/identity-readme)\n- [DefaultAzureCredential Documentation](https://learn.microsoft.com/python/api/azure-identity/azure.identity.defaultazurecredential)\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-1340918120/README.md"}
```

### 🔧 8. bash

**Input:**

```json
{"command":"python3 -m py_compile keyvault_crud.py \u0026\u0026 echo \"✓ Script syntax is valid\"","description":"Check Python syntax of the script"}
```

### 🔧 9. view

**Input:**

```json
{"path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-1340918120"}
```

## Generated Files

- `README.md`
- `__pycache__/keyvault_crud.cpython-312.pyc`
- `keyvault_crud.py`
- `requirements.txt`

## Copilot Response

I'll help you create a Python script for Azure Key Vault CRUD operations using the Azure SDK documentation. Let me first check the documentation and then create the script.

Perfect! Now I have the information I need. Let me also check for the purge operation documentation.

Great! Now I have all the information I need. Let me create a comprehensive Python script that demonstrates all CRUD operations with proper error handling.

Now let me create a requirements.txt file to show the required pip packages:

Let me also create a README to provide additional context:

Perfect! Let me verify the script is syntactically correct:

Excellent! Let me create a summary document showing what's in the script:

Perfect! I've created a complete Azure Key Vault CRUD operations script based on the official Azure SDK for Python documentation. Here's what I've delivered:

## 📁 Files Created

1. **keyvault_crud.py** - Main script demonstrating all 4 CRUD operations
2. **requirements.txt** - Required pip packages
3. **README.md** - Complete usage instructions and documentation

## 🔑 Key Features

The script performs all CRUD operations as requested:

1. **CREATE**: Uses `set_secret()` to create "my-secret" with value "my-secret-value"
2. **READ**: Uses `get_secret()` to retrieve and print the secret value
3. **UPDATE**: Uses `set_secret()` again to update to "updated-value" (creates new version)
4. **DELETE**: Uses `begin_delete_secret()` for soft-delete, then `purge_deleted_secret()` for permanent deletion

## 🛡️ Error Handling

Includes comprehensive error handling for:
- Missing environment variables
- Authentication failures (`ClientAuthenticationError`)
- Resource not found (`ResourceNotFoundError`)
- HTTP errors with status codes
- Permission errors (403 with helpful message)

## 📦 Required Packages

```
azure-keyvault-secrets
azure-identity
```

## 🚀 Usage

```bash
pip install -r requirements.txt
export AZURE_KEY_VAULT_URL="https://your-vault.vault.azure.net/"
python keyvault_crud.py
```

The script uses `DefaultAzureCredential` for authentication, which supports Azure CLI, managed identity, environment variables, and interactive browser login.

## Verification

**Result:** ✅ PASS

**Summary:** Complete and correct implementation of Azure Key Vault CRUD operations with proper authentication, error handling, and all required SDK methods.

**Reasoning:**

The code fully satisfies all requirements. It correctly implements all four CRUD operations: (1) CREATE using set_secret() with 'my-secret' and 'my-secret-value', (2) READ using get_secret() to retrieve and print the value, (3) UPDATE using set_secret() with 'updated-value', and (4) DELETE using begin_delete_secret() followed by purge_deleted_secret(). The script properly uses DefaultAzureCredential for authentication and creates a SecretClient with vault URL and credential. It correctly handles soft-delete by calling .result() on the delete poller to wait for completion before purging. Error handling is comprehensive, covering ClientAuthenticationError, ResourceNotFoundError, HttpResponseError, and generic exceptions. The requirements.txt includes both required packages (azure-keyvault-secrets and azure-identity). The code is syntactically valid Python, well-structured, and includes helpful output formatting. All method names match the expected SDK methods, and the implementation follows Azure SDK best practices.

---

[← Back to Summary](../../../../../../summary.md)
