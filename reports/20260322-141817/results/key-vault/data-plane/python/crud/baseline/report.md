# Evaluation Report: key-vault-dp-python-crud

**Config:** baseline | **Result:** ✅ PASSED | **Duration:** 75.6s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-crud` |
| Config | baseline |
| Result | ✅ PASSED |
| Duration | 75.6s |
| Timestamp | 2026-03-22T21:18:17Z |
| Files Generated | 3 |
| Event Count | 68 |

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

The user wants me to write a Python script that performs CRUD operations on Azure Key Vault secrets using the Azure SDK for Python. They specifically want:

1. Create a secret
2. Read the secret
3. Update the secret
4. Delete and purge the secret

They want me to use DefaultAzureCredential and include proper error handling. They also want to know the required pip packages.

I should:
1. Look up the Azure SDK for Python documentation for Key Vault
2. Write a complete script with all CRUD operations
3. Include error handling
4. Show the required packages

Let me start by fetching the documentation and then writing the script.

## Tool Calls

### 🔧 1. report_intent

**Input:**

```json
{"intent":"Researching Azure SDK documentation"}
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
{"file_text":"\"\"\"\nAzure Key Vault CRUD Operations Demo\nDemonstrates Create, Read, Update, and Delete operations for secrets.\n\nRequired packages:\n    pip install azure-keyvault-secrets azure-identity\n\"\"\"\n\nimport os\nimport sys\nfrom azure.identity import DefaultAzureCredential\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.core.exceptions import ResourceNotFoundError, HttpResponseError\n\n\ndef main():\n    # Get the Key Vault URL from environment variable\n    vault_url = os.environ.get(\"VAULT_URL\")\n    \n    if not vault_url:\n        print(\"Error: VAULT_URL environment variable is not set\")\n        print(\"Please set it to your Key Vault URL (e.g., https://my-vault.vault.azure.net/)\")\n        sys.exit(1)\n    \n    try:\n        # Authenticate using DefaultAzureCredential\n        credential = DefaultAzureCredential()\n        \n        # Create the SecretClient\n        client = SecretClient(vault_url=vault_url, credential=credential)\n        \n        print(f\"Connected to Key Vault: {vault_url}\\n\")\n        \n        # 1. CREATE - Set a new secret\n        print(\"1. CREATE: Creating secret 'my-secret'...\")\n        try:\n            secret = client.set_secret(\"my-secret\", \"my-secret-value\")\n            print(f\"   ✓ Secret created successfully\")\n            print(f\"   - Name: {secret.name}\")\n            print(f\"   - Version: {secret.properties.version}\")\n            print()\n        except HttpResponseError as e:\n            print(f\"   ✗ Failed to create secret: {e.message}\")\n            sys.exit(1)\n        \n        # 2. READ - Retrieve the secret\n        print(\"2. READ: Reading secret 'my-secret'...\")\n        try:\n            retrieved_secret = client.get_secret(\"my-secret\")\n            print(f\"   ✓ Secret retrieved successfully\")\n            print(f\"   - Name: {retrieved_secret.name}\")\n            print(f\"   - Value: {retrieved_secret.value}\")\n            print(f\"   - Version: {retrieved_secret.properties.version}\")\n            print()\n        except ResourceNotFoundError:\n            print(f\"   ✗ Secret 'my-secret' not found\")\n            sys.exit(1)\n        except HttpResponseError as e:\n            print(f\"   ✗ Failed to retrieve secret: {e.message}\")\n            sys.exit(1)\n        \n        # 3. UPDATE - Update the secret value\n        print(\"3. UPDATE: Updating secret 'my-secret' to new value...\")\n        try:\n            updated_secret = client.set_secret(\"my-secret\", \"updated-value\")\n            print(f\"   ✓ Secret updated successfully\")\n            print(f\"   - Name: {updated_secret.name}\")\n            print(f\"   - New Value: {updated_secret.value}\")\n            print(f\"   - New Version: {updated_secret.properties.version}\")\n            print()\n        except HttpResponseError as e:\n            print(f\"   ✗ Failed to update secret: {e.message}\")\n            sys.exit(1)\n        \n        # 4. DELETE - Delete and purge the secret\n        print(\"4. DELETE: Deleting secret 'my-secret'...\")\n        try:\n            # Begin delete operation (returns a poller)\n            delete_poller = client.begin_delete_secret(\"my-secret\")\n            deleted_secret = delete_poller.result()  # Wait for deletion to complete\n            print(f\"   ✓ Secret deleted successfully\")\n            print(f\"   - Name: {deleted_secret.name}\")\n            print(f\"   - Deleted Date: {deleted_secret.deleted_date}\")\n            print(f\"   - Scheduled Purge Date: {deleted_secret.scheduled_purge_date}\")\n            print()\n        except ResourceNotFoundError:\n            print(f\"   ✗ Secret 'my-secret' not found\")\n            sys.exit(1)\n        except HttpResponseError as e:\n            print(f\"   ✗ Failed to delete secret: {e.message}\")\n            sys.exit(1)\n        \n        # PURGE - Permanently delete the secret (soft-delete enabled vaults)\n        print(\"   PURGE: Permanently purging deleted secret 'my-secret'...\")\n        try:\n            client.purge_deleted_secret(\"my-secret\")\n            print(f\"   ✓ Secret purged successfully (permanently deleted)\")\n            print()\n        except ResourceNotFoundError:\n            print(f\"   ✗ Deleted secret 'my-secret' not found for purging\")\n            sys.exit(1)\n        except HttpResponseError as e:\n            print(f\"   ✗ Failed to purge secret: {e.message}\")\n            print(f\"   Note: Purging may fail if soft-delete is not enabled or vault doesn't support purge\")\n            sys.exit(1)\n        \n        print(\"All CRUD operations completed successfully!\")\n        \n    except Exception as e:\n        print(f\"Unexpected error: {str(e)}\")\n        sys.exit(1)\n\n\nif __name__ == \"__main__\":\n    main()\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-826285646/keyvault_crud.py"}
```

### 🔧 5. create

**Input:**

```json
{"file_text":"azure-keyvault-secrets\nazure-identity\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-826285646/requirements.txt"}
```

### 🔧 6. create

**Input:**

```json
{"file_text":"# Azure Key Vault CRUD Operations Demo\n\nThis script demonstrates all four CRUD operations on Azure Key Vault secrets using the Azure SDK for Python.\n\n## Required Packages\n\nInstall the required packages using pip:\n\n```bash\npip install -r requirements.txt\n```\n\nOr install individually:\n\n```bash\npip install azure-keyvault-secrets azure-identity\n```\n\n## Prerequisites\n\n1. An Azure subscription\n2. An Azure Key Vault with soft-delete enabled\n3. Appropriate permissions (secrets/set, secrets/get, secrets/delete, secrets/purge)\n4. Authentication configured for DefaultAzureCredential\n\n## Authentication\n\nThe script uses `DefaultAzureCredential`, which attempts authentication using multiple methods in the following order:\n\n1. Environment variables (AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET)\n2. Managed Identity (if running on Azure)\n3. Azure CLI (if logged in via `az login`)\n4. Azure PowerShell\n5. Interactive browser\n\nFor local development, the easiest method is to use Azure CLI:\n\n```bash\naz login\n```\n\n## Usage\n\nSet the VAULT_URL environment variable to your Key Vault URL:\n\n```bash\nexport VAULT_URL=\"https://your-key-vault-name.vault.azure.net/\"\n```\n\nThen run the script:\n\n```bash\npython keyvault_crud.py\n```\n\n## Operations Performed\n\nThe script performs the following operations in order:\n\n1. **CREATE** - Creates a secret named \"my-secret\" with value \"my-secret-value\"\n2. **READ** - Retrieves and displays the secret value\n3. **UPDATE** - Updates the secret to a new value \"updated-value\"\n4. **DELETE** - Deletes the secret (soft-delete)\n5. **PURGE** - Permanently deletes the secret (requires soft-delete enabled vault)\n\n## Expected Output\n\n```\nConnected to Key Vault: https://your-key-vault.vault.azure.net/\n\n1. CREATE: Creating secret 'my-secret'...\n   ✓ Secret created successfully\n   - Name: my-secret\n   - Version: abc123...\n\n2. READ: Reading secret 'my-secret'...\n   ✓ Secret retrieved successfully\n   - Name: my-secret\n   - Value: my-secret-value\n   - Version: abc123...\n\n3. UPDATE: Updating secret 'my-secret' to new value...\n   ✓ Secret updated successfully\n   - Name: my-secret\n   - New Value: updated-value\n   - New Version: def456...\n\n4. DELETE: Deleting secret 'my-secret'...\n   ✓ Secret deleted successfully\n   - Name: my-secret\n   - Deleted Date: ...\n   - Scheduled Purge Date: ...\n\n   PURGE: Permanently purging deleted secret 'my-secret'...\n   ✓ Secret purged successfully (permanently deleted)\n\nAll CRUD operations completed successfully!\n```\n\n## Error Handling\n\nThe script includes comprehensive error handling for:\n- Missing VAULT_URL environment variable\n- Authentication failures\n- Resource not found errors\n- HTTP response errors\n- Permission issues\n\n## Notes\n\n- The UPDATE operation creates a new version of the secret (Azure Key Vault maintains version history)\n- Purging is only necessary in vaults with soft-delete enabled\n- After deletion, secrets can be recovered before the scheduled purge date (unless purged)\n- Purge operations are permanent and cannot be undone\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-826285646/README.md"}
```

## Generated Files

- `README.md`
- `keyvault_crud.py`
- `requirements.txt`

## Copilot Response

I'll help you create a Python script for Azure Key Vault CRUD operations using the official Azure SDK documentation.

Now let me fetch information about purging secrets:

Perfect! Now I have all the information needed to create the script. Let me create a complete Python script with all CRUD operations:

Now let me create a requirements.txt file and a README with usage instructions:

Perfect! I've created a complete Azure Key Vault CRUD operations script based on the official Azure SDK for Python documentation. Here's what I've provided:

## Created Files:

1. **keyvault_crud.py** - Main script with all 4 CRUD operations
2. **requirements.txt** - Required packages
3. **README.md** - Complete documentation

## Script Features:

✅ **CREATE** - Creates "my-secret" with value "my-secret-value" using `set_secret()`
✅ **READ** - Retrieves the secret using `get_secret()` and prints its value
✅ **UPDATE** - Updates to "updated-value" using `set_secret()` (creates new version)
✅ **DELETE** - Deletes using `begin_delete_secret().result()` (waits for completion)
✅ **PURGE** - Permanently deletes using `purge_deleted_secret()`

## Key Features:

- Uses `DefaultAzureCredential` for authentication as requested
- Comprehensive error handling with try/except blocks
- Clear console output with status indicators (✓/✗)
- Handles soft-delete enabled vaults correctly
- Properly waits for async delete operation to complete

## Required Packages:

```bash
pip install azure-keyvault-secrets azure-identity
```

## Usage:

```bash
export VAULT_URL="https://your-vault.vault.azure.net/"
python keyvault_crud.py
```

The script is fully documented and follows Azure SDK best practices from the official documentation!

## Verification

**Result:** ✅ PASS

**Summary:** All CRUD operations correctly implemented with proper authentication, error handling, and required packages.

**Reasoning:**

The generated code fully meets all requirements from the prompt. It correctly implements all four CRUD operations: (1) Creates secret 'my-secret' with value 'my-secret-value' using set_secret(), (2) Reads and prints the secret value using get_secret(), (3) Updates the secret to 'updated-value' using set_secret() again, (4) Deletes the secret using begin_delete_secret() and waits for completion with result(), then purges it using purge_deleted_secret(). Authentication is correctly implemented with DefaultAzureCredential. The SecretClient is properly instantiated with vault_url and credential parameters. Error handling includes ResourceNotFoundError and HttpResponseError exceptions as expected. The requirements.txt file correctly lists both azure-keyvault-secrets and azure-identity packages. The code is syntactically valid Python with proper imports, correct method signatures, and logical flow. The README provides comprehensive documentation including installation instructions, prerequisites, and usage examples. All SDK methods match Azure's official documentation patterns for Key Vault operations.

---

[← Back to Summary](../../../../../../summary.md)
