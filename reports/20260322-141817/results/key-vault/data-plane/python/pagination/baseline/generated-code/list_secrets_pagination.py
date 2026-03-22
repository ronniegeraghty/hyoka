"""
Azure Key Vault Secrets Pagination Example

This script demonstrates how to list secrets from an Azure Key Vault
that contains hundreds of secrets using the ItemPaged pattern with pagination.

Based on Azure SDK for Python documentation:
- https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets
- https://learn.microsoft.com/en-us/python/api/azure-core/azure.core.paging.itempaged

Required packages:
    pip install azure-keyvault-secrets azure-identity
"""

import os
from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient


def list_secrets_with_pagination(vault_url: str):
    """
    List all enabled secrets from a Key Vault using pagination.
    
    Args:
        vault_url: The URL of the Azure Key Vault (e.g., https://my-vault.vault.azure.net/)
    """
    # Create SecretClient with DefaultAzureCredential
    # DefaultAzureCredential automatically tries multiple authentication methods
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    try:
        # list_properties_of_secrets() returns an ItemPaged[SecretProperties] object
        # This doesn't include the actual secret values, only metadata
        secret_properties = client.list_properties_of_secrets()
        
        # Process secrets page by page using by_page()
        # by_page() returns an iterator of pages (each page is itself an iterator)
        page_iterator = secret_properties.by_page()
        
        page_count = 0
        total_secrets = 0
        enabled_secrets = 0
        
        print(f"Listing secrets from: {vault_url}\n")
        print("=" * 80)
        
        # Iterate through each page
        for page in page_iterator:
            page_count += 1
            secrets_in_page = 0
            
            print(f"\n--- Page {page_count} ---")
            
            # Iterate through secrets in the current page
            for secret_property in page:
                total_secrets += 1
                secrets_in_page += 1
                
                # Filter to show only enabled secrets
                if secret_property.enabled:
                    enabled_secrets += 1
                    
                    # Print secret details
                    print(f"\nSecret Name: {secret_property.name}")
                    print(f"  Content Type: {secret_property.content_type or 'Not set'}")
                    print(f"  Created On: {secret_property.created_on}")
                    print(f"  Enabled: {secret_property.enabled}")
                    
                    # Additional metadata available:
                    # print(f"  Updated On: {secret_property.updated_on}")
                    # print(f"  Expires On: {secret_property.expires_on}")
                    # print(f"  Version: {secret_property.version}")
            
            print(f"\nSecrets in page {page_count}: {secrets_in_page}")
        
        # Print summary
        print("\n" + "=" * 80)
        print(f"\nSummary:")
        print(f"  Total pages processed: {page_count}")
        print(f"  Total secrets found: {total_secrets}")
        print(f"  Enabled secrets: {enabled_secrets}")
        print(f"  Disabled secrets: {total_secrets - enabled_secrets}")
        
    finally:
        # Close the client to release resources
        client.close()
        credential.close()


def list_secrets_alternative_method(vault_url: str):
    """
    Alternative approach: iterate through all secrets without explicit pagination.
    
    The ItemPaged object is itself iterable and handles pagination automatically
    behind the scenes. This is simpler but gives less control over page boundaries.
    
    Args:
        vault_url: The URL of the Azure Key Vault
    """
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    try:
        # list_properties_of_secrets() returns ItemPaged[SecretProperties]
        secret_properties = client.list_properties_of_secrets()
        
        enabled_count = 0
        
        print(f"\nListing enabled secrets (alternative method):\n")
        
        # ItemPaged is iterable - you can iterate directly without by_page()
        # This automatically handles pagination internally
        for secret_property in secret_properties:
            # Filter for enabled secrets only
            if secret_property.enabled:
                enabled_count += 1
                print(f"{enabled_count}. {secret_property.name} "
                      f"(Created: {secret_property.created_on})")
        
        print(f"\nTotal enabled secrets: {enabled_count}")
        
    finally:
        client.close()
        credential.close()


if __name__ == "__main__":
    # Get vault URL from environment variable
    # Example: export VAULT_URL="https://my-key-vault.vault.azure.net/"
    vault_url = os.environ.get("VAULT_URL")
    
    if not vault_url:
        print("Error: Please set the VAULT_URL environment variable")
        print("Example: export VAULT_URL='https://my-key-vault.vault.azure.net/'")
        exit(1)
    
    # Method 1: Explicit pagination with by_page()
    print("=" * 80)
    print("METHOD 1: Using by_page() for explicit pagination control")
    print("=" * 80)
    list_secrets_with_pagination(vault_url)
    
    # Method 2: Automatic pagination (simpler but less control)
    print("\n\n" + "=" * 80)
    print("METHOD 2: Automatic pagination (ItemPaged iterator)")
    print("=" * 80)
    list_secrets_alternative_method(vault_url)
