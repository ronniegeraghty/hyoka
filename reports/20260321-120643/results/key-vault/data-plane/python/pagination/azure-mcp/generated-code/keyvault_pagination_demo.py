#!/usr/bin/env python3
"""
Azure Key Vault Secrets Pagination Demo

This script demonstrates how the azure-keyvault-secrets SDK handles pagination
for large vaults with hundreds of secrets using the ItemPaged pattern.

Key concepts:
- SecretClient with DefaultAzureCredential
- ItemPaged pattern returned by list_properties_of_secrets()
- by_page() method for page-based iteration
- Filtering enabled secrets only
"""

import os
from datetime import datetime
from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential


def list_secrets_with_pagination(vault_url: str) -> None:
    """
    List all enabled secrets from Azure Key Vault using pagination.
    
    Args:
        vault_url: The URL of the Azure Key Vault (e.g., https://my-vault.vault.azure.net/)
    """
    # Create SecretClient with DefaultAzureCredential
    # DefaultAzureCredential tries multiple authentication methods:
    # - Environment variables, Managed Identity, Azure CLI, etc.
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    print(f"Connecting to Key Vault: {vault_url}")
    print("=" * 80)
    
    # list_properties_of_secrets() returns an ItemPaged[SecretProperties] object
    # This doesn't include secret values - only metadata
    secret_properties_paged = client.list_properties_of_secrets()
    
    # Use by_page() to iterate through results page by page
    # This is efficient for large vaults as it fetches data in chunks
    page_number = 0
    total_secrets = 0
    enabled_secrets = 0
    
    print("\nProcessing secrets page by page...\n")
    
    # by_page() returns an iterator of pages, where each page is an iterator of items
    for page in secret_properties_paged.by_page():
        page_number += 1
        page_secret_count = 0
        
        print(f"--- Page {page_number} ---")
        
        # Iterate through secrets in the current page
        for secret_property in page:
            page_secret_count += 1
            total_secrets += 1
            
            # Filter to show only enabled secrets
            if secret_property.enabled:
                enabled_secrets += 1
                
                # Extract properties
                name = secret_property.name
                content_type = secret_property.content_type or "N/A"
                created_on = secret_property.created_on
                
                # Format the created date
                if created_on:
                    created_date_str = created_on.strftime("%Y-%m-%d %H:%M:%S UTC")
                else:
                    created_date_str = "N/A"
                
                # Print secret information
                print(f"  Secret Name: {name}")
                print(f"    Content Type: {content_type}")
                print(f"    Created On: {created_date_str}")
                print(f"    Enabled: Yes")
                print()
        
        print(f"Secrets in this page: {page_secret_count}")
        print()
    
    # Summary
    print("=" * 80)
    print(f"\nSummary:")
    print(f"  Total pages processed: {page_number}")
    print(f"  Total secrets found: {total_secrets}")
    print(f"  Enabled secrets: {enabled_secrets}")
    print(f"  Disabled secrets: {total_secrets - enabled_secrets}")
    
    # Clean up
    client.close()
    credential.close()


def demonstrate_pagination_with_continuation_token(vault_url: str) -> None:
    """
    Demonstrate pagination with continuation tokens for resuming iteration.
    
    This shows how to pause and resume pagination, useful for:
    - Processing large datasets in batches
    - Handling interruptions or rate limits
    - Distributing work across multiple processes
    
    Args:
        vault_url: The URL of the Azure Key Vault
    """
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    print("\n" + "=" * 80)
    print("Demonstrating continuation tokens...")
    print("=" * 80)
    
    secret_properties_paged = client.list_properties_of_secrets()
    
    # Get the first page
    pages_iterator = secret_properties_paged.by_page()
    first_page = next(pages_iterator)
    
    print("\nProcessing first page only:")
    count = 0
    for secret_property in first_page:
        if secret_property.enabled:
            count += 1
            print(f"  {count}. {secret_property.name}")
    
    print(f"\nProcessed {count} enabled secrets from first page")
    print("(In a real scenario, you could save the continuation token and resume later)")
    
    # Clean up
    client.close()
    credential.close()


def main():
    """
    Main entry point for the script.
    """
    # Get vault URL from environment variable
    # Set this with: export VAULT_URL="https://your-vault-name.vault.azure.net/"
    vault_url = os.environ.get("VAULT_URL")
    
    if not vault_url:
        print("Error: VAULT_URL environment variable is not set.")
        print("Please set it to your Azure Key Vault URL:")
        print('  export VAULT_URL="https://your-vault-name.vault.azure.net/"')
        print("\nAlso ensure DefaultAzureCredential can authenticate:")
        print("  - Set AZURE_CLIENT_ID, AZURE_CLIENT_SECRET, AZURE_TENANT_ID, or")
        print("  - Use Azure CLI: az login, or")
        print("  - Use Managed Identity when running in Azure")
        return 1
    
    try:
        # Main pagination demonstration
        list_secrets_with_pagination(vault_url)
        
        # Bonus: Continuation token demonstration
        demonstrate_pagination_with_continuation_token(vault_url)
        
        print("\n✓ Script completed successfully!")
        return 0
        
    except Exception as e:
        print(f"\n✗ Error: {e}")
        print(f"Error type: {type(e).__name__}")
        return 1


if __name__ == "__main__":
    exit(main())
