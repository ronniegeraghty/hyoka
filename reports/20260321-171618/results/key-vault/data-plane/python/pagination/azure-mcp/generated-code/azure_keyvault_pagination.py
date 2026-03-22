#!/usr/bin/env python3
"""
Azure Key Vault Secrets Pagination Demo

This script demonstrates how to list all secrets in an Azure Key Vault
with hundreds of secrets using proper pagination techniques.

Required packages:
    pip install azure-keyvault-secrets azure-identity
"""

import os
from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient


def main():
    """
    List all enabled secrets from an Azure Key Vault using pagination.
    
    Prerequisites:
    - Set VAULT_URL environment variable (e.g., "https://my-vault.vault.azure.net/")
    - Configure Azure authentication (DefaultAzureCredential will use environment variables,
      managed identity, Azure CLI, or other available authentication methods)
    """
    
    # Get vault URL from environment
    vault_url = os.environ.get("VAULT_URL")
    if not vault_url:
        raise ValueError("VAULT_URL environment variable is not set")
    
    # Create credential and secret client
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    print(f"Connecting to Key Vault: {vault_url}\n")
    print("=" * 80)
    
    # Example 1: Basic iteration (simplest approach)
    print("\n1. BASIC ITERATION")
    print("-" * 80)
    print("This approach iterates through all secrets automatically.")
    print("The SDK handles pagination behind the scenes.\n")
    
    secret_count = 0
    secret_properties = client.list_properties_of_secrets()
    
    for secret_property in secret_properties:
        # Filter to show only enabled secrets
        if secret_property.enabled:
            secret_count += 1
            print(f"Secret #{secret_count}:")
            print(f"  Name:         {secret_property.name}")
            print(f"  Content Type: {secret_property.content_type or 'Not set'}")
            print(f"  Created On:   {secret_property.created_on}")
            print(f"  Enabled:      {secret_property.enabled}")
            print()
    
    print(f"Total enabled secrets (basic iteration): {secret_count}\n")
    
    # Example 2: Page-by-page iteration using by_page()
    print("=" * 80)
    print("\n2. PAGE-BY-PAGE ITERATION")
    print("-" * 80)
    print("This approach explicitly processes secrets in pages.")
    print("Useful for displaying progress, implementing rate limiting, or batch processing.\n")
    
    secret_properties = client.list_properties_of_secrets()
    pages = secret_properties.by_page()
    
    page_count = 0
    total_secrets = 0
    enabled_secrets = 0
    
    for page in pages:
        page_count += 1
        page_secrets = list(page)
        total_secrets += len(page_secrets)
        
        print(f"Processing Page {page_count} ({len(page_secrets)} secrets)")
        
        for secret_property in page_secrets:
            if secret_property.enabled:
                enabled_secrets += 1
                print(f"  - {secret_property.name}")
                print(f"    Content Type: {secret_property.content_type or 'Not set'}")
                print(f"    Created:      {secret_property.created_on}")
        
        print()
    
    print(f"Summary:")
    print(f"  Total pages processed:   {page_count}")
    print(f"  Total secrets found:     {total_secrets}")
    print(f"  Enabled secrets:         {enabled_secrets}")
    print(f"  Disabled/filtered:       {total_secrets - enabled_secrets}\n")
    
    # Example 3: Using continuation tokens
    print("=" * 80)
    print("\n3. CONTINUATION TOKEN USAGE")
    print("-" * 80)
    print("This approach demonstrates how to use continuation tokens")
    print("for resumable pagination (e.g., saving state between runs).\n")
    
    secret_properties = client.list_properties_of_secrets()
    pages = secret_properties.by_page()
    
    page_num = 0
    continuation_token = None
    
    # Process first 2 pages to demonstrate continuation
    for page in pages:
        page_num += 1
        page_secrets = list(page)
        
        print(f"Page {page_num}: {len(page_secrets)} secrets")
        
        # Show first few secrets from this page
        for i, secret_property in enumerate(page_secrets[:3]):
            if secret_property.enabled:
                print(f"  {i+1}. {secret_property.name}")
        
        if len(page_secrets) > 3:
            print(f"  ... and {len(page_secrets) - 3} more secrets")
        
        # Get continuation token for resuming later
        if hasattr(page, 'continuation_token'):
            continuation_token = page.continuation_token
            print(f"  Continuation token available: {continuation_token[:50]}..." if continuation_token else "  No continuation token")
        
        print()
        
        # Stop after 2 pages for demo purposes
        if page_num >= 2:
            break
    
    print(f"Processed {page_num} pages for demonstration.\n")
    
    # Resume from continuation token (if available)
    if continuation_token:
        print("Resuming from continuation token...\n")
        secret_properties = client.list_properties_of_secrets()
        resumed_pages = secret_properties.by_page(continuation_token=continuation_token)
        
        for i, page in enumerate(resumed_pages):
            if i == 0:  # Show just the first resumed page
                page_secrets = list(page)
                print(f"Resumed page has {len(page_secrets)} secrets")
                break
    
    print("\n" + "=" * 80)
    print("\nPagination demonstration complete!")
    print("\nKey Takeaways:")
    print("1. ItemPaged provides automatic iteration over all items")
    print("2. Use by_page() for explicit page-by-page processing")
    print("3. Continuation tokens enable resumable pagination")
    print("4. Filter properties like 'enabled' to process only relevant secrets")
    print("5. Secret values are NOT included in list operations for performance")
    print("   (Use client.get_secret(name) to retrieve values)")
    

if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\n\nOperation cancelled by user.")
    except Exception as e:
        print(f"\nError: {e}")
        print("\nTroubleshooting:")
        print("1. Ensure VAULT_URL environment variable is set")
        print("2. Verify Azure authentication is configured (Azure CLI, environment variables, etc.)")
        print("3. Check that you have 'secrets/list' permission in the Key Vault")
        raise
