#!/usr/bin/env python3
"""
Azure Key Vault Secrets Pagination Demo

This script demonstrates how to use the Azure Key Vault Secrets SDK
to list all secrets in a vault with hundreds of secrets using pagination.

Based on Azure SDK for Python documentation:
- https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/
- https://learn.microsoft.com/en-us/python/api/azure-core/azure.core.paging.itempaged
"""

from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential
from datetime import datetime


def list_secrets_with_pagination(vault_url: str) -> None:
    """
    List all enabled secrets in a Key Vault using page-by-page iteration.
    
    Args:
        vault_url: The URL of the Azure Key Vault (e.g., "https://my-vault.vault.azure.net/")
    """
    # Create SecretClient with DefaultAzureCredential
    # DefaultAzureCredential tries multiple authentication methods automatically
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    try:
        # list_properties_of_secrets() returns ItemPaged[SecretProperties]
        # This lists secret metadata (not values) and requires secrets/list permission
        secret_properties_paged = client.list_properties_of_secrets()
        
        print(f"Listing secrets from vault: {vault_url}")
        print("=" * 80)
        
        # Process secrets page by page using by_page()
        # by_page() returns an iterator of pages, where each page is an iterator of items
        page_num = 0
        total_secrets = 0
        enabled_secrets = 0
        
        for page in secret_properties_paged.by_page():
            page_num += 1
            page_secrets = 0
            
            print(f"\n--- Page {page_num} ---")
            
            # Iterate through secrets in this page
            for secret_properties in page:
                page_secrets += 1
                total_secrets += 1
                
                # Filter to show only enabled secrets
                if secret_properties.enabled:
                    enabled_secrets += 1
                    
                    # Extract properties from SecretProperties object
                    name = secret_properties.name
                    content_type = secret_properties.content_type or "N/A"
                    created_on = secret_properties.created_on
                    
                    # Format the created date
                    if created_on:
                        created_date_str = created_on.strftime("%Y-%m-%d %H:%M:%S UTC")
                    else:
                        created_date_str = "N/A"
                    
                    # Print secret information
                    print(f"  Secret Name:    {name}")
                    print(f"  Content Type:   {content_type}")
                    print(f"  Created Date:   {created_date_str}")
                    print(f"  Enabled:        {secret_properties.enabled}")
                    print()
            
            print(f"Secrets in this page: {page_secrets}")
        
        # Summary
        print("=" * 80)
        print(f"Total secrets processed: {total_secrets}")
        print(f"Enabled secrets: {enabled_secrets}")
        print(f"Disabled secrets: {total_secrets - enabled_secrets}")
        print(f"Total pages: {page_num}")
        
    finally:
        # Close the client to clean up resources
        client.close()


def list_secrets_simple_iteration(vault_url: str) -> None:
    """
    Alternative approach: List all enabled secrets using simple iteration.
    
    This is simpler but doesn't give you explicit page control.
    ItemPaged can be iterated directly without calling by_page().
    
    Args:
        vault_url: The URL of the Azure Key Vault
    """
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    try:
        print(f"\nSimple iteration approach for vault: {vault_url}")
        print("=" * 80)
        
        enabled_count = 0
        
        # Direct iteration over ItemPaged - pagination happens automatically
        for secret_properties in client.list_properties_of_secrets():
            if secret_properties.enabled:
                enabled_count += 1
                print(f"Enabled Secret: {secret_properties.name}")
        
        print(f"\nTotal enabled secrets: {enabled_count}")
        
    finally:
        client.close()


if __name__ == "__main__":
    import sys
    
    # Example usage
    # Replace with your Key Vault URL
    # Format: https://<vault-name>.vault.azure.net/
    
    if len(sys.argv) > 1:
        vault_url = sys.argv[1]
    else:
        # Default example - replace with your vault URL
        vault_url = "https://my-keyvault.vault.azure.net/"
        print("Usage: python list_keyvault_secrets_paginated.py <vault-url>")
        print(f"Using default vault URL: {vault_url}")
        print("Replace with your actual vault URL.\n")
    
    # Demonstrate page-by-page iteration
    list_secrets_with_pagination(vault_url)
    
    # Uncomment to see the simple iteration approach
    # list_secrets_simple_iteration(vault_url)
