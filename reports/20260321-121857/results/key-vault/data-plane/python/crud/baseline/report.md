# Evaluation Report: key-vault-dp-python-crud

**Config:** baseline | **Result:** ✅ PASSED | **Duration:** 99.1s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-crud` |
| Config | baseline |
| Result | ✅ PASSED |
| Duration | 99.1s |
| Timestamp | 2026-03-21T19:18:57Z |
| Files Generated | 3 |
| Event Count | 73 |

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

The user wants me to create a Python script that performs CRUD operations on Azure Key Vault secrets using the Azure SDK for Python. They specifically want me to use only the Azure SDK documentation.

Let me break down what I need to do:
1. Look up the Azure SDK for Python documentation for Key Vault
2. Write a script that:
   - Creates a secret called "my-secret" with value "my-secret-value"
   - Reads the secret back and prints its value
   - Updates the secret to a new value "updated-value"
   - Deletes the secret and purges it
3. Use DefaultAzureCredential for authentication
4. Include proper error handling
5. Show required pip packages

I should fetch the Azure SDK documentation to ensure I'm using the correct API and following best practices.

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
{"file_text":"\"\"\"\nAzure Key Vault Secrets CRUD Operations Demo\n\nThis script demonstrates all four CRUD operations on Azure Key Vault secrets:\n1. Create a new secret\n2. Read the secret back\n3. Update the secret to a new value\n4. Delete and purge the secret\n\nRequirements:\n    pip install azure-keyvault-secrets azure-identity\n\nAuthentication:\n    Uses DefaultAzureCredential which supports multiple authentication methods:\n    - Environment variables (AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET)\n    - Managed Identity (when running on Azure)\n    - Azure CLI (az login)\n    - Visual Studio Code\n    - Azure PowerShell\n    - Interactive browser\n\nPrerequisites:\n    - An Azure Key Vault with the URL set in the VAULT_URL environment variable\n    - Appropriate permissions: secrets/set, secrets/get, secrets/delete, secrets/purge\n    - Soft-delete must be enabled on the vault for purge operation\n\"\"\"\n\nimport os\nimport sys\nfrom azure.identity import DefaultAzureCredential\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.core.exceptions import (\n    ResourceNotFoundError,\n    HttpResponseError,\n    ClientAuthenticationError\n)\n\n\ndef main():\n    \"\"\"Main function demonstrating CRUD operations on Key Vault secrets.\"\"\"\n    \n    # Get the vault URL from environment variable\n    vault_url = os.environ.get(\"VAULT_URL\")\n    if not vault_url:\n        print(\"Error: VAULT_URL environment variable is not set.\")\n        print(\"Example: export VAULT_URL='https://your-vault-name.vault.azure.net/'\")\n        sys.exit(1)\n    \n    print(f\"Connecting to Key Vault: {vault_url}\\n\")\n    \n    # Initialize the credential and client\n    try:\n        credential = DefaultAzureCredential()\n        client = SecretClient(vault_url=vault_url, credential=credential)\n    except ClientAuthenticationError as e:\n        print(f\"Authentication Error: {e.message}\")\n        print(\"\\nPlease ensure you are authenticated using one of:\")\n        print(\"  - Azure CLI: az login\")\n        print(\"  - Environment variables: AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET\")\n        print(\"  - Managed Identity (when running on Azure)\")\n        sys.exit(1)\n    except Exception as e:\n        print(f\"Error initializing client: {e}\")\n        sys.exit(1)\n    \n    secret_name = \"my-secret\"\n    \n    try:\n        # ========================================\n        # 1. CREATE - Set a new secret\n        # ========================================\n        print(\"=\" * 60)\n        print(\"1. CREATE - Creating a new secret\")\n        print(\"=\" * 60)\n        \n        initial_value = \"my-secret-value\"\n        print(f\"Setting secret '{secret_name}' with value '{initial_value}'...\")\n        \n        secret = client.set_secret(secret_name, initial_value)\n        \n        print(f\"✓ Secret created successfully!\")\n        print(f\"  Name: {secret.name}\")\n        print(f\"  Value: {secret.value}\")\n        print(f\"  Version: {secret.properties.version}\")\n        print(f\"  Created: {secret.properties.created_on}\")\n        print()\n        \n        # ========================================\n        # 2. READ - Retrieve the secret\n        # ========================================\n        print(\"=\" * 60)\n        print(\"2. READ - Retrieving the secret\")\n        print(\"=\" * 60)\n        \n        print(f\"Getting secret '{secret_name}'...\")\n        \n        retrieved_secret = client.get_secret(secret_name)\n        \n        print(f\"✓ Secret retrieved successfully!\")\n        print(f\"  Name: {retrieved_secret.name}\")\n        print(f\"  Value: {retrieved_secret.value}\")\n        print(f\"  Version: {retrieved_secret.properties.version}\")\n        print(f\"  Enabled: {retrieved_secret.properties.enabled}\")\n        print()\n        \n        # ========================================\n        # 3. UPDATE - Update the secret value\n        # ========================================\n        print(\"=\" * 60)\n        print(\"3. UPDATE - Updating the secret value\")\n        print(\"=\" * 60)\n        \n        updated_value = \"updated-value\"\n        print(f\"Updating secret '{secret_name}' to new value '{updated_value}'...\")\n        \n        # set_secret creates a new version when the secret already exists\n        updated_secret = client.set_secret(secret_name, updated_value)\n        \n        print(f\"✓ Secret updated successfully!\")\n        print(f\"  Name: {updated_secret.name}\")\n        print(f\"  Value: {updated_secret.value}\")\n        print(f\"  New Version: {updated_secret.properties.version}\")\n        print(f\"  Updated: {updated_secret.properties.updated_on}\")\n        print()\n        \n        # Optional: Update secret properties (metadata) without changing the value\n        print(f\"Updating secret properties (metadata)...\")\n        \n        updated_properties = client.update_secret_properties(\n            secret_name,\n            content_type=\"text/plain\",\n            enabled=True\n        )\n        \n        print(f\"✓ Secret properties updated!\")\n        print(f\"  Content Type: {updated_properties.content_type}\")\n        print(f\"  Enabled: {updated_properties.enabled}\")\n        print()\n        \n        # ========================================\n        # 4. DELETE - Delete and purge the secret\n        # ========================================\n        print(\"=\" * 60)\n        print(\"4. DELETE - Deleting and purging the secret\")\n        print(\"=\" * 60)\n        \n        print(f\"Deleting secret '{secret_name}'...\")\n        \n        # begin_delete_secret returns a poller for long-running operations\n        delete_poller = client.begin_delete_secret(secret_name)\n        \n        # Wait for the deletion to complete\n        deleted_secret = delete_poller.result()\n        \n        print(f\"✓ Secret deleted successfully!\")\n        print(f\"  Name: {deleted_secret.name}\")\n        print(f\"  Deleted On: {deleted_secret.deleted_date}\")\n        print(f\"  Scheduled Purge Date: {deleted_secret.scheduled_purge_date}\")\n        print(f\"  Recovery ID: {deleted_secret.recovery_id}\")\n        print()\n        \n        # Purge the deleted secret (permanent deletion)\n        # This is only possible in vaults with soft-delete enabled\n        print(f\"Purging deleted secret '{secret_name}' (permanent deletion)...\")\n        \n        client.purge_deleted_secret(secret_name)\n        \n        print(f\"✓ Secret purged successfully!\")\n        print(f\"  The secret has been permanently deleted and cannot be recovered.\")\n        print()\n        \n        # ========================================\n        # Verification\n        # ========================================\n        print(\"=\" * 60)\n        print(\"Verification - Attempting to retrieve the deleted secret\")\n        print(\"=\" * 60)\n        \n        try:\n            client.get_secret(secret_name)\n            print(\"⚠ Warning: Secret still exists (unexpected)\")\n        except ResourceNotFoundError:\n            print(f\"✓ Confirmed: Secret '{secret_name}' no longer exists\")\n        print()\n        \n        print(\"=\" * 60)\n        print(\"CRUD Operations Completed Successfully!\")\n        print(\"=\" * 60)\n        \n    except HttpResponseError as e:\n        print(f\"\\n❌ HTTP Error: {e.message}\")\n        print(f\"   Status Code: {e.status_code}\")\n        print(f\"   Error Code: {e.error.code if hasattr(e, 'error') else 'N/A'}\")\n        \n        if e.status_code == 403:\n            print(\"\\n   This usually means insufficient permissions.\")\n            print(\"   Required permissions: secrets/set, secrets/get, secrets/delete, secrets/purge\")\n        \n        sys.exit(1)\n        \n    except ResourceNotFoundError as e:\n        print(f\"\\n❌ Resource Not Found: {e.message}\")\n        sys.exit(1)\n        \n    except Exception as e:\n        print(f\"\\n❌ Unexpected Error: {type(e).__name__}: {e}\")\n        sys.exit(1)\n        \n    finally:\n        # Clean up - close the credential\n        credential.close()\n\n\nif __name__ == \"__main__\":\n    main()\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-432507038/key_vault_crud.py"}
```

