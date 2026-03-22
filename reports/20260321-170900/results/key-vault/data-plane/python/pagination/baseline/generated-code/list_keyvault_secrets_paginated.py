#!/usr/bin/env python3
"""
Azure Key Vault Secrets Pagination Example

This script demonstrates how to list secrets from an Azure Key Vault
that contains hundreds of secrets using the ItemPaged pagination pattern.

Required packages:
- azure-keyvault-secrets
- azure-identity
"""

from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient


def list_secrets_with_pagination(vault_url: str):
    """
    List all enabled secrets in a Key Vault using pagination.
    
    Args:
        vault_url: The URL of the Azure Key Vault (e.g., https://myvault.vault.azure.net/)
    """
    # Initialize SecretClient with DefaultAzureCredential
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    print(f"Listing secrets from: {vault_url}")
    print("=" * 80)
    
    # Get ItemPaged iterator for secret properties
    # list_properties_of_secrets returns ItemPaged[SecretProperties]
    secret_properties = client.list_properties_of_secrets()
    
    # Process secrets page by page using by_page()
    page_count = 0
    total_secrets = 0
    enabled_secrets = 0
    
    # by_page() returns an iterator of pages (each page is an iterator of SecretProperties)
    for page in secret_properties.by_page():
        page_count += 1
        secrets_in_page = 0
        
        print(f"\n--- Page {page_count} ---")
        
        # Iterate through secrets in the current page
        for secret_property in page:
            secrets_in_page += 1
            total_secrets += 1
            
            # Filter to show only enabled secrets
            if secret_property.enabled:
                enabled_secrets += 1
                
                # Print secret details
                name = secret_property.name
                content_type = secret_property.content_type or "N/A"
                created_on = secret_property.created_on
                
                # Format created_on date
                if created_on:
                    created_date = created_on.strftime("%Y-%m-%d %H:%M:%S UTC")
                else:
                    created_date = "N/A"
                
                print(f"  Secret: {name}")
                print(f"    Content Type: {content_type}")
                print(f"    Created On: {created_date}")
                print(f"    Enabled: {secret_property.enabled}")
        
        print(f"Secrets in this page: {secrets_in_page}")
    
    # Print summary
    print("\n" + "=" * 80)
    print(f"Summary:")
    print(f"  Total pages processed: {page_count}")
    print(f"  Total secrets found: {total_secrets}")
    print(f"  Enabled secrets: {enabled_secrets}")
    print(f"  Disabled secrets: {total_secrets - enabled_secrets}")


def list_secrets_simple_iteration(vault_url: str):
    """
    Alternative approach: List secrets using simple iteration without explicit pagination.
    
    ItemPaged can be iterated directly without calling by_page().
    This is useful when you don't need page-level control.
    
    Args:
        vault_url: The URL of the Azure Key Vault
    """
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    print(f"\nSimple iteration of secrets from: {vault_url}")
    print("=" * 80)
    
    # Direct iteration over ItemPaged (pagination happens transparently)
    enabled_count = 0
    
    for secret_property in client.list_properties_of_secrets():
        if secret_property.enabled:
            enabled_count += 1
            print(f"  {secret_property.name} - Created: {secret_property.created_on}")
    
    print(f"\nTotal enabled secrets: {enabled_count}")


if __name__ == "__main__":
    import os
    
    # Get vault URL from environment variable
    vault_url = os.environ.get("AZURE_KEYVAULT_URL")
    
    if not vault_url:
        print("Error: AZURE_KEYVAULT_URL environment variable not set")
        print("Example: export AZURE_KEYVAULT_URL='https://myvault.vault.azure.net/'")
        exit(1)
    
    try:
        # Demonstrate page-by-page iteration
        list_secrets_with_pagination(vault_url)
        
        # Uncomment to see simple iteration approach
        # list_secrets_simple_iteration(vault_url)
        
    except Exception as e:
        print(f"\nError: {e}")
        print("\nMake sure you have:")
        print("1. Set AZURE_KEYVAULT_URL environment variable")
        print("2. Appropriate Azure credentials configured (az login, managed identity, etc.)")
        print("3. 'secrets/list' permission in the Key Vault access policy")
        exit(1)
