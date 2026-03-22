# Azure Key Vault Secrets Pagination Demo

This script demonstrates how to handle pagination when listing secrets from an Azure Key Vault with hundreds of secrets.

## Required Packages

Install the required packages using pip:

```bash
pip install azure-keyvault-secrets azure-identity
```

Or using the requirements file:

```bash
pip install -r requirements.txt
```

## Package Details

- **azure-keyvault-secrets** (>=4.8.0): Azure Key Vault Secrets client library
- **azure-identity** (>=1.15.0): Azure authentication library for DefaultAzureCredential

## Pagination Pattern Explained

### ItemPaged Object

The `list_properties_of_secrets()` method returns an `ItemPaged[SecretProperties]` object:

```python
secret_properties = client.list_properties_of_secrets()
```

### Two Ways to Iterate

1. **Explicit Pagination with by_page()** (Recommended for large vaults):
   ```python
   for page in secret_properties.by_page():
       for secret in page:
           # Process each secret
   ```

2. **Simple Iteration** (Auto-pagination):
   ```python
   for secret in secret_properties:
       # Pagination happens automatically
   ```

## Configuration

Before running, update the `KEY_VAULT_URL` in the script:

```python
KEY_VAULT_URL = "https://<your-key-vault-name>.vault.azure.net/"
```

## Authentication

The script uses `DefaultAzureCredential`, which attempts authentication in this order:

1. Environment variables (AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET)
2. Managed Identity (when running in Azure)
3. Azure CLI credentials
4. Azure PowerShell credentials
5. Interactive browser authentication

## Required Permissions

Your Azure identity needs these Key Vault permissions:
- **List** - to list secret names
- **Get** - to retrieve secret properties

## Running the Script

```bash
python list_secrets_paginated.py
```

## Output

The script will display:
- Secrets grouped by page
- For each enabled secret:
  - Secret name
  - Content type
  - Created date
  - Enabled status
- Summary statistics (total secrets, enabled/disabled counts, total pages)
