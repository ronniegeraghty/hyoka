#!/usr/bin/env python3
"""
Azure Key Vault Secrets Pagination Demo

This script demonstrates how to list secrets from an Azure Key Vault
that contains hundreds of secrets using the ItemPaged pattern with by_page().

Prerequisites:
1. An Azure Key Vault with secrets
2. Environment variable VAULT_URL set to your Key Vault URL
3. Proper Azure credentials configured for DefaultAzureCredential

Required packages:
    pip install azure-keyvault-secrets azure-identity
"""

import os
from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential


def main():
    # Get vault URL from environment variable
    vault_url = os.environ.get("VAULT_URL")
    if not vault_url:
        print("Error: VAULT_URL environment variable is not set")
        print("Example: export VAULT_URL='https://your-vault-name.vault.azure.net/'")
        return
    
    # Create a SecretClient using DefaultAzureCredential
    # DefaultAzureCredential will try multiple authentication methods:
    # - Environment variables
    # - Managed identity
    # - Azure CLI credentials
    # - Azure PowerShell credentials
    # - Interactive browser authentication
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    print(f"Connecting to Key Vault: {vault_url}\n")
    print("=" * 80)
    print("Listing all enabled secrets with pagination")
    print("=" * 80)
    
    # list_properties_of_secrets() returns an ItemPaged[SecretProperties] object
    # This is an iterable that supports pagination
    secret_properties_paged = client.list_properties_of_secrets()
    
    # Use by_page() to iterate through secrets in pages
    # This is useful for large vaults as it processes secrets in manageable chunks
    page_number = 0
    total_secrets = 0
    enabled_secrets = 0
    
    # by_page() returns an iterator of pages, where each page is an iterator of SecretProperties
    for page in secret_properties_paged.by_page():
        page_number += 1
        page_secrets = 0
        
        print(f"\n--- Page {page_number} ---")
        
        # Iterate through each secret in this page
        for secret_properties in page:
            page_secrets += 1
            total_secrets += 1
            
            # Filter to show only enabled secrets
            if secret_properties.enabled:
                enabled_secrets += 1
                
                # Extract the properties we want to display
                name = secret_properties.name
                content_type = secret_properties.content_type or "Not set"
                created_on = secret_properties.created_on
                
                # Format the created date
                if created_on:
                    created_date_str = created_on.strftime("%Y-%m-%d %H:%M:%S UTC")
                else:
                    created_date_str = "Unknown"
                
                # Print the secret information
                print(f"  Secret: {name}")
                print(f"    Content Type: {content_type}")
                print(f"    Created: {created_date_str}")
                print(f"    Enabled: {secret_properties.enabled}")
                print()
        
        print(f"Processed {page_secrets} secrets in this page")
    
    # Print summary statistics
    print("\n" + "=" * 80)
    print("Summary")
    print("=" * 80)
    print(f"Total pages processed: {page_number}")
    print(f"Total secrets found: {total_secrets}")
    print(f"Enabled secrets: {enabled_secrets}")
    print(f"Disabled secrets: {total_secrets - enabled_secrets}")
    
    # Close the credential when done
    credential.close()
    print("\nDone!")


def demonstrate_alternative_patterns():
    """
    Additional examples of working with ItemPaged pagination patterns.
    """
    vault_url = os.environ.get("VAULT_URL", "")
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    print("\n" + "=" * 80)
    print("Alternative Pattern 1: Simple iteration without by_page()")
    print("=" * 80)
    
    # You can also iterate directly without using by_page()
    # This is simpler but gives you less control over pagination
    count = 0
    for secret_properties in client.list_properties_of_secrets():
        if secret_properties.enabled:
            count += 1
            print(f"{count}. {secret_properties.name}")
            if count >= 5:  # Just show first 5 for demo
                print("  ...")
                break
    
    print("\n" + "=" * 80)
    print("Alternative Pattern 2: Using continuation tokens")
    print("=" * 80)
    
    # You can use continuation tokens to resume pagination
    # This is useful for long-running operations or when processing in batches
    secret_properties_paged = client.list_properties_of_secrets()
    pages = secret_properties_paged.by_page()
    
    # Get the first page
    first_page = next(pages)
    first_page_list = list(first_page)
    print(f"First page has {len(first_page_list)} secrets")
    
    # Get continuation token (if available)
    # Note: continuation_token is available on the page iterator
    # You can save this token and use it later to resume from this point
    # continuation_token = pages.continuation_token
    # Later, you can resume: client.list_properties_of_secrets().by_page(continuation_token)
    
    credential.close()


if __name__ == "__main__":
    try:
        main()
        
        # Uncomment to see alternative patterns:
        # demonstrate_alternative_patterns()
        
    except KeyboardInterrupt:
        print("\n\nOperation cancelled by user")
    except Exception as e:
        print(f"\nError: {e}")
        print("\nMake sure:")
        print("1. VAULT_URL environment variable is set")
        print("2. You have proper Azure credentials configured")
        print("3. You have 'secrets/list' permission on the Key Vault")
        raise
