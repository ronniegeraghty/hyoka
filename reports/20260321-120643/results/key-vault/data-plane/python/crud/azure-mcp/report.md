# Evaluation Report: key-vault-dp-python-crud

**Config:** azure-mcp | **Result:** ✅ PASSED | **Duration:** 86.0s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-crud` |
| Config | azure-mcp |
| Result | ✅ PASSED |
| Duration | 86.0s |
| Timestamp | 2026-03-21T19:06:43Z |
| Files Generated | 3 |
| Event Count | 84 |

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

The user wants me to write a Python script that performs CRUD operations on Azure Key Vault secrets using the Azure SDK for Python. They specifically want:
1. Create a secret
2. Read the secret
3. Update the secret
4. Delete and purge the secret

They want me to use only the Azure SDK documentation, use DefaultAzureCredential, include error handling, and show required pip packages.

This is a straightforward coding task - I should:
1. Look up the Azure SDK for Python Key Vault documentation
2. Write a script that demonstrates all four operations
3. Include proper error handling
4. Document the required packages

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
{"file_text":"#!/usr/bin/env python3\n\"\"\"\nAzure Key Vault Secrets CRUD Operations Demo\n\nThis script demonstrates all four CRUD operations on Azure Key Vault secrets:\n1. CREATE - Set a new secret\n2. READ - Retrieve the secret\n3. UPDATE - Update the secret value\n4. DELETE - Delete and purge the secret\n\nRequirements:\n    pip install azure-keyvault-secrets azure-identity\n\nEnvironment Variables:\n    VAULT_URL - Your Azure Key Vault URL (e.g., https://my-vault.vault.azure.net/)\n\nAuthentication:\n    Uses DefaultAzureCredential which attempts authentication via:\n    - Environment variables\n    - Managed Identity\n    - Azure CLI\n    - Azure PowerShell\n    - Interactive browser\n\"\"\"\n\nimport os\nimport sys\nfrom azure.identity import DefaultAzureCredential\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.core.exceptions import (\n    ResourceNotFoundError,\n    HttpResponseError,\n    AzureError\n)\n\n\ndef main():\n    \"\"\"Main function demonstrating CRUD operations on Azure Key Vault secrets.\"\"\"\n    \n    # Get vault URL from environment variable\n    vault_url = os.environ.get(\"VAULT_URL\")\n    if not vault_url:\n        print(\"Error: VAULT_URL environment variable is not set\")\n        print(\"Example: export VAULT_URL='https://my-vault.vault.azure.net/'\")\n        sys.exit(1)\n    \n    print(f\"Connecting to Key Vault: {vault_url}\")\n    print(\"-\" * 80)\n    \n    try:\n        # Initialize credential and client\n        credential = DefaultAzureCredential()\n        client = SecretClient(vault_url=vault_url, credential=credential)\n        \n        secret_name = \"my-secret\"\n        \n        # ============================================================\n        # 1. CREATE - Set a new secret\n        # ============================================================\n        print(\"\\n1. CREATE - Setting a new secret\")\n        print(\"-\" * 80)\n        try:\n            secret = client.set_secret(secret_name, \"my-secret-value\")\n            print(f\"✓ Secret created successfully!\")\n            print(f\"  Name: {secret.name}\")\n            print(f\"  Value: {secret.value}\")\n            print(f\"  Version: {secret.properties.version}\")\n            print(f\"  Created: {secret.properties.created_on}\")\n        except HttpResponseError as e:\n            print(f\"✗ Failed to create secret: {e.message}\")\n            raise\n        \n        # ============================================================\n        # 2. READ - Retrieve the secret\n        # ============================================================\n        print(\"\\n2. READ - Retrieving the secret\")\n        print(\"-\" * 80)\n        try:\n            retrieved_secret = client.get_secret(secret_name)\n            print(f\"✓ Secret retrieved successfully!\")\n            print(f\"  Name: {retrieved_secret.name}\")\n            print(f\"  Value: {retrieved_secret.value}\")\n            print(f\"  Version: {retrieved_secret.properties.version}\")\n            print(f\"  Content Type: {retrieved_secret.properties.content_type}\")\n            print(f\"  Enabled: {retrieved_secret.properties.enabled}\")\n        except ResourceNotFoundError:\n            print(f\"✗ Secret '{secret_name}' not found\")\n            raise\n        except HttpResponseError as e:\n            print(f\"✗ Failed to retrieve secret: {e.message}\")\n            raise\n        \n        # ============================================================\n        # 3. UPDATE - Update the secret to a new value\n        # ============================================================\n        print(\"\\n3. UPDATE - Updating the secret value\")\n        print(\"-\" * 80)\n        try:\n            # Update the secret value (creates a new version)\n            updated_secret = client.set_secret(secret_name, \"updated-value\")\n            print(f\"✓ Secret value updated successfully!\")\n            print(f\"  Name: {updated_secret.name}\")\n            print(f\"  New Value: {updated_secret.value}\")\n            print(f\"  New Version: {updated_secret.properties.version}\")\n            print(f\"  Updated: {updated_secret.properties.updated_on}\")\n            \n            # Optional: Update secret properties (metadata)\n            # This updates metadata without changing the value\n            print(\"\\n  Updating secret properties...\")\n            updated_properties = client.update_secret_properties(\n                secret_name,\n                content_type=\"text/plain\",\n                enabled=True\n            )\n            print(f\"✓ Secret properties updated!\")\n            print(f\"  Content Type: {updated_properties.content_type}\")\n            print(f\"  Enabled: {updated_properties.enabled}\")\n            \n        except HttpResponseError as e:\n            print(f\"✗ Failed to update secret: {e.message}\")\n            raise\n        \n        # ============================================================\n        # 4. DELETE - Delete and purge the secret\n        # ============================================================\n        print(\"\\n4. DELETE - Deleting and purging the secret\")\n        print(\"-\" * 80)\n        try:\n            # Delete the secret (soft delete)\n            print(f\"  Deleting secret '{secret_name}'...\")\n            delete_operation = client.begin_delete_secret(secret_name)\n            deleted_secret = delete_operation.result()\n            print(f\"✓ Secret deleted successfully!\")\n            print(f\"  Name: {deleted_secret.name}\")\n            print(f\"  Deleted Date: {deleted_secret.deleted_date}\")\n            print(f\"  Scheduled Purge Date: {deleted_secret.scheduled_purge_date}\")\n            print(f\"  Recovery ID: {deleted_secret.recovery_id}\")\n            \n            # Purge the secret permanently (only works with soft-delete enabled)\n            print(f\"\\n  Purging deleted secret '{secret_name}'...\")\n            client.purge_deleted_secret(secret_name)\n            print(f\"✓ Secret purged successfully!\")\n            print(f\"  The secret has been permanently deleted and cannot be recovered.\")\n            \n        except ResourceNotFoundError:\n            print(f\"✗ Secret '{secret_name}' not found for deletion\")\n            raise\n        except HttpResponseError as e:\n            if \"not currently in a deleted state\" in str(e):\n                print(f\"✗ Secret is not in deleted state (may already be purged)\")\n            else:\n                print(f\"✗ Failed to delete/purge secret: {e.message}\")\n            raise\n        \n        print(\"\\n\" + \"=\" * 80)\n        print(\"All CRUD operations completed successfully!\")\n        print(\"=\" * 80)\n        \n    except AzureError as e:\n        print(f\"\\n✗ Azure error occurred: {e}\")\n        sys.exit(1)\n    except Exception as e:\n        print(f\"\\n✗ Unexpected error occurred: {e}\")\n        sys.exit(1)\n    finally:\n        # Clean up\n        try:\n            credential.close()\n        except:\n            pass\n\n\nif __name__ == \"__main__\":\n    main()\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-1364908405/keyvault_crud_demo.py"}
```

