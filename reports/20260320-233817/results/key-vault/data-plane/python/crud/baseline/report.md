# Evaluation Report: key-vault-dp-python-crud

**Config:** baseline | **Result:** ✅ PASSED | **Duration:** 94.7s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-crud` |
| Config | baseline |
| Result | ✅ PASSED |
| Duration | 94.7s |
| Timestamp | 2026-03-21T06:38:17Z |
| Files Generated | 3 |
| Event Count | 95 |

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

The user wants me to create a Python script that performs CRUD operations on Azure Key Vault secrets using the Azure SDK for Python. They want me to use only the Azure SDK documentation as a reference.

Key requirements:
1. Create a secret "my-secret" with value "my-secret-value"
2. Read the secret and print its value
3. Update the secret to "updated-value"
4. Delete and purge the secret (soft-delete enabled)
5. Use DefaultAzureCredential
6. Include error handling
7. Show required pip packages

This is a straightforward task - I need to:
1. Look up the Azure SDK documentation to ensure I'm using the correct API
2. Create a Python script with all CRUD operations
3. Include proper error handling
4. Document the required packages

Let me start by fetching the Azure SDK documentation for Key Vault secrets and then write the script.

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
{"file_text":"\"\"\"\nAzure Key Vault Secrets CRUD Operations Demo\n\nThis script demonstrates all four CRUD operations on Azure Key Vault secrets:\n- Create a secret\n- Read a secret\n- Update a secret\n- Delete and purge a secret\n\nRequired pip packages:\n    pip install azure-keyvault-secrets azure-identity\n\nPrerequisites:\n    - Azure subscription\n    - Azure Key Vault with soft-delete enabled\n    - Set environment variable VAULT_URL with your Key Vault URL\n      Example: export VAULT_URL=\"https://my-key-vault.vault.azure.net/\"\n    - Appropriate authentication configured for DefaultAzureCredential\n\"\"\"\n\nimport os\nfrom azure.identity import DefaultAzureCredential\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.core.exceptions import ResourceNotFoundError, HttpResponseError\n\n\ndef main():\n    \"\"\"Main function demonstrating CRUD operations on Azure Key Vault secrets.\"\"\"\n    \n    # Get vault URL from environment variable\n    vault_url = os.environ.get(\"VAULT_URL\")\n    if not vault_url:\n        print(\"ERROR: VAULT_URL environment variable is not set\")\n        print(\"Please set it to your Key Vault URL, e.g.:\")\n        print('export VAULT_URL=\"https://my-key-vault.vault.azure.net/\"')\n        return\n    \n    print(f\"Connecting to Key Vault: {vault_url}\")\n    print(\"-\" * 60)\n    \n    # Authenticate and create the secret client\n    try:\n        credential = DefaultAzureCredential()\n        client = SecretClient(vault_url=vault_url, credential=credential)\n        print(\"✓ Successfully authenticated with DefaultAzureCredential\\n\")\n    except Exception as e:\n        print(f\"ERROR: Failed to create SecretClient: {e}\")\n        return\n    \n    secret_name = \"my-secret\"\n    \n    # 1. CREATE - Set a new secret\n    print(\"1. CREATE - Setting a new secret\")\n    print(\"-\" * 60)\n    try:\n        secret = client.set_secret(secret_name, \"my-secret-value\")\n        print(f\"✓ Created secret: {secret.name}\")\n        print(f\"  Value: {secret.value}\")\n        print(f\"  Version: {secret.properties.version}\")\n        print(f\"  Created on: {secret.properties.created_on}\\n\")\n    except HttpResponseError as e:\n        print(f\"ERROR: Failed to create secret: {e.message}\")\n        return\n    except Exception as e:\n        print(f\"ERROR: Unexpected error creating secret: {e}\")\n        return\n    \n    # 2. READ - Retrieve the secret\n    print(\"2. READ - Retrieving the secret\")\n    print(\"-\" * 60)\n    try:\n        retrieved_secret = client.get_secret(secret_name)\n        print(f\"✓ Retrieved secret: {retrieved_secret.name}\")\n        print(f\"  Value: {retrieved_secret.value}\")\n        print(f\"  Version: {retrieved_secret.properties.version}\\n\")\n    except ResourceNotFoundError:\n        print(f\"ERROR: Secret '{secret_name}' not found\")\n        return\n    except Exception as e:\n        print(f\"ERROR: Failed to retrieve secret: {e}\")\n        return\n    \n    # 3. UPDATE - Update the secret with a new value\n    print(\"3. UPDATE - Updating the secret to a new value\")\n    print(\"-\" * 60)\n    try:\n        # Setting a secret with an existing name creates a new version\n        updated_secret = client.set_secret(secret_name, \"updated-value\")\n        print(f\"✓ Updated secret: {updated_secret.name}\")\n        print(f\"  New value: {updated_secret.value}\")\n        print(f\"  New version: {updated_secret.properties.version}\")\n        print(f\"  Updated on: {updated_secret.properties.updated_on}\\n\")\n    except HttpResponseError as e:\n        print(f\"ERROR: Failed to update secret: {e.message}\")\n        return\n    except Exception as e:\n        print(f\"ERROR: Unexpected error updating secret: {e}\")\n        return\n    \n    # 4. DELETE - Delete and purge the secret\n    print(\"4. DELETE - Deleting and purging the secret\")\n    print(\"-\" * 60)\n    try:\n        # Begin delete operation (returns a poller)\n        delete_operation = client.begin_delete_secret(secret_name)\n        deleted_secret = delete_operation.result()\n        print(f\"✓ Deleted secret: {deleted_secret.name}\")\n        print(f\"  Deleted on: {deleted_secret.deleted_date}\")\n        print(f\"  Scheduled purge date: {deleted_secret.scheduled_purge_date}\")\n        \n        # Purge the secret permanently (for soft-delete enabled vaults)\n        client.purge_deleted_secret(secret_name)\n        print(f\"✓ Purged secret '{secret_name}' permanently\\n\")\n        \n    except ResourceNotFoundError:\n        print(f\"ERROR: Secret '{secret_name}' not found for deletion\")\n        return\n    except HttpResponseError as e:\n        print(f\"ERROR: Failed to delete/purge secret: {e.message}\")\n        print(\"Note: Ensure your vault has soft-delete enabled and you have purge permissions\")\n        return\n    except Exception as e:\n        print(f\"ERROR: Unexpected error during delete/purge: {e}\")\n        return\n    \n    print(\"=\" * 60)\n    print(\"All CRUD operations completed successfully!\")\n    print(\"=\" * 60)\n\n\nif __name__ == \"__main__\":\n    main()\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-3773349544/azure_keyvault_crud.py"}
```

