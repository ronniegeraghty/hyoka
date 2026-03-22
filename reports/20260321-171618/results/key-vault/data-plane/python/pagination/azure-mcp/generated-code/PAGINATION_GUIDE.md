# Azure Key Vault Pagination - Technical Deep Dive

## How the ItemPaged Pattern Works

Based on the Azure SDK for Python documentation, here's how pagination works in Azure Key Vault:

### Architecture

```
SecretClient.list_properties_of_secrets()
    ↓
Returns: ItemPaged[SecretProperties]
    ↓
ItemPaged provides two iteration modes:
    1. Direct iteration (automatic pagination)
    2. Page-by-page iteration via by_page()
```

### Direct Iteration (Automatic)

```python
secrets = client.list_properties_of_secrets()

# The ItemPaged object handles pagination automatically
for secret in secrets:
    print(secret.name)
```

**How it works:**
- The SDK fetches pages automatically as you iterate
- No need to manage continuation tokens manually
- Simplest approach for processing all items sequentially

### Page-by-Page Iteration

```python
secrets = client.list_properties_of_secrets()
pages = secrets.by_page()

for page in pages:
    # page is an iterator of SecretProperties
    for secret in page:
        print(secret.name)
```

**How it works:**
- `by_page()` returns an iterator of pages
- Each page is itself an iterator of items
- Allows you to:
  - Track progress per page
  - Implement batch processing
  - Add delays between pages for rate limiting
  - Process pages in parallel (if needed)

### Continuation Tokens

```python
secrets = client.list_properties_of_secrets()
pages = secrets.by_page()

# Get first page
first_page = next(pages)
items = list(first_page)

# Save continuation token
token = first_page.continuation_token

# Later, resume from where you left off
secrets = client.list_properties_of_secrets()
resumed_pages = secrets.by_page(continuation_token=token)
```

**Use cases:**
- Long-running operations that might be interrupted
- Processing large vaults in multiple sessions
- Checkpointing for fault tolerance
- Distributed processing

## Performance Characteristics

### What's Included in List Operations

✓ **Included** (Metadata):
- Secret name
- Enabled status
- Content type
- Created date
- Updated date
- Expiration date
- Tags
- Version
- Vault URL

✗ **NOT Included**:
- Secret values (must use `get_secret()` for values)

### Why This Matters

For a vault with 1000 secrets:

**List operation:**
- Single API call per page (page size determined by service)
- Returns only metadata (~1-2 KB per secret)
- Fast: processes all 1000 secrets in seconds

**Getting values:**
- Would require 1000 individual `get_secret()` calls
- Only do this if you actually need the values
- Consider parallel processing for large batches

## Best Practices

### 1. Choose the Right Iteration Method

Use **direct iteration** when:
- Processing all secrets sequentially
- No need for progress tracking
- Simplicity is preferred

Use **by_page()** when:
- Need to display progress
- Implementing rate limiting
- Batch processing with checkpoints
- Need to count items per page

Use **continuation tokens** when:
- Operation might take very long
- Need fault tolerance
- Want to resume after interruption
- Distributed processing

### 2. Filter Early

```python
# Good: Filter during iteration
for secret in client.list_properties_of_secrets():
    if secret.enabled and secret.content_type == "application/json":
        process(secret)

# Also good: Filter by page for batch processing
pages = client.list_properties_of_secrets().by_page()
for page in pages:
    enabled_secrets = [s for s in page if s.enabled]
    batch_process(enabled_secrets)
```

### 3. Handle Large Result Sets

For vaults with thousands of secrets:

```python
import time

pages = client.list_properties_of_secrets().by_page()

for page_num, page in enumerate(pages, 1):
    secrets = list(page)
    print(f"Processing page {page_num} ({len(secrets)} items)")
    
    # Process the page
    for secret in secrets:
        if secret.enabled:
            # Your processing logic
            pass
    
    # Optional: Add delay between pages to avoid rate limits
    if page_num % 10 == 0:
        time.sleep(0.1)
```

### 4. Error Handling

```python
from azure.core.exceptions import ResourceNotFoundError, ServiceRequestError

try:
    secrets = client.list_properties_of_secrets()
    for secret in secrets:
        print(secret.name)
except ResourceNotFoundError:
    print("Vault not found or no access")
except ServiceRequestError as e:
    print(f"Network error: {e}")
except Exception as e:
    print(f"Unexpected error: {e}")
```

## Comparison with Other Azure SDK List Operations

The ItemPaged pattern is consistent across Azure SDKs:

| Service | Method | Returns |
|---------|--------|---------|
| Key Vault Secrets | `list_properties_of_secrets()` | `ItemPaged[SecretProperties]` |
| Key Vault Keys | `list_properties_of_keys()` | `ItemPaged[KeyProperties]` |
| Storage Blobs | `list_blobs()` | `ItemPaged[BlobProperties]` |
| Cosmos DB | `query_items()` | `ItemPaged[Dict]` |

All support:
- Direct iteration
- `by_page()` for page control
- Continuation tokens

## Code Examples from Official Samples

The official Azure SDK samples demonstrate:

```python
# From: azure-sdk-for-python/sdk/keyvault/azure-keyvault-secrets/samples/list_operations.py

# Simple iteration
secrets = client.list_properties_of_secrets()
for secret in secrets:
    retrieved_secret = client.get_secret(secret.name)
    print(f"Secret: {retrieved_secret.name}")

# List versions of a specific secret
secret_versions = client.list_properties_of_secret_versions("secret-name")
for version in secret_versions:
    print(f"Version: {version.version}")
```

## Summary

The Azure Key Vault Python SDK provides three levels of pagination control:

1. **Automatic** - Direct iteration (easiest)
2. **Manual** - Page-by-page via `by_page()` (more control)
3. **Resumable** - Continuation tokens (fault-tolerant)

Choose based on your use case:
- Small vaults → Direct iteration
- Progress tracking → by_page()
- Long operations → Continuation tokens
- Always filter early to process only what you need
- Remember: list operations return metadata only, not values