### 🔧 5. create

**Input:**

```json
{"file_text":"azure-keyvault-secrets\nazure-identity\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-432507038/requirements.txt"}
```

### 🔧 6. create

**Input:**

```json
{"file_text":"# Azure Key Vault Secrets CRUD Operations\n\nThis script demonstrates all four CRUD operations on Azure Key Vault secrets using the Azure SDK for Python.\n\n## Operations Demonstrated\n\n1. **CREATE** - Create a new secret called \"my-secret\" with value \"my-secret-value\"\n2. **READ** - Read the secret back and print its value\n3. **UPDATE** - Update the secret to a new value \"updated-value\"\n4. **DELETE** - Delete the secret and purge it (for soft-delete enabled vaults)\n\n## Installation\n\nInstall the required packages using pip:\n\n```bash\npip install -r requirements.txt\n```\n\nOr install packages individually:\n\n```bash\npip install azure-keyvault-secrets azure-identity\n```\n\n## Prerequisites\n\n1. **Azure Key Vault**: You need an existing Azure Key Vault with:\n   - Soft-delete enabled (required for purge operation)\n   - The vault URL (e.g., `https://your-vault-name.vault.azure.net/`)\n\n2. **Authentication**: Configure one of the following authentication methods:\n   \n   - **Azure CLI** (easiest for local development):\n     ```bash\n     az login\n     ```\n   \n   - **Environment Variables** (for service principal):\n     ```bash\n     export AZURE_CLIENT_ID=\"your-client-id\"\n     export AZURE_TENANT_ID=\"your-tenant-id\"\n     export AZURE_CLIENT_SECRET=\"your-client-secret\"\n     ```\n   \n   - **Managed Identity** (when running on Azure resources)\n\n3. **Permissions**: Your identity needs the following Key Vault permissions:\n   - `secrets/set` - To create and update secrets\n   - `secrets/get` - To read secrets\n   - `secrets/delete` - To delete secrets\n   - `secrets/purge` - To permanently delete secrets\n\n## Usage\n\nSet the VAULT_URL environment variable and run the script:\n\n```bash\nexport VAULT_URL=\"https://your-vault-name.vault.azure.net/\"\npython key_vault_crud.py\n```\n\n## Expected Output\n\nThe script will:\n\n1. Connect to your Key Vault\n2. Create a secret named \"my-secret\" with value \"my-secret-value\"\n3. Retrieve and display the secret\n4. Update the secret to value \"updated-value\"\n5. Update secret metadata (content type)\n6. Delete the secret (soft-delete)\n7. Purge the secret (permanent deletion)\n8. Verify the secret no longer exists\n\n## Error Handling\n\nThe script includes comprehensive error handling for:\n\n- Missing VAULT_URL environment variable\n- Authentication failures\n- Insufficient permissions (403 errors)\n- Resource not found errors\n- General HTTP errors\n\n## Python Version\n\nRequires Python 3.9 or later.\n\n## Documentation\n\nBased on official Azure SDK for Python documentation:\n- [Azure Key Vault Secrets Client Library](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)\n- [SecretClient API Reference](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient)\n- [DefaultAzureCredential](https://learn.microsoft.com/en-us/python/api/azure-identity/azure.identity.defaultazurecredential)\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-432507038/README.md"}
```

## Generated Files

- `README.md`
- `key_vault_crud.py`
- `requirements.txt`

## Copilot Response

I'll help you create a Python script for CRUD operations on Azure Key Vault secrets using the official Azure SDK documentation.

Let me fetch additional information about purging deleted secrets:

Perfect! Now I have all the information I need. Let me create a comprehensive Python script that demonstrates all four CRUD operations on Azure Key Vault secrets:

Now let me create a requirements.txt file with the necessary dependencies:

Now let me create a README with usage instructions:

Perfect! I've created a complete Azure Key Vault CRUD operations script based on the official Azure SDK for Python documentation. Here's what I've delivered:

## 📦 Files Created

1. **`key_vault_crud.py`** - Main script demonstrating all CRUD operations
2. **`requirements.txt`** - Required pip packages
3. **`README.md`** - Usage instructions and documentation

## ✨ Key Features

The script performs all four CRUD operations:

1. **CREATE** - Creates secret "my-secret" with value "my-secret-value" using `set_secret()`
2. **READ** - Retrieves the secret using `get_secret()` and prints its value
3. **UPDATE** - Updates to "updated-value" using `set_secret()` (creates new version) and updates metadata with `update_secret_properties()`
4. **DELETE** - Deletes with `begin_delete_secret()` and purges with `purge_deleted_secret()`

## 🔐 Authentication

Uses `DefaultAzureCredential` which supports:
- Azure CLI (`az login`)
- Environment variables (service principal)
- Managed Identity
- Visual Studio Code
- Azure PowerShell

## 🛡️ Error Handling

Includes comprehensive exception handling for:
- Missing environment variables
- Authentication failures
- Permission errors (403)
- Resource not found errors
- General HTTP errors

## 📋 Required Packages

```
azure-keyvault-secrets
azure-identity
```

## 🚀 Usage

```bash
pip install -r requirements.txt
export VAULT_URL="https://your-vault-name.vault.azure.net/"
python key_vault_crud.py
```

All code is based directly on the official Azure SDK for Python documentation with proper error handling and clear comments explaining each operation.

## Verification

**Result:** ✅ PASS

**Summary:** Code correctly implements all four CRUD operations with proper authentication, error handling, and soft-delete management as specified.

**Reasoning:**

The generated code fully meets all requirements from the original prompt. It correctly implements all four CRUD operations: (1) CREATE uses set_secret() to create 'my-secret' with value 'my-secret-value', (2) READ uses get_secret() to retrieve and print the secret value, (3) UPDATE uses set_secret() again with 'updated-value', (4) DELETE uses begin_delete_secret() followed by purge_deleted_secret(). The code properly uses DefaultAzureCredential for authentication and creates a SecretClient with vault URL and credential. It handles soft-delete correctly by calling delete_poller.result() to wait for deletion to complete before purging. Error handling is comprehensive, including ResourceNotFoundError, HttpResponseError, and ClientAuthenticationError with helpful messages. The requirements.txt file correctly lists both azure-keyvault-secrets and azure-identity packages. The code is syntactically valid Python with clear structure, proper exception handling, and verification steps. It exceeds expectations with detailed documentation, helpful output formatting, and additional features like update_secret_properties().

---

[← Back to Summary](../../../../../../summary.md)
