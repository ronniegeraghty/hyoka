#!/usr/bin/env python3
"""
Azure Key Vault Secrets Pagination Example

This script demonstrates how to list all secrets in an Azure Key Vault
using pagination with the azure-keyvault-secrets SDK. It handles vaults
with hundreds of secrets efficiently using the ItemPaged pattern.

Required packages:
    pip install azure-keyvault-secrets azure-identity
"""

import os
from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient


def list_secrets_with_pagination(vault_url: str):
    """
    List all enabled secrets in a Key Vault using pagination.
    
    Args:
        vault_url: The URL of the Azure Key Vault (e.g., https://my-vault.vault.azure.net/)
    """
    # Create SecretClient with DefaultAzureCredential
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    print(f"Listing secrets from: {vault_url}\n")
    print("=" * 80)
    
    # Get ItemPaged iterator for secret properties
    # Note: list_properties_of_secrets() returns ItemPaged[SecretProperties]
    # It does NOT include secret values, only metadata
    secret_properties_paged = client.list_properties_of_secrets()
    
    # Process secrets page by page using by_page()
    page_count = 0
    total_secrets = 0
    enabled_secrets = 0
    
    # by_page() returns an iterator of pages (each page is itself an iterator)
    for page in secret_properties_paged.by_page():
        page_count += 1
        secrets_in_page = 0
        
        print(f"\n--- Page {page_count} ---")
        
        # Iterate through each secret in the current page
        for secret_property in page:
            secrets_in_page += 1
            total_secrets += 1
            
            # Filter to show only enabled secrets
            if secret_property.enabled:
                enabled_secrets += 1
                
                # Extract and format the properties
                name = secret_property.name
                content_type = secret_property.content_type or "N/A"
                created_on = secret_property.created_on
                
                # Format created date
                if created_on:
                    created_date_str = created_on.strftime("%Y-%m-%d %H:%M:%S UTC")
                else:
                    created_date_str = "N/A"
                
                # Print secret information
                print(f"  Secret: {name}")
                print(f"    Content Type: {content_type}")
                print(f"    Created: {created_date_str}")
                print(f"    Enabled: {secret_property.enabled}")
                print()
        
        print(f"  Secrets in this page: {secrets_in_page}")
    
    # Print summary
    print("=" * 80)
    print(f"\nSummary:")
    print(f"  Total pages processed: {page_count}")
    print(f"  Total secrets found: {total_secrets}")
    print(f"  Enabled secrets: {enabled_secrets}")
    print(f"  Disabled secrets: {total_secrets - enabled_secrets}")


def list_secrets_simple_iteration(vault_url: str):
    """
    Alternative approach: List secrets using simple iteration (without explicit pagination).
    
    The ItemPaged object can be iterated directly without using by_page().
    This approach is simpler but gives less control over page boundaries.
    
    Args:
        vault_url: The URL of the Azure Key Vault
    """
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    print(f"\nSimple iteration approach (no explicit pagination):\n")
    print("=" * 80)
    
    # Iterate directly through the ItemPaged object
    # Pagination happens automatically in the background
    enabled_count = 0
    
    for secret_property in client.list_properties_of_secrets():
        if secret_property.enabled:
            enabled_count += 1
            print(f"Secret: {secret_property.name}")
            print(f"  Content Type: {secret_property.content_type or 'N/A'}")
            print(f"  Created: {secret_property.created_on}")
            print()
    
    print(f"Total enabled secrets: {enabled_count}")


if __name__ == "__main__":
    # Get vault URL from environment variable
    vault_url = os.environ.get("AZURE_KEYVAULT_URL")
    
    if not vault_url:
        print("Error: AZURE_KEYVAULT_URL environment variable not set")
        print("\nUsage:")
        print("  export AZURE_KEYVAULT_URL='https://your-vault.vault.azure.net/'")
        print("  python list_keyvault_secrets_paginated.py")
        exit(1)
    
    try:
        # Demonstrate pagination with by_page()
        list_secrets_with_pagination(vault_url)
        
        # Uncomment to see the simple iteration approach
        # list_secrets_simple_iteration(vault_url)
        
    except Exception as e:
        print(f"\nError: {e}")
        print("\nMake sure you have:")
        print("  1. Set AZURE_KEYVAULT_URL environment variable")
        print("  2. Authenticated with Azure (az login)")
        print("  3. Have 'List' permission on secrets in the Key Vault")
        exit(1)
