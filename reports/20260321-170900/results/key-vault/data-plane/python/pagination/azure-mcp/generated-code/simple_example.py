#!/usr/bin/env python3
"""
Minimal example showing Azure Key Vault pagination mechanics
"""

from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential

vault_url = "https://your-vault.vault.azure.net/"  # Replace with your vault

# Initialize client
credential = DefaultAzureCredential()
client = SecretClient(vault_url=vault_url, credential=credential)

print("=== Method 1: Page-by-page iteration (explicit control) ===\n")

# Get ItemPaged object
secrets_paged = client.list_properties_of_secrets()

# Iterate page by page
for page_num, page in enumerate(secrets_paged.by_page(), start=1):
    print(f"Page {page_num}:")
    
    # Each page is an iterator of SecretProperties objects
    for secret_props in page:
        if secret_props.enabled:  # Filter for enabled secrets only
            print(f"  - {secret_props.name}")
            print(f"    Content Type: {secret_props.content_type or 'N/A'}")
            print(f"    Created: {secret_props.created_on}")
    
    print()

print("\n=== Method 2: Simple iteration (automatic pagination) ===\n")

# Direct iteration - pagination happens automatically behind the scenes
for secret_props in client.list_properties_of_secrets():
    if secret_props.enabled:
        print(f"{secret_props.name} - Created: {secret_props.created_on}")

# Clean up
client.close()

print("\nNote: list_properties_of_secrets() returns metadata only.")
print("To get secret values, call: client.get_secret(name)")
