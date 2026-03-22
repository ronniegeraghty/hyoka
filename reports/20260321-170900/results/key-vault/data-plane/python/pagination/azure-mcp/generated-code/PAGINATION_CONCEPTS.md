# Azure Key Vault Pagination - Key Concepts

## How Pagination Works in azure-keyvault-secrets

### The ItemPaged Pattern

```python
from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential

client = SecretClient(vault_url="https://vault.vault.azure.net/", 
                     credential=DefaultAzureCredential())

# Returns ItemPaged[SecretProperties] - not a list!
secrets = client.list_properties_of_secrets()
```

### ItemPaged is NOT a list

- It's an **iterator** that fetches data lazily
- Only loads one page at a time into memory
- Automatically handles continuation tokens
- Server determines page size (not configurable by client)

### Two Iteration Methods

#### 1. Direct Iteration (Automatic)
```python
# Pagination is completely hidden
for secret in client.list_properties_of_secrets():
    print(secret.name)
```

**When to use**: Simple scripts where you don't need page boundaries

#### 2. Page-by-page Iteration (Explicit)
```python
# Get explicit control over pages
for page in client.list_properties_of_secrets().by_page():
    # page is an iterator of items
    for secret in page:
        print(secret.name)
```

**When to use**: 
- Progress tracking (e.g., "Processing page 5 of N")
- Checkpointing/resume capability
- Batch processing with page boundaries
- Performance monitoring per page

### Continuation Tokens

For resuming pagination from a specific point:

```python
# Start fresh
pages = client.list_properties_of_secrets().by_page()

# Get first page
first_page = next(pages)
for secret in first_page:
    print(secret.name)

# Get continuation token to resume later
token = pages.continuation_token

# Later... resume from where you left off
resumed_pages = client.list_properties_of_secrets().by_page(continuation_token=token)
for page in resumed_pages:
    for secret in page:
        print(secret.name)
```

### Performance Characteristics

For a vault with 500 secrets:

- **Pages**: ~5-20 pages (depends on server-side page size)
- **HTTP requests**: One per page
- **Memory**: Only current page in memory (~25-100 items)
- **Latency**: ~100-500ms per page (network dependent)

### What You Get: SecretProperties (NOT Secret Values)

`list_properties_of_secrets()` returns **metadata only**:

```python
for secret in client.list_properties_of_secrets():
    # Available attributes:
    secret.name           # str
    secret.enabled        # bool
    secret.content_type   # str | None
    secret.created_on     # datetime | None
    secret.updated_on     # datetime | None
    secret.expires_on     # datetime | None
    secret.not_before     # datetime | None
    secret.tags           # dict | None
    secret.version        # str
    secret.vault_url      # str
    
    # NOT available - must call get_secret():
    # secret.value  ❌ This doesn't exist on SecretProperties
```

To get the actual secret value:
```python
secret_with_value = client.get_secret(secret.name)
print(secret_with_value.value)  # The actual secret
```

### Filtering During Iteration

Always filter early to improve performance:

```python
# ✅ Good - filter while iterating (low memory)
enabled_secrets = [s.name for s in client.list_properties_of_secrets() 
                   if s.enabled]

# ❌ Bad - loading all into memory first
all_secrets = list(client.list_properties_of_secrets())  # Could be huge!
enabled_secrets = [s.name for s in all_secrets if s.enabled]
```

### Error Handling

```python
from azure.core.exceptions import AzureError

try:
    for page_num, page in enumerate(client.list_properties_of_secrets().by_page(), 1):
        try:
            for secret in page:
                print(secret.name)
        except AzureError as page_error:
            print(f"Error on page {page_num}: {page_error}")
            # Can continue to next page or abort
            continue
except AzureError as e:
    print(f"Failed to list secrets: {e}")
```

### Best Practices

1. **Don't convert to list unnecessarily**: `list(ItemPaged)` defeats pagination
2. **Use context manager**: `with SecretClient(...) as client:`
3. **Filter early**: Check `enabled` during iteration, not after
4. **Don't fetch values unless needed**: `get_secret()` is much slower than `list_properties_of_secrets()`
5. **Use by_page() for large vaults**: Better progress tracking and error recovery

### Complete Example

```python
from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential

vault_url = "https://my-vault.vault.azure.net/"

with SecretClient(vault_url=vault_url, credential=DefaultAzureCredential()) as client:
    # Process page by page
    for page_num, page in enumerate(client.list_properties_of_secrets().by_page(), 1):
        print(f"\n--- Page {page_num} ---")
        page_count = 0
        
        for secret in page:
            # Filter for enabled secrets only
            if secret.enabled:
                page_count += 1
                print(f"{secret.name:30} | {secret.content_type or 'N/A':15} | {secret.created_on}")
        
        print(f"Enabled secrets in page: {page_count}")
```

### Comparison with Other SDKs

**JavaScript/TypeScript**: Similar pattern with `PagedAsyncIterableIterator`
**C#/.NET**: Uses `Pageable<T>` or `AsyncPageable<T>`
**Java**: Uses `PagedIterable<T>` or `PagedFlux<T>`

All follow the same concept: lazy-loading iterators with `by_page()` methods.
