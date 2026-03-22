#!/usr/bin/env python3
"""
Azure Key Vault Secret Pagination Example

This script demonstrates how to list secrets from an Azure Key Vault with pagination
using the azure-keyvault-secrets SDK. It uses the ItemPaged pattern to efficiently
handle large vaults containing hundreds of secrets.

Based on Azure SDK for Python documentation:
- https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/
- https://learn.microsoft.com/en-us/python/api/azure-core/azure.core.paging.itempaged
"""

from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential
from datetime import datetime


def list_secrets_with_pagination(vault_url: str):
    """
    List all enabled secrets in an Azure Key Vault using pagination.
    
    Args:
        vault_url: The URL of the Azure Key Vault (e.g., "https://myvault.vault.azure.net/")
    """
    # Create a SecretClient using DefaultAzureCredential
    # DefaultAzureCredential automatically tries multiple authentication methods:
    # - Environment variables, Managed Identity, Azure CLI, etc.
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    print(f"Listing secrets from: {vault_url}\n")
    print("=" * 80)
    
    try:
        # list_properties_of_secrets() returns an ItemPaged[SecretProperties]
        # This method lists identifiers and attributes, but NOT the actual secret values
        # Requires secrets/list permission
        secret_properties = client.list_properties_of_secrets()
        
        # Process secrets page by page using by_page()
        # by_page() returns an iterator of pages, where each page is itself an iterator
        page_count = 0
        total_secrets = 0
        total_enabled = 0
        
        for page in secret_properties.by_page():
            page_count += 1
            secrets_in_page = 0
            
            print(f"\n--- Page {page_count} ---")
            
            # Iterate through each secret in this page
            for secret_prop in page:
                secrets_in_page += 1
                total_secrets += 1
                
                # Filter to show only enabled secrets
                if secret_prop.enabled:
                    total_enabled += 1
                    
                    # Extract the requested properties
                    name = secret_prop.name or "N/A"
                    content_type = secret_prop.content_type or "None"
                    created_on = secret_prop.created_on
                    
                    # Format the created date
                    if created_on:
                        created_date_str = created_on.strftime("%Y-%m-%d %H:%M:%S UTC")
                    else:
                        created_date_str = "N/A"
                    
                    # Print the secret information
                    print(f"  Secret: {name}")
                    print(f"    Content Type: {content_type}")
                    print(f"    Created: {created_date_str}")
                    print(f"    Enabled: {secret_prop.enabled}")
                    print()
            
            print(f"Secrets in this page: {secrets_in_page}")
        
        # Print summary
        print("=" * 80)
        print(f"\nSummary:")
        print(f"  Total pages processed: {page_count}")
        print(f"  Total secrets found: {total_secrets}")
        print(f"  Enabled secrets: {total_enabled}")
        print(f"  Disabled secrets: {total_secrets - total_enabled}")
        
    except Exception as e:
        print(f"Error listing secrets: {type(e).__name__}: {e}")
    finally:
        # Close the client connection
        client.close()


def list_secrets_simple_iteration(vault_url: str):
    """
    Alternative approach: Iterate through all secrets without explicit page handling.
    
    The ItemPaged object can be iterated directly without calling by_page().
    This is simpler but provides less control over pagination.
    
    Args:
        vault_url: The URL of the Azure Key Vault
    """
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    print(f"\nSimple iteration (no explicit pagination):")
    print("=" * 80)
    
    try:
        # list_properties_of_secrets() returns ItemPaged[SecretProperties]
        secret_properties = client.list_properties_of_secrets()
        
        # Iterate directly through the ItemPaged object
        # Pages are fetched automatically as needed
        enabled_count = 0
        
        for secret_prop in secret_properties:
            if secret_prop.enabled:
                enabled_count += 1
                print(f"  {secret_prop.name} - Created: {secret_prop.created_on}")
        
        print(f"\nTotal enabled secrets: {enabled_count}")
        
    except Exception as e:
        print(f"Error: {type(e).__name__}: {e}")
    finally:
        client.close()


if __name__ == "__main__":
    import sys
    
    # Example usage
    # Replace with your Key Vault URL
    if len(sys.argv) > 1:
        vault_url = sys.argv[1]
    else:
        # Default example URL (replace with your actual vault)
        vault_url = "https://your-vault-name.vault.azure.net/"
        print("Usage: python list_key_vault_secrets.py <vault-url>")
        print(f"Using default vault URL: {vault_url}")
        print("Replace with your actual Key Vault URL\n")
    
    # Demonstrate pagination with by_page()
    list_secrets_with_pagination(vault_url)
    
    # Uncomment to see the simple iteration approach
    # list_secrets_simple_iteration(vault_url)
