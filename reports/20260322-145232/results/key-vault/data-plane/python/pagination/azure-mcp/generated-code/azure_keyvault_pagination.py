#!/usr/bin/env python3
"""
Azure Key Vault Secrets Pagination Example

This script demonstrates how to handle pagination when listing secrets
from an Azure Key Vault that contains hundreds of secrets.

Key concepts demonstrated:
1. SecretClient with DefaultAzureCredential
2. ItemPaged pattern for iterating through secrets
3. Processing secrets in pages using by_page()
4. Filtering enabled secrets only
5. Extracting secret properties (name, content type, created date)

Prerequisites:
- An Azure Key Vault with secrets
- azure-keyvault-secrets and azure-identity packages installed
- Authentication configured for DefaultAzureCredential
"""

import os
from datetime import datetime
from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential


def list_secrets_simple(client):
    """
    Basic iteration through secrets using the ItemPaged pattern.
    
    This method automatically handles pagination behind the scenes.
    Each secret is retrieved one at a time as you iterate.
    """
    print("\n" + "=" * 80)
    print("METHOD 1: Simple iteration (automatic pagination)")
    print("=" * 80)
    
    secret_count = 0
    enabled_count = 0
    
    # list_properties_of_secrets returns an ItemPaged[SecretProperties] object
    # Iterating over it automatically handles pagination
    secret_properties = client.list_properties_of_secrets()
    
    for secret in secret_properties:
        secret_count += 1
        
        # Filter to show only enabled secrets
        if secret.enabled:
            enabled_count += 1
            
            # Format the created date
            created_date = secret.created_on.strftime("%Y-%m-%d %H:%M:%S") if secret.created_on else "N/A"
            
            print(f"\nSecret #{enabled_count}:")
            print(f"  Name:         {secret.name}")
            print(f"  Content Type: {secret.content_type if secret.content_type else 'Not set'}")
            print(f"  Created:      {created_date}")
            print(f"  Enabled:      {secret.enabled}")
    
    print(f"\n\nTotal secrets processed: {secret_count}")
    print(f"Enabled secrets: {enabled_count}")


def list_secrets_by_page(client):
    """
    Process secrets page by page using the by_page() method.
    
    This approach is more efficient for large vaults as it:
    - Processes secrets in batches
    - Allows you to see how many API calls are made
    - Enables better control over pagination flow
    """
    print("\n" + "=" * 80)
    print("METHOD 2: Page-by-page iteration (explicit pagination)")
    print("=" * 80)
    
    secret_count = 0
    enabled_count = 0
    page_count = 0
    
    # Get the ItemPaged object
    secret_properties = client.list_properties_of_secrets()
    
    # Use by_page() to get an iterator of pages
    # Each page is itself an iterator of SecretProperties objects
    pages = secret_properties.by_page()
    
    for page in pages:
        page_count += 1
        secrets_in_page = 0
        
        print(f"\n--- Processing Page {page_count} ---")
        
        # Iterate through secrets in this page
        for secret in page:
            secret_count += 1
            secrets_in_page += 1
            
            # Filter to show only enabled secrets
            if secret.enabled:
                enabled_count += 1
                
                # Format the created date
                created_date = secret.created_on.strftime("%Y-%m-%d %H:%M:%S") if secret.created_on else "N/A"
                
                print(f"\n  Secret #{enabled_count}:")
                print(f"    Name:         {secret.name}")
                print(f"    Content Type: {secret.content_type if secret.content_type else 'Not set'}")
                print(f"    Created:      {created_date}")
                print(f"    Enabled:      {secret.enabled}")
        
        print(f"\nPage {page_count} contained {secrets_in_page} secrets")
    
    print(f"\n\nSummary:")
    print(f"  Total pages: {page_count}")
    print(f"  Total secrets processed: {secret_count}")
    print(f"  Enabled secrets: {enabled_count}")


def list_secrets_with_continuation_token(client):
    """
    Use continuation tokens to resume pagination from a specific point.
    
    This is useful for:
    - Resuming after an interruption
    - Distributing work across multiple processes
    - Implementing custom pagination controls
    """
    print("\n" + "=" * 80)
    print("METHOD 3: Pagination with continuation token")
    print("=" * 80)
    
    secret_properties = client.list_properties_of_secrets()
    
    # Get the first page
    pages = secret_properties.by_page()
    first_page = next(pages, None)
    
    if first_page is None:
        print("No secrets found in the vault.")
        return
    
    # Process first page
    print("\n--- First Page ---")
    first_page_count = 0
    for secret in first_page:
        if secret.enabled:
            first_page_count += 1
            print(f"  {secret.name}")
    
    print(f"First page had {first_page_count} enabled secrets")
    
    # Get continuation token from the page
    continuation_token = getattr(first_page, 'continuation_token', None)
    
    if continuation_token:
        print(f"\nContinuation token retrieved: {continuation_token[:50]}...")
        print("\n--- Resuming from second page using continuation token ---")
        
        # Resume pagination from the continuation token
        remaining_secrets = client.list_properties_of_secrets()
        remaining_pages = remaining_secrets.by_page(continuation_token=continuation_token)
        
        remaining_count = 0
        for page in remaining_pages:
            for secret in page:
                if secret.enabled:
                    remaining_count += 1
        
        print(f"Remaining pages had {remaining_count} enabled secrets")
        print(f"Total enabled secrets: {first_page_count + remaining_count}")
    else:
        print("\nNo continuation token (all secrets fit in one page)")


def main():
    """
    Main function to demonstrate all pagination methods.
    """
    # Get the Key Vault URL from environment variable
    vault_url = os.environ.get("VAULT_URL")
    
    if not vault_url:
        print("ERROR: Please set the VAULT_URL environment variable.")
        print("Example: export VAULT_URL='https://your-vault-name.vault.azure.net/'")
        return
    
    print("Azure Key Vault Secrets Pagination Demo")
    print("=" * 80)
    print(f"Vault URL: {vault_url}")
    
    # Create a SecretClient using DefaultAzureCredential
    # DefaultAzureCredential tries multiple authentication methods:
    # 1. Environment variables (AZURE_CLIENT_ID, AZURE_CLIENT_SECRET, AZURE_TENANT_ID)
    # 2. Managed Identity (if running on Azure)
    # 3. Azure CLI credentials (az login)
    # 4. Azure PowerShell credentials
    # 5. Interactive browser authentication
    print("\nAuthenticating with DefaultAzureCredential...")
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    try:
        # Demonstrate different pagination methods
        
        # Method 1: Simple iteration (ItemPaged handles pagination automatically)
        list_secrets_simple(client)
        
        # Method 2: Explicit page-by-page processing
        list_secrets_by_page(client)
        
        # Method 3: Using continuation tokens for resumable pagination
        list_secrets_with_continuation_token(client)
        
    except Exception as e:
        print(f"\nERROR: {type(e).__name__}: {e}")
        print("\nTroubleshooting:")
        print("1. Verify VAULT_URL is correct")
        print("2. Ensure you have 'secrets/list' permission on the Key Vault")
        print("3. Check authentication (try 'az login' if using Azure CLI)")
        print("4. Verify network connectivity to Azure")
    
    finally:
        # Clean up: close the credential
        credential.close()
    
    print("\n" + "=" * 80)
    print("Demo completed!")
    print("=" * 80)


if __name__ == "__main__":
    main()
