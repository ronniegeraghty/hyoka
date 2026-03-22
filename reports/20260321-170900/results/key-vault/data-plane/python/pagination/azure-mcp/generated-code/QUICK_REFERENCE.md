# Quick Reference: Azure Key Vault Pagination

## Installation
```bash
pip install azure-keyvault-secrets azure-identity
```

## Basic Setup
```python
from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential

client = SecretClient(
    vault_url="https://vault-name.vault.azure.net/",
    credential=DefaultAzureCredential()
)
```

## Iteration Methods

### Simple (Automatic Pagination)
```python
for secret in client.list_properties_of_secrets():
    if secret.enabled:
        print(secret.name)
```

### Page-by-page (Explicit Control)
```python
for page in client.list_properties_of_secrets().by_page():
    for secret in page:
        if secret.enabled:
            print(secret.name)
```

### With Page Numbers
```python
for page_num, page in enumerate(client.list_properties_of_secrets().by_page(), start=1):
    print(f"Page {page_num}")
    for secret in page:
        print(f"  {secret.name}")
```

## Available Properties

From `SecretProperties` object:
- `name` - Secret name
- `enabled` - True/False
- `content_type` - Optional string
- `created_on` - datetime (UTC)
- `updated_on` - datetime (UTC)
- `expires_on` - Optional datetime
- `tags` - Dictionary
- `version` - Version ID
- `vault_url` - Vault URL

**Note**: Use `client.get_secret(name)` to get actual secret value

## Authentication (Local Dev)
```bash
az login
az account set --subscription "subscription-id"
```

## Grant Permissions
```bash
az keyvault set-policy \
  --name vault-name \
  --object-id $(az ad signed-in-user show --query id -o tsv) \
  --secret-permissions list
```

## Common Patterns

### Count enabled secrets
```python
count = sum(1 for s in client.list_properties_of_secrets() if s.enabled)
```

### Filter by content type
```python
certs = [s for s in client.list_properties_of_secrets() 
         if s.content_type == "application/x-pkcs12"]
```

### Get recently created
```python
from datetime import datetime, timedelta
week_ago = datetime.now() - timedelta(days=7)

recent = [s for s in client.list_properties_of_secrets()
          if s.created_on and s.created_on > week_ago]
```

## Remember

✅ DO:
- Use `by_page()` for hundreds of secrets
- Filter during iteration
- Close client or use context manager
- Check `enabled` property

❌ DON'T:
- Convert ItemPaged to list: `list(secrets)`
- Call `get_secret()` in tight loops
- Forget to check permissions
- Assume page size is consistent

## File Summary

- **list_keyvault_secrets_paginated.py** - Full featured script with page tracking
- **simple_example.py** - Minimal working example
- **README.md** - Complete documentation
- **PAGINATION_CONCEPTS.md** - Deep dive into pagination mechanics
- **requirements.txt** - Package dependencies
