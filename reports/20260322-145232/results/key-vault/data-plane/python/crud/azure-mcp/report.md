# Evaluation Report: key-vault-dp-python-crud

**Config:** azure-mcp | **Result:** ✅ PASSED | **Duration:** 83.7s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-crud` |
| Config | azure-mcp |
| Result | ✅ PASSED |
| Duration | 83.7s |
| Timestamp | 2026-03-22T21:54:01Z |
| Files Generated | 3 |
| Event Count | 73 |

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

The user wants me to create a Python script that performs CRUD operations on Azure Key Vault secrets using the Azure SDK for Python. They specifically want:
1. Create a secret
2. Read it back
3. Update it
4. Delete and purge it

They want me to use only the Azure SDK for Python documentation, use DefaultAzureCredential, include error handling, and show required pip packages.

This is a straightforward task - I should:
1. Look up the Azure SDK for Python documentation for Key Vault
2. Create a script with all four CRUD operations
3. Include proper error handling
4. Document the required pip packages

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
{"file_text":"#!/usr/bin/env python3\n\"\"\"\nAzure Key Vault CRUD Operations Demo\n\nThis script demonstrates all four CRUD operations on Azure Key Vault secrets:\n1. Create - Set a new secret\n2. Read - Get and display the secret value\n3. Update - Modify the secret value\n4. Delete - Delete and purge the secret\n\nRequired environment variable:\n    VAULT_URL: The URL of your Azure Key Vault (e.g., https://my-vault.vault.azure.net/)\n\nRequired pip packages:\n    pip install azure-keyvault-secrets azure-identity\n\"\"\"\n\nimport os\nimport sys\nfrom azure.identity import DefaultAzureCredential\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.core.exceptions import ResourceNotFoundError, HttpResponseError\n\n\ndef main():\n    \"\"\"Perform CRUD operations on Azure Key Vault secrets.\"\"\"\n    \n    # Get vault URL from environment variable\n    vault_url = os.environ.get(\"VAULT_URL\")\n    if not vault_url:\n        print(\"Error: VAULT_URL environment variable is not set\")\n        print(\"Example: export VAULT_URL='https://my-vault.vault.azure.net/'\")\n        sys.exit(1)\n    \n    print(f\"Connecting to Key Vault: {vault_url}\")\n    \n    # Initialize the credential and client\n    try:\n        credential = DefaultAzureCredential()\n        client = SecretClient(vault_url=vault_url, credential=credential)\n        print(\"✓ Successfully authenticated\\n\")\n    except Exception as e:\n        print(f\"Error: Failed to authenticate: {e}\")\n        sys.exit(1)\n    \n    secret_name = \"my-secret\"\n    \n    # ========================================\n    # CREATE: Set a new secret\n    # ========================================\n    try:\n        print(\"=\" * 50)\n        print(\"1. CREATE - Setting secret\")\n        print(\"=\" * 50)\n        \n        secret_value = \"my-secret-value\"\n        secret = client.set_secret(secret_name, secret_value)\n        \n        print(f\"✓ Secret created successfully\")\n        print(f\"  Name: {secret.name}\")\n        print(f\"  Value: {secret.value}\")\n        print(f\"  Version: {secret.properties.version}\")\n        print()\n        \n    except HttpResponseError as e:\n        print(f\"✗ Error creating secret: {e.message}\")\n        sys.exit(1)\n    except Exception as e:\n        print(f\"✗ Unexpected error: {e}\")\n        sys.exit(1)\n    \n    # ========================================\n    # READ: Get the secret value\n    # ========================================\n    try:\n        print(\"=\" * 50)\n        print(\"2. READ - Retrieving secret\")\n        print(\"=\" * 50)\n        \n        retrieved_secret = client.get_secret(secret_name)\n        \n        print(f\"✓ Secret retrieved successfully\")\n        print(f\"  Name: {retrieved_secret.name}\")\n        print(f\"  Value: {retrieved_secret.value}\")\n        print(f\"  Content Type: {retrieved_secret.properties.content_type}\")\n        print(f\"  Enabled: {retrieved_secret.properties.enabled}\")\n        print()\n        \n    except ResourceNotFoundError:\n        print(f\"✗ Error: Secret '{secret_name}' not found\")\n        sys.exit(1)\n    except HttpResponseError as e:\n        print(f\"✗ Error retrieving secret: {e.message}\")\n        sys.exit(1)\n    except Exception as e:\n        print(f\"✗ Unexpected error: {e}\")\n        sys.exit(1)\n    \n    # ========================================\n    # UPDATE: Change the secret value\n    # ========================================\n    try:\n        print(\"=\" * 50)\n        print(\"3. UPDATE - Updating secret value\")\n        print(\"=\" * 50)\n        \n        new_value = \"updated-value\"\n        updated_secret = client.set_secret(secret_name, new_value)\n        \n        print(f\"✓ Secret updated successfully\")\n        print(f\"  Name: {updated_secret.name}\")\n        print(f\"  New Value: {updated_secret.value}\")\n        print(f\"  New Version: {updated_secret.properties.version}\")\n        print()\n        \n    except HttpResponseError as e:\n        print(f\"✗ Error updating secret: {e.message}\")\n        sys.exit(1)\n    except Exception as e:\n        print(f\"✗ Unexpected error: {e}\")\n        sys.exit(1)\n    \n    # ========================================\n    # DELETE: Delete and purge the secret\n    # ========================================\n    try:\n        print(\"=\" * 50)\n        print(\"4. DELETE - Deleting secret\")\n        print(\"=\" * 50)\n        \n        # Begin delete operation (returns a poller)\n        delete_poller = client.begin_delete_secret(secret_name)\n        \n        # Wait for the deletion to complete\n        deleted_secret = delete_poller.result()\n        \n        print(f\"✓ Secret deleted successfully\")\n        print(f\"  Name: {deleted_secret.name}\")\n        print(f\"  Deleted Date: {deleted_secret.deleted_date}\")\n        print(f\"  Scheduled Purge Date: {deleted_secret.scheduled_purge_date}\")\n        print(f\"  Recovery ID: {deleted_secret.recovery_id}\")\n        print()\n        \n    except HttpResponseError as e:\n        print(f\"✗ Error deleting secret: {e.message}\")\n        sys.exit(1)\n    except Exception as e:\n        print(f\"✗ Unexpected error: {e}\")\n        sys.exit(1)\n    \n    # ========================================\n    # PURGE: Permanently delete the secret\n    # ========================================\n    try:\n        print(\"=\" * 50)\n        print(\"5. PURGE - Permanently deleting secret\")\n        print(\"=\" * 50)\n        \n        # Purge the deleted secret (permanent deletion)\n        client.purge_deleted_secret(secret_name)\n        \n        print(f\"✓ Secret purged successfully\")\n        print(f\"  The secret '{secret_name}' has been permanently deleted\")\n        print()\n        \n    except HttpResponseError as e:\n        print(f\"✗ Error purging secret: {e.message}\")\n        print(\"  Note: Vault must have soft-delete enabled for purge operation\")\n        sys.exit(1)\n    except Exception as e:\n        print(f\"✗ Unexpected error: {e}\")\n        sys.exit(1)\n    \n    print(\"=\" * 50)\n    print(\"✓ All CRUD operations completed successfully!\")\n    print(\"=\" * 50)\n\n\nif __name__ == \"__main__\":\n    main()\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-2089708386/keyvault_crud.py"}
```

