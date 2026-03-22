# Azure Key Vault Pagination - Quick Reference

## Required Packages

```bash
pip install azure-keyvault-secrets azure-identity
```

## Key Imports

```python
from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient
```

## Initialize Client

```python
credential = DefaultAzureCredential()
client = SecretClient(vault_url="https://your-vault.vault.azure.net/", credential=credential)
```

## Pagination Patterns

### Pattern 1: Simple Iteration (Recommended)

ItemPaged handles all pagination automatically:

```python
secret_properties = client.list_properties_of_secrets()

for secret in secret_properties:
    print(f"{secret.name}: {secret.created_on}")
```

### Pattern 2: Page-by-Page Processing

Process secrets in explicit pages:

```python
secret_properties = client.list_properties_of_secrets()
pages = secret_properties.by_page()

for page in pages:
    # Each page is an iterator
    for secret in page:
        print(secret.name)
```

### Pattern 3: Continuation Tokens

Resume pagination from a saved state:

```python
# Initial request
pages = client.list_properties_of_secrets().by_page()
first_page = next(pages)

# Save token
token = pages.continuation_token

# Resume later
resumed = client.list_properties_of_secrets().by_page(continuation_token=token)
next_page = next(resumed)
```

## Filtering Enabled Secrets

```python
secret_properties = client.list_properties_of_secrets()

for secret in secret_properties:
    if secret.enabled:
        print(f"Name: {secret.name}")
        print(f"Content Type: {secret.content_type}")
        print(f"Created: {secret.created_on}")
```

## SecretProperties Attributes

Properties returned by `list_properties_of_secrets()`:

| Attribute | Type | Description |
|-----------|------|-------------|
| `name` | str | Secret name |
| `enabled` | bool | Whether secret is enabled |
| `content_type` | str | Content type (optional) |
| `created_on` | datetime | Creation timestamp (UTC) |
| `updated_on` | datetime | Last update timestamp (UTC) |
| `version` | str | Secret version |
| `tags` | dict | Custom metadata |
| `expires_on` | datetime | Expiration date (optional) |
| `not_before` | datetime | Not valid before date (optional) |

**Important**: Values are NOT included. Use `client.get_secret(name)` to retrieve values.

## Authentication Methods

DefaultAzureCredential tries these in order:

1. Environment variables (AZURE_CLIENT_ID, AZURE_CLIENT_SECRET, AZURE_TENANT_ID)
2. Managed Identity (on Azure VMs, App Services, etc.)
3. Azure CLI (`az login`)
4. Visual Studio Code
5. Azure PowerShell

## Required Permissions

- **secrets/list**: Required to list secret properties

## Common Use Cases

### Large Vault Processing

```python
pages = client.list_properties_of_secrets().by_page()

for page_num, page in enumerate(pages, 1):
    secrets = list(page)
    print(f"Processing page {page_num} with {len(secrets)} secrets")
    
    # Batch process secrets in this page
    for secret in secrets:
        # Your processing logic
        pass
```

### Count All Secrets

```python
secret_properties = client.list_properties_of_secrets()
total = sum(1 for _ in secret_properties)
print(f"Total secrets: {total}")
```

### Filter by Content Type

```python
secret_properties = client.list_properties_of_secrets()

json_secrets = [s for s in secret_properties if s.content_type == "application/json"]
print(f"Found {len(json_secrets)} JSON secrets")
```

## Error Handling

```python
from azure.core.exceptions import ResourceNotFoundError, HttpResponseError

try:
    secret_properties = client.list_properties_of_secrets()
    for secret in secret_properties:
        print(secret.name)
except HttpResponseError as e:
    print(f"Error listing secrets: {e.message}")
```

## Performance Tips

1. **Use by_page()** for batch processing large numbers of secrets
2. **Don't fetch values in loops** - list operations don't include values for performance
3. **Enable logging** for debugging: `client = SecretClient(..., logging_enable=True)`
4. **Close credentials** when done: `credential.close()`

## Official Documentation Links

- [Azure Key Vault Secrets Overview](https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme)
- [SecretClient Reference](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient)
- [ItemPaged Reference](https://learn.microsoft.com/en-us/python/api/azure-core/azure.core.paging.itempaged)
- [SecretProperties Reference](https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretproperties)