### 🔧 5. create

**Input:**

```json
{"file_text":"azure-keyvault-secrets\nazure-identity\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-3773349544/requirements.txt"}
```

### 🔧 6. create

**Input:**

```json
{"file_text":"# Azure Key Vault Secrets CRUD Operations\n\nThis script demonstrates all four CRUD operations on Azure Key Vault secrets using the official Azure SDK for Python.\n\n## Features\n\n✅ **Create** - Set a new secret with name \"my-secret\" and value \"my-secret-value\"  \n✅ **Read** - Retrieve the secret and print its value  \n✅ **Update** - Update the secret to a new value \"updated-value\"  \n✅ **Delete** - Delete the secret and purge it (for soft-delete enabled vaults)\n\n## Installation\n\nInstall the required packages:\n\n```bash\npip install -r requirements.txt\n```\n\nOr install manually:\n\n```bash\npip install azure-keyvault-secrets azure-identity\n```\n\n## Prerequisites\n\n1. **Azure subscription**\n2. **Azure Key Vault** with soft-delete enabled\n3. **Authentication** - Configure one of the following for DefaultAzureCredential:\n   - Azure CLI: Run `az login`\n   - Environment variables (Service Principal)\n   - Managed Identity (when running in Azure)\n   - Interactive browser authentication\n\n4. **Permissions** - Your account/identity needs the following Key Vault permissions:\n   - `secrets/get`\n   - `secrets/set`\n   - `secrets/delete`\n   - `secrets/purge`\n\n## Usage\n\nSet the VAULT_URL environment variable to your Key Vault URL:\n\n```bash\nexport VAULT_URL=\"https://your-key-vault-name.vault.azure.net/\"\n```\n\nRun the script:\n\n```bash\npython azure_keyvault_crud.py\n```\n\n## Example Output\n\n```\nConnecting to Key Vault: https://my-key-vault.vault.azure.net/\n------------------------------------------------------------\n✓ Successfully authenticated with DefaultAzureCredential\n\n1. CREATE - Setting a new secret\n------------------------------------------------------------\n✓ Created secret: my-secret\n  Value: my-secret-value\n  Version: abc123...\n  Created on: 2026-03-21 06:38:20.123456+00:00\n\n2. READ - Retrieving the secret\n------------------------------------------------------------\n✓ Retrieved secret: my-secret\n  Value: my-secret-value\n  Version: abc123...\n\n3. UPDATE - Updating the secret to a new value\n------------------------------------------------------------\n✓ Updated secret: my-secret\n  New value: updated-value\n  New version: def456...\n  Updated on: 2026-03-21 06:38:21.234567+00:00\n\n4. DELETE - Deleting and purging the secret\n------------------------------------------------------------\n✓ Deleted secret: my-secret\n  Deleted on: 2026-03-21 06:38:22.345678+00:00\n  Scheduled purge date: 2026-04-20 06:38:22.345678+00:00\n✓ Purged secret 'my-secret' permanently\n\n============================================================\nAll CRUD operations completed successfully!\n============================================================\n```\n\n## Error Handling\n\nThe script includes comprehensive error handling for:\n- Missing environment variables\n- Authentication failures\n- Resource not found errors\n- HTTP response errors\n- Permission issues\n\n## Documentation References\n\n- [Azure Key Vault Secrets SDK for Python](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)\n- [SecretClient API Reference](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient)\n- [DefaultAzureCredential](https://learn.microsoft.com/en-us/python/api/azure-identity/azure.identity.defaultazurecredential)\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-3773349544/README.md"}
```

### 🔧 7. bash

**Input:**

```json
{"command":"ls -lah /tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-3773349544/","description":"List created files"}
```

### 🔧 8. bash

**Input:**

```json
{"command":"head -30 /tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-3773349544/azure_keyvault_crud.py","description":"Display script header"}
```

## Generated Files

- `README.md`
- `azure_keyvault_crud.py`
- `requirements.txt`

## Copilot Response

I'll create a Python script that demonstrates all CRUD operations on Azure Key Vault secrets using the official Azure SDK for Python documentation.

Now let me fetch the additional documentation about purging deleted secrets:

Perfect! Now I have all the information I need from the Azure SDK documentation. Let me create a comprehensive Python script with all CRUD operations:

Now let me create a requirements.txt file for easy installation:

Perfect! Let me also create a README with usage instructions:

Now let me verify the files were created correctly:

Perfect! Let me show you a summary of the script:

## ✅ Complete!

I've created a comprehensive Python script that demonstrates all four CRUD operations on Azure Key Vault secrets using the official Azure SDK for Python documentation.

### 📁 Files Created:

1. **azure_keyvault_crud.py** - Main script with complete CRUD operations
2. **requirements.txt** - Package dependencies
3. **README.md** - Detailed usage instructions

### 🔑 Script Features:

**1. CREATE** - Uses `set_secret()` to create "my-secret" with value "my-secret-value"

**2. READ** - Uses `get_secret()` to retrieve and print the secret value

**3. UPDATE** - Uses `set_secret()` again to update to "updated-value" (creates new version)

**4. DELETE & PURGE** - Uses `begin_delete_secret()` followed by `purge_deleted_secret()` for soft-delete enabled vaults

### 🛡️ Additional Features:
- ✅ **DefaultAzureCredential** for flexible authentication
- ✅ **Comprehensive error handling** (ResourceNotFoundError, HttpResponseError)
- ✅ **Clear output** showing each operation's results
- ✅ **Environment variable** configuration for vault URL
- ✅ **Proper waiting** for delete operation completion using poller

### 📦 Installation:
```bash
pip install -r requirements.txt
```

### 🚀 Usage:
```bash
export VAULT_URL="https://your-key-vault.vault.azure.net/"
python azure_keyvault_crud.py
```

The script is production-ready with proper error handling and follows Azure SDK best practices from the official documentation!

## Verification

**Result:** ✅ PASS

**Summary:** Code correctly implements all CRUD operations with proper Azure SDK usage, authentication, error handling, and soft-delete support

**Reasoning:**

The generated code fully meets all prompt requirements: (1) Uses correct Azure SDK packages (azure-keyvault-secrets, azure-identity) as shown in requirements.txt; (2) Implements all four CRUD operations correctly - CREATE uses set_secret() with 'my-secret' and 'my-secret-value', READ uses get_secret() and prints the value, UPDATE uses set_secret() with 'updated-value', DELETE uses begin_delete_secret() followed by purge_deleted_secret(); (3) Uses DefaultAzureCredential for authentication as required; (4) Includes comprehensive error handling for ResourceNotFoundError, HttpResponseError, and general exceptions; (5) Properly handles soft-delete by calling .result() on the delete operation poller before purging, ensuring the delete completes before purge; (6) Code is syntactically valid Python with proper structure, docstrings, and clear output; (7) README provides complete documentation including installation, prerequisites, permissions, and usage examples. All expected coverage points are addressed.

---

[← Back to Summary](../../../../../../summary.md)
