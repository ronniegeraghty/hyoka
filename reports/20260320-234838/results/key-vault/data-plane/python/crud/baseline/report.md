# Evaluation Report: key-vault-dp-python-crud

**Config:** baseline | **Result:** ✅ PASSED | **Duration:** 100.6s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-crud` |
| Config | baseline |
| Result | ✅ PASSED |
| Duration | 100.6s |
| Timestamp | 2026-03-21T06:48:38Z |
| Files Generated | 4 |
| Event Count | 94 |

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

The user wants me to write a Python script that performs CRUD operations on Azure Key Vault secrets using the Azure SDK for Python. They specifically mention:
1. Create a secret
2. Read the secret
3. Update the secret
4. Delete and purge the secret

They want me to use DefaultAzureCredential for authentication and include proper error handling and show required pip packages.

The user emphasizes "Using only the Azure SDK for Python documentation" which means I should refer to official Azure SDK documentation to ensure accuracy.

This is a straightforward task - I need to:
1. Look up the Azure SDK for Python Key Vault documentation
2. Write a script that demonstrates CRUD operations
3. Include error handling
4. List the required pip packages

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

### 🔧 4. report_intent

**Input:**

```json
{"intent":"Creating CRUD script"}
```

### 🔧 5. create

**Input:**

```json
{"file_text":"#!/usr/bin/env python3\n\"\"\"\nAzure Key Vault Secrets CRUD Operations Demo\n\nThis script demonstrates all four CRUD operations on Azure Key Vault secrets:\n1. Create - Set a new secret\n2. Read - Retrieve the secret value\n3. Update - Modify the secret value\n4. Delete - Delete and purge the secret\n\nRequired pip packages:\n    pip install azure-keyvault-secrets azure-identity\n\nPrerequisites:\n    - Azure subscription\n    - Azure Key Vault with soft-delete enabled\n    - Proper authentication configured (Azure CLI login, managed identity, etc.)\n    - Key Vault access policies granting: secrets/set, secrets/get, secrets/delete, secrets/purge\n    - Environment variable VAULT_URL set to your Key Vault URL\n\"\"\"\n\nimport os\nimport sys\nfrom azure.identity import DefaultAzureCredential\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.core.exceptions import (\n    ResourceNotFoundError,\n    HttpResponseError,\n    ClientAuthenticationError\n)\n\n\ndef main():\n    \"\"\"Perform CRUD operations on Azure Key Vault secrets.\"\"\"\n    \n    # Get vault URL from environment variable\n    vault_url = os.environ.get(\"VAULT_URL\")\n    if not vault_url:\n        print(\"Error: VAULT_URL environment variable not set\")\n        print(\"Example: export VAULT_URL='https://my-key-vault.vault.azure.net/'\")\n        sys.exit(1)\n    \n    print(f\"Connecting to Key Vault: {vault_url}\")\n    \n    try:\n        # Authenticate using DefaultAzureCredential\n        credential = DefaultAzureCredential()\n        \n        # Create the SecretClient\n        client = SecretClient(vault_url=vault_url, credential=credential)\n        \n        secret_name = \"my-secret\"\n        \n        # 1. CREATE - Set a new secret\n        print(\"\\n\" + \"=\"*60)\n        print(\"1. CREATE - Setting a new secret\")\n        print(\"=\"*60)\n        try:\n            secret = client.set_secret(secret_name, \"my-secret-value\")\n            print(f\"✓ Secret created successfully\")\n            print(f\"  Name: {secret.name}\")\n            print(f\"  Value: {secret.value}\")\n            print(f\"  Version: {secret.properties.version}\")\n            print(f\"  Created: {secret.properties.created_on}\")\n        except HttpResponseError as e:\n            print(f\"✗ Failed to create secret: {e.message}\")\n            raise\n        \n        # 2. READ - Retrieve the secret\n        print(\"\\n\" + \"=\"*60)\n        print(\"2. READ - Retrieving the secret\")\n        print(\"=\"*60)\n        try:\n            retrieved_secret = client.get_secret(secret_name)\n            print(f\"✓ Secret retrieved successfully\")\n            print(f\"  Name: {retrieved_secret.name}\")\n            print(f\"  Value: {retrieved_secret.value}\")\n            print(f\"  Version: {retrieved_secret.properties.version}\")\n        except ResourceNotFoundError:\n            print(f\"✗ Secret '{secret_name}' not found\")\n            raise\n        except HttpResponseError as e:\n            print(f\"✗ Failed to retrieve secret: {e.message}\")\n            raise\n        \n        # 3. UPDATE - Set a new value (creates a new version)\n        print(\"\\n\" + \"=\"*60)\n        print(\"3. UPDATE - Updating the secret value\")\n        print(\"=\"*60)\n        try:\n            updated_secret = client.set_secret(secret_name, \"updated-value\")\n            print(f\"✓ Secret updated successfully\")\n            print(f\"  Name: {updated_secret.name}\")\n            print(f\"  New Value: {updated_secret.value}\")\n            print(f\"  New Version: {updated_secret.properties.version}\")\n            print(f\"  Updated: {updated_secret.properties.updated_on}\")\n        except HttpResponseError as e:\n            print(f\"✗ Failed to update secret: {e.message}\")\n            raise\n        \n        # 4. DELETE - Delete and purge the secret\n        print(\"\\n\" + \"=\"*60)\n        print(\"4. DELETE - Deleting and purging the secret\")\n        print(\"=\"*60)\n        try:\n            # Begin delete operation (returns a poller)\n            print(f\"  Initiating deletion of '{secret_name}'...\")\n            delete_operation = client.begin_delete_secret(secret_name)\n            \n            # Wait for deletion to complete\n            deleted_secret = delete_operation.result()\n            print(f\"✓ Secret deleted successfully\")\n            print(f\"  Name: {deleted_secret.name}\")\n            print(f\"  Deleted Date: {deleted_secret.deleted_date}\")\n            print(f\"  Scheduled Purge Date: {deleted_secret.scheduled_purge_date}\")\n            print(f\"  Recovery ID: {deleted_secret.recovery_id}\")\n            \n            # Purge the deleted secret (permanent deletion)\n            print(f\"\\n  Purging deleted secret '{secret_name}'...\")\n            client.purge_deleted_secret(secret_name)\n            print(f\"✓ Secret purged successfully (permanently deleted)\")\n            \n        except HttpResponseError as e:\n            print(f\"✗ Failed to delete/purge secret: {e.message}\")\n            raise\n        \n        print(\"\\n\" + \"=\"*60)\n        print(\"All CRUD operations completed successfully!\")\n        print(\"=\"*60)\n        \n    except ClientAuthenticationError as e:\n        print(f\"\\n✗ Authentication failed: {e.message}\")\n        print(\"  Make sure you're logged in via Azure CLI or have proper credentials configured\")\n        sys.exit(1)\n    except HttpResponseError as e:\n        print(f\"\\n✗ HTTP error occurred: {e.message}\")\n        print(f\"  Status code: {e.status_code}\")\n        sys.exit(1)\n    except Exception as e:\n        print(f\"\\n✗ Unexpected error: {str(e)}\")\n        sys.exit(1)\n    finally:\n        # Close the credential to clean up resources\n        credential.close()\n\n\nif __name__ == \"__main__\":\n    main()\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-1034887953/keyvault_crud.py"}
```

### 🔧 6. create

**Input:**

```json
{"file_text":"azure-keyvault-secrets\u003e=4.7.0\nazure-identity\u003e=1.12.0\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-1034887953/requirements.txt"}
```

### 🔧 7. create

**Input:**

```json
{"file_text":"# Azure Key Vault Secrets CRUD Operations\n\nThis script demonstrates all four CRUD operations (Create, Read, Update, Delete) on Azure Key Vault secrets using the Azure SDK for Python.\n\n## Prerequisites\n\n1. **Python 3.9 or later**\n2. **Azure subscription** with an Azure Key Vault\n3. **Key Vault with soft-delete enabled** (default for new vaults)\n4. **Proper permissions**: Your identity needs the following Key Vault access policies:\n   - secrets/set\n   - secrets/get\n   - secrets/delete\n   - secrets/purge\n\n## Installation\n\nInstall the required packages:\n\n```bash\npip install -r requirements.txt\n```\n\nOr install directly:\n\n```bash\npip install azure-keyvault-secrets azure-identity\n```\n\n## Authentication Setup\n\nThe script uses `DefaultAzureCredential`, which attempts authentication through multiple methods in order:\n\n1. **Azure CLI** (easiest for local development):\n   ```bash\n   az login\n   ```\n\n2. **Managed Identity** (for Azure-hosted resources)\n3. **Environment variables** (for service principals)\n4. **Visual Studio Code** authentication\n5. **And more...**\n\nFor production, use Managed Identity. For local development, use Azure CLI login.\n\n## Configuration\n\nSet the `VAULT_URL` environment variable to your Key Vault URL:\n\n```bash\nexport VAULT_URL=\"https://your-key-vault-name.vault.azure.net/\"\n```\n\n## Usage\n\nRun the script:\n\n```bash\npython keyvault_crud.py\n```\n\n## What the Script Does\n\n1. **CREATE**: Sets a new secret named \"my-secret\" with value \"my-secret-value\"\n2. **READ**: Retrieves the secret and displays its value\n3. **UPDATE**: Updates the secret to a new value \"updated-value\" (creates a new version)\n4. **DELETE**: Deletes the secret and purges it permanently (soft-delete enabled vault)\n\n## Output Example\n\n```\nConnecting to Key Vault: https://your-vault.vault.azure.net/\n\n============================================================\n1. CREATE - Setting a new secret\n============================================================\n✓ Secret created successfully\n  Name: my-secret\n  Value: my-secret-value\n  Version: abc123...\n  Created: 2026-03-21 06:48:00\n\n============================================================\n2. READ - Retrieving the secret\n============================================================\n✓ Secret retrieved successfully\n  Name: my-secret\n  Value: my-secret-value\n  Version: abc123...\n\n============================================================\n3. UPDATE - Updating the secret value\n============================================================\n✓ Secret updated successfully\n  Name: my-secret\n  New Value: updated-value\n  New Version: def456...\n  Updated: 2026-03-21 06:48:01\n\n============================================================\n4. DELETE - Deleting and purging the secret\n============================================================\n  Initiating deletion of 'my-secret'...\n✓ Secret deleted successfully\n  Name: my-secret\n  Deleted Date: 2026-03-21 06:48:02\n  Scheduled Purge Date: 2026-06-19 06:48:02\n  Recovery ID: ...\n\n  Purging deleted secret 'my-secret'...\n✓ Secret purged successfully (permanently deleted)\n\n============================================================\nAll CRUD operations completed successfully!\n============================================================\n```\n\n## Error Handling\n\nThe script includes comprehensive error handling for:\n- Missing environment variables\n- Authentication failures\n- Resource not found errors\n- HTTP errors\n- General exceptions\n\n## References\n\n- [Azure Key Vault Secrets Python SDK Documentation](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)\n- [DefaultAzureCredential Documentation](https://learn.microsoft.com/en-us/python/api/azure-identity/azure.identity.defaultazurecredential)\n- [Azure Key Vault Documentation](https://learn.microsoft.com/en-us/azure/key-vault/)\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-1034887953/README.md"}
```

### 🔧 8. bash

**Input:**

```json
{"command":"cd /tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-1034887953 \u0026\u0026 ls -la","description":"List created files"}
```

### 🔧 9. bash

**Input:**

```json
{"command":"cd /tmp/azsdk-prompt-eval-key-vault-dp-python-crud-baseline-1034887953 \u0026\u0026 python3 -m py_compile keyvault_crud.py \u0026\u0026 echo \"✓ Syntax is valid\"","description":"Verify Python syntax"}
```

## Generated Files

- `README.md`
- `__pycache__/keyvault_crud.cpython-312.pyc`
- `keyvault_crud.py`
- `requirements.txt`

## Copilot Response

I'll help you create a Python script that performs CRUD operations on Azure Key Vault secrets using the Azure SDK for Python. Let me first fetch the official documentation to ensure accuracy.

Great! Now let me fetch additional information about purging deleted secrets:

Perfect! Now I have all the information needed. Let me create a comprehensive Python script demonstrating all four CRUD operations:

Now let me create a requirements.txt file and a README with setup instructions:

Perfect! Let me verify the script by viewing it to ensure everything is correct:

Perfect! Let me verify the script syntax is correct:

Excellent! I've successfully created a comprehensive Python script that demonstrates all four CRUD operations on Azure Key Vault secrets. Here's what I've delivered:

## 📋 Created Files

1. **`keyvault_crud.py`** - Main script demonstrating:
   - ✅ **CREATE**: Set a new secret "my-secret" with value "my-secret-value"
   - ✅ **READ**: Retrieve and display the secret value
   - ✅ **UPDATE**: Update secret to "updated-value" (creates new version)
   - ✅ **DELETE**: Delete the secret and purge it permanently

2. **`requirements.txt`** - Required packages:
   - `azure-keyvault-secrets>=4.7.0`
   - `azure-identity>=1.12.0`

3. **`README.md`** - Complete setup and usage instructions

## 🎯 Key Features

- **DefaultAzureCredential** authentication (works with Azure CLI, Managed Identity, etc.)
- **Comprehensive error handling** for all common scenarios:
  - Missing environment variables
  - Authentication failures
  - Resource not found errors
  - HTTP errors
- **Clear console output** with visual indicators (✓/✗)
- **Detailed information** printed for each operation
- **Production-ready** with proper exception handling and resource cleanup

## 🚀 Usage

```bash
# Install dependencies
pip install -r requirements.txt

