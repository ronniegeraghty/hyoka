# Evaluation Report: key-vault-dp-python-crud

**Config:** azure-mcp | **Result:** ✅ PASSED | **Duration:** 95.9s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-crud` |
| Config | azure-mcp |
| Result | ✅ PASSED |
| Duration | 95.9s |
| Timestamp | 2026-03-21T06:38:46Z |
| Files Generated | 3 |
| Event Count | 68 |

## Configuration

- **model:** claude-sonnet-4.5
- **name:** azure-mcp

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

The user wants me to write a Python script that demonstrates CRUD operations on Azure Key Vault secrets using the Azure SDK for Python. They want me to use only the Azure SDK for Python documentation as a reference.

Let me break down what's needed:
1. A script that performs Create, Read, Update, Delete operations on Key Vault secrets
2. Use DefaultAzureCredential for authentication
3. Include proper error handling
4. Show required pip packages
5. Work with soft-delete enabled vault (so need to purge after delete)

I should:
1. First fetch the Azure SDK for Python documentation for Key Vault
2. Write a complete script based on the official documentation
3. Include all necessary imports and error handling
4. Document the required packages

Let me start by fetching the official documentation.

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
{"max_length":10000,"url":"https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient?view=azure-python"}
```

### 🔧 4. create

**Input:**

```json
{"file_text":"#!/usr/bin/env python3\n\"\"\"\nAzure Key Vault Secrets CRUD Operations Demo\n\nThis script demonstrates all four CRUD operations on Azure Key Vault secrets:\n1. Create a new secret\n2. Read the secret back\n3. Update the secret to a new value\n4. Delete and purge the secret (for soft-delete enabled vaults)\n\nRequired environment variable:\n    AZURE_KEY_VAULT_URL: The URL of your Azure Key Vault\n                         (e.g., https://my-key-vault.vault.azure.net/)\n\nAuthentication uses DefaultAzureCredential, which supports multiple authentication methods:\n- Environment variables (AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET)\n- Azure CLI (az login)\n- Azure PowerShell\n- Managed Identity\n- Interactive browser\n\nRequired pip packages:\n    pip install azure-keyvault-secrets azure-identity\n\"\"\"\n\nimport os\nimport sys\nfrom azure.identity import DefaultAzureCredential\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.core.exceptions import (\n    ResourceNotFoundError,\n    HttpResponseError,\n    ClientAuthenticationError\n)\n\n\ndef main():\n    \"\"\"Main function demonstrating CRUD operations on Azure Key Vault secrets.\"\"\"\n    \n    # Get the Key Vault URL from environment variable\n    vault_url = os.environ.get(\"AZURE_KEY_VAULT_URL\")\n    if not vault_url:\n        print(\"ERROR: AZURE_KEY_VAULT_URL environment variable is not set.\")\n        print(\"Example: export AZURE_KEY_VAULT_URL='https://my-key-vault.vault.azure.net/'\")\n        sys.exit(1)\n    \n    print(f\"Using Key Vault: {vault_url}\\n\")\n    \n    # Initialize credential and client\n    try:\n        credential = DefaultAzureCredential()\n        client = SecretClient(vault_url=vault_url, credential=credential)\n        print(\"✓ Successfully initialized SecretClient with DefaultAzureCredential\\n\")\n    except ClientAuthenticationError as e:\n        print(f\"ERROR: Authentication failed: {e.message}\")\n        print(\"\\nPlease ensure you are authenticated using one of these methods:\")\n        print(\"  - Azure CLI: az login\")\n        print(\"  - Environment variables: AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET\")\n        print(\"  - Managed Identity (when running in Azure)\")\n        sys.exit(1)\n    except Exception as e:\n        print(f\"ERROR: Failed to initialize client: {e}\")\n        sys.exit(1)\n    \n    secret_name = \"my-secret\"\n    \n    try:\n        # ========================================\n        # 1. CREATE - Set a new secret\n        # ========================================\n        print(\"=\" * 60)\n        print(\"1. CREATE - Creating a new secret\")\n        print(\"=\" * 60)\n        \n        initial_value = \"my-secret-value\"\n        secret = client.set_secret(secret_name, initial_value)\n        \n        print(f\"✓ Secret created successfully!\")\n        print(f\"  Name:    {secret.name}\")\n        print(f\"  Value:   {secret.value}\")\n        print(f\"  Version: {secret.properties.version}\")\n        print(f\"  Created: {secret.properties.created_on}\")\n        print()\n        \n        # ========================================\n        # 2. READ - Retrieve the secret\n        # ========================================\n        print(\"=\" * 60)\n        print(\"2. READ - Retrieving the secret\")\n        print(\"=\" * 60)\n        \n        retrieved_secret = client.get_secret(secret_name)\n        \n        print(f\"✓ Secret retrieved successfully!\")\n        print(f\"  Name:    {retrieved_secret.name}\")\n        print(f\"  Value:   {retrieved_secret.value}\")\n        print(f\"  Version: {retrieved_secret.properties.version}\")\n        print()\n        \n        # ========================================\n        # 3. UPDATE - Update the secret value\n        # ========================================\n        print(\"=\" * 60)\n        print(\"3. UPDATE - Updating the secret to a new value\")\n        print(\"=\" * 60)\n        \n        updated_value = \"updated-value\"\n        # set_secret creates a new version when the name already exists\n        updated_secret = client.set_secret(secret_name, updated_value)\n        \n        print(f\"✓ Secret updated successfully!\")\n        print(f\"  Name:        {updated_secret.name}\")\n        print(f\"  New Value:   {updated_secret.value}\")\n        print(f\"  New Version: {updated_secret.properties.version}\")\n        print(f\"  Updated:     {updated_secret.properties.updated_on}\")\n        print()\n        \n        # Verify the update\n        verified_secret = client.get_secret(secret_name)\n        print(f\"  Verification: Current value is '{verified_secret.value}'\")\n        print()\n        \n        # ========================================\n        # 4. DELETE - Delete and purge the secret\n        # ========================================\n        print(\"=\" * 60)\n        print(\"4. DELETE - Deleting and purging the secret\")\n        print(\"=\" * 60)\n        \n        # Delete the secret (soft delete)\n        print(f\"Deleting secret '{secret_name}'...\")\n        delete_poller = client.begin_delete_secret(secret_name)\n        deleted_secret = delete_poller.result()\n        \n        print(f\"✓ Secret deleted successfully!\")\n        print(f\"  Name:           {deleted_secret.name}\")\n        print(f\"  Deleted Date:   {deleted_secret.deleted_date}\")\n        print(f\"  Recovery ID:    {deleted_secret.recovery_id}\")\n        print(f\"  Scheduled Purge: {deleted_secret.scheduled_purge_date}\")\n        print()\n        \n        # Purge the secret (permanent deletion for soft-delete enabled vaults)\n        print(f\"Purging secret '{secret_name}' permanently...\")\n        client.purge_deleted_secret(secret_name)\n        \n        print(f\"✓ Secret purged successfully!\")\n        print(f\"  The secret '{secret_name}' has been permanently deleted.\")\n        print()\n        \n        print(\"=\" * 60)\n        print(\"✓ All CRUD operations completed successfully!\")\n        print(\"=\" * 60)\n        \n    except ResourceNotFoundError as e:\n        print(f\"ERROR: Resource not found: {e.message}\")\n    except HttpResponseError as e:\n        print(f\"ERROR: HTTP response error: {e.message}\")\n        if \"Operation purge is not allowed\" in str(e):\n            print(\"\\nNote: Purge operation may not be allowed if:\")\n            print(\"  - The vault does not have soft-delete enabled\")\n            print(\"  - The vault's purge protection is enabled\")\n            print(\"  - The secret's recovery level does not support purging\")\n    except Exception as e:\n        print(f\"ERROR: An unexpected error occurred: {e}\")\n        import traceback\n        traceback.print_exc()\n        sys.exit(1)\n    finally:\n        # Clean up\n        credential.close()\n        client.close()\n\n\nif __name__ == \"__main__\":\n    main()\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-2832097541/keyvault_crud_demo.py"}
```