### 🔧 5. create

**Input:**

```json
{"file_text":"# Azure Key Vault Secrets SDK for Python\nazure-keyvault-secrets\u003e=4.8.0\n\n# Azure Identity library for authentication\nazure-identity\u003e=1.15.0\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-1364908405/requirements.txt"}
```

### 🔧 6. create

**Input:**

```json
{"file_text":"# Azure Key Vault Secrets CRUD Demo\n\nThis script demonstrates all four CRUD operations (Create, Read, Update, Delete) on Azure Key Vault secrets using the Azure SDK for Python.\n\n## Prerequisites\n\n1. **Python 3.9 or later**\n2. **An Azure Key Vault** with soft-delete enabled\n3. **Appropriate permissions**: Your Azure identity needs the following Key Vault secret permissions:\n   - `secrets/set` - to create/update secrets\n   - `secrets/get` - to read secrets\n   - `secrets/delete` - to delete secrets\n   - `secrets/purge` - to permanently delete secrets\n\n## Installation\n\nInstall the required Python packages:\n\n```bash\npip install -r requirements.txt\n```\n\nOr install directly:\n\n```bash\npip install azure-keyvault-secrets azure-identity\n```\n\n## Authentication\n\nThe script uses `DefaultAzureCredential` which attempts authentication in the following order:\n\n1. **Environment variables** - Set these for service principal authentication:\n   - `AZURE_CLIENT_ID`\n   - `AZURE_TENANT_ID`\n   - `AZURE_CLIENT_SECRET`\n\n2. **Managed Identity** - Automatically used when running on Azure services\n\n3. **Azure CLI** - Run `az login` first:\n   ```bash\n   az login\n   ```\n\n4. **Azure PowerShell** - Run `Connect-AzAccount` first\n\n5. **Interactive browser** - Opens a browser window for authentication\n\n## Usage\n\n1. Set the Key Vault URL environment variable:\n   ```bash\n   export VAULT_URL=\"https://your-vault-name.vault.azure.net/\"\n   ```\n\n2. Run the script:\n   ```bash\n   python keyvault_crud_demo.py\n   ```\n\n## What the Script Does\n\n### 1. CREATE\nCreates a new secret named \"my-secret\" with value \"my-secret-value\"\n\n### 2. READ\nRetrieves the secret and displays its value and properties\n\n### 3. UPDATE\n- Updates the secret to a new value \"updated-value\" (creates a new version)\n- Updates secret metadata (content type, enabled status)\n\n### 4. DELETE\n- Soft-deletes the secret (can be recovered within retention period)\n- Purges the secret permanently (irreversible deletion)\n\n## Example Output\n\n```\nConnecting to Key Vault: https://my-vault.vault.azure.net/\n--------------------------------------------------------------------------------\n\n1. CREATE - Setting a new secret\n--------------------------------------------------------------------------------\n✓ Secret created successfully!\n  Name: my-secret\n  Value: my-secret-value\n  Version: abc123...\n  Created: 2026-03-21 19:00:00\n\n2. READ - Retrieving the secret\n--------------------------------------------------------------------------------\n✓ Secret retrieved successfully!\n  Name: my-secret\n  Value: my-secret-value\n  Version: abc123...\n\n3. UPDATE - Updating the secret value\n--------------------------------------------------------------------------------\n✓ Secret value updated successfully!\n  Name: my-secret\n  New Value: updated-value\n  New Version: def456...\n\n4. DELETE - Deleting and purging the secret\n--------------------------------------------------------------------------------\n✓ Secret deleted successfully!\n✓ Secret purged successfully!\n  The secret has been permanently deleted and cannot be recovered.\n\n================================================================================\nAll CRUD operations completed successfully!\n================================================================================\n```\n\n## Error Handling\n\nThe script includes comprehensive error handling for:\n- Missing environment variables\n- Authentication failures\n- Permission issues\n- Resource not found errors\n- Azure service errors\n\n## References\n\n- [Azure Key Vault Secrets SDK Documentation](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)\n- [Azure Identity Documentation](https://learn.microsoft.com/en-us/python/api/overview/azure/identity-readme)\n- [DefaultAzureCredential Documentation](https://learn.microsoft.com/en-us/python/api/azure-identity/azure.identity.defaultazurecredential)\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-1364908405/README.md"}
```

### 🔧 7. bash

**Input:**

```json
{"command":"ls -lh","description":"List created files"}
```

## Generated Files

- `README.md`
- `keyvault_crud_demo.py`
- `requirements.txt`

## Copilot Response

I'll help you create a Python script for Azure Key Vault CRUD operations using the official Azure SDK documentation. Let me first check the Azure SDK documentation and then write the script.

Now let me check the documentation for purging a deleted secret:

Perfect! Now I have all the information I need. Let me create a comprehensive Python script that demonstrates all four CRUD operations on Azure Key Vault secrets:

Now let me create a requirements.txt file with the necessary packages:

Finally, let me create a README with usage instructions:

Perfect! Let me verify the files were created:

Excellent! I've created a complete Azure Key Vault CRUD demo script based on the official Azure SDK documentation. Here's what I've provided:

## 📦 Files Created

1. **`keyvault_crud_demo.py`** - Main Python script demonstrating all CRUD operations
2. **`requirements.txt`** - Required pip packages
3. **`README.md`** - Comprehensive usage instructions

## 🔑 Key Features

**CRUD Operations Implemented:**
1. **CREATE** - Sets secret "my-secret" with value "my-secret-value"
2. **READ** - Retrieves and displays the secret value
3. **UPDATE** - Updates to "updated-value" and modifies metadata
4. **DELETE** - Soft-deletes then purges the secret permanently

**Additional Features:**
- ✅ Uses `DefaultAzureCredential` for flexible authentication
- ✅ Comprehensive error handling for all Azure exceptions
- ✅ Detailed console output showing each operation's results
- ✅ Proper cleanup with credential closure
- ✅ Comments and documentation throughout

## 🚀 Quick Start

```bash
# Install dependencies
pip install azure-keyvault-secrets azure-identity

