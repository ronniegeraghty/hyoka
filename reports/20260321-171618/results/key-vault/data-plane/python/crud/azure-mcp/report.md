# Evaluation Report: key-vault-dp-python-crud

**Config:** azure-mcp | **Result:** ✅ PASSED | **Duration:** 56.4s

## Overview

| Field | Value |
|-------|-------|
| Prompt ID | `key-vault-dp-python-crud` |
| Config | azure-mcp |
| Result | ✅ PASSED |
| Duration | 56.4s |
| Timestamp | 2026-03-22T00:17:28Z |
| Files Generated | 3 |
| Event Count | 41 |

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

The user wants me to create a Python script that demonstrates CRUD operations on Azure Key Vault secrets using the Azure SDK for Python. They want:
1. Create a secret
2. Read the secret
3. Update the secret
4. Delete and purge the secret

They want me to use only the Azure SDK documentation, use DefaultAzureCredential, include error handling, and show required pip packages.

This is a straightforward task - I need to write a Python script. Let me start by creating this script with proper documentation and error handling.

I should:
1. Report my intent
2. Create the Python script with all CRUD operations
3. Include comments about required packages
4. Add proper error handling

## Tool Calls

### 🔧 1. report_intent

**Input:**

```json
{"intent":"Creating Key Vault CRUD script"}
```

### 🔧 2. create

**Input:**

```json
{"file_text":"\"\"\"\nAzure Key Vault Secrets CRUD Operations Demo\n\nRequired packages:\n    pip install azure-keyvault-secrets azure-identity\n\"\"\"\n\nfrom azure.keyvault.secrets import SecretClient\nfrom azure.identity import DefaultAzureCredential\nfrom azure.core.exceptions import ResourceNotFoundError, HttpResponseError\nimport sys\n\n\ndef main():\n    # Replace with your Key Vault URL\n    # Format: https://\u003cyour-key-vault-name\u003e.vault.azure.net/\n    key_vault_url = \"https://\u003cyour-key-vault-name\u003e.vault.azure.net/\"\n    \n    # Check if Key Vault URL is provided as command line argument\n    if len(sys.argv) \u003e 1:\n        key_vault_url = sys.argv[1]\n    \n    if \"\u003cyour-key-vault-name\u003e\" in key_vault_url:\n        print(\"Error: Please provide a valid Key Vault URL\")\n        print(\"Usage: python keyvault_crud.py https://\u003cyour-key-vault-name\u003e.vault.azure.net/\")\n        sys.exit(1)\n    \n    try:\n        # Authenticate using DefaultAzureCredential\n        credential = DefaultAzureCredential()\n        \n        # Create SecretClient\n        client = SecretClient(vault_url=key_vault_url, credential=credential)\n        \n        print(f\"Connected to Key Vault: {key_vault_url}\\n\")\n        \n        # 1. CREATE - Set a new secret\n        print(\"=\" * 60)\n        print(\"1. CREATE - Setting secret 'my-secret'\")\n        print(\"=\" * 60)\n        secret_name = \"my-secret\"\n        secret_value = \"my-secret-value\"\n        \n        secret = client.set_secret(secret_name, secret_value)\n        print(f\"✓ Secret created successfully\")\n        print(f\"  Name: {secret.name}\")\n        print(f\"  Value: {secret.value}\")\n        print(f\"  Version: {secret.properties.version}\")\n        print()\n        \n        # 2. READ - Retrieve the secret\n        print(\"=\" * 60)\n        print(\"2. READ - Getting secret 'my-secret'\")\n        print(\"=\" * 60)\n        \n        retrieved_secret = client.get_secret(secret_name)\n        print(f\"✓ Secret retrieved successfully\")\n        print(f\"  Name: {retrieved_secret.name}\")\n        print(f\"  Value: {retrieved_secret.value}\")\n        print(f\"  Version: {retrieved_secret.properties.version}\")\n        print(f\"  Created: {retrieved_secret.properties.created_on}\")\n        print()\n        \n        # 3. UPDATE - Update the secret value\n        print(\"=\" * 60)\n        print(\"3. UPDATE - Updating secret 'my-secret'\")\n        print(\"=\" * 60)\n        new_value = \"updated-value\"\n        \n        updated_secret = client.set_secret(secret_name, new_value)\n        print(f\"✓ Secret updated successfully\")\n        print(f\"  Name: {updated_secret.name}\")\n        print(f\"  Value: {updated_secret.value}\")\n        print(f\"  New Version: {updated_secret.properties.version}\")\n        print()\n        \n        # 4. DELETE - Delete and purge the secret\n        print(\"=\" * 60)\n        print(\"4. DELETE - Deleting secret 'my-secret'\")\n        print(\"=\" * 60)\n        \n        # Begin delete operation (soft delete)\n        deleted_secret = client.begin_delete_secret(secret_name).result()\n        print(f\"✓ Secret deleted (soft-deleted)\")\n        print(f\"  Name: {deleted_secret.name}\")\n        print(f\"  Deleted on: {deleted_secret.deleted_date}\")\n        print(f\"  Scheduled purge: {deleted_secret.scheduled_purge_date}\")\n        print()\n        \n        # Purge the deleted secret permanently\n        print(\"Purging secret permanently...\")\n        client.purge_deleted_secret(secret_name)\n        print(f\"✓ Secret purged successfully\")\n        print()\n        \n        print(\"=\" * 60)\n        print(\"All CRUD operations completed successfully!\")\n        print(\"=\" * 60)\n        \n    except ResourceNotFoundError as e:\n        print(f\"Error: Resource not found - {e.message}\")\n        sys.exit(1)\n    except HttpResponseError as e:\n        print(f\"Error: HTTP response error - {e.message}\")\n        print(f\"Status code: {e.status_code}\")\n        sys.exit(1)\n    except Exception as e:\n        print(f\"Error: An unexpected error occurred - {str(e)}\")\n        sys.exit(1)\n    finally:\n        # Close the credential\n        credential.close()\n\n\nif __name__ == \"__main__\":\n    main()\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-13383758/keyvault_crud.py"}
```

