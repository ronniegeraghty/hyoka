# Evaluation Report: key-vault-dp-python-crud

**Config:** baseline | **Result:** ✅ PASSED | **Duration:** 60.9s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-crud` |
| Config | baseline |
| Result | ✅ PASSED |
| Duration | 60.9s |
| Timestamp | 2026-03-22T21:25:36Z |
| Files Generated | 3 |
| Event Count | 41 |

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

The user wants me to write a Python script that demonstrates CRUD operations on Azure Key Vault secrets using the Azure SDK for Python. They specifically want:

1. Create a secret
2. Read the secret
3. Update the secret
4. Delete and purge the secret
5. Use DefaultAzureCredential
6. Include error handling
7. Show required pip packages

I should use only the Azure SDK for Python documentation to ensure accuracy. This is a straightforward task - I'll create a Python script that demonstrates these operations.

Let me start by creating the script with proper imports, error handling, and comments.

## Tool Calls

### 🔧 1. report_intent

**Input:**

```json
{"intent":"Creating Azure Key Vault CRUD script"}
```

### 🔧 2. create

**Input:**

```json
{"file_text":"#!/usr/bin/env python3\n\"\"\"\nAzure Key Vault Secrets CRUD Operations Demo\n\nThis script demonstrates all four CRUD operations on Azure Key Vault secrets:\n- Create a secret\n- Read a secret\n- Update a secret\n- Delete and purge a secret\n\nRequired pip packages:\n    pip install azure-keyvault-secrets azure-identity\n\"\"\"\n\nimport os\nimport sys\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.identity import DefaultAzureCredential\nfrom azure.core.exceptions import ResourceNotFoundError, HttpResponseError\n\n\ndef main():\n    # Get Key Vault URL from environment variable\n    vault_url = os.getenv(\"AZURE_KEY_VAULT_URL\")\n    \n    if not vault_url:\n        print(\"Error: AZURE_KEY_VAULT_URL environment variable is not set\", file=sys.stderr)\n        print(\"Example: export AZURE_KEY_VAULT_URL='https://your-vault-name.vault.azure.net/'\", file=sys.stderr)\n        sys.exit(1)\n    \n    try:\n        # Authenticate using DefaultAzureCredential\n        credential = DefaultAzureCredential()\n        \n        # Create a SecretClient\n        client = SecretClient(vault_url=vault_url, credential=credential)\n        \n        secret_name = \"my-secret\"\n        \n        # 1. CREATE: Create a new secret\n        print(f\"Creating secret '{secret_name}'...\")\n        try:\n            created_secret = client.set_secret(secret_name, \"my-secret-value\")\n            print(f\"✓ Secret created successfully\")\n            print(f\"  Name: {created_secret.name}\")\n            print(f\"  Value: {created_secret.value}\")\n            print(f\"  Version: {created_secret.properties.version}\")\n            print()\n        except HttpResponseError as e:\n            print(f\"✗ Failed to create secret: {e.message}\", file=sys.stderr)\n            sys.exit(1)\n        \n        # 2. READ: Read the secret back\n        print(f\"Reading secret '{secret_name}'...\")\n        try:\n            retrieved_secret = client.get_secret(secret_name)\n            print(f\"✓ Secret retrieved successfully\")\n            print(f\"  Name: {retrieved_secret.name}\")\n            print(f\"  Value: {retrieved_secret.value}\")\n            print(f\"  Version: {retrieved_secret.properties.version}\")\n            print()\n        except ResourceNotFoundError:\n            print(f\"✗ Secret '{secret_name}' not found\", file=sys.stderr)\n            sys.exit(1)\n        except HttpResponseError as e:\n            print(f\"✗ Failed to read secret: {e.message}\", file=sys.stderr)\n            sys.exit(1)\n        \n        # 3. UPDATE: Update the secret to a new value\n        print(f\"Updating secret '{secret_name}'...\")\n        try:\n            updated_secret = client.set_secret(secret_name, \"updated-value\")\n            print(f\"✓ Secret updated successfully\")\n            print(f\"  Name: {updated_secret.name}\")\n            print(f\"  Value: {updated_secret.value}\")\n            print(f\"  Version: {updated_secret.properties.version}\")\n            print()\n        except HttpResponseError as e:\n            print(f\"✗ Failed to update secret: {e.message}\", file=sys.stderr)\n            sys.exit(1)\n        \n        # 4. DELETE: Delete the secret\n        print(f\"Deleting secret '{secret_name}'...\")\n        try:\n            deleted_secret = client.begin_delete_secret(secret_name)\n            # Wait for the delete operation to complete\n            deleted_secret_result = deleted_secret.result()\n            print(f\"✓ Secret deleted successfully\")\n            print(f\"  Name: {deleted_secret_result.name}\")\n            print(f\"  Deleted on: {deleted_secret_result.deleted_date}\")\n            print(f\"  Scheduled purge: {deleted_secret_result.scheduled_purge_date}\")\n            print()\n        except ResourceNotFoundError:\n            print(f\"✗ Secret '{secret_name}' not found\", file=sys.stderr)\n            sys.exit(1)\n        except HttpResponseError as e:\n            print(f\"✗ Failed to delete secret: {e.message}\", file=sys.stderr)\n            sys.exit(1)\n        \n        # PURGE: Purge the deleted secret (for soft-delete enabled vaults)\n        print(f\"Purging secret '{secret_name}'...\")\n        try:\n            client.purge_deleted_secret(secret_name)\n            print(f\"✓ Secret purged successfully\")\n            print(f\"  The secret '{secret_name}' has been permanently deleted\")\n        except ResourceNotFoundError:\n            print(f\"✗ Deleted secret '{secret_name}' not found\", file=sys.stderr)\n            sys.exit(1)\n        except HttpResponseError as e:\n            print(f\"✗ Failed to purge secret: {e.message}\", file=sys.stderr)\n            print(f\"  Note: Purging may require additional permissions or the vault may not have soft-delete enabled\", file=sys.stderr)\n            sys.exit(1)\n    \n    except Exception as e:\n        print(f\"✗ Unexpected error: {str(e)}\", file=sys.stderr)\n        sys.exit(1)\n    finally:\n        # Close the credential\n        credential.close()\n    \n    print(\"\\n✓ All CRUD operations completed successfully!\")\n\n\nif __name__ == \"__main__\":\n    main()\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-590729329/key_vault_crud.py"}
```

### 🔧 3. create

**Input:**

```json
{"file_text":"azure-keyvault-secrets\nazure-identity\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-590729329/requirements.txt"}
```

### 🔧 4. create

**Input:**

```json
{"file_text":"# Azure Key Vault Secrets CRUD Operations\n\nThis script demonstrates all four CRUD operations on Azure Key Vault secrets using the Azure SDK for Python.\n\n## Prerequisites\n\n1. An Azure subscription\n2. An Azure Key Vault with soft-delete enabled\n3. Appropriate permissions to create, read, update, delete, and purge secrets\n4. Python 3.7 or later\n\n## Installation\n\nInstall the required packages:\n\n```bash\npip install -r requirements.txt\n```\n\nOr install them individually:\n\n```bash\npip install azure-keyvault-secrets azure-identity\n```\n\n## Authentication\n\nThis script uses `DefaultAzureCredential`, which attempts authentication through multiple methods in order:\n\n1. Environment variables\n2. Managed Identity\n3. Visual Studio Code\n4. Azure CLI\n5. Azure PowerShell\n6. Interactive browser\n\nThe easiest way to authenticate locally is using Azure CLI:\n\n```bash\naz login\n```\n\n## Configuration\n\nSet the Key Vault URL as an environment variable:\n\n```bash\nexport AZURE_KEY_VAULT_URL='https://your-vault-name.vault.azure.net/'\n```\n\nOn Windows (PowerShell):\n\n```powershell\n$env:AZURE_KEY_VAULT_URL='https://your-vault-name.vault.azure.net/'\n```\n\n## Usage\n\nRun the script:\n\n```bash\npython key_vault_crud.py\n```\n\n## Operations Performed\n\n1. **CREATE**: Creates a new secret called \"my-secret\" with value \"my-secret-value\"\n2. **READ**: Reads the secret back and prints its value\n3. **UPDATE**: Updates the secret to a new value \"updated-value\"\n4. **DELETE**: Deletes the secret (soft-delete)\n5. **PURGE**: Permanently purges the deleted secret\n\n## Required Permissions\n\nYour Azure identity needs the following Key Vault permissions:\n\n- Get (secrets)\n- List (secrets)\n- Set (secrets)\n- Delete (secrets)\n- Purge (secrets)\n\n## Notes\n\n- The Key Vault must have soft-delete enabled for the purge operation\n- Purging a secret permanently deletes it and cannot be undone\n- Each update creates a new version of the secret\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-590729329/README.md"}
```

## Generated Files

- `README.md`
- `key_vault_crud.py`
- `requirements.txt`

## Copilot Response

I'll create a Python script demonstrating all four CRUD operations on Azure Key Vault secrets using the Azure SDK for Python.

Perfect! I've created a complete Azure Key Vault secrets CRUD script with the following files:

## 📄 Files Created

1. **key_vault_crud.py** - Main script demonstrating all CRUD operations
2. **requirements.txt** - Required pip packages
3. **README.md** - Documentation and usage instructions

## 🔑 Key Features

The script performs all four CRUD operations:

1. **CREATE** - Uses `client.set_secret()` to create "my-secret"
2. **READ** - Uses `client.get_secret()` to retrieve the value
3. **UPDATE** - Uses `client.set_secret()` again with new value
4. **DELETE** - Uses `client.begin_delete_secret()` for soft-delete
5. **PURGE** - Uses `client.purge_deleted_secret()` for permanent deletion

## ✨ Features Included

- ✅ **DefaultAzureCredential** for authentication
- ✅ **Comprehensive error handling** with specific exception types
- ✅ **Clear output** with status indicators
- ✅ **Environment variable configuration** for vault URL
- ✅ **Proper resource cleanup**

## 📦 Required Packages

```bash
pip install azure-keyvault-secrets azure-identity
```

## 🚀 Usage

```bash
export AZURE_KEY_VAULT_URL='https://your-vault-name.vault.azure.net/'
python key_vault_crud.py
```

The script will authenticate using DefaultAzureCredential (works with Azure CLI, Managed Identity, etc.) and perform all operations sequentially with clear output for each step.

## Verification

**Result:** ✅ PASS

**Summary:** Code fully meets all requirements with correct CRUD operations, authentication, error handling, and documentation.

**Reasoning:**

The generated code fully satisfies all requirements: (1) It performs all four CRUD operations correctly - creates 'my-secret' with 'my-secret-value', reads and prints it, updates to 'updated-value', deletes and purges. (2) Uses correct Azure SDK packages (azure-keyvault-secrets, azure-identity) shown in requirements.txt. (3) Implements DefaultAzureCredential for authentication as specified. (4) Creates SecretClient properly with vault_url and credential. (5) Uses correct SDK methods: set_secret() for create/update, get_secret() for read, begin_delete_secret() for delete, and purge_deleted_secret() for purge. (6) Handles soft-delete correctly by calling .result() on the delete poller to wait for completion before purging. (7) Includes comprehensive error handling with ResourceNotFoundError and HttpResponseError. (8) Code is syntactically valid Python with proper imports, structure, and will execute correctly. (9) Excellent documentation in README.md covering installation, authentication, configuration, and permissions. The code demonstrates production-quality practices including environment variable configuration, detailed error messages, and proper credential cleanup.

---

[← Back to Summary](../../../../../../summary.md)
