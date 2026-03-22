# Evaluation Report: key-vault-dp-python-crud

**Config:** azure-mcp | **Result:** ✅ PASSED | **Duration:** 98.5s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-crud` |
| Config | azure-mcp |
| Result | ✅ PASSED |
| Duration | 98.5s |
| Timestamp | 2026-03-21T06:14:30Z |
| Files Generated | 3 |
| Event Count | 79 |

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

They want me to use only the Azure SDK for Python documentation, use DefaultAzureCredential, include error handling, and show required pip packages.

This is a straightforward task - I need to:
1. Look up the Azure SDK for Python documentation for Key Vault
2. Write a Python script with the CRUD operations
3. Include error handling
4. Document the required pip packages

Let me start by fetching the Azure SDK documentation for Key Vault secrets and then create the script.

## Tool Calls

### 🔧 1. report_intent

**Input:**

```json
{"intent":"Creating Key Vault CRUD script"}
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
{"file_text":"#!/usr/bin/env python3\n\"\"\"\nAzure Key Vault Secrets CRUD Operations Demo\n\nThis script demonstrates all four CRUD operations on Azure Key Vault secrets:\n1. Create a new secret\n2. Read the secret back\n3. Update the secret to a new value\n4. Delete and purge the secret\n\nRequirements:\n    pip install azure-keyvault-secrets azure-identity\n\nPrerequisites:\n    - Azure Key Vault with soft-delete enabled\n    - Set VAULT_URL environment variable (e.g., https://my-vault.vault.azure.net/)\n    - Configure authentication (Azure CLI, managed identity, etc. for DefaultAzureCredential)\n    - Required permissions: secrets/set, secrets/get, secrets/delete, secrets/purge\n\"\"\"\n\nimport os\nimport sys\nimport time\nfrom azure.identity import DefaultAzureCredential\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.core.exceptions import (\n    ResourceNotFoundError,\n    HttpResponseError,\n    ClientAuthenticationError\n)\n\n\ndef main():\n    \"\"\"Main function demonstrating CRUD operations on Azure Key Vault secrets.\"\"\"\n    \n    # Get vault URL from environment variable\n    vault_url = os.environ.get(\"VAULT_URL\")\n    if not vault_url:\n        print(\"ERROR: VAULT_URL environment variable not set\")\n        print(\"Example: export VAULT_URL='https://my-vault.vault.azure.net/'\")\n        sys.exit(1)\n    \n    print(f\"Connecting to Azure Key Vault: {vault_url}\\n\")\n    \n    # Initialize the credential and secret client\n    try:\n        credential = DefaultAzureCredential()\n        client = SecretClient(vault_url=vault_url, credential=credential)\n        print(\"✓ Successfully created SecretClient with DefaultAzureCredential\\n\")\n    except ClientAuthenticationError as e:\n        print(f\"ERROR: Authentication failed: {e.message}\")\n        sys.exit(1)\n    except Exception as e:\n        print(f\"ERROR: Failed to create SecretClient: {e}\")\n        sys.exit(1)\n    \n    secret_name = \"my-secret\"\n    \n    try:\n        # ===================================================================\n        # 1. CREATE - Set a new secret\n        # ===================================================================\n        print(\"=\" * 70)\n        print(\"1. CREATE - Setting a new secret\")\n        print(\"=\" * 70)\n        \n        secret_value = \"my-secret-value\"\n        print(f\"Creating secret '{secret_name}' with value '{secret_value}'...\")\n        \n        secret = client.set_secret(secret_name, secret_value)\n        \n        print(f\"✓ Secret created successfully!\")\n        print(f\"  Name: {secret.name}\")\n        print(f\"  Value: {secret.value}\")\n        print(f\"  Version: {secret.properties.version}\")\n        print(f\"  Created: {secret.properties.created_on}\")\n        print()\n        \n        # ===================================================================\n        # 2. READ - Retrieve the secret\n        # ===================================================================\n        print(\"=\" * 70)\n        print(\"2. READ - Retrieving the secret\")\n        print(\"=\" * 70)\n        \n        print(f\"Reading secret '{secret_name}'...\")\n        \n        retrieved_secret = client.get_secret(secret_name)\n        \n        print(f\"✓ Secret retrieved successfully!\")\n        print(f\"  Name: {retrieved_secret.name}\")\n        print(f\"  Value: {retrieved_secret.value}\")\n        print(f\"  Version: {retrieved_secret.properties.version}\")\n        print(f\"  Content Type: {retrieved_secret.properties.content_type}\")\n        print(f\"  Enabled: {retrieved_secret.properties.enabled}\")\n        print()\n        \n        # ===================================================================\n        # 3. UPDATE - Update the secret to a new value\n        # ===================================================================\n        print(\"=\" * 70)\n        print(\"3. UPDATE - Updating the secret value\")\n        print(\"=\" * 70)\n        \n        new_value = \"updated-value\"\n        print(f\"Updating secret '{secret_name}' to new value '{new_value}'...\")\n        \n        # set_secret creates a new version when the secret already exists\n        updated_secret = client.set_secret(secret_name, new_value)\n        \n        print(f\"✓ Secret updated successfully!\")\n        print(f\"  Name: {updated_secret.name}\")\n        print(f\"  New Value: {updated_secret.value}\")\n        print(f\"  New Version: {updated_secret.properties.version}\")\n        print(f\"  Updated: {updated_secret.properties.updated_on}\")\n        print()\n        \n        # Optionally, update secret properties (metadata) without changing the value\n        print(f\"Updating secret properties (content type)...\")\n        updated_properties = client.update_secret_properties(\n            secret_name,\n            content_type=\"text/plain\"\n        )\n        print(f\"✓ Secret properties updated!\")\n        print(f\"  Content Type: {updated_properties.content_type}\")\n        print()\n        \n        # ===================================================================\n        # 4. DELETE - Delete and purge the secret\n        # ===================================================================\n        print(\"=\" * 70)\n        print(\"4. DELETE - Deleting and purging the secret\")\n        print(\"=\" * 70)\n        \n        print(f\"Deleting secret '{secret_name}'...\")\n        \n        # begin_delete_secret returns a poller for long-running operations\n        delete_poller = client.begin_delete_secret(secret_name)\n        \n        # Wait for the deletion to complete\n        deleted_secret = delete_poller.result()\n        \n        print(f\"✓ Secret deleted successfully!\")\n        print(f\"  Name: {deleted_secret.name}\")\n        print(f\"  Deleted Date: {deleted_secret.deleted_date}\")\n        print(f\"  Scheduled Purge Date: {deleted_secret.scheduled_purge_date}\")\n        print(f\"  Recovery ID: {deleted_secret.recovery_id}\")\n        print()\n        \n        # Wait a moment to ensure deletion is complete before purging\n        print(\"Waiting for deletion to propagate...\")\n        time.sleep(2)\n        \n        # Purge the deleted secret (permanent deletion)\n        print(f\"Purging deleted secret '{secret_name}'...\")\n        client.purge_deleted_secret(secret_name)\n        \n        print(f\"✓ Secret purged successfully!\")\n        print(f\"  The secret '{secret_name}' has been permanently deleted.\")\n        print()\n        \n        # Verify the secret no longer exists\n        print(\"Verifying secret is completely removed...\")\n        try:\n            client.get_secret(secret_name)\n            print(\"WARNING: Secret still exists!\")\n        except ResourceNotFoundError:\n            print(\"✓ Confirmed: Secret no longer exists in the vault.\")\n        print()\n        \n    except ResourceNotFoundError as e:\n        print(f\"ERROR: Resource not found: {e.message}\")\n        sys.exit(1)\n    except HttpResponseError as e:\n        print(f\"ERROR: HTTP error occurred: {e.message}\")\n        print(f\"Status code: {e.status_code}\")\n        sys.exit(1)\n    except Exception as e:\n        print(f\"ERROR: An unexpected error occurred: {e}\")\n        sys.exit(1)\n    finally:\n        # Clean up - close the credential\n        credential.close()\n    \n    print(\"=\" * 70)\n    print(\"All CRUD operations completed successfully!\")\n    print(\"=\" * 70)\n\n\nif __name__ == \"__main__\":\n    main()\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-4220189447/keyvault_crud.py"}
```

### 🔧 5. create

**Input:**

```json
{"file_text":"azure-keyvault-secrets\u003e=4.7.0\nazure-identity\u003e=1.12.0\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-4220189447/requirements.txt"}
```

### 🔧 6. create

**Input:**

```json
{"file_text":"# Azure Key Vault Secrets CRUD Demo\n\nThis script demonstrates all four CRUD (Create, Read, Update, Delete) operations on Azure Key Vault secrets using the Azure SDK for Python.\n\n## Features\n\n- **Create**: Creates a new secret called \"my-secret\" with value \"my-secret-value\"\n- **Read**: Retrieves the secret and prints its value\n- **Update**: Updates the secret to a new value \"updated-value\"\n- **Delete**: Deletes the secret and purges it (for vaults with soft-delete enabled)\n\n## Prerequisites\n\n1. **Python 3.9 or later**\n2. **Azure Key Vault** with soft-delete enabled\n3. **Azure authentication** configured (one of the following):\n   - Azure CLI: `az login`\n   - Managed Identity (if running on Azure resources)\n   - Environment variables (AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET)\n   - Visual Studio Code Azure Account extension\n4. **Required permissions** on the Key Vault:\n   - secrets/set\n   - secrets/get\n   - secrets/delete\n   - secrets/purge\n\n## Installation\n\nInstall the required packages:\n\n```bash\npip install -r requirements.txt\n```\n\nOr install directly:\n\n```bash\npip install azure-keyvault-secrets azure-identity\n```\n\n## Configuration\n\nSet the VAULT_URL environment variable to your Azure Key Vault URL:\n\n```bash\nexport VAULT_URL='https://your-vault-name.vault.azure.net/'\n```\n\n## Usage\n\nRun the script:\n\n```bash\npython keyvault_crud.py\n```\n\nOr make it executable and run directly:\n\n```bash\nchmod +x keyvault_crud.py\n./keyvault_crud.py\n```\n\n## Example Output\n\n```\nConnecting to Azure Key Vault: https://your-vault.vault.azure.net/\n\n✓ Successfully created SecretClient with DefaultAzureCredential\n\n======================================================================\n1. CREATE - Setting a new secret\n======================================================================\nCreating secret 'my-secret' with value 'my-secret-value'...\n✓ Secret created successfully!\n  Name: my-secret\n  Value: my-secret-value\n  Version: abc123...\n  Created: 2026-03-21 06:14:33\n\n======================================================================\n2. READ - Retrieving the secret\n======================================================================\nReading secret 'my-secret'...\n✓ Secret retrieved successfully!\n  Name: my-secret\n  Value: my-secret-value\n  Version: abc123...\n\n======================================================================\n3. UPDATE - Updating the secret value\n======================================================================\nUpdating secret 'my-secret' to new value 'updated-value'...\n✓ Secret updated successfully!\n  Name: my-secret\n  New Value: updated-value\n  New Version: def456...\n\n======================================================================\n4. DELETE - Deleting and purging the secret\n======================================================================\nDeleting secret 'my-secret'...\n✓ Secret deleted successfully!\nPurging deleted secret 'my-secret'...\n✓ Secret purged successfully!\n\n======================================================================\nAll CRUD operations completed successfully!\n======================================================================\n```\n\n## Error Handling\n\nThe script includes comprehensive error handling for:\n- Missing environment variables\n- Authentication failures\n- Resource not found errors\n- HTTP response errors\n- Unexpected exceptions\n\n## Authentication Methods\n\n`DefaultAzureCredential` attempts authentication through multiple methods in order:\n1. Environment variables\n2. Managed Identity\n3. Visual Studio Code\n4. Azure CLI\n5. Azure PowerShell\n6. Interactive browser\n\n## References\n\n- [Azure Key Vault Secrets SDK Documentation](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)\n- [Azure Identity SDK Documentation](https://learn.microsoft.com/en-us/python/api/overview/azure/identity-readme)\n- [Azure Key Vault Overview](https://learn.microsoft.com/en-us/azure/key-vault/general/overview)\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-4220189447/README.md"}
```

### 🔧 7. bash

**Input:**

```json
{"command":"ls -lh /tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-4220189447/","description":"List created files"}
```

## Generated Files

- `README.md`
- `keyvault_crud.py`
- `requirements.txt`

## Copilot Response

I'll help you create a Python script for Azure Key Vault CRUD operations using the Azure SDK documentation. Let me first fetch the relevant documentation and then create the script.

Now let me fetch information about purging deleted secrets:

Perfect! Now I have all the information I need. Let me create a comprehensive Python script with CRUD operations for Azure Key Vault secrets:

Now let me create a requirements.txt file and a README with usage instructions:

Perfect! Let me verify the script and show you what was created:

Perfect! I've created a complete Azure Key Vault CRUD operations script based on the official Azure SDK for Python documentation. Here's what was created:

## 📁 Files Created

1. **keyvault_crud.py** - Main script with all CRUD operations
2. **requirements.txt** - Python package dependencies
3. **README.md** - Complete usage documentation

## 🔑 Key Features

The script demonstrates all four CRUD operations:

1. **CREATE**: Creates secret "my-secret" with value "my-secret-value"
2. **READ**: Retrieves and prints the secret value
3. **UPDATE**: Updates the secret to "updated-value" (creates new version)
4. **DELETE**: Deletes and purges the secret from soft-delete vault

## ✨ Additional Features

- **DefaultAzureCredential** for automatic authentication
- **Comprehensive error handling** (ResourceNotFoundError, HttpResponseError, etc.)
- **Proper cleanup** with credential closing
- **Detailed output** showing each operation's results
- **Property updates** demonstration (content type metadata)

## 📦 Required Packages

```bash
pip install azure-keyvault-secrets azure-identity
```

## 🚀 Usage

```bash
export VAULT_URL='https://your-vault.vault.azure.net/'
python keyvault_crud.py
```

The script includes all best practices from the Azure SDK documentation including proper error handling, resource cleanup, and waiting for long-running operations to complete.

## Verification

**Result:** ✅ PASS

**Summary:** Code fully implements all CRUD operations with correct Azure SDK usage, proper authentication, error handling, and soft-delete support as required.

**Reasoning:**

The generated code fully meets all requirements from the original prompt. It correctly implements all four CRUD operations: (1) Creates 'my-secret' with 'my-secret-value' using set_secret(), (2) Reads it back with get_secret() and prints the value, (3) Updates to 'updated-value' using set_secret() again, (4) Deletes with begin_delete_secret() and purges with purge_deleted_secret(). The code uses DefaultAzureCredential for authentication as required, creates a SecretClient with vault URL and credential, and includes comprehensive error handling for ResourceNotFoundError, HttpResponseError, and ClientAuthenticationError. The requirements.txt correctly specifies azure-keyvault-secrets and azure-identity packages. The code properly handles soft-delete by calling result() on the delete poller to wait for completion and includes a time.sleep(2) before purging. The script is syntactically valid Python, well-structured with clear section markers, includes detailed output for each operation, and even goes beyond requirements by demonstrating update_secret_properties() and verifying the secret is removed. The README provides excellent documentation covering prerequisites, installation, configuration, usage, and authentication methods.

---

[← Back to Summary](../../../../../../summary.md)
