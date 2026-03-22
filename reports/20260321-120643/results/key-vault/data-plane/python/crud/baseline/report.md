# Evaluation Report: key-vault-dp-python-crud

**Config:** baseline | **Result:** ✅ PASSED | **Duration:** 84.1s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-crud` |
| Config | baseline |
| Result | ✅ PASSED |
| Duration | 84.1s |
| Timestamp | 2026-03-21T19:06:43Z |
| Files Generated | 3 |
| Event Count | 84 |

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
2. Read it back
3. Update it
4. Delete and purge it

They want:
- Use of DefaultAzureCredential
- Proper error handling
- Required pip packages listed

I should:
1. Look up the official Azure SDK documentation
2. Write a complete script based on the documentation
3. Include error handling
4. Document the pip packages needed

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
{"max_length":10000,"url":"https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient?view=azure-python"}
```

### 🔧 4. create

**Input:**

```json
{"file_text":"\"\"\"\nAzure Key Vault Secrets CRUD Operations Demo\n\nThis script demonstrates all four CRUD operations on Azure Key Vault secrets:\n1. Create a new secret\n2. Read the secret back\n3. Update the secret to a new value\n4. Delete and purge the secret\n\nRequired pip packages:\n    pip install azure-keyvault-secrets azure-identity\n\nPrerequisites:\n- An Azure Key Vault instance\n- Set the VAULT_URL environment variable to your vault URL\n  (e.g., https://my-key-vault.vault.azure.net/)\n- Appropriate Azure credentials configured for DefaultAzureCredential\n- Required permissions: secrets/set, secrets/get, secrets/delete, secrets/purge\n\"\"\"\n\nimport os\nimport sys\nfrom azure.identity import DefaultAzureCredential\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.core.exceptions import ResourceNotFoundError, HttpResponseError\n\n\ndef main():\n    # Get vault URL from environment variable\n    vault_url = os.environ.get(\"VAULT_URL\")\n    if not vault_url:\n        print(\"Error: VAULT_URL environment variable is not set\")\n        print(\"Example: export VAULT_URL='https://my-key-vault.vault.azure.net/'\")\n        sys.exit(1)\n\n    # Authenticate using DefaultAzureCredential\n    try:\n        credential = DefaultAzureCredential()\n        client = SecretClient(vault_url=vault_url, credential=credential)\n        print(f\"Connected to Key Vault: {vault_url}\\n\")\n    except Exception as e:\n        print(f\"Error: Failed to authenticate or create client: {e}\")\n        sys.exit(1)\n\n    secret_name = \"my-secret\"\n\n    # ========================================================================\n    # 1. CREATE: Set a new secret\n    # ========================================================================\n    print(\"=\" * 70)\n    print(\"1. CREATE: Setting a new secret\")\n    print(\"=\" * 70)\n    try:\n        secret = client.set_secret(secret_name, \"my-secret-value\")\n        print(f\"✓ Secret created successfully!\")\n        print(f\"  Name: {secret.name}\")\n        print(f\"  Value: {secret.value}\")\n        print(f\"  Version: {secret.properties.version}\")\n        print()\n    except HttpResponseError as e:\n        print(f\"✗ Failed to create secret: {e.message}\")\n        sys.exit(1)\n    except Exception as e:\n        print(f\"✗ Unexpected error during secret creation: {e}\")\n        sys.exit(1)\n\n    # ========================================================================\n    # 2. READ: Get the secret back\n    # ========================================================================\n    print(\"=\" * 70)\n    print(\"2. READ: Retrieving the secret\")\n    print(\"=\" * 70)\n    try:\n        retrieved_secret = client.get_secret(secret_name)\n        print(f\"✓ Secret retrieved successfully!\")\n        print(f\"  Name: {retrieved_secret.name}\")\n        print(f\"  Value: {retrieved_secret.value}\")\n        print(f\"  Version: {retrieved_secret.properties.version}\")\n        print()\n    except ResourceNotFoundError:\n        print(f\"✗ Secret '{secret_name}' not found\")\n        sys.exit(1)\n    except HttpResponseError as e:\n        print(f\"✗ Failed to retrieve secret: {e.message}\")\n        sys.exit(1)\n    except Exception as e:\n        print(f\"✗ Unexpected error during secret retrieval: {e}\")\n        sys.exit(1)\n\n    # ========================================================================\n    # 3. UPDATE: Set a new value for the secret\n    # ========================================================================\n    print(\"=\" * 70)\n    print(\"3. UPDATE: Updating the secret to a new value\")\n    print(\"=\" * 70)\n    try:\n        # set_secret creates a new version when the secret name already exists\n        updated_secret = client.set_secret(secret_name, \"updated-value\")\n        print(f\"✓ Secret updated successfully!\")\n        print(f\"  Name: {updated_secret.name}\")\n        print(f\"  Value: {updated_secret.value}\")\n        print(f\"  New Version: {updated_secret.properties.version}\")\n        print()\n    except HttpResponseError as e:\n        print(f\"✗ Failed to update secret: {e.message}\")\n        sys.exit(1)\n    except Exception as e:\n        print(f\"✗ Unexpected error during secret update: {e}\")\n        sys.exit(1)\n\n    # ========================================================================\n    # 4. DELETE: Delete and purge the secret\n    # ========================================================================\n    print(\"=\" * 70)\n    print(\"4. DELETE: Deleting and purging the secret\")\n    print(\"=\" * 70)\n    try:\n        # Delete the secret (soft-delete)\n        print(f\"Deleting secret '{secret_name}'...\")\n        delete_poller = client.begin_delete_secret(secret_name)\n        deleted_secret = delete_poller.result()\n        print(f\"✓ Secret soft-deleted successfully!\")\n        print(f\"  Name: {deleted_secret.name}\")\n        print(f\"  Deleted Date: {deleted_secret.deleted_date}\")\n        print(f\"  Scheduled Purge Date: {deleted_secret.scheduled_purge_date}\")\n        print()\n\n        # Purge the secret (permanent deletion)\n        print(f\"Purging secret '{secret_name}' permanently...\")\n        client.purge_deleted_secret(secret_name)\n        print(f\"✓ Secret purged successfully!\")\n        print(f\"  The secret '{secret_name}' has been permanently deleted.\")\n        print()\n    except HttpResponseError as e:\n        print(f\"✗ Failed to delete/purge secret: {e.message}\")\n        sys.exit(1)\n    except Exception as e:\n        print(f\"✗ Unexpected error during secret deletion/purge: {e}\")\n        sys.exit(1)\n\n    print(\"=\" * 70)\n    print(\"All CRUD operations completed successfully!\")\n    print(\"=\" * 70)\n\n\nif __name__ == \"__main__\":\n    main()\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-2548782539/keyvault_crud_demo.py"}
```

