#!/usr/bin/env python3
"""
Azure Key Vault Secrets Pagination Example

This script demonstrates how to list and paginate through hundreds of secrets
in an Azure Key Vault using the azure-keyvault-secrets SDK.

It shows:
- Using SecretClient with DefaultAzureCredential
- Iterating through secrets using the ItemPaged pattern
- Processing secrets in pages using by_page()
- Filtering for enabled secrets only
"""

from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential


def list_secrets_paginated(vault_url: str):
    """
    List all enabled secrets in a Key Vault using pagination.
    
    Args:
        vault_url: The URL of the Azure Key Vault (e.g., 'https://myvault.vault.azure.net/')
    """
    # Create a SecretClient using DefaultAzureCredential
    # DefaultAzureCredential will try multiple authentication methods:
    # - Environment variables
    # - Managed Identity
    # - Azure CLI
    # - Azure PowerShell
    # - Interactive browser
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    print(f"Listing secrets from: {vault_url}\n")
    print("=" * 80)
    
    # list_properties_of_secrets() returns an ItemPaged[SecretProperties]
    # This doesn't include secret values, only metadata
    secret_properties = client.list_properties_of_secrets()
    
    # Process secrets page by page using by_page()
    # This is efficient for large vaults with hundreds of secrets
    page_number = 0
    total_secrets = 0
    enabled_secrets = 0
    
    # by_page() returns an iterator of pages, where each page is an iterator of items
    for page in secret_properties.by_page():
        page_number += 1
        page_secrets = 0
        
        print(f"\n--- Page {page_number} ---")
        
        # Iterate through secrets in the current page
        for secret_property in page:
            page_secrets += 1
            total_secrets += 1
            
            # Filter: only process enabled secrets
            if secret_property.enabled:
                enabled_secrets += 1
                
                # Extract properties
                name = secret_property.name
                content_type = secret_property.content_type or "Not set"
                created_on = secret_property.created_on
                
                # Format the created date
                if created_on:
                    created_date_str = created_on.strftime("%Y-%m-%d %H:%M:%S UTC")
                else:
                    created_date_str = "Unknown"
                
                # Print secret information
                print(f"  Name:         {name}")
                print(f"  Content Type: {content_type}")
                print(f"  Created:      {created_date_str}")
                print(f"  Enabled:      {secret_property.enabled}")
                print()
        
        print(f"Secrets in this page: {page_secrets}")
    
    # Print summary
    print("=" * 80)
    print(f"\nSummary:")
    print(f"  Total pages processed: {page_number}")
    print(f"  Total secrets found:   {total_secrets}")
    print(f"  Enabled secrets:       {enabled_secrets}")
    print(f"  Disabled secrets:      {total_secrets - enabled_secrets}")


def list_secrets_simple(vault_url: str):
    """
    Alternative approach: iterate through all secrets without explicit pagination.
    
    This is simpler but ItemPaged handles pagination automatically behind the scenes.
    For very large vaults, explicit pagination with by_page() is more efficient.
    
    Args:
        vault_url: The URL of the Azure Key Vault
    """
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    print(f"\nSimple iteration (automatic pagination):")
    print("=" * 80)
    
    enabled_count = 0
    
    # ItemPaged can be used as a simple iterator
    # Pagination happens automatically behind the scenes
    for secret_property in client.list_properties_of_secrets():
        if secret_property.enabled:
            enabled_count += 1
            print(f"  {secret_property.name} - {secret_property.created_on}")
    
    print(f"\nEnabled secrets: {enabled_count}")


if __name__ == "__main__":
    import os
    import sys
    
    # Get the vault URL from environment variable or command line
    vault_url = os.environ.get("AZURE_KEYVAULT_URL")
    
    if len(sys.argv) > 1:
        vault_url = sys.argv[1]
    
    if not vault_url:
        print("Error: Please provide the Key Vault URL")
        print("\nUsage:")
        print("  1. Set environment variable:")
        print("     export AZURE_KEYVAULT_URL='https://myvault.vault.azure.net/'")
        print("     python list_keyvault_secrets_paginated.py")
        print("\n  2. Pass as command line argument:")
        print("     python list_keyvault_secrets_paginated.py https://myvault.vault.azure.net/")
        sys.exit(1)
    
    # Ensure the URL has the correct format
    if not vault_url.startswith("https://"):
        vault_url = f"https://{vault_url}"
    if not vault_url.endswith("/"):
        vault_url = f"{vault_url}/"
    
    try:
        # Demonstrate paginated iteration
        list_secrets_paginated(vault_url)
        
        # Uncomment to see the simple iteration approach
        # list_secrets_simple(vault_url)
        
    except Exception as e:
        print(f"\nError: {e}")
        print("\nTroubleshooting:")
        print("  1. Ensure you're authenticated (az login)")
        print("  2. Verify you have 'List' permission on secrets")
        print("  3. Check the vault URL is correct")
        sys.exit(1)