### 🔧 5. create

**Input:**

```json
{"file_text":"azure-keyvault-secrets\nazure-identity\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-2832097541/requirements.txt"}
```

### 🔧 6. create

**Input:**

```json
{"file_text":"# Azure Key Vault Secrets CRUD Demo\n\nThis script demonstrates all four CRUD operations on Azure Key Vault secrets using the Azure SDK for Python.\n\n## Prerequisites\n\n1. **Python 3.9 or later**\n2. **An Azure Key Vault** with soft-delete enabled (recommended)\n3. **Proper authentication** configured (see below)\n\n## Installation\n\nInstall the required packages:\n\n```bash\npip install -r requirements.txt\n```\n\nOr install directly:\n\n```bash\npip install azure-keyvault-secrets azure-identity\n```\n\n## Required Packages\n\n- `azure-keyvault-secrets`: Azure Key Vault Secrets client library\n- `azure-identity`: Azure authentication library (provides DefaultAzureCredential)\n\n## Configuration\n\nSet the Key Vault URL as an environment variable:\n\n```bash\nexport AZURE_KEY_VAULT_URL=\"https://your-key-vault-name.vault.azure.net/\"\n```\n\n## Authentication\n\nThe script uses `DefaultAzureCredential`, which automatically tries multiple authentication methods in order:\n\n1. **Environment Variables**:\n   ```bash\n   export AZURE_CLIENT_ID=\"your-client-id\"\n   export AZURE_TENANT_ID=\"your-tenant-id\"\n   export AZURE_CLIENT_SECRET=\"your-client-secret\"\n   ```\n\n2. **Azure CLI** (easiest for local development):\n   ```bash\n   az login\n   ```\n\n3. **Managed Identity** (when running in Azure)\n\n4. **Interactive Browser** (fallback)\n\n## Permissions Required\n\nEnsure your Azure identity has the following Key Vault permissions:\n\n- `secrets/set` - To create and update secrets\n- `secrets/get` - To read secrets\n- `secrets/delete` - To delete secrets\n- `secrets/purge` - To permanently delete secrets (only needed for soft-delete enabled vaults)\n\nYou can assign these permissions using Azure RBAC role \"Key Vault Secrets Officer\" or via Access Policies.\n\n## Usage\n\nRun the script:\n\n```bash\npython keyvault_crud_demo.py\n```\n\n## What the Script Does\n\nThe script performs these operations in sequence:\n\n1. **CREATE**: Creates a secret named \"my-secret\" with value \"my-secret-value\"\n2. **READ**: Retrieves the secret and prints its value\n3. **UPDATE**: Updates the secret to a new value \"updated-value\" (creates a new version)\n4. **DELETE**: Soft-deletes the secret\n5. **PURGE**: Permanently deletes the secret (for soft-delete enabled vaults)\n\n## Output Example\n\n```\nUsing Key Vault: https://my-key-vault.vault.azure.net/\n\n✓ Successfully initialized SecretClient with DefaultAzureCredential\n\n============================================================\n1. CREATE - Creating a new secret\n============================================================\n✓ Secret created successfully!\n  Name:    my-secret\n  Value:   my-secret-value\n  Version: abc123...\n  Created: 2026-03-21 06:30:00.123456+00:00\n\n...\n```\n\n## Error Handling\n\nThe script includes comprehensive error handling for:\n\n- Missing environment variables\n- Authentication failures\n- Resource not found errors\n- HTTP response errors\n- Permission errors\n- Purge operation restrictions\n\n## Notes\n\n- If your vault has **purge protection** enabled, the purge operation will fail. This is a safety feature.\n- If your vault does **not** have soft-delete enabled, the delete operation is permanent and purge is not needed.\n- The script uses `begin_delete_secret()` which returns a poller to wait for deletion to complete.\n- Secret values are stored as strings in Azure Key Vault.\n\n## Reference\n\nBased on official Azure SDK for Python documentation:\n- [Azure Key Vault Secrets README](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)\n- [SecretClient API Reference](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient)\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-2832097541/README.md"}
```

## Generated Files

- `README.md`
- `keyvault_crud_demo.py`
- `requirements.txt`

## Copilot Response

I'll help you create a script for Azure Key Vault CRUD operations using the official Azure SDK for Python documentation.

Let me also check the documentation for purging deleted secrets:

Perfect! Now I have all the information I need. Let me create a comprehensive script that demonstrates all CRUD operations on Azure Key Vault secrets:

Now let me create a requirements.txt file and a README with setup instructions:

Perfect! I've created a complete Azure Key Vault CRUD demo script based on the official Azure SDK for Python documentation. Here's what I've provided:

## 📁 Files Created

1. **`keyvault_crud_demo.py`** - Main script with all CRUD operations
2. **`requirements.txt`** - Required pip packages
3. **`README.md`** - Complete setup and usage instructions

## 🔑 Key Features

The script demonstrates all four CRUD operations:

1. **CREATE**: Creates "my-secret" with value "my-secret-value"
2. **READ**: Retrieves and prints the secret value
3. **UPDATE**: Updates to "updated-value" (creates new version)
4. **DELETE & PURGE**: Soft-deletes then permanently purges the secret

## 📦 Required Packages

```bash
pip install azure-keyvault-secrets azure-identity
```

## 🛡️ Authentication & Error Handling

- Uses `DefaultAzureCredential` (supports Azure CLI, environment vars, managed identity)
- Comprehensive error handling for authentication, permissions, and API errors
- Clear error messages with troubleshooting hints

## 🚀 Usage

```bash
export AZURE_KEY_VAULT_URL="https://your-vault.vault.azure.net/"
python keyvault_crud_demo.py
```

The script is production-ready with proper cleanup, detailed logging, and follows all Azure SDK best practices from the official documentation!

## Verification

**Result:** ✅ PASS

**Summary:** Complete implementation with all CRUD operations, correct SDK usage, proper error handling, and comprehensive documentation.

**Reasoning:**

The generated code fully meets all requirements. It correctly implements all four CRUD operations: (1) CREATE using set_secret() with 'my-secret' and 'my-secret-value', (2) READ using get_secret() and printing the value, (3) UPDATE using set_secret() with 'updated-value', and (4) DELETE and PURGE using begin_delete_secret() with proper polling (delete_poller.result()) before purge_deleted_secret(). Authentication uses DefaultAzureCredential as specified. The requirements.txt includes both azure-keyvault-secrets and azure-identity packages. Error handling is comprehensive, covering ResourceNotFoundError, HttpResponseError, ClientAuthenticationError, and general exceptions. The code is syntactically valid Python with proper imports, structure, and resource cleanup. The documentation in README.md is thorough, covering installation, authentication methods, permissions, and usage. All expected SDK methods are used correctly, and the soft-delete flow is properly handled by waiting for deletion to complete before purging.

---

[← Back to Summary](../../../../../../summary.md)
