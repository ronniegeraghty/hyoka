#!/usr/bin/env python3
"""
Azure Key Vault Secrets Pagination Demo

This script demonstrates how the azure-keyvault-secrets SDK handles pagination
when listing secrets from an Azure Key Vault containing hundreds of secrets.

Based on the official Azure SDK for Python documentation:
https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme
"""

import os
from datetime import datetime
from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential


def main():
    """
    Demonstrates pagination patterns for Azure Key Vault secrets.
    
    Key concepts:
    - list_properties_of_secrets() returns an ItemPaged[SecretProperties] object
    - ItemPaged supports iteration and the by_page() method for pagination
    - by_page() returns an iterator of pages (each page is also an iterator)
    - Filtering can be done during iteration to show only enabled secrets
    """
    
    # Get Key Vault URL from environment variable
    # Set VAULT_URL environment variable: export VAULT_URL="https://your-vault.vault.azure.net/"
    vault_url = os.environ.get("VAULT_URL")
    if not vault_url:
        print("ERROR: Please set the VAULT_URL environment variable")
        print("Example: export VAULT_URL='https://your-vault.vault.azure.net/'")
        return
    
    # Authenticate using DefaultAzureCredential
    # This supports multiple authentication methods (environment variables, managed identity, CLI, etc.)
    # For local development, use: az login
    credential = DefaultAzureCredential()
    
    # Create the SecretClient
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    print(f"Connected to Key Vault: {vault_url}")
    print("=" * 80)
    
    # Method 1: Simple iteration (automatic pagination)
    print("\n=== Method 1: Simple Iteration (ItemPaged) ===")
    print("The SDK handles pagination automatically when iterating.\n")
    
    secret_count = 0
    enabled_count = 0
    
    # list_properties_of_secrets() returns ItemPaged[SecretProperties]
    # Note: This does NOT return secret values, only metadata
    secret_properties = client.list_properties_of_secrets()
    
    for secret in secret_properties:
        secret_count += 1
        
        # Filter: Only process enabled secrets
        if secret.enabled:
            enabled_count += 1
            
            # Format the created date
            created_date = secret.created_on.strftime("%Y-%m-%d %H:%M:%S") if secret.created_on else "N/A"
            
            # Get content_type (may be None)
            content_type = secret.content_type if secret.content_type else "Not set"
            
            print(f"Secret #{enabled_count}:")
            print(f"  Name:         {secret.name}")
            print(f"  Content Type: {content_type}")
            print(f"  Created On:   {created_date}")
            print(f"  Enabled:      {secret.enabled}")
            print()
    
    print(f"Total secrets found: {secret_count}")
    print(f"Enabled secrets: {enabled_count}")
    
    # Method 2: Pagination with by_page()
    print("\n" + "=" * 80)
    print("\n=== Method 2: Explicit Pagination using by_page() ===")
    print("Process secrets page by page for better control and performance.\n")
    
    # Get a new iterator
    secret_properties = client.list_properties_of_secrets()
    
    # Use by_page() to get an iterator of pages
    # Each page is itself an iterator of SecretProperties objects
    pages = secret_properties.by_page()
    
    page_number = 0
    total_secrets = 0
    total_enabled = 0
    
    for page in pages:
        page_number += 1
        page_secret_count = 0
        page_enabled_count = 0
        
        print(f"--- Page {page_number} ---")
        
        # Iterate through secrets in this page
        for secret in page:
            page_secret_count += 1
            total_secrets += 1
            
            # Filter: Only count enabled secrets
            if secret.enabled:
                page_enabled_count += 1
                total_enabled += 1
                
                # Format the created date
                created_date = secret.created_on.strftime("%Y-%m-%d %H:%M:%S") if secret.created_on else "N/A"
                
                # Get content_type (may be None)
                content_type = secret.content_type if secret.content_type else "Not set"
                
                print(f"  {secret.name}")
                print(f"    Content Type: {content_type}")
                print(f"    Created On:   {created_date}")
        
        print(f"\nPage {page_number} summary:")
        print(f"  Secrets in this page: {page_secret_count}")
        print(f"  Enabled in this page: {page_enabled_count}")
        print()
    
    print(f"Total pages processed: {page_number}")
    print(f"Total secrets across all pages: {total_secrets}")
    print(f"Total enabled secrets: {total_enabled}")
    
    # Method 3: Demonstrate continuation token (for resuming pagination)
    print("\n" + "=" * 80)
    print("\n=== Method 3: Using Continuation Tokens ===")
    print("Continuation tokens allow resuming pagination from a specific point.\n")
    
    secret_properties = client.list_properties_of_secrets()
    pages = secret_properties.by_page()
    
    # Get the first page
    first_page = next(pages)
    first_page_list = list(first_page)
    
    print(f"First page has {len(first_page_list)} secrets")
    if first_page_list:
        print(f"First secret: {first_page_list[0].name}")
    
    # Get continuation token from the page iterator
    # Note: The continuation_token attribute is available on the page iterator
    # In real scenarios, you would save this token to resume later
    try:
        continuation_token = pages.continuation_token
        if continuation_token:
            print(f"\nContinuation token available: {continuation_token[:50]}...")
            
            # Create a new iterator starting from the continuation token
            resumed_pages = client.list_properties_of_secrets().by_page(
                continuation_token=continuation_token
            )
            
            # Get the next page using the continuation token
            next_page = next(resumed_pages)
            next_page_list = list(next_page)
            
            print(f"Resumed page has {len(next_page_list)} secrets")
            if next_page_list:
                print(f"First secret in resumed page: {next_page_list[0].name}")
        else:
            print("\nNo continuation token (only one page of results)")
    except (AttributeError, StopIteration):
        print("\nNo more pages available or continuation token not accessible")
    
    print("\n" + "=" * 80)
    print("\nPagination demonstration complete!")
    print("\nKey Takeaways:")
    print("1. list_properties_of_secrets() returns ItemPaged[SecretProperties]")
    print("2. ItemPaged can be iterated directly (SDK handles pagination)")
    print("3. Use by_page() for explicit page-by-page processing")
    print("4. Each page is an iterator of SecretProperties objects")
    print("5. Secret values are NOT included - use get_secret() to retrieve values")
    print("6. Continuation tokens enable resuming pagination")
    
    # Clean up
    client.close()
    credential.close()


if __name__ == "__main__":
    main()