### 🔧 3. create

**Input:**

```json
{"file_text":"azure-keyvault-secrets\u003e=4.7.0\nazure-identity\u003e=1.15.0\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-13383758/requirements.txt"}
```

### 🔧 4. create

**Input:**

```json
{"file_text":"# Azure Key Vault Secrets CRUD Operations\n\nThis script demonstrates all four CRUD operations on Azure Key Vault secrets using the Azure SDK for Python.\n\n## Prerequisites\n\n1. An Azure Key Vault with soft-delete enabled\n2. Azure credentials configured (Azure CLI login, environment variables, or managed identity)\n3. Appropriate permissions on the Key Vault (Secret Officer role or equivalent)\n\n## Installation\n\nInstall required packages:\n\n```bash\npip install -r requirements.txt\n```\n\nOr install directly:\n\n```bash\npip install azure-keyvault-secrets azure-identity\n```\n\n## Usage\n\nRun the script with your Key Vault URL:\n\n```bash\npython keyvault_crud.py https://\u003cyour-key-vault-name\u003e.vault.azure.net/\n```\n\n## What the Script Does\n\n1. **CREATE**: Creates a new secret named \"my-secret\" with value \"my-secret-value\"\n2. **READ**: Retrieves the secret and displays its value\n3. **UPDATE**: Updates the secret to a new value \"updated-value\"\n4. **DELETE**: Soft-deletes the secret, then purges it permanently\n\n## Authentication\n\nThe script uses `DefaultAzureCredential` which attempts authentication in this order:\n1. Environment variables\n2. Managed identity\n3. Azure CLI\n4. Azure PowerShell\n5. Interactive browser\n\n## Error Handling\n\nThe script includes proper error handling for:\n- Resource not found errors\n- HTTP response errors\n- General exceptions\n\n## Notes\n\n- The script requires a Key Vault with soft-delete enabled for the purge operation\n- Purging a secret permanently removes it and cannot be undone\n- The script closes the credential properly in the finally block\n","path":"/tmp/azsdk-prompt-eval-key-vault-dp-python-crud-azure-mcp-13383758/README.md"}
```

## Generated Files

- `README.md`
- `keyvault_crud.py`
- `requirements.txt`

## Copilot Response

I'll create a Python script demonstrating all CRUD operations on Azure Key Vault secrets using the Azure SDK for Python.

Perfect! I've created a complete Python script demonstrating all CRUD operations on Azure Key Vault secrets. Here's what I've included:

## Files Created:

1. **keyvault_crud.py** - Main script with all CRUD operations:
   - ✅ **CREATE**: Sets "my-secret" with value "my-secret-value"
   - ✅ **READ**: Retrieves and displays the secret
   - ✅ **UPDATE**: Updates to "updated-value" 
   - ✅ **DELETE**: Soft-deletes then purges the secret

2. **requirements.txt** - Required packages:
   - `azure-keyvault-secrets>=4.7.0`
   - `azure-identity>=1.15.0`

3. **README.md** - Usage instructions and documentation

## Key Features:

- Uses `DefaultAzureCredential` for authentication (tries multiple methods automatically)
- Comprehensive error handling for `ResourceNotFoundError`, `HttpResponseError`, and general exceptions
- Clean output showing each operation's result
- Proper resource cleanup (closes credential in finally block)
- Accepts Key Vault URL via command line argument
- Detailed comments throughout

## Usage:

```bash
pip install -r requirements.txt
python keyvault_crud.py https://<your-key-vault-name>.vault.azure.net/
```

The script follows Azure SDK best practices and includes all error handling as requested!

## Verification

**Result:** ✅ PASS

**Summary:** Code fully implements all CRUD operations with correct SDK usage, proper authentication, comprehensive error handling, and correct soft-delete handling.

**Reasoning:**

The generated code fully meets all requirements. It correctly implements all four CRUD operations: (1) CREATE using set_secret() with 'my-secret' and 'my-secret-value', (2) READ using get_secret() to retrieve and print the value, (3) UPDATE using set_secret() again with 'updated-value', and (4) DELETE using begin_delete_secret().result() followed by purge_deleted_secret(). The code properly uses DefaultAzureCredential for authentication and creates a SecretClient with the vault URL. Error handling is comprehensive with try-except-finally blocks catching ResourceNotFoundError, HttpResponseError, and generic exceptions. The requirements.txt file correctly specifies azure-keyvault-secrets and azure-identity packages. Critically, the soft-delete scenario is handled correctly by calling .result() on the delete poller to wait for completion before purging. The code is syntactically valid Python, includes helpful documentation in README.md, proper imports, and would execute successfully given valid Azure credentials and Key Vault access.

---

[← Back to Summary](../../../../../../summary.md)
