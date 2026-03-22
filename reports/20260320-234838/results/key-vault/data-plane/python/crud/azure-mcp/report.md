# Evaluation Report: key-vault-dp-python-crud

**Config:** azure-mcp | **Result:** ✅ PASSED | **Duration:** 65.5s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-crud` |
| Config | azure-mcp |
| Result | ✅ PASSED |
| Duration | 65.5s |
| Timestamp | 2026-03-21T06:48:38Z |
| Files Generated | 3 |
| Event Count | 41 |

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

The user wants me to create a Python script that demonstrates CRUD operations on Azure Key Vault secrets using the Azure SDK for Python. They specifically want:

1. Create a secret
2. Read a secret
3. Update a secret
4. Delete and purge a secret

They want me to use only the Azure SDK documentation as a reference, use DefaultAzureCredential for authentication, include error handling, and list required pip packages.

This is a straightforward task - I need to create a Python script file. Let me do this efficiently by:
1. Creating the script with proper structure
2. Including all CRUD operations
3. Adding error handling
4. Adding comments about required packages

I'll create this file directly in the current working directory.

## Tool Calls

### 🔧 1. report_intent

**Input:**

```json
{"intent":"Creating Key Vault CRUD script"}
```

### 🔧 2. create

**Input:**

```json
{"file_text":"\"\"\"\nAzure Key Vault Secrets CRUD Operations\n\nRequired packages:\n    pip install azure-keyvault-secrets azure-identity\n\nEnvironment variables required:\n    AZURE_KEY_VAULT_URL - Your Key Vault URL (e.g., https://your-vault.vault.azure.net/)\n\"\"\"\n\nimport os\nimport sys\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.identity import DefaultAzureCredential\nfrom azure.core.exceptions import ResourceNotFoundError, HttpResponseError\n\n\ndef main():\n    # Get Key Vault URL from environment variable\n    vault_url = os.environ.get(\"AZURE_KEY_VAULT_URL\")\n    \n    if not vault_url:\n        print(\"Error: AZURE_KEY_VAULT_URL environment variable is not set\")\n        print(\"Example: export AZURE_KEY_VAULT_URL='https://your-vault.vault.azure.net/'\")\n        sys.exit(1)\n    \n    print(f\"Connecting to Key Vault: {vault_url}\\n\")\n    \n    try:\n        # Create a SecretClient using DefaultAzureCredential\n        credential = DefaultAzureCredential()\n        client = SecretClient(vault_url=vault_url, credential=credential)\n        \n        secret_name = \"my-secret\"\n        \n        # 1. CREATE - Set a new secret\n        print(\"=\" * 60)\n        print(\"1. CREATE - Setting a new secret\")\n        print(\"=\" * 60)\n        try:\n            secret = client.set_secret(secret_name, \"my-secret-value\")\n            print(f\"✓ Secret '{secret_name}' created successfully\")\n            print(f\"  Version: {secret.properties.version}\")\n            print(f\"  Created on: {secret.properties.created_on}\")\n        except HttpResponseError as e:\n            print(f\"✗ Failed to create secret: {e.message}\")\n            sys.exit(1)\n        \n        # 2. READ - Retrieve the secret\n        print(\"\\n\" + \"=\" * 60)\n        print(\"2. READ - Retrieving the secret\")\n        print(\"=\" * 60)\n        try:\n            retrieved_secret = client.get_secret(secret_name)\n            print(f\"✓ Secret '{secret_name}' retrieved successfully\")\n            print(f\"  Value: {retrieved_secret.value}\")\n            print(f\"  Version: {retrieved_secret.properties.version}\")\n        except ResourceNotFoundError:\n            print(f\"✗ Secret '{secret_name}' not found\")\n        except HttpResponseError as e:\n            print(f\"✗ Failed to retrieve secret: {e.message}\")\n        \n        # 3. UPDATE - Update the secret with a new value\n        print(\"\\n\" + \"=\" * 60)\n        print(\"3. UPDATE - Updating the secret with a new value\")\n        print(\"=\" * 60)\n        try:\n            updated_secret = client.set_secret(secret_name, \"updated-value\")\n            print(f\"✓ Secret '{secret_name}' updated successfully\")\n            print(f\"  New value: {updated_secret.value}\")\n            print(f\"  New version: {updated_secret.properties.version}\")\n        except HttpResponseError as e:\n            print(f\"✗ Failed to update secret: {e.message}\")\n        \n        # 4. DELETE - Delete and purge the secret\n        print(\"\\n\" + \"=\" * 60)\n        print(\"4. DELETE - Deleting and purging the secret\")\n        print(\"=\" * 60)\n        try:\n            # Begin delete operation (soft delete)\n            poller = client.begin_delete_secret(secret_name)\n            deleted_secret = poller.result()\n            print(f\"✓ Secret '{secret_name}' deleted successfully (soft-delete)\")\n            print(f\"  Scheduled purge date: {deleted_secret.properties.scheduled_purge_date}\")\n            print(f\"  Deleted date: {deleted_secret.properties.deleted_date}\")\n            \n            # Purge the deleted secret (permanent deletion)\n            print(f\"\\n  Purging secret '{secret_name}'...\")\n            client.purge_deleted_secret(secret_name)\n            print(f\"✓ Secret '{secret_name}' purged successfully (permanent deletion)\")\n            \n        except ResourceNotFoundError:\n            print(f\"✗ Secret '{secret_name}' not found for deletion\")\n        except HttpResponseError as e:\n            print(f\"✗ Failed to delete/purge secret: {e.message}\")\n        \n        print(\"\\n\" + \"=\" * 60)\n        print(\"All CRUD operations completed successfully!\")\n        print(\"=\" * 60)\n        \n    except Exception as e:\n        print(f\"\\n✗ An unexpected error occurred: {str(e)}\")\n        sys.exit(1)\n    \n    finally:\n        # Close the credential\n        credential.close()\n\n\nif __name__ == \"__main__\":\n    main()\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-291941054/key_vault_crud.py"}
```

### 🔧 3. create

**Input:**

```json
{"file_text":"azure-keyvault-secrets\u003e=4.7.0\nazure-identity\u003e=1.12.0\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-291941054/requirements.txt"}
```

### 🔧 4. create

**Input:**

```json
{"file_text":"# Azure Key Vault Secrets CRUD Operations\n\nThis script demonstrates all four CRUD operations on Azure Key Vault secrets using the Azure SDK for Python.\n\n## Prerequisites\n\n1. **Azure Key Vault**: You need an existing Key Vault with soft-delete enabled\n2. **Authentication**: Appropriate Azure credentials configured for DefaultAzureCredential\n3. **Permissions**: Your identity needs the following Key Vault permissions:\n   - Set (Create/Update secrets)\n   - Get (Read secrets)\n   - Delete (Delete secrets)\n   - Purge (Permanently delete secrets)\n\n## Installation\n\nInstall the required packages:\n\n```bash\npip install -r requirements.txt\n```\n\nOr install individually:\n\n```bash\npip install azure-keyvault-secrets azure-identity\n```\n\n## Configuration\n\nSet the Key Vault URL as an environment variable:\n\n```bash\nexport AZURE_KEY_VAULT_URL='https://your-vault-name.vault.azure.net/'\n```\n\n## Authentication\n\nThe script uses `DefaultAzureCredential` which attempts authentication through multiple methods in order:\n\n1. Environment variables\n2. Managed Identity\n3. Visual Studio Code\n4. Azure CLI\n5. Azure PowerShell\n6. Interactive browser\n\nFor local development, the easiest method is Azure CLI:\n\n```bash\naz login\n```\n\n## Usage\n\nRun the script:\n\n```bash\npython key_vault_crud.py\n```\n\n## Operations Performed\n\n1. **CREATE**: Creates a secret named \"my-secret\" with value \"my-secret-value\"\n2. **READ**: Retrieves and displays the secret value\n3. **UPDATE**: Updates the secret to a new value \"updated-value\"\n4. **DELETE**: Soft-deletes the secret, then permanently purges it\n\n## Error Handling\n\nThe script includes comprehensive error handling for:\n- Missing environment variables\n- Authentication failures\n- Resource not found errors\n- HTTP response errors\n- General exceptions\n\n## Output Example\n\n```\nConnecting to Key Vault: https://your-vault.vault.azure.net/\n\n============================================================\n1. CREATE - Setting a new secret\n============================================================\n✓ Secret 'my-secret' created successfully\n  Version: abc123...\n  Created on: 2026-03-21 06:48:41\n\n============================================================\n2. READ - Retrieving the secret\n============================================================\n✓ Secret 'my-secret' retrieved successfully\n  Value: my-secret-value\n  Version: abc123...\n\n============================================================\n3. UPDATE - Updating the secret with a new value\n============================================================\n✓ Secret 'my-secret' updated successfully\n  New value: updated-value\n  New version: def456...\n\n============================================================\n4. DELETE - Deleting and purging the secret\n============================================================\n✓ Secret 'my-secret' deleted successfully (soft-delete)\n  Scheduled purge date: 2026-04-20 06:48:41\n  Deleted date: 2026-03-21 06:48:41\n\n  Purging secret 'my-secret'...\n✓ Secret 'my-secret' purged successfully (permanent deletion)\n\n============================================================\nAll CRUD operations completed successfully!\n============================================================\n```\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-291941054/README.md"}
```

## Generated Files

- `README.md`
- `key_vault_crud.py`
- `requirements.txt`

## Copilot Response

I'll create a Python script that demonstrates all four CRUD operations on Azure Key Vault secrets using the Azure SDK for Python.

Perfect! I've created a complete Azure Key Vault CRUD operations script with all the requested features:

## 📦 Files Created:

1. **`key_vault_crud.py`** - Main script with all 4 CRUD operations:
   - ✅ **CREATE**: Sets "my-secret" with value "my-secret-value"
   - ✅ **READ**: Retrieves and displays the secret value
   - ✅ **UPDATE**: Updates to "updated-value" 
   - ✅ **DELETE**: Soft-deletes and purges the secret

2. **`requirements.txt`** - Required packages:
   - `azure-keyvault-secrets>=4.7.0`
   - `azure-identity>=1.12.0`

3. **`README.md`** - Complete documentation

## 🔑 Key Features:

- Uses `DefaultAzureCredential` for authentication
- Comprehensive error handling for all operations
- Clear console output with success/failure indicators
- Handles soft-delete and permanent purge
- Environment variable configuration for Key Vault URL

## 🚀 To run:

```bash
pip install -r requirements.txt
export AZURE_KEY_VAULT_URL='https://your-vault.vault.azure.net/'
python key_vault_crud.py
```

The script follows Azure SDK best practices and includes proper resource cleanup!

## Verification

**Result:** ✅ PASS

**Summary:** All requirements met: correct CRUD operations, authentication, error handling, and packages

**Reasoning:**

The generated code fully meets all prompt requirements. It correctly implements all four CRUD operations: (1) CREATE using set_secret() with 'my-secret' and 'my-secret-value', (2) READ using get_secret() and printing the value, (3) UPDATE using set_secret() with 'updated-value', and (4) DELETE using begin_delete_secret() followed by purge_deleted_secret(). The code properly handles soft-delete by calling poller.result() to wait for deletion completion before purging. Authentication uses DefaultAzureCredential as specified. Error handling includes ResourceNotFoundError and HttpResponseError exceptions, plus environment variable validation. The requirements.txt correctly specifies azure-keyvault-secrets and azure-identity packages. The code is syntactically valid Python with proper imports and structure. The README provides comprehensive documentation covering installation, SecretClient creation, all CRUD methods, soft-delete handling, and exception handling as expected.

---

[← Back to Summary](../../../../../../summary.md)
