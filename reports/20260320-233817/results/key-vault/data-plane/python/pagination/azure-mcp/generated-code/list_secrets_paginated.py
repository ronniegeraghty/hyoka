"""
Azure Key Vault Secrets Pagination Demo

This script demonstrates how to list secrets from an Azure Key Vault
that contains hundreds of secrets using the ItemPaged pattern with pagination.
"""

from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential

# Configuration
KEY_VAULT_URL = "https://<your-key-vault-name>.vault.azure.net/"


def list_secrets_with_pagination(vault_url: str):
    """
    List all enabled secrets from Key Vault using pagination.
    
    The list_properties_of_secrets() method returns an ItemPaged[SecretProperties]
    object that supports pagination through the by_page() method.
    """
    
    # Create SecretClient with DefaultAzureCredential
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    print(f"Listing secrets from: {vault_url}\n")
    print("=" * 80)
    
    # Get ItemPaged iterator for secret properties
    # This returns an ItemPaged[SecretProperties] object
    secret_properties = client.list_properties_of_secrets()
    
    # Process secrets page by page using by_page()
    # This is efficient for large vaults with hundreds of secrets
    page_number = 1
    total_secrets = 0
    enabled_secrets = 0
    
    for page in secret_properties.by_page():
        print(f"\n--- Page {page_number} ---\n")
        
        secrets_in_page = 0
        
        # Iterate through secrets in this page
        for secret_property in page:
            secrets_in_page += 1
            total_secrets += 1
            
            # Filter: only show enabled secrets
            if secret_property.enabled:
                enabled_secrets += 1
                
                # Extract properties
                name = secret_property.name
                content_type = secret_property.content_type or "N/A"
                created_on = secret_property.created_on
                
                # Print secret information
                print(f"Secret Name: {name}")
                print(f"  Content Type: {content_type}")
                print(f"  Created On: {created_on}")
                print(f"  Enabled: {secret_property.enabled}")
                print()
        
        print(f"Secrets in this page: {secrets_in_page}")
        page_number += 1
    
    # Summary
    print("=" * 80)
    print(f"\nSummary:")
    print(f"  Total secrets processed: {total_secrets}")
    print(f"  Enabled secrets: {enabled_secrets}")
    print(f"  Disabled secrets: {total_secrets - enabled_secrets}")
    print(f"  Total pages: {page_number - 1}")


def list_secrets_simple_iteration(vault_url: str):
    """
    Alternative: Simple iteration without explicit pagination.
    
    The ItemPaged object is iterable and handles pagination automatically
    behind the scenes. This is simpler but gives less control.
    """
    
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    print(f"\nSimple iteration (auto-pagination):\n")
    print("=" * 80)
    
    enabled_count = 0
    
    # ItemPaged is iterable - pagination happens automatically
    for secret_property in client.list_properties_of_secrets():
        if secret_property.enabled:
            enabled_count += 1
            print(f"{secret_property.name} - Created: {secret_property.created_on}")
    
    print(f"\nTotal enabled secrets: {enabled_count}")


if __name__ == "__main__":
    # Replace with your Key Vault URL
    # vault_url = "https://my-keyvault.vault.azure.net/"
    
    try:
        # Method 1: Explicit pagination with by_page()
        list_secrets_with_pagination(KEY_VAULT_URL)
        
        # Method 2: Simple iteration (uncomment to use)
        # list_secrets_simple_iteration(KEY_VAULT_URL)
        
    except Exception as e:
        print(f"Error: {e}")
        print("\nMake sure to:")
        print("1. Replace KEY_VAULT_URL with your actual Key Vault URL")
        print("2. Have appropriate Azure credentials configured")
        print("3. Have the required permissions (Get, List) on the Key Vault")
