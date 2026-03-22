# Evaluation Report: key-vault-dp-python-crud

**Config:** baseline | **Result:** ✅ PASSED | **Duration:** 74.4s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-crud` |
| Config | baseline |
| Result | ✅ PASSED |
| Duration | 74.4s |
| Timestamp | 2026-03-21T17:27:40Z |
| Files Generated | 2 |
| Event Count | 61 |

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

The user wants me to write a Python script that performs CRUD operations on Azure Key Vault secrets using the Azure SDK for Python. They want me to use only the Azure SDK documentation as a reference.

This is a straightforward task - I need to:
1. Look up the Azure Key Vault Python SDK documentation
2. Write a script that performs Create, Read, Update, Delete operations on secrets
3. Use DefaultAzureCredential for authentication
4. Include error handling
5. Document the required pip packages

Let me start by fetching the Azure SDK documentation for Key Vault secrets.

## Tool Calls

### 🔧 1. report_intent

**Input:**

```json
{"intent":"Researching Azure SDK documentation"}
```

### 🔧 2. web_fetch

**Input:**

```json
{"max_length":15000,"url":"https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme?view=azure-python"}
```

### 🔧 3. web_fetch

**Input:**

```json
{"max_length":15000,"url":"https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient?view=azure-python"}
```

### 🔧 4. create

**Input:**

```json
{"file_text":"\"\"\"\nAzure Key Vault Secrets CRUD Operations Demo\n\nThis script demonstrates all four CRUD operations on Azure Key Vault secrets:\n1. Create a new secret\n2. Read the secret back\n3. Update the secret to a new value\n4. Delete and purge the secret\n\nPrerequisites:\n- Azure Key Vault with soft-delete enabled\n- Appropriate permissions (secrets/get, secrets/set, secrets/delete, secrets/purge)\n- Azure credentials configured (via Azure CLI, environment variables, or managed identity)\n\nRequired packages:\n    pip install azure-keyvault-secrets azure-identity\n\"\"\"\n\nimport os\nimport sys\nfrom azure.identity import DefaultAzureCredential\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.core.exceptions import ResourceNotFoundError, HttpResponseError\n\n\ndef main():\n    # Get vault URL from environment variable\n    vault_url = os.environ.get(\"VAULT_URL\")\n    \n    if not vault_url:\n        print(\"Error: VAULT_URL environment variable is not set\")\n        print(\"Example: export VAULT_URL='https://your-vault-name.vault.azure.net/'\")\n        sys.exit(1)\n    \n    print(f\"Connecting to Key Vault: {vault_url}\\n\")\n    \n    # Create credential and client\n    try:\n        credential = DefaultAzureCredential()\n        client = SecretClient(vault_url=vault_url, credential=credential)\n    except Exception as e:\n        print(f\"Error creating Key Vault client: {e}\")\n        sys.exit(1)\n    \n    secret_name = \"my-secret\"\n    \n    # 1. CREATE - Set a new secret\n    print(\"=\" * 60)\n    print(\"1. CREATE - Creating new secret\")\n    print(\"=\" * 60)\n    try:\n        secret = client.set_secret(secret_name, \"my-secret-value\")\n        print(f\"✓ Secret created successfully\")\n        print(f\"  Name: {secret.name}\")\n        print(f\"  Value: {secret.value}\")\n        print(f\"  Version: {secret.properties.version}\")\n        print()\n    except HttpResponseError as e:\n        print(f\"✗ Error creating secret: {e.message}\")\n        sys.exit(1)\n    \n    # 2. READ - Retrieve the secret\n    print(\"=\" * 60)\n    print(\"2. READ - Retrieving secret\")\n    print(\"=\" * 60)\n    try:\n        retrieved_secret = client.get_secret(secret_name)\n        print(f\"✓ Secret retrieved successfully\")\n        print(f\"  Name: {retrieved_secret.name}\")\n        print(f\"  Value: {retrieved_secret.value}\")\n        print(f\"  Version: {retrieved_secret.properties.version}\")\n        print(f\"  Content Type: {retrieved_secret.properties.content_type}\")\n        print(f\"  Enabled: {retrieved_secret.properties.enabled}\")\n        print()\n    except ResourceNotFoundError:\n        print(f\"✗ Secret '{secret_name}' not found\")\n        sys.exit(1)\n    except HttpResponseError as e:\n        print(f\"✗ Error retrieving secret: {e.message}\")\n        sys.exit(1)\n    \n    # 3. UPDATE - Update the secret to a new value\n    print(\"=\" * 60)\n    print(\"3. UPDATE - Updating secret to new value\")\n    print(\"=\" * 60)\n    try:\n        # set_secret creates a new version with the updated value\n        updated_secret = client.set_secret(secret_name, \"updated-value\")\n        print(f\"✓ Secret updated successfully\")\n        print(f\"  Name: {updated_secret.name}\")\n        print(f\"  Value: {updated_secret.value}\")\n        print(f\"  New Version: {updated_secret.properties.version}\")\n        print()\n        \n        # Optionally, update secret properties (metadata) without changing the value\n        # This demonstrates updating properties like content_type, enabled, tags, etc.\n        print(\"  Updating secret metadata...\")\n        updated_properties = client.update_secret_properties(\n            secret_name,\n            content_type=\"text/plain\",\n            tags={\"environment\": \"demo\", \"purpose\": \"crud-example\"}\n        )\n        print(f\"  ✓ Metadata updated\")\n        print(f\"    Content Type: {updated_properties.content_type}\")\n        print(f\"    Tags: {updated_properties.tags}\")\n        print()\n    except HttpResponseError as e:\n        print(f\"✗ Error updating secret: {e.message}\")\n        sys.exit(1)\n    \n    # 4. DELETE - Delete and purge the secret\n    print(\"=\" * 60)\n    print(\"4. DELETE - Deleting and purging secret\")\n    print(\"=\" * 60)\n    try:\n        # begin_delete_secret returns a poller for long-running delete operation\n        print(f\"  Deleting secret '{secret_name}'...\")\n        delete_poller = client.begin_delete_secret(secret_name)\n        \n        # Wait for deletion to complete\n        deleted_secret = delete_poller.result()\n        print(f\"✓ Secret deleted successfully\")\n        print(f\"  Name: {deleted_secret.name}\")\n        print(f\"  Deleted Date: {deleted_secret.deleted_date}\")\n        print(f\"  Recovery ID: {deleted_secret.recovery_id}\")\n        print(f\"  Scheduled Purge Date: {deleted_secret.scheduled_purge_date}\")\n        print()\n        \n        # Purge the secret permanently (only available with soft-delete enabled)\n        print(f\"  Purging secret '{secret_name}' permanently...\")\n        client.purge_deleted_secret(secret_name)\n        print(f\"✓ Secret purged successfully\")\n        print(f\"  The secret has been permanently deleted and cannot be recovered.\")\n        print()\n        \n    except ResourceNotFoundError:\n        print(f\"✗ Secret '{secret_name}' not found for deletion\")\n        sys.exit(1)\n    except HttpResponseError as e:\n        print(f\"✗ Error deleting/purging secret: {e.message}\")\n        print(f\"  Note: Purge requires soft-delete to be enabled on the vault\")\n        sys.exit(1)\n    \n    print(\"=\" * 60)\n    print(\"All CRUD operations completed successfully!\")\n    print(\"=\" * 60)\n\n\nif __name__ == \"__main__\":\n    main()\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-838645146/keyvault_crud_demo.py"}
```

### 🔧 5. create

**Input:**

```json
{"file_text":"# Azure Key Vault Secrets CRUD Operations Demo\n\nThis script demonstrates all four CRUD operations on Azure Key Vault secrets using the Azure SDK for Python.\n\n## Required Packages\n\nInstall the required packages using pip:\n\n```bash\npip install azure-keyvault-secrets azure-identity\n```\n\n**Package Details:**\n- `azure-keyvault-secrets`: Azure Key Vault Secrets client library\n- `azure-identity`: Azure authentication library (includes DefaultAzureCredential)\n\n## Prerequisites\n\n1. **Azure Key Vault**: An existing Azure Key Vault with soft-delete enabled\n2. **Python**: Python 3.9 or later\n3. **Permissions**: Your credential needs the following permissions:\n   - `secrets/get` - Read secrets\n   - `secrets/set` - Create and update secrets\n   - `secrets/delete` - Delete secrets\n   - `secrets/purge` - Purge deleted secrets\n\n## Authentication\n\nThis script uses `DefaultAzureCredential`, which attempts authentication through multiple methods in the following order:\n\n1. Environment variables (AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET)\n2. Managed Identity\n3. Visual Studio Code\n4. Azure CLI (`az login`)\n5. Azure PowerShell\n6. Interactive browser\n\nFor local development, the easiest method is to use Azure CLI:\n\n```bash\naz login\n```\n\n## Usage\n\nSet the `VAULT_URL` environment variable to your Key Vault URL:\n\n```bash\nexport VAULT_URL='https://your-vault-name.vault.azure.net/'\n```\n\nRun the script:\n\n```bash\npython keyvault_crud_demo.py\n```\n\n## Operations Performed\n\nThe script performs the following operations:\n\n### 1. CREATE\nCreates a new secret named \"my-secret\" with value \"my-secret-value\" using `set_secret()`.\n\n### 2. READ\nRetrieves the secret using `get_secret()` and displays its properties.\n\n### 3. UPDATE\n- Updates the secret value to \"updated-value\" using `set_secret()` (creates a new version)\n- Updates metadata (content type and tags) using `update_secret_properties()`\n\n### 4. DELETE\n- Soft-deletes the secret using `begin_delete_secret()` (returns a poller)\n- Permanently purges the deleted secret using `purge_deleted_secret()`\n\n## Error Handling\n\nThe script includes comprehensive error handling for:\n- Missing environment variables\n- Authentication failures\n- Resource not found errors\n- HTTP response errors\n- Permission issues\n\n## Notes\n\n- **Soft-delete**: The purge operation requires soft-delete to be enabled on your Key Vault\n- **Versions**: Each call to `set_secret()` with an existing secret name creates a new version\n- **Polling**: `begin_delete_secret()` returns a poller because deletion may take several seconds\n- **Permissions**: If you lack purge permissions, the script will fail at the purge step\n\n## References\n\n- [Azure Key Vault Secrets Python SDK Documentation](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)\n- [SecretClient API Reference](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient)\n- [DefaultAzureCredential Documentation](https://learn.microsoft.com/en-us/python/api/azure-identity/azure.identity.defaultazurecredential)\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-838645146/README.md"}
```

## Generated Files

- `README.md`
- `keyvault_crud_demo.py`

## Copilot Response

I'll help you create a Python script for Azure Key Vault secret CRUD operations using the official Azure SDK documentation.

Now let me check for information about purging deleted secrets:

Perfect! Now I have all the information I need from the Azure SDK documentation. Let me create a comprehensive Python script:

Now let me create a README file with additional information:

Perfect! I've created a complete Python script that demonstrates all four CRUD operations on Azure Key Vault secrets. Here's a summary:

## ✅ Created Files

1. **`keyvault_crud_demo.py`** - Main script with all CRUD operations
2. **`README.md`** - Comprehensive documentation

## 🔑 Key Features

**CRUD Operations:**
1. **CREATE**: Uses `set_secret()` to create "my-secret" with value "my-secret-value"
2. **READ**: Uses `get_secret()` to retrieve and display the secret
3. **UPDATE**: 
   - Uses `set_secret()` again to update value to "updated-value" (creates new version)
   - Uses `update_secret_properties()` to update metadata (content type, tags)
4. **DELETE**: 
   - Uses `begin_delete_secret()` for soft-delete (returns poller)
   - Uses `purge_deleted_secret()` for permanent deletion

**Additional Features:**
- ✅ DefaultAzureCredential for authentication
- ✅ Comprehensive error handling (ResourceNotFoundError, HttpResponseError)
- ✅ Clear console output with visual formatting
- ✅ Proper handling of long-running operations (delete poller)
- ✅ Environment variable validation

**Required Packages:**
```bash
pip install azure-keyvault-secrets azure-identity
```

The script is ready to run once you set the `VAULT_URL` environment variable and have proper Azure credentials configured!

## Verification

**Result:** ✅ PASS

**Summary:** Code fully implements all four CRUD operations with correct Azure SDK usage, authentication, and error handling.

**Reasoning:**

The generated code successfully meets all requirements: (1) Creates secret 'my-secret' with value 'my-secret-value' using set_secret(), (2) Reads and prints the secret value using get_secret(), (3) Updates to 'updated-value' using set_secret(), (4) Deletes with begin_delete_secret() and waits for completion with .result(), then purges with purge_deleted_secret(). Uses DefaultAzureCredential for authentication and SecretClient with vault URL. Includes proper error handling for ResourceNotFoundError and HttpResponseError. README documents required packages (azure-keyvault-secrets and azure-identity). Code is syntactically valid Python with correct Azure SDK API usage. The soft-delete workflow is properly handled by waiting for the delete operation to complete before purging. All expected coverage points are addressed including installation instructions, client creation, all four CRUD methods, and exception handling.

---

[← Back to Summary](../../../../../../summary.md)
