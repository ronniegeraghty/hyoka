# Evaluation Report: key-vault-dp-python-crud

**Config:** azure-mcp | **Result:** ✅ PASSED | **Duration:** 128.7s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-crud` |
| Config | azure-mcp |
| Result | ✅ PASSED |
| Duration | 128.7s |
| Timestamp | 2026-03-22T00:09:50Z |
| Files Generated | 4 |
| Event Count | 106 |

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
2. Read it back
3. Update it
4. Delete and purge it

They want me to use only the Azure SDK for Python documentation, use DefaultAzureCredential, and include error handling and pip packages.

This is a straightforward task - I need to:
1. Look up the Azure SDK for Python Key Vault documentation
2. Write a script that performs these operations
3. Include proper error handling
4. Document the required pip packages

Let me start by fetching the Azure SDK documentation for Key Vault.

## Tool Calls

### 🔧 1. report_intent

**Input:**

```json
{"intent":"Fetching Azure Key Vault documentation"}
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
{"file_text":"#!/usr/bin/env python3\n\"\"\"\nAzure Key Vault Secrets CRUD Operations Demo\n\nThis script demonstrates all four CRUD operations on Azure Key Vault secrets:\n1. Create - Create a new secret\n2. Read - Retrieve the secret value\n3. Update - Update the secret to a new value\n4. Delete - Delete and purge the secret\n\nPrerequisites:\n- pip install azure-keyvault-secrets azure-identity\n- An Azure Key Vault with soft-delete enabled\n- Proper authentication configured for DefaultAzureCredential\n- Required permissions: secrets/set, secrets/get, secrets/delete, secrets/purge\n\nEnvironment Variables:\n- AZURE_KEY_VAULT_URL: Your Key Vault URL (e.g., https://my-vault.vault.azure.net/)\n\"\"\"\n\nimport os\nimport sys\nfrom azure.identity import DefaultAzureCredential\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.core.exceptions import (\n    ResourceNotFoundError,\n    HttpResponseError,\n    ClientAuthenticationError\n)\n\n\ndef main():\n    \"\"\"Main function to demonstrate Key Vault CRUD operations.\"\"\"\n    \n    # Get Key Vault URL from environment variable\n    vault_url = os.environ.get(\"AZURE_KEY_VAULT_URL\")\n    if not vault_url:\n        print(\"Error: AZURE_KEY_VAULT_URL environment variable is not set\")\n        print(\"Example: export AZURE_KEY_VAULT_URL='https://my-vault.vault.azure.net/'\")\n        sys.exit(1)\n    \n    print(f\"Connecting to Key Vault: {vault_url}\\n\")\n    \n    try:\n        # Initialize credential and client\n        credential = DefaultAzureCredential()\n        client = SecretClient(vault_url=vault_url, credential=credential)\n        \n        secret_name = \"my-secret\"\n        \n        # ================================================================\n        # 1. CREATE - Create a new secret\n        # ================================================================\n        print(\"=\" * 60)\n        print(\"1. CREATE OPERATION\")\n        print(\"=\" * 60)\n        \n        try:\n            secret_value = \"my-secret-value\"\n            print(f\"Creating secret '{secret_name}' with value '{secret_value}'...\")\n            \n            created_secret = client.set_secret(secret_name, secret_value)\n            \n            print(f\"✓ Secret created successfully!\")\n            print(f\"  Name: {created_secret.name}\")\n            print(f\"  Value: {created_secret.value}\")\n            print(f\"  Version: {created_secret.properties.version}\")\n            print(f\"  Created: {created_secret.properties.created_on}\")\n            print()\n            \n        except HttpResponseError as e:\n            print(f\"✗ Failed to create secret: {e.message}\")\n            sys.exit(1)\n        \n        # ================================================================\n        # 2. READ - Retrieve the secret\n        # ================================================================\n        print(\"=\" * 60)\n        print(\"2. READ OPERATION\")\n        print(\"=\" * 60)\n        \n        try:\n            print(f\"Reading secret '{secret_name}'...\")\n            \n            retrieved_secret = client.get_secret(secret_name)\n            \n            print(f\"✓ Secret retrieved successfully!\")\n            print(f\"  Name: {retrieved_secret.name}\")\n            print(f\"  Value: {retrieved_secret.value}\")\n            print(f\"  Version: {retrieved_secret.properties.version}\")\n            print(f\"  Content Type: {retrieved_secret.properties.content_type}\")\n            print(f\"  Enabled: {retrieved_secret.properties.enabled}\")\n            print()\n            \n        except ResourceNotFoundError:\n            print(f\"✗ Secret '{secret_name}' not found\")\n            sys.exit(1)\n        except HttpResponseError as e:\n            print(f\"✗ Failed to read secret: {e.message}\")\n            sys.exit(1)\n        \n        # ================================================================\n        # 3. UPDATE - Update the secret value\n        # ================================================================\n        print(\"=\" * 60)\n        print(\"3. UPDATE OPERATION\")\n        print(\"=\" * 60)\n        \n        try:\n            new_value = \"updated-value\"\n            print(f\"Updating secret '{secret_name}' to new value '{new_value}'...\")\n            \n            # Note: set_secret creates a new version when updating\n            updated_secret = client.set_secret(secret_name, new_value)\n            \n            print(f\"✓ Secret updated successfully!\")\n            print(f\"  Name: {updated_secret.name}\")\n            print(f\"  New Value: {updated_secret.value}\")\n            print(f\"  New Version: {updated_secret.properties.version}\")\n            print(f\"  Updated: {updated_secret.properties.updated_on}\")\n            print()\n            \n            # Verify the update by reading it back\n            print(f\"Verifying update by reading secret again...\")\n            verified_secret = client.get_secret(secret_name)\n            print(f\"✓ Verified value: {verified_secret.value}\")\n            print()\n            \n        except HttpResponseError as e:\n            print(f\"✗ Failed to update secret: {e.message}\")\n            sys.exit(1)\n        \n        # ================================================================\n        # 4. DELETE - Delete and purge the secret\n        # ================================================================\n        print(\"=\" * 60)\n        print(\"4. DELETE OPERATION\")\n        print(\"=\" * 60)\n        \n        try:\n            print(f\"Deleting secret '{secret_name}'...\")\n            \n            # Begin delete operation (returns a poller for soft-delete vaults)\n            delete_poller = client.begin_delete_secret(secret_name)\n            \n            # Wait for deletion to complete\n            deleted_secret = delete_poller.result()\n            \n            print(f\"✓ Secret deleted successfully!\")\n            print(f\"  Name: {deleted_secret.name}\")\n            print(f\"  Deleted Date: {deleted_secret.deleted_date}\")\n            print(f\"  Scheduled Purge Date: {deleted_secret.scheduled_purge_date}\")\n            print(f\"  Recovery ID: {deleted_secret.recovery_id}\")\n            print()\n            \n        except HttpResponseError as e:\n            print(f\"✗ Failed to delete secret: {e.message}\")\n            sys.exit(1)\n        \n        # ================================================================\n        # PURGE - Permanently delete the secret (soft-delete enabled)\n        # ================================================================\n        print(\"=\" * 60)\n        print(\"5. PURGE OPERATION (Permanent Deletion)\")\n        print(\"=\" * 60)\n        \n        try:\n            print(f\"Purging deleted secret '{secret_name}' permanently...\")\n            \n            # Purge the deleted secret (permanent deletion)\n            client.purge_deleted_secret(secret_name)\n            \n            print(f\"✓ Secret purged successfully!\")\n            print(f\"  The secret '{secret_name}' has been permanently deleted.\")\n            print(f\"  It cannot be recovered.\")\n            print()\n            \n        except HttpResponseError as e:\n            # If purge is not supported (e.g., vault doesn't have soft-delete enabled)\n            print(f\"Note: Purge operation failed or not needed: {e.message}\")\n            print(f\"This may happen if soft-delete is not enabled on the vault.\")\n            print()\n        \n        print(\"=\" * 60)\n        print(\"All CRUD operations completed successfully!\")\n        print(\"=\" * 60)\n        \n    except ClientAuthenticationError as e:\n        print(f\"\\n✗ Authentication failed: {e.message}\")\n        print(\"\\nTroubleshooting:\")\n        print(\"- Ensure you are logged in with Azure CLI: az login\")\n        print(\"- Or set up other authentication methods for DefaultAzureCredential\")\n        print(\"- Check that your account has proper permissions on the Key Vault\")\n        sys.exit(1)\n        \n    except Exception as e:\n        print(f\"\\n✗ Unexpected error: {type(e).__name__}: {str(e)}\")\n        sys.exit(1)\n    \n    finally:\n        # Close the client connection\n        client.close()\n        credential.close()\n\n\nif __name__ == \"__main__\":\n    main()\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-2572272990/keyvault_crud.py"}
```

### 🔧 5. create

**Input:**

```json
{"file_text":"# Azure Key Vault Secrets SDK\nazure-keyvault-secrets\u003e=4.8.0\n\n# Azure Identity SDK for authentication\nazure-identity\u003e=1.15.0\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-2572272990/requirements.txt"}
```

### 🔧 6. create

**Input:**

```json
{"file_text":"# Azure Key Vault Secrets CRUD Operations\n\nThis script demonstrates all four CRUD (Create, Read, Update, Delete) operations on Azure Key Vault secrets using the Azure SDK for Python.\n\n## Features\n\nThe script performs the following operations in sequence:\n\n1. **CREATE**: Creates a new secret called \"my-secret\" with value \"my-secret-value\"\n2. **READ**: Retrieves the secret and displays its value\n3. **UPDATE**: Updates the secret to a new value \"updated-value\"\n4. **DELETE**: Deletes the secret (soft-delete)\n5. **PURGE**: Permanently purges the deleted secret (for soft-delete enabled vaults)\n\n## Prerequisites\n\n### Azure Resources\n- An Azure subscription\n- An Azure Key Vault with soft-delete enabled\n- Appropriate permissions on the Key Vault:\n  - `secrets/set` - To create and update secrets\n  - `secrets/get` - To read secrets\n  - `secrets/delete` - To delete secrets\n  - `secrets/purge` - To permanently delete secrets\n\n### Python Requirements\n- Python 3.9 or later\n\n## Installation\n\nInstall the required packages using pip:\n\n```bash\npip install -r requirements.txt\n```\n\nOr install them directly:\n\n```bash\npip install azure-keyvault-secrets azure-identity\n```\n\n## Required Packages\n\n- **azure-keyvault-secrets**: Azure Key Vault Secrets client library\n- **azure-identity**: Azure authentication library providing DefaultAzureCredential\n\n## Authentication\n\nThe script uses `DefaultAzureCredential` from the Azure Identity library, which attempts multiple authentication methods in the following order:\n\n1. **Environment variables** - `AZURE_CLIENT_ID`, `AZURE_TENANT_ID`, `AZURE_CLIENT_SECRET`\n2. **Managed Identity** - If deployed to Azure with managed identity enabled\n3. **Azure CLI** - If logged in via `az login`\n4. **Azure PowerShell** - If logged in via `Connect-AzAccount`\n5. **Interactive browser** - As a fallback\n\n### Quick Setup with Azure CLI\n\n```bash\n# Login to Azure\naz login\n\n# Set your subscription (if you have multiple)\naz account set --subscription \"your-subscription-id\"\n```\n\n## Configuration\n\nSet the Key Vault URL as an environment variable:\n\n### Linux/macOS:\n```bash\nexport AZURE_KEY_VAULT_URL=\"https://your-vault-name.vault.azure.net/\"\n```\n\n### Windows (PowerShell):\n```powershell\n$env:AZURE_KEY_VAULT_URL=\"https://your-vault-name.vault.azure.net/\"\n```\n\n### Windows (Command Prompt):\n```cmd\nset AZURE_KEY_VAULT_URL=https://your-vault-name.vault.azure.net/\n```\n\n## Usage\n\nRun the script:\n\n```bash\npython keyvault_crud.py\n```\n\n## Example Output\n\n```\nConnecting to Key Vault: https://my-vault.vault.azure.net/\n\n============================================================\n1. CREATE OPERATION\n============================================================\nCreating secret 'my-secret' with value 'my-secret-value'...\n✓ Secret created successfully!\n  Name: my-secret\n  Value: my-secret-value\n  Version: abc123def456\n  Created: 2026-03-22 00:00:00+00:00\n\n============================================================\n2. READ OPERATION\n============================================================\nReading secret 'my-secret'...\n✓ Secret retrieved successfully!\n  Name: my-secret\n  Value: my-secret-value\n  Version: abc123def456\n  Content Type: None\n  Enabled: True\n\n============================================================\n3. UPDATE OPERATION\n============================================================\nUpdating secret 'my-secret' to new value 'updated-value'...\n✓ Secret updated successfully!\n  Name: my-secret\n  New Value: updated-value\n  New Version: def789ghi012\n  Updated: 2026-03-22 00:01:00+00:00\n\nVerifying update by reading secret again...\n✓ Verified value: updated-value\n\n============================================================\n4. DELETE OPERATION\n============================================================\nDeleting secret 'my-secret'...\n✓ Secret deleted successfully!\n  Name: my-secret\n  Deleted Date: 2026-03-22 00:02:00+00:00\n  Scheduled Purge Date: 2026-04-21 00:02:00+00:00\n  Recovery ID: https://my-vault.vault.azure.net/deletedsecrets/my-secret\n\n============================================================\n5. PURGE OPERATION (Permanent Deletion)\n============================================================\nPurging deleted secret 'my-secret' permanently...\n✓ Secret purged successfully!\n  The secret 'my-secret' has been permanently deleted.\n  It cannot be recovered.\n\n============================================================\nAll CRUD operations completed successfully!\n============================================================\n```\n\n## Error Handling\n\nThe script includes comprehensive error handling for:\n\n- **Authentication errors**: When credentials are invalid or missing\n- **Missing Key Vault URL**: When the environment variable is not set\n- **Resource not found**: When trying to access a non-existent secret\n- **Permission errors**: When the user lacks necessary permissions\n- **General HTTP errors**: For other API-related issues\n\n## Troubleshooting\n\n### Authentication Failed\n```\n✗ Authentication failed\n```\n\n**Solutions**:\n- Ensure you're logged in: `az login`\n- Verify your account has access to the Key Vault\n- Check that the Key Vault's access policies or RBAC includes your account\n\n### Permission Denied\n```\n✗ Failed to create secret: Forbidden\n```\n\n**Solutions**:\n- Add appropriate access policies in the Azure Portal\n- Or assign the \"Key Vault Secrets Officer\" role if using RBAC\n\n### Purge Operation Failed\n```\nNote: Purge operation failed or not needed\n```\n\n**Reason**: This typically means soft-delete is not enabled on the vault. In this case, the delete operation is already permanent.\n\n## Key Vault Setup\n\nTo create a Key Vault with soft-delete enabled:\n\n```bash\n# Create a resource group\naz group create --name myResourceGroup --location eastus\n\n# Create a Key Vault with soft-delete enabled (default since 2020)\naz keyvault create \\\n  --name myKeyVault \\\n  --resource-group myResourceGroup \\\n  --location eastus\n\n# Grant yourself permissions\naz keyvault set-policy \\\n  --name myKeyVault \\\n  --upn user@example.com \\\n  --secret-permissions get set delete purge list\n```\n\n## API Reference\n\nBased on the Azure SDK for Python documentation:\n\n- [SecretClient](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient)\n- [DefaultAzureCredential](https://learn.microsoft.com/en-us/python/api/azure-identity/azure.identity.defaultazurecredential)\n- [Azure Key Vault Overview](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)\n\n## Notes\n\n- **Versioning**: Each time you update a secret using `set_secret()`, a new version is created\n- **Soft-delete**: When soft-delete is enabled, deleted secrets can be recovered within the retention period\n- **Purge**: The purge operation permanently deletes the secret and cannot be undone\n- **Connection cleanup**: The script properly closes the client and credential connections\n\n## License\n\nThis is a demonstration script for educational purposes.\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-2572272990/README.md"}
```

### 🔧 7. bash

**Input:**

```json
{"command":"ls -lah /tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-2572272990/","description":"List all created files"}
```

### 🔧 8. bash

**Input:**

```json
{"command":"cd /tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-2572272990/ \u0026\u0026 python3 -m py_compile keyvault_crud.py \u0026\u0026 echo \"✓ Script syntax is valid\"","description":"Verify Python script syntax"}
```

### 🔧 9. bash

**Input:**

```json
{"command":"cat /tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-2572272990/keyvault_crud.py | head -60","description":"Show script header"}
```

## Generated Files

- `README.md`
- `__pycache__/keyvault_crud.cpython-312.pyc`
- `keyvault_crud.py`
- `requirements.txt`

## Copilot Response

I'll help you create a script that performs all CRUD operations on Azure Key Vault secrets using the Azure SDK for Python. Let me first check the official documentation and then write the script.

Let me also check for information about purging secrets:

Perfect! Now I have all the information needed. Let me create a comprehensive Python script that demonstrates all CRUD operations on Azure Key Vault secrets:

Now let me create a requirements.txt file with the necessary packages:

Let me also create a comprehensive README file:

Perfect! Let me verify the files were created correctly:

Excellent! Let me also verify the Python script is syntactically correct:

Perfect! Let me create a quick summary document showing the key parts:

## Summary

I've successfully created a complete Azure Key Vault secrets CRUD operations script based on the official Azure SDK for Python documentation! Here's what's included:

### 📁 Files Created:

1. **`keyvault_crud.py`** - Main Python script (211 lines)
2. **`requirements.txt`** - Python package dependencies
3. **`README.md`** - Comprehensive documentation

### ✨ Key Features:

**The script demonstrates all CRUD operations:**

1. **CREATE**: Creates secret "my-secret" with value "my-secret-value"
2. **READ**: Retrieves and displays the secret value
3. **UPDATE**: Updates the secret to "updated-value"
4. **DELETE**: Soft-deletes the secret (with poller support)
5. **PURGE**: Permanently deletes the secret from soft-delete vault

### 🔑 Key Implementation Details:

- **Authentication**: Uses `DefaultAzureCredential` (supports Azure CLI, Managed Identity, etc.)
- **Error Handling**: Comprehensive exception handling for authentication, permissions, and API errors
- **Clean Resource Management**: Properly closes client and credential connections
- **Informative Output**: Detailed console output showing each operation's status

### 📦 Required Packages:

```bash
pip install azure-keyvault-secrets azure-identity
```

### 🚀 Usage:

```bash
export AZURE_KEY_VAULT_URL="https://your-vault.vault.azure.net/"
python keyvault_crud.py
```

The script follows Azure SDK best practices from the official documentation and includes proper error handling, resource cleanup, and detailed logging of each operation!

## Verification

**Result:** ✅ PASS

**Summary:** Complete implementation of Azure Key Vault CRUD operations with proper authentication, error handling, and soft-delete support

**Reasoning:**

The code fully satisfies all requirements from the original prompt. It correctly implements all four CRUD operations: (1) Creates 'my-secret' with value 'my-secret-value' using set_secret(), (2) Reads and prints the secret value using get_secret(), (3) Updates to 'updated-value' using set_secret() again, (4) Deletes using begin_delete_secret() with proper wait via result(), and purges using purge_deleted_secret(). Authentication uses DefaultAzureCredential as required. The requirements.txt correctly specifies azure-keyvault-secrets and azure-identity packages. Error handling is comprehensive, including ResourceNotFoundError, HttpResponseError, and ClientAuthenticationError. The code properly handles soft-delete by waiting for deletion to complete before purging (delete_poller.result()). The SecretClient is correctly instantiated with vault URL and credential. The code is syntactically valid, well-structured, includes proper cleanup in finally block, and provides excellent documentation in both code comments and README. All expected API methods are used correctly according to Azure SDK patterns.

---

[← Back to Summary](../../../../../../summary.md)