### 🔧 5. create

**Input:**

```json
{"file_text":"# Azure Key Vault Secrets CRUD Demo Requirements\n# Install with: pip install -r requirements.txt\n\n# Azure Key Vault Secrets client library\nazure-keyvault-secrets\u003e=4.10.0\n\n# Azure Identity library for authentication\nazure-identity\u003e=1.12.0\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-2548782539/requirements.txt"}
```

### 🔧 6. create

**Input:**

```json
{"file_text":"# Azure Key Vault Secrets CRUD Operations Demo\n\nThis Python script demonstrates all four CRUD operations on Azure Key Vault secrets using the official Azure SDK for Python.\n\n## Operations Demonstrated\n\n1. **CREATE**: Set a new secret called \"my-secret\" with value \"my-secret-value\"\n2. **READ**: Retrieve the secret and print its value\n3. **UPDATE**: Update the secret to a new value \"updated-value\"\n4. **DELETE**: Delete the secret and purge it (for soft-delete enabled vaults)\n\n## Prerequisites\n\n- Python 3.9 or later\n- An Azure Key Vault instance with soft-delete enabled\n- Azure credentials configured for authentication\n- Required Key Vault permissions:\n  - `secrets/set`\n  - `secrets/get`\n  - `secrets/delete`\n  - `secrets/purge`\n\n## Installation\n\nInstall the required packages:\n\n```bash\npip install -r requirements.txt\n```\n\nOr install directly:\n\n```bash\npip install azure-keyvault-secrets azure-identity\n```\n\n## Configuration\n\nSet the `VAULT_URL` environment variable to your Azure Key Vault URL:\n\n```bash\nexport VAULT_URL='https://your-key-vault-name.vault.azure.net/'\n```\n\n## Authentication\n\nThe script uses `DefaultAzureCredential` which automatically tries multiple authentication methods in order:\n\n1. Environment variables\n2. Managed Identity\n3. Azure CLI credentials\n4. Azure PowerShell credentials\n5. Interactive browser authentication\n\nFor local development, the easiest method is to authenticate using Azure CLI:\n\n```bash\naz login\n```\n\n## Usage\n\nRun the script:\n\n```bash\npython keyvault_crud_demo.py\n```\n\n## Expected Output\n\n```\nConnected to Key Vault: https://your-key-vault-name.vault.azure.net/\n\n======================================================================\n1. CREATE: Setting a new secret\n======================================================================\n✓ Secret created successfully!\n  Name: my-secret\n  Value: my-secret-value\n  Version: abc123...\n\n======================================================================\n2. READ: Retrieving the secret\n======================================================================\n✓ Secret retrieved successfully!\n  Name: my-secret\n  Value: my-secret-value\n  Version: abc123...\n\n======================================================================\n3. UPDATE: Updating the secret to a new value\n======================================================================\n✓ Secret updated successfully!\n  Name: my-secret\n  Value: updated-value\n  New Version: def456...\n\n======================================================================\n4. DELETE: Deleting and purging the secret\n======================================================================\nDeleting secret 'my-secret'...\n✓ Secret soft-deleted successfully!\n  Name: my-secret\n  Deleted Date: 2026-03-21 19:06:46\n  Scheduled Purge Date: 2026-06-19 19:06:46\n\nPurging secret 'my-secret' permanently...\n✓ Secret purged successfully!\n  The secret 'my-secret' has been permanently deleted.\n\n======================================================================\nAll CRUD operations completed successfully!\n======================================================================\n```\n\n## Error Handling\n\nThe script includes comprehensive error handling for:\n\n- Missing environment variables\n- Authentication failures\n- Permission issues\n- Resource not found errors\n- HTTP response errors\n\n## Notes\n\n- The `set_secret()` method creates a new secret if it doesn't exist, or creates a new version if it does exist\n- For vaults without soft-delete enabled, the `begin_delete_secret()` operation is permanent and `purge_deleted_secret()` is not needed\n- Each secret update creates a new version while preserving the history\n- Purging a secret is permanent and cannot be recovered\n\n## References\n\n- [Azure Key Vault Secrets Documentation](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)\n- [SecretClient API Reference](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient)\n- [DefaultAzureCredential Documentation](https://learn.microsoft.com/en-us/python/api/azure-identity/azure.identity.defaultazurecredential)\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-2548782539/README.md"}
```

### 🔧 7. bash

**Input:**

```json
{"command":"ls -la /tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-2548782539/","description":"List all created files"}
```

## Generated Files

- `README.md`
- `keyvault_crud_demo.py`
- `requirements.txt`

## Copilot Response

I'll help you create a Python script for Azure Key Vault CRUD operations using the official Azure SDK documentation.

Let me fetch more information about purging deleted secrets:

Perfect! Now I have all the information needed from the Azure SDK documentation. Let me create a complete Python script that demonstrates all CRUD operations:

Now let me create a requirements.txt file for the pip packages:

Perfect! Let me also create a README to explain how to use the script:

Perfect! Let me view the created files to confirm everything is correct:

## Summary

I've created a complete Azure Key Vault CRUD operations demo based on the official Azure SDK for Python documentation. Here's what was delivered:

### Files Created:

1. **`keyvault_crud_demo.py`** - Main script demonstrating all four CRUD operations:
   - **CREATE**: Uses `set_secret()` to create "my-secret" with value "my-secret-value"
   - **READ**: Uses `get_secret()` to retrieve and print the secret value
   - **UPDATE**: Uses `set_secret()` again to update to "updated-value" (creates new version)
   - **DELETE**: Uses `begin_delete_secret()` for soft-delete, then `purge_deleted_secret()` for permanent deletion

2. **`requirements.txt`** - Lists required packages:
   - `azure-keyvault-secrets>=4.10.0`
   - `azure-identity>=1.12.0`

3. **`README.md`** - Complete documentation with usage instructions

### Key Features:

✅ **Based on official Azure SDK documentation** - All code examples taken directly from Microsoft's documentation  
✅ **DefaultAzureCredential** - Uses recommended authentication method  
✅ **Comprehensive error handling** - Catches `ResourceNotFoundError`, `HttpResponseError`, and general exceptions  
✅ **Clear output** - Formatted sections showing each operation's results  
✅ **Production-ready** - Environment variable configuration, proper error messages, exit codes  

The script is ready to run once you set the `VAULT_URL` environment variable and have Azure credentials configured!

## Verification

**Result:** ✅ PASS

**Summary:** Complete implementation with all CRUD operations, proper authentication, comprehensive error handling, and correct soft-delete workflow

**Reasoning:**

The generated code fully meets all requirements from the original prompt. It correctly implements all four CRUD operations: (1) CREATE - uses set_secret() to create 'my-secret' with value 'my-secret-value', (2) READ - uses get_secret() to retrieve and print the secret value, (3) UPDATE - uses set_secret() again to update to 'updated-value', and (4) DELETE - uses begin_delete_secret() followed by purge_deleted_secret() for complete deletion. The code properly uses DefaultAzureCredential for authentication and creates a SecretClient with vault URL. Error handling is comprehensive, including ResourceNotFoundError, HttpResponseError, and generic exceptions for each operation. The soft-delete workflow is correctly implemented by calling .result() on the delete poller to wait for completion before purging. The requirements.txt includes both required packages (azure-keyvault-secrets and azure-identity). The code is syntactically valid Python with proper imports, structure, and would execute successfully given valid Azure credentials and vault URL. The README provides excellent documentation covering prerequisites, installation, authentication, and usage.

---

[← Back to Summary](../../../../../../summary.md)
