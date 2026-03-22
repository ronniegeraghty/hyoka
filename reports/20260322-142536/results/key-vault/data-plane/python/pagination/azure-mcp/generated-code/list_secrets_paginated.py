#!/usr/bin/env python3
"""
Azure Key Vault Secrets Pagination Example

This script demonstrates how to list secrets from an Azure Key Vault
that contains hundreds of secrets using the ItemPaged pattern with pagination.

Required environment variable:
    AZURE_KEY_VAULT_URL: The URL of your Azure Key Vault
                         (e.g., https://my-vault.vault.azure.net/)
"""

import os
from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient


def main():
    # Get the Key Vault URL from environment variable
    vault_url = os.environ.get("AZURE_KEY_VAULT_URL")
    if not vault_url:
        raise ValueError(
            "AZURE_KEY_VAULT_URL environment variable must be set. "
            "Example: https://my-vault.vault.azure.net/"
        )

    # Create a SecretClient using DefaultAzureCredential
    # DefaultAzureCredential automatically tries multiple authentication methods:
    # - Environment variables
    # - Managed Identity
    # - Azure CLI
    # - Azure PowerShell
    # - Interactive browser
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)

    print(f"Connecting to Key Vault: {vault_url}")
    print("=" * 80)

    # list_properties_of_secrets() returns an ItemPaged[SecretProperties] object
    # ItemPaged is an iterator that automatically handles pagination
    secret_properties = client.list_properties_of_secrets()

    # Method 1: Iterate through all secrets (automatic pagination)
    # The ItemPaged iterator automatically fetches additional pages as needed
    print("\n=== Method 1: Simple iteration (automatic pagination) ===\n")
    
    total_secrets = 0
    enabled_secrets = 0
    
    for secret_property in secret_properties:
        # Filter to show only enabled secrets
        if secret_property.enabled:
            enabled_secrets += 1
            print(f"Name: {secret_property.name}")
            print(f"  Content Type: {secret_property.content_type or 'Not set'}")
            print(f"  Created: {secret_property.created_on}")
            print(f"  Enabled: {secret_property.enabled}")
            print()
        
        total_secrets += 1

    print(f"Total secrets: {total_secrets}")
    print(f"Enabled secrets: {enabled_secrets}")

    # Method 2: Process secrets page by page using by_page()
    # This approach gives you explicit control over pagination
    print("\n" + "=" * 80)
    print("=== Method 2: Explicit page-by-page processing using by_page() ===\n")

    # Get a fresh iterator for the second demonstration
    secret_properties = client.list_properties_of_secrets()
    
    # by_page() returns an iterator of pages
    # Each page is itself an iterator of SecretProperties objects
    pages = secret_properties.by_page()
    
    page_number = 0
    total_secrets = 0
    enabled_secrets = 0
    
    for page in pages:
        page_number += 1
        secrets_in_page = 0
        enabled_in_page = 0
        
        print(f"--- Page {page_number} ---")
        
        # Iterate through secrets in this page
        for secret_property in page:
            secrets_in_page += 1
            total_secrets += 1
            
            # Filter to show only enabled secrets
            if secret_property.enabled:
                enabled_in_page += 1
                enabled_secrets += 1
                
                print(f"  [{secrets_in_page}] {secret_property.name}")
                print(f"      Content Type: {secret_property.content_type or 'Not set'}")
                print(f"      Created: {secret_property.created_on}")
                print(f"      Enabled: {secret_property.enabled}")
        
        print(f"\nSecrets in page {page_number}: {secrets_in_page}")
        print(f"Enabled secrets in page {page_number}: {enabled_in_page}")
        print()
    
    print(f"Total pages processed: {page_number}")
    print(f"Total secrets: {total_secrets}")
    print(f"Total enabled secrets: {enabled_secrets}")

    # Method 3: Using continuation tokens for resumable pagination
    print("\n" + "=" * 80)
    print("=== Method 3: Using continuation tokens (resumable pagination) ===\n")
    
    # Get a fresh iterator
    secret_properties = client.list_properties_of_secrets()
    
    # Start pagination with an optional continuation_token
    # Pass None to start from the beginning
    pages = secret_properties.by_page(continuation_token=None)
    
    page_count = 0
    for page in pages:
        page_count += 1
        
        # Get the continuation token for this page
        # This token can be used to resume iteration from the next page
        continuation_token = page.continuation_token
        
        secrets_in_page = sum(1 for _ in page)
        
        print(f"Page {page_count}: {secrets_in_page} secrets")
        
        if continuation_token:
            print(f"  Continuation token available (can resume from next page)")
        else:
            print(f"  No continuation token (this is the last page)")
        
        # In a real application, you could save the continuation_token
        # and use it later to resume pagination:
        # pages = client.list_properties_of_secrets().by_page(continuation_token=saved_token)
        
        # For this demo, we'll only process a few pages
        if page_count >= 3:
            print("\n(Stopping after 3 pages for demo purposes)")
            break
    
    print(f"\nProcessed {page_count} page(s)")

    # Close the credential when done
    credential.close()
    client.close()


if __name__ == "__main__":
    try:
        main()
    except Exception as e:
        print(f"Error: {e}")
        raise
