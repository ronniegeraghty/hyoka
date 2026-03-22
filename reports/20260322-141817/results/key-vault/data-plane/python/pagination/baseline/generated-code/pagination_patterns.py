"""
Quick Reference: Azure Key Vault Secrets Pagination Patterns

Based on official Azure SDK for Python documentation.
"""

# ============================================================================
# REQUIRED IMPORTS
# ============================================================================
from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient

# ============================================================================
# BASIC SETUP
# ============================================================================
vault_url = "https://my-vault.vault.azure.net/"
credential = DefaultAzureCredential()
client = SecretClient(vault_url=vault_url, credential=credential)

# ============================================================================
# PATTERN 1: EXPLICIT PAGINATION WITH by_page()
# ============================================================================
# Returns: ItemPaged[SecretProperties]
secret_properties = client.list_properties_of_secrets()

# Get page iterator
pages = secret_properties.by_page()

# Iterate through pages
for page in pages:
    print("--- New Page ---")
    for secret in page:
        print(f"  {secret.name}")

# ============================================================================
# PATTERN 2: AUTOMATIC PAGINATION (SIMPLE ITERATION)
# ============================================================================
# ItemPaged is directly iterable - handles pagination automatically
secret_properties = client.list_properties_of_secrets()

for secret in secret_properties:
    print(secret.name)

# ============================================================================
# PATTERN 3: FILTERING WITH PAGINATION
# ============================================================================
secret_properties = client.list_properties_of_secrets()

# Filter for enabled secrets only
for secret in secret_properties:
    if secret.enabled:
        print(f"{secret.name} - Created: {secret.created_on}")

# ============================================================================
# PATTERN 4: PAGE-BY-PAGE WITH CONTINUATION TOKEN
# ============================================================================
# by_page() accepts a continuation_token for resuming from a specific point
pages = secret_properties.by_page(continuation_token=None)

for page in pages:
    for secret in page:
        print(secret.name)
    
    # Get continuation token if you need to pause and resume
    # continuation_token = page.continuation_token

# ============================================================================
# ACCESSING SecretProperties ATTRIBUTES
# ============================================================================
secret_properties = client.list_properties_of_secrets()

for secret in secret_properties:
    # Common attributes
    print(f"Name: {secret.name}")
    print(f"Content Type: {secret.content_type}")
    print(f"Created On: {secret.created_on}")
    print(f"Updated On: {secret.updated_on}")
    print(f"Enabled: {secret.enabled}")
    print(f"Expires On: {secret.expires_on}")
    print(f"Version: {secret.version}")
    print(f"ID: {secret.id}")
    print(f"Tags: {secret.tags}")

# ============================================================================
# OTHER RELATED LIST METHODS
# ============================================================================

# List all versions of a specific secret (also returns ItemPaged)
versions = client.list_properties_of_secret_versions("secret-name")
for version in versions:
    print(f"Version: {version.version}, Created: {version.created_on}")

# List deleted secrets (soft-delete enabled vaults only)
deleted_secrets = client.list_deleted_secrets()
for deleted in deleted_secrets:
    print(f"Deleted: {deleted.name}, Deleted on: {deleted.deleted_date}")

# ============================================================================
# RESOURCE CLEANUP
# ============================================================================
client.close()
credential.close()

# Or use context managers:
# with SecretClient(vault_url=vault_url, credential=credential) as client:
#     secret_properties = client.list_properties_of_secrets()
#     for secret in secret_properties:
#         print(secret.name)