# Set your Key Vault URL
export VAULT_URL="https://your-key-vault-name.vault.azure.net/"

# Run the script
python keyvault_crud.py
```

The script follows Azure SDK for Python best practices and is based entirely on the official Microsoft documentation!

## Verification

**Result:** ✅ PASS

**Summary:** Complete implementation of all four CRUD operations with correct SDK usage, authentication, error handling, and soft-delete management.

**Reasoning:**

The generated code fully meets all requirements. It correctly implements all four CRUD operations: (1) Creates a secret named 'my-secret' with value 'my-secret-value' using set_secret(), (2) Reads the secret back using get_secret() and prints its value, (3) Updates the secret to 'updated-value' using set_secret() again, and (4) Deletes and purges the secret using begin_delete_secret() followed by purge_deleted_secret(). The code properly uses DefaultAzureCredential for authentication and creates a SecretClient with vault URL. It includes comprehensive error handling for ResourceNotFoundError, HttpResponseError, and ClientAuthenticationError. The soft-delete handling is correct, calling delete_operation.result() to wait for deletion completion before purging. The requirements.txt file correctly specifies azure-keyvault-secrets and azure-identity packages. The code is syntactically valid Python and includes clear documentation in both the script and README.md. All expected SDK methods are used correctly.

---

[← Back to Summary](../../../../../../summary.md)