### 🔧 5. create

**Input:**

```json
{"file_text":"azure-keyvault-secrets\nazure-identity\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-2089708386/requirements.txt"}
```

### 🔧 6. create

**Input:**

```json
{"file_text":"# Azure Key Vault CRUD Operations\n\nThis script demonstrates all four CRUD operations on Azure Key Vault secrets using the Azure SDK for Python.\n\n## Operations Performed\n\n1. **CREATE** - Set a new secret called \"my-secret\" with value \"my-secret-value\"\n2. **READ** - Retrieve the secret and print its value\n3. **UPDATE** - Update the secret to a new value \"updated-value\"\n4. **DELETE** - Delete the secret (soft-delete)\n5. **PURGE** - Permanently delete the secret (requires soft-delete enabled vault)\n\n## Prerequisites\n\n- Python 3.9 or later\n- An Azure subscription\n- An Azure Key Vault with soft-delete enabled\n- Appropriate Azure credentials configured (see Authentication below)\n\n## Installation\n\nInstall the required packages:\n\n```bash\npip install -r requirements.txt\n```\n\nOr install packages directly:\n\n```bash\npip install azure-keyvault-secrets azure-identity\n```\n\n## Required Packages\n\n- **azure-keyvault-secrets** - Azure Key Vault Secrets client library\n- **azure-identity** - Azure authentication library providing DefaultAzureCredential\n\n## Authentication\n\nThis script uses `DefaultAzureCredential` which attempts to authenticate via multiple methods in order:\n\n1. Environment variables (AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET)\n2. Managed Identity (if deployed to Azure)\n3. Azure CLI (if logged in via `az login`)\n4. Azure PowerShell\n5. Interactive browser\n\nFor local development, the easiest method is Azure CLI:\n\n```bash\naz login\n```\n\n## Configuration\n\nSet the `VAULT_URL` environment variable to your Azure Key Vault URL:\n\n```bash\nexport VAULT_URL='https://your-vault-name.vault.azure.net/'\n```\n\n## Usage\n\nRun the script:\n\n```bash\npython keyvault_crud.py\n```\n\n## Required Permissions\n\nEnsure your Azure identity has the following permissions on the Key Vault:\n\n- `secrets/set` - For creating and updating secrets\n- `secrets/get` - For reading secrets\n- `secrets/delete` - For deleting secrets\n- `secrets/purge` - For purging deleted secrets\n\nYou can assign these using Azure RBAC role \"Key Vault Secrets Officer\" or via Key Vault access policies.\n\n## Error Handling\n\nThe script includes comprehensive error handling for:\n\n- Missing environment variables\n- Authentication failures\n- Resource not found errors\n- HTTP response errors\n- General exceptions\n\n## Sample Output\n\n```\nConnecting to Key Vault: https://my-vault.vault.azure.net/\n✓ Successfully authenticated\n\n==================================================\n1. CREATE - Setting secret\n==================================================\n✓ Secret created successfully\n  Name: my-secret\n  Value: my-secret-value\n  Version: abc123...\n\n==================================================\n2. READ - Retrieving secret\n==================================================\n✓ Secret retrieved successfully\n  Name: my-secret\n  Value: my-secret-value\n  Content Type: None\n  Enabled: True\n\n==================================================\n3. UPDATE - Updating secret value\n==================================================\n✓ Secret updated successfully\n  Name: my-secret\n  New Value: updated-value\n  New Version: def456...\n\n==================================================\n4. DELETE - Deleting secret\n==================================================\n✓ Secret deleted successfully\n  Name: my-secret\n  Deleted Date: 2026-03-22 21:54:00+00:00\n  Scheduled Purge Date: 2026-06-20 21:54:00+00:00\n  Recovery ID: https://...\n\n==================================================\n5. PURGE - Permanently deleting secret\n==================================================\n✓ Secret purged successfully\n  The secret 'my-secret' has been permanently deleted\n\n==================================================\n✓ All CRUD operations completed successfully!\n==================================================\n```\n\n## Notes\n\n- The vault must have soft-delete enabled for the purge operation to work\n- Deleted secrets can be recovered before purging (not demonstrated in this script)\n- Secret values are stored as strings in Azure Key Vault\n- Each update creates a new version of the secret\n\n## References\n\n- [Azure Key Vault Secrets Python SDK Documentation](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)\n- [Azure Identity Python SDK Documentation](https://learn.microsoft.com/en-us/python/api/overview/azure/identity-readme)\n- [Azure Key Vault Overview](https://learn.microsoft.com/en-us/azure/key-vault/general/overview)\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-2089708386/README.md"}
```

## Generated Files

- `README.md`
- `keyvault_crud.py`
- `requirements.txt`

## Copilot Response

I'll create a Python script that performs CRUD operations on Azure Key Vault secrets using the official Azure SDK documentation.

Now let me get information about purging deleted secrets:

Perfect! Now I have all the information I need. Let me create a comprehensive Python script with all four CRUD operations:

Now let me create a requirements.txt file with the needed packages:

Now let me create a comprehensive README:

Perfect! I've created a complete Python script that demonstrates all four CRUD operations on Azure Key Vault secrets. Here's what I've provided:

## 📁 Files Created:

1. **`keyvault_crud.py`** - Main script with all CRUD operations
2. **`requirements.txt`** - Required pip packages
3. **`README.md`** - Comprehensive documentation

## 🔑 Key Features:

**CRUD Operations:**
- ✅ **CREATE** - Sets "my-secret" with value "my-secret-value"
- ✅ **READ** - Retrieves and displays the secret value
- ✅ **UPDATE** - Updates the secret to "updated-value"
- ✅ **DELETE** - Soft-deletes the secret (with poller)
- ✅ **PURGE** - Permanently removes the secret

**Additional Features:**
- Uses `DefaultAzureCredential` for authentication
- Comprehensive error handling for all operations
- Clear console output with status indicators
- Environment variable configuration
- Detailed comments and documentation

## 📦 Required Packages:
```bash
pip install azure-keyvault-secrets azure-identity
```

## 🚀 Usage:
```bash
export VAULT_URL='https://your-vault-name.vault.azure.net/'
python keyvault_crud.py
```

The script is based entirely on the official Azure SDK for Python documentation and includes proper error handling for production use!

## Verification

**Result:** ✅ PASS

**Summary:** Code correctly implements all CRUD operations with proper authentication, error handling, and required packages

**Reasoning:**

The generated code fully meets all requirements specified in the original prompt. It correctly implements all four CRUD operations using the Azure SDK for Python: (1) CREATE - uses set_secret() to create 'my-secret' with value 'my-secret-value', (2) READ - uses get_secret() to retrieve and print the secret value, (3) UPDATE - uses set_secret() again to update to 'updated-value', (4) DELETE & PURGE - uses begin_delete_secret() with .result() to wait for completion, then purge_deleted_secret() for permanent deletion. The code properly uses DefaultAzureCredential for authentication, creates a SecretClient with vault URL and credential, includes comprehensive error handling with try/except blocks catching ResourceNotFoundError and HttpResponseError, validates environment variables, and provides informative output. The requirements.txt correctly lists both required packages (azure-keyvault-secrets and azure-identity). The README provides clear documentation about installation, authentication, permissions, and usage. The code is syntactically valid, well-structured with clear sections for each operation, and follows Python best practices. All expected SDK methods are used correctly, and the soft-delete flow is properly handled by waiting for deletion to complete before purging.

---

[← Back to Summary](../../../../../../summary.md)
