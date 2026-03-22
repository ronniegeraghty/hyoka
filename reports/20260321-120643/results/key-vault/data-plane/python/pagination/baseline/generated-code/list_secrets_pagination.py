#!/usr/bin/env python3
"""
Azure Key Vault Secrets - Pagination Example

This script demonstrates how to list all secrets in an Azure Key Vault
that contains hundreds of secrets using the ItemPaged pattern with pagination.

Based on official Azure SDK for Python documentation:
- https://learn.microsoft.com/python/api/azure-keyvault-secrets/
- https://learn.microsoft.com/python/api/azure-core/azure.core.paging.itempaged

Prerequisites:
1. Azure Key Vault with secrets
2. Proper authentication configured for DefaultAzureCredential
3. Required packages installed (see requirements below)
"""

import os
from datetime import datetime
from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential
from azure.core.exceptions import ResourceNotFoundError


def main():
    """
    Main function to demonstrate pagination with Azure Key Vault secrets.
    """
    
    # Get vault URL from environment variable
    vault_url = os.environ.get("VAULT_URL")
    if not vault_url:
        print("Error: VAULT_URL environment variable not set")
        print("Example: export VAULT_URL='https://your-vault-name.vault.azure.net/'")
        return
    
    print(f"Connecting to Key Vault: {vault_url}")
    print("-" * 80)
    
    # Create credential and client
    # DefaultAzureCredential will try multiple authentication methods:
    # 1. Environment variables (AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET)
    # 2. Managed Identity
    # 3. Azure CLI credentials
    # 4. Azure PowerShell credentials
    # 5. Interactive browser authentication
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    try:
        # Example 1: Simple iteration through all secrets (handles pagination automatically)
        print("\n=== Example 1: Simple iteration (automatic pagination) ===\n")
        simple_iteration_example(client)
        
        # Example 2: Manual pagination using by_page() for better control
        print("\n=== Example 2: Manual pagination with by_page() ===\n")
        manual_pagination_example(client)
        
        # Example 3: Filtering enabled secrets only with pagination
        print("\n=== Example 3: Filtering enabled secrets with pagination ===\n")
        filter_enabled_secrets_example(client)
        
    except ResourceNotFoundError as e:
        print(f"Resource not found: {e.message}")
    except Exception as e:
        print(f"An error occurred: {type(e).__name__}: {e}")
    finally:
        # Clean up
        credential.close()
        client.close()


def simple_iteration_example(client: SecretClient):
    """
    Example 1: Simple iteration through secrets.
    
    The list_properties_of_secrets() method returns an ItemPaged[SecretProperties] object.
    ItemPaged handles pagination automatically when you iterate over it.
    Behind the scenes, it fetches pages as needed.
    """
    secret_count = 0
    
    # list_properties_of_secrets() returns ItemPaged[SecretProperties]
    # Note: This does NOT include secret values, only properties/metadata
    secret_properties = client.list_properties_of_secrets()
    
    for secret_property in secret_properties:
        secret_count += 1
        print(f"{secret_count}. Name: {secret_property.name}")
        print(f"   Content Type: {secret_property.content_type or 'Not set'}")
        print(f"   Created: {format_datetime(secret_property.created_on)}")
        print(f"   Enabled: {secret_property.enabled}")
        print()
    
    print(f"Total secrets found: {secret_count}")


def manual_pagination_example(client: SecretClient):
    """
    Example 2: Manual pagination using by_page().
    
    The by_page() method returns an iterator of pages, where each page
    is itself an iterator of items. This gives you more control over
    pagination and allows you to process secrets page by page.
    
    This is useful for:
    - Monitoring progress in large vaults
    - Implementing custom batching logic
    - Handling rate limits or quotas
    - Debugging pagination issues
    """
    total_secrets = 0
    page_count = 0
    
    # Get an iterator of pages
    # Each page is an iterator of SecretProperties objects
    secret_pages = client.list_properties_of_secrets().by_page()
    
    for page in secret_pages:
        page_count += 1
        page_secrets = list(page)  # Convert page iterator to list to count items
        page_size = len(page_secrets)
        
        print(f"--- Page {page_count} ({page_size} secrets) ---")
        
        for secret_property in page_secrets:
            total_secrets += 1
            print(f"  {total_secrets}. {secret_property.name}")
            print(f"     Content Type: {secret_property.content_type or 'Not set'}")
            print(f"     Created: {format_datetime(secret_property.created_on)}")
            print(f"     Enabled: {secret_property.enabled}")
        
        print()
    
    print(f"Total pages: {page_count}")
    print(f"Total secrets: {total_secrets}")


def filter_enabled_secrets_example(client: SecretClient):
    """
    Example 3: Filter to show only enabled secrets using pagination.
    
    Demonstrates how to combine pagination with filtering logic.
    This processes secrets page by page and only displays enabled secrets.
    """
    enabled_count = 0
    disabled_count = 0
    page_count = 0
    
    # Process secrets page by page
    secret_pages = client.list_properties_of_secrets().by_page()
    
    for page in secret_pages:
        page_count += 1
        
        # Filter enabled secrets in this page
        enabled_in_page = [s for s in page if s.enabled]
        disabled_in_page_count = sum(1 for s in page if not s.enabled)
        
        if enabled_in_page:
            print(f"--- Page {page_count} ---")
            
            for secret_property in enabled_in_page:
                enabled_count += 1
                print(f"  {enabled_count}. Name: {secret_property.name}")
                print(f"     Content Type: {secret_property.content_type or 'Not set'}")
                print(f"     Created: {format_datetime(secret_property.created_on)}")
                print()
        
        disabled_count += disabled_in_page_count
    
    print(f"Summary:")
    print(f"  Total enabled secrets: {enabled_count}")
    print(f"  Total disabled secrets: {disabled_count}")
    print(f"  Total secrets: {enabled_count + disabled_count}")
    print(f"  Pages processed: {page_count}")


def format_datetime(dt: datetime) -> str:
    """
    Format datetime for display.
    
    Args:
        dt: DateTime object to format
        
    Returns:
        Formatted datetime string
    """
    if dt is None:
        return "N/A"
    return dt.strftime("%Y-%m-%d %H:%M:%S UTC")


if __name__ == "__main__":
    main()
