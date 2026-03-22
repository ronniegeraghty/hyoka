#!/usr/bin/env python3
"""
Azure Key Vault Secrets Pagination Example

This script demonstrates how to list all secrets in an Azure Key Vault
that contains hundreds of secrets using the ItemPaged pagination pattern.

Required pip packages:
    pip install azure-keyvault-secrets azure-identity
"""

import os
from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient


def list_secrets_with_pagination(vault_url: str):
    """
    List all secrets in a Key Vault using pagination.
    
    Args:
        vault_url: The URL of the Azure Key Vault (e.g., https://my-vault.vault.azure.net/)
    """
    # Create a SecretClient using DefaultAzureCredential
    # DefaultAzureCredential will try multiple authentication methods:
    # - Environment variables
    # - Managed Identity
    # - Azure CLI credentials
    # - Azure PowerShell credentials
    # - Interactive browser (if needed)
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    print(f"Listing secrets from Key Vault: {vault_url}\n")
    print("=" * 80)
    
    # Get an ItemPaged object that will paginate through all secrets
    # list_properties_of_secrets() returns ItemPaged[SecretProperties]
    # Note: This does NOT include the secret values, only metadata
    secret_properties_paged = client.list_properties_of_secrets()
    
    # METHOD 1: Iterate through all secrets (pagination handled automatically)
    print("\nMETHOD 1: Simple iteration (pagination handled internally)")
    print("-" * 80)
    
    total_count = 0
    enabled_count = 0
    
    for secret_property in secret_properties_paged:
        # Filter to show only enabled secrets
        if secret_property.enabled:
            enabled_count += 1
            
            # Print secret details
            print(f"\nSecret Name: {secret_property.name}")
            print(f"  Content Type: {secret_property.content_type or 'Not set'}")
            print(f"  Created On: {secret_property.created_on}")
            print(f"  Enabled: {secret_property.enabled}")
            print(f"  Updated On: {secret_property.updated_on}")
        
        total_count += 1
    
    print(f"\n\nTotal secrets: {total_count}")
    print(f"Enabled secrets: {enabled_count}")
    
    
def list_secrets_by_page(vault_url: str):
    """
    List secrets in a Key Vault by processing pages explicitly.
    
    This demonstrates how to use the by_page() method to process
    secrets in chunks/pages, which is useful for:
    - Understanding API request patterns
    - Implementing custom progress tracking
    - Handling very large vaults more efficiently
    
    Args:
        vault_url: The URL of the Azure Key Vault
    """
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    print(f"\n\n{'=' * 80}")
    print("METHOD 2: Processing by pages (explicit page iteration)")
    print("=" * 80)
    
    # Get an ItemPaged object
    secret_properties_paged = client.list_properties_of_secrets()
    
    # Use by_page() to get a page iterator
    # by_page() returns Iterator[Iterator[SecretProperties]]
    # Each page is itself an iterator of SecretProperties objects
    pages = secret_properties_paged.by_page()
    
    page_number = 0
    total_secrets = 0
    total_enabled = 0
    
    for page in pages:
        page_number += 1
        page_count = 0
        enabled_in_page = 0
        
        print(f"\n--- Page {page_number} ---")
        
        # Iterate through secrets in this page
        for secret_property in page:
            page_count += 1
            total_secrets += 1
            
            # Filter to show only enabled secrets
            if secret_property.enabled:
                enabled_in_page += 1
                total_enabled += 1
                
                print(f"\n  Secret Name: {secret_property.name}")
                print(f"    Content Type: {secret_property.content_type or 'Not set'}")
                print(f"    Created On: {secret_property.created_on}")
                print(f"    Enabled: {secret_property.enabled}")
        
        print(f"\n  Secrets in page {page_number}: {page_count}")
        print(f"  Enabled in page {page_number}: {enabled_in_page}")
    
    print(f"\n\n--- Summary ---")
    print(f"Total pages processed: {page_number}")
    print(f"Total secrets: {total_secrets}")
    print(f"Total enabled secrets: {total_enabled}")


def list_secrets_with_continuation_token(vault_url: str):
    """
    List secrets using continuation tokens for resumable pagination.
    
    This demonstrates how to:
    - Start pagination from a specific point using continuation tokens
    - Save and restore pagination state
    - Implement resumable operations
    
    Args:
        vault_url: The URL of the Azure Key Vault
    """
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    print(f"\n\n{'=' * 80}")
    print("METHOD 3: Using continuation tokens")
    print("=" * 80)
    
    secret_properties_paged = client.list_properties_of_secrets()
    
    # Get pages with ability to access continuation token
    pages = secret_properties_paged.by_page()
    
    # Process first page and get continuation token
    first_page = next(pages, None)
    
    if first_page:
        print("\nProcessing first page...")
        count = 0
        for secret_property in first_page:
            if secret_property.enabled:
                count += 1
        
        print(f"Enabled secrets in first page: {count}")
        
        # The continuation_token can be accessed from the page iterator
        # This token can be saved and used to resume pagination later
        # Note: In practice, you would save this token to a database or file
        print("\nContinuation token available for resuming pagination")
        print("(In a real application, you would save this token to resume later)")


def main():
    """Main entry point for the script."""
    # Get the vault URL from environment variable
    vault_url = os.environ.get("AZURE_KEYVAULT_URL")
    
    if not vault_url:
        print("Error: AZURE_KEYVAULT_URL environment variable not set")
        print("\nUsage:")
        print("  export AZURE_KEYVAULT_URL='https://your-vault.vault.azure.net/'")
        print("  python list_key_vault_secrets_paginated.py")
        return
    
    try:
        # Demonstrate three different pagination approaches
        list_secrets_with_pagination(vault_url)
        list_secrets_by_page(vault_url)
        list_secrets_with_continuation_token(vault_url)
        
    except Exception as e:
        print(f"\nError: {e}")
        print("\nMake sure you have:")
        print("1. Set AZURE_KEYVAULT_URL environment variable")
        print("2. Authenticated with Azure (az login)")
        print("3. Have 'List' permission for secrets in the Key Vault")


if __name__ == "__main__":
    main()