# Set your vault URL
export VAULT_URL="https://your-vault-name.vault.azure.net/"

# Authenticate (if using Azure CLI)
az login

# Run the script
python keyvault_crud_demo.py
```

The script follows all Azure SDK best practices from the official documentation and includes proper error handling for common scenarios like missing secrets, authentication failures, and permission issues.

## Verification

**Result:** ✅ PASS

**Summary:** Code correctly implements all CRUD operations using Azure SDK with proper authentication, error handling, and soft-delete management.

**Reasoning:**

The generated code fully satisfies all requirements: (1) Implements all four CRUD operations correctly - CREATE uses set_secret() with 'my-secret'/'my-secret-value', READ uses get_secret() and prints the value, UPDATE uses set_secret() with 'updated-value', DELETE uses begin_delete_secret() followed by purge_deleted_secret(). (2) Uses correct Azure SDK packages - azure-keyvault-secrets and azure-identity are specified in requirements.txt with appropriate versions. (3) Properly uses DefaultAzureCredential for authentication and creates SecretClient with vault URL and credential. (4) Includes comprehensive error handling for ResourceNotFoundError, HttpResponseError, and AzureError. (5) Correctly handles soft-delete scenario by calling begin_delete_secret().result() to wait for deletion to complete before calling purge_deleted_secret(). (6) Code is syntactically valid Python with proper imports, structure, and API usage. The implementation goes beyond minimum requirements by also demonstrating update_secret_properties() and providing detailed documentation in the README.

---

[← Back to Summary](../../../../../../summary.md)
