# Evaluation Report: key-vault-dp-python-crud

**Config:** baseline | **Result:** ✅ PASSED | **Duration:** 52.7s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-crud` |
| Config | baseline |
| Result | ✅ PASSED |
| Duration | 52.7s |
| Timestamp | 2026-03-22T00:09:00Z |
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

They want me to use only Azure SDK documentation, DefaultAzureCredential, and include error handling and pip packages.

This is a straightforward request - I need to create a Python script file. Let me start by reporting my intent and creating the script.

## Tool Calls

### 🔧 1. report_intent

**Input:**

```json
{"intent":"Creating Key Vault CRUD script"}
```

### 🔧 2. create

**Input:**

```json
{"file_text":"\"\"\"\nAzure Key Vault Secret CRUD Operations\n\nThis script demonstrates Create, Read, Update, and Delete operations\non Azure Key Vault secrets using the Azure SDK for Python.\n\nRequired packages:\n    pip install azure-keyvault-secrets azure-identity\n\"\"\"\n\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.identity import DefaultAzureCredential\nfrom azure.core.exceptions import ResourceNotFoundError, HttpResponseError\nimport sys\n\n\ndef main():\n    # Replace with your Key Vault URL\n    # Format: https://\u003cyour-key-vault-name\u003e.vault.azure.net/\n    key_vault_url = input(\"Enter your Key Vault URL: \").strip()\n    \n    if not key_vault_url:\n        print(\"Error: Key Vault URL is required\")\n        sys.exit(1)\n    \n    try:\n        # Authenticate using DefaultAzureCredential\n        credential = DefaultAzureCredential()\n        \n        # Create a SecretClient\n        client = SecretClient(vault_url=key_vault_url, credential=credential)\n        \n        print(f\"\\nConnected to Key Vault: {key_vault_url}\")\n        print(\"-\" * 60)\n        \n        # 1. CREATE: Create a new secret\n        print(\"\\n1. CREATE: Creating secret 'my-secret'...\")\n        secret_name = \"my-secret\"\n        secret_value = \"my-secret-value\"\n        \n        created_secret = client.set_secret(secret_name, secret_value)\n        print(f\"   ✓ Secret created: {created_secret.name}\")\n        print(f\"   - Value: {created_secret.value}\")\n        print(f\"   - Version: {created_secret.properties.version}\")\n        \n        # 2. READ: Read the secret back\n        print(\"\\n2. READ: Reading secret 'my-secret'...\")\n        retrieved_secret = client.get_secret(secret_name)\n        print(f\"   ✓ Secret retrieved: {retrieved_secret.name}\")\n        print(f\"   - Value: {retrieved_secret.value}\")\n        print(f\"   - Version: {retrieved_secret.properties.version}\")\n        \n        # 3. UPDATE: Update the secret to a new value\n        print(\"\\n3. UPDATE: Updating secret 'my-secret'...\")\n        new_value = \"updated-value\"\n        updated_secret = client.set_secret(secret_name, new_value)\n        print(f\"   ✓ Secret updated: {updated_secret.name}\")\n        print(f\"   - New Value: {updated_secret.value}\")\n        print(f\"   - New Version: {updated_secret.properties.version}\")\n        \n        # Verify the update\n        verification = client.get_secret(secret_name)\n        print(f\"   - Verified Value: {verification.value}\")\n        \n        # 4. DELETE: Delete and purge the secret\n        print(\"\\n4. DELETE: Deleting secret 'my-secret'...\")\n        poller = client.begin_delete_secret(secret_name)\n        deleted_secret = poller.result()\n        print(f\"   ✓ Secret deleted: {deleted_secret.name}\")\n        print(f\"   - Scheduled purge date: {deleted_secret.scheduled_purge_date}\")\n        print(f\"   - Deleted date: {deleted_secret.deleted_date}\")\n        \n        # Purge the deleted secret (for soft-delete enabled vaults)\n        print(\"\\n   Purging deleted secret...\")\n        client.purge_deleted_secret(secret_name)\n        print(f\"   ✓ Secret purged permanently\")\n        \n        print(\"\\n\" + \"-\" * 60)\n        print(\"All CRUD operations completed successfully!\")\n        \n    except ResourceNotFoundError as e:\n        print(f\"\\n✗ Error: Resource not found - {e.message}\")\n        sys.exit(1)\n    except HttpResponseError as e:\n        print(f\"\\n✗ HTTP Error: {e.status_code} - {e.message}\")\n        if e.status_code == 401:\n            print(\"  Hint: Check your authentication credentials\")\n        elif e.status_code == 403:\n            print(\"  Hint: Check your Key Vault access permissions\")\n        sys.exit(1)\n    except Exception as e:\n        print(f\"\\n✗ Unexpected error: {type(e).__name__} - {str(e)}\")\n        sys.exit(1)\n    finally:\n        # Clean up credential\n        if 'credential' in locals():\n            credential.close()\n\n\nif __name__ == \"__main__\":\n    main()\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-1294133265/key_vault_crud.py"}
```

### 🔧 3. create

**Input:**

```json
{"file_text":"azure-keyvault-secrets\u003e=4.7.0\nazure-identity\u003e=1.12.0\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-1294133265/requirements.txt"}
```

### 🔧 4. create

**Input:**

```json
{"file_text":"# Azure Key Vault Secret CRUD Operations\n\nThis script demonstrates all four CRUD operations on Azure Key Vault secrets using the Azure SDK for Python.\n\n## Prerequisites\n\n1. An Azure subscription\n2. An Azure Key Vault with soft-delete enabled\n3. Appropriate permissions (Secret Get, Set, Delete, Purge)\n4. Python 3.7 or higher\n\n## Installation\n\nInstall the required packages:\n\n```bash\npip install -r requirements.txt\n```\n\nOr install individually:\n\n```bash\npip install azure-keyvault-secrets azure-identity\n```\n\n## Authentication\n\nThis script uses `DefaultAzureCredential`, which attempts multiple authentication methods in order:\n1. Environment variables\n2. Managed Identity\n3. Azure CLI\n4. Azure PowerShell\n5. Interactive browser\n\nFor local development, the easiest method is Azure CLI:\n\n```bash\naz login\n```\n\n## Usage\n\nRun the script:\n\n```bash\npython key_vault_crud.py\n```\n\nWhen prompted, enter your Key Vault URL in the format:\n```\nhttps://\u003cyour-key-vault-name\u003e.vault.azure.net/\n```\n\n## Operations Performed\n\n1. **CREATE**: Creates a secret named \"my-secret\" with value \"my-secret-value\"\n2. **READ**: Retrieves the secret and prints its value\n3. **UPDATE**: Updates the secret to a new value \"updated-value\"\n4. **DELETE**: Soft-deletes the secret and then purges it permanently\n\n## Error Handling\n\nThe script includes error handling for:\n- Resource not found errors\n- HTTP errors (401 Unauthorized, 403 Forbidden, etc.)\n- General exceptions\n\n## Required Permissions\n\nEnsure your identity has the following Key Vault permissions:\n- `Secret Get`\n- `Secret Set`\n- `Secret Delete`\n- `Secret Purge`\n\nYou can assign these via Azure RBAC role \"Key Vault Secrets Officer\" or through access policies.\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-1294133265/README.md"}
```

## Generated Files

- `README.md`
- `key_vault_crud.py`
- `requirements.txt`

## Copilot Response

I'll create a Python script that demonstrates all four CRUD operations on Azure Key Vault secrets using the Azure SDK for Python.

Perfect! I've created a complete Azure Key Vault CRUD script with the following files:

## 📁 Files Created

1. **key_vault_crud.py** - Main script with all four CRUD operations:
   - ✅ **CREATE**: Creates "my-secret" with value "my-secret-value"
   - ✅ **READ**: Retrieves and prints the secret value
   - ✅ **UPDATE**: Updates to "updated-value"
   - ✅ **DELETE**: Soft-deletes and purges the secret

2. **requirements.txt** - Required packages:
   - `azure-keyvault-secrets>=4.7.0`
   - `azure-identity>=1.12.0`

3. **README.md** - Complete documentation with setup and usage instructions

## 🔑 Key Features

- Uses `DefaultAzureCredential` for flexible authentication
- Comprehensive error handling for common issues (401, 403, ResourceNotFoundError)
- Clean output showing each operation's result
- Proper resource cleanup
- Handles soft-delete enabled vaults with purge operation

## 🚀 Quick Start

```bash
pip install -r requirements.txt
python key_vault_crud.py
```

The script will prompt for your Key Vault URL and then execute all four CRUD operations in sequence!

## Verification

**Result:** ✅ PASS

**Summary:** Code fully implements all CRUD operations with correct SDK usage, proper authentication, and comprehensive error handling.

**Reasoning:**

The code fully meets all requirements from the original prompt. It correctly implements all four CRUD operations: (1) Creates 'my-secret' with value 'my-secret-value' using set_secret(), (2) Reads it back with get_secret() and prints the value, (3) Updates it to 'updated-value' using set_secret() again, (4) Deletes and purges using begin_delete_secret() followed by purge_deleted_secret(). It uses DefaultAzureCredential for authentication as required. The code properly handles soft-delete by using poller.result() to wait for delete completion before purging. Error handling includes ResourceNotFoundError and HttpResponseError with helpful hints for common issues (401/403). The requirements.txt file specifies both required packages (azure-keyvault-secrets and azure-identity). The code is syntactically valid Python, uses correct SDK methods, and includes proper imports. The README provides clear installation instructions. All expected coverage points are addressed: package installation, SecretClient creation with vault URL and credential, all four required methods, soft-delete handling, and exception handling.

---

[← Back to Summary](../../../../../../summary.md)
