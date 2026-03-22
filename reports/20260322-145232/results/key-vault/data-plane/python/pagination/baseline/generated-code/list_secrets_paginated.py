#!/usr/bin/env python3
"""
Azure Key Vault Secrets - Pagination Demonstration

This script demonstrates how to handle pagination when listing secrets in an Azure Key Vault
that contains hundreds of secrets using the Azure SDK for Python.

Prerequisites:
1. An Azure Key Vault with secrets
2. azure-keyvault-secrets and azure-identity packages installed
3. Proper Azure credentials configured for DefaultAzureCredential

For more information:
- SecretClient: https://learn.microsoft.com/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient
- ItemPaged: https://learn.microsoft.com/python/api/azure-core/azure.core.paging.itempaged
"""

import os
from datetime import datetime
from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential


def main():
    """
    Demonstrates pagination patterns for listing secrets in Azure Key Vault.
    """
    
    # Initialize the SecretClient with DefaultAzureCredential
    # DefaultAzureCredential tries multiple authentication methods (environment variables,
    # managed identity, Azure CLI, etc.)
    vault_url = os.environ.get("VAULT_URL")
    if not vault_url:
        raise ValueError("VAULT_URL environment variable must be set")
    
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    print(f"Connected to Key Vault: {vault_url}\n")
    print("=" * 80)
    
    # Method 1: Simple iteration (SDK handles pagination automatically)
    print("\n[Method 1] Simple iteration - SDK handles pagination automatically")
    print("-" * 80)
    demonstrate_simple_iteration(client)
    
    # Method 2: Page-by-page iteration with explicit pagination control
    print("\n[Method 2] Page-by-page iteration using by_page()")
    print("-" * 80)
    demonstrate_page_iteration(client)
    
    # Method 3: Page-by-page with continuation token
    print("\n[Method 3] Page-by-page with continuation token")
    print("-" * 80)
    demonstrate_continuation_token(client)
    
    print("\n" + "=" * 80)
    print("Pagination demonstration complete!")


def demonstrate_simple_iteration(client: SecretClient):
    """
    Method 1: Simple iteration over secrets.
    
    list_properties_of_secrets() returns an ItemPaged[SecretProperties] object.
    When you iterate over it directly, the SDK automatically handles pagination
    in the background, fetching additional pages as needed.
    
    This is the simplest approach but provides less control over pagination.
    """
    print("Iterating through all secrets (filtering for enabled secrets only)...\n")
    
    # list_properties_of_secrets returns ItemPaged[SecretProperties]
    # Note: This does NOT return secret values, only metadata
    secret_properties = client.list_properties_of_secrets()
    
    count = 0
    enabled_count = 0
    
    # Direct iteration - SDK handles pagination automatically
    for secret_property in secret_properties:
        count += 1
        
        # Filter to show only enabled secrets
        if secret_property.enabled:
            enabled_count += 1
            
            # Format the created date
            created_date = format_datetime(secret_property.created_on)
            
            # SecretProperties has content_type, which may be None
            content_type = secret_property.content_type or "Not specified"
            
            print(f"  Name: {secret_property.name}")
            print(f"    Content Type: {content_type}")
            print(f"    Created: {created_date}")
            print(f"    Enabled: {secret_property.enabled}")
            print()
    
    print(f"Total secrets: {count}")
    print(f"Enabled secrets: {enabled_count}")
    print(f"Disabled secrets: {count - enabled_count}")


def demonstrate_page_iteration(client: SecretClient):
    """
    Method 2: Iterate page by page using by_page().
    
    The by_page() method returns an iterator of pages, where each page is itself
    an iterator of items. This gives you explicit control over pagination and
    allows you to process secrets in batches.
    
    This is useful when:
    - You want to display progress per page
    - You need to process secrets in batches
    - You want to control memory usage with large result sets
    """
    print("Processing secrets page by page...\n")
    
    # Get the ItemPaged object
    secret_properties = client.list_properties_of_secrets()
    
    # Use by_page() to get an iterator of pages
    # Each page is an iterator of SecretProperties
    pages = secret_properties.by_page()
    
    page_num = 0
    total_count = 0
    total_enabled = 0
    
    # Iterate through each page
    for page in pages:
        page_num += 1
        page_count = 0
        page_enabled = 0
        
        print(f"--- Page {page_num} ---")
        
        # Iterate through secrets in this page
        for secret_property in page:
            page_count += 1
            total_count += 1
            
            # Filter to show only enabled secrets
            if secret_property.enabled:
                page_enabled += 1
                total_enabled += 1
                
                created_date = format_datetime(secret_property.created_on)
                content_type = secret_property.content_type or "Not specified"
                
                print(f"  Name: {secret_property.name}")
                print(f"    Content Type: {content_type}")
                print(f"    Created: {created_date}")
                print()
        
        print(f"Secrets in this page: {page_count}")
        print(f"Enabled secrets in this page: {page_enabled}")
        print()
    
    print(f"Total pages: {page_num}")
    print(f"Total secrets: {total_count}")
    print(f"Total enabled secrets: {total_enabled}")


def demonstrate_continuation_token(client: SecretClient):
    """
    Method 3: Using continuation tokens for resumable pagination.
    
    Continuation tokens allow you to save your position in the result set
    and resume later. This is useful for:
    - Long-running operations that might be interrupted
    - Implementing "load more" functionality in UIs
    - Distributing work across multiple processes
    
    The continuation_token can be retrieved from the page iterator's
    continuation_token attribute.
    """
    print("Demonstrating continuation tokens...\n")
    
    # Get the ItemPaged object
    secret_properties = client.list_properties_of_secrets()
    
    # Get pages iterator
    pages = secret_properties.by_page()
    
    # Process first page
    first_page = next(pages, None)
    
    if first_page is None:
        print("No secrets found in vault.")
        return
    
    print("--- First Page ---")
    count = 0
    for secret_property in first_page:
        if secret_property.enabled:
            count += 1
            created_date = format_datetime(secret_property.created_on)
            content_type = secret_property.content_type or "Not specified"
            
            print(f"  Name: {secret_property.name}")
            print(f"    Content Type: {content_type}")
            print(f"    Created: {created_date}")
            print()
    
    print(f"Enabled secrets in first page: {count}")
    
    # Get the continuation token from the current page
    # This token can be saved and used later to resume pagination
    continuation_token = pages.continuation_token
    
    if continuation_token:
        print(f"\nContinuation token available: {continuation_token[:50]}...")
        print("(This token could be saved and used to resume pagination later)")
        print("\n--- Resuming from continuation token ---")
        
        # Create a new pages iterator starting from the continuation token
        resumed_pages = client.list_properties_of_secrets().by_page(
            continuation_token=continuation_token
        )
        
        # Process the next page using the continuation token
        next_page = next(resumed_pages, None)
        
        if next_page:
            print("Successfully resumed pagination from saved position")
            resume_count = 0
            for secret_property in next_page:
                if secret_property.enabled:
                    resume_count += 1
            print(f"Enabled secrets in resumed page: {resume_count}")
        else:
            print("No more pages available")
    else:
        print("\nNo continuation token - only one page of results")


def format_datetime(dt: datetime | None) -> str:
    """
    Format a datetime object for display.
    
    Args:
        dt: DateTime object to format, or None
        
    Returns:
        Formatted string representation of the datetime
    """
    if dt is None:
        return "N/A"
    return dt.strftime("%Y-%m-%d %H:%M:%S UTC")


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\n\nOperation cancelled by user")
    except Exception as e:
        print(f"\nError: {e}")
        raise
