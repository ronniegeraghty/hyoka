#!/usr/bin/env python3
"""
Azure Key Vault Secrets Pagination Example

This script demonstrates how to list secrets from an Azure Key Vault
with hundreds of secrets using the ItemPaged pagination pattern.

Based on Azure SDK for Python documentation:
- https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme
- https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/
"""

import os
from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient


def list_secrets_with_pagination(vault_url: str) -> None:
    """
    List all enabled secrets from Key Vault using pagination.
    
    Args:
        vault_url: The URL of the Azure Key Vault (e.g., https://my-vault.vault.azure.net/)
    """
    # Create credential and client
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    try:
        # Get ItemPaged object for secret properties
        # list_properties_of_secrets returns ItemPaged[SecretProperties]
        secret_properties_paged = client.list_properties_of_secrets()
        
        # Process secrets page by page using by_page()
        # by_page() returns an iterator of pages (each page is an iterator of items)
        page_iterator = secret_properties_paged.by_page()
        
        page_count = 0
        total_secrets = 0
        enabled_secrets = 0
        
        print(f"Listing secrets from vault: {vault_url}\n")
        print("=" * 80)
        
        for page in page_iterator:
            page_count += 1
            secrets_in_page = 0
            
            print(f"\n--- Page {page_count} ---\n")
            
            # Iterate through secrets in this page
            for secret_property in page:
                secrets_in_page += 1
                total_secrets += 1
                
                # Filter to show only enabled secrets
                if secret_property.enabled:
                    enabled_secrets += 1
                    
                    # Extract properties
                    name = secret_property.name
                    content_type = secret_property.content_type or "Not specified"
                    created_on = secret_property.created_on
                    
                    # Format created date
                    created_date_str = created_on.strftime("%Y-%m-%d %H:%M:%S UTC") if created_on else "Unknown"
                    
                    # Print secret information
                    print(f"Secret Name:    {name}")
                    print(f"Content Type:   {content_type}")
                    print(f"Created Date:   {created_date_str}")
                    print(f"Enabled:        {secret_property.enabled}")
                    print("-" * 80)
            
            print(f"\nSecrets in page {page_count}: {secrets_in_page}")
        
        # Print summary
        print("\n" + "=" * 80)
        print(f"\nSummary:")
        print(f"  Total pages processed:     {page_count}")
        print(f"  Total secrets found:       {total_secrets}")
        print(f"  Enabled secrets (shown):   {enabled_secrets}")
        print(f"  Disabled secrets (hidden): {total_secrets - enabled_secrets}")
        
    finally:
        # Close the client
        client.close()
        credential.close()


def main():
    """Main entry point."""
    # Get vault URL from environment variable
    vault_url = os.environ.get("AZURE_KEY_VAULT_URL")
    
    if not vault_url:
        print("Error: AZURE_KEY_VAULT_URL environment variable is not set.")
        print("\nUsage:")
        print("  export AZURE_KEY_VAULT_URL='https://your-vault.vault.azure.net/'")
        print("  python list_secrets_paginated.py")
        return 1
    
    try:
        list_secrets_with_pagination(vault_url)
        return 0
    except Exception as e:
        print(f"\nError: {e}")
        return 1


if __name__ == "__main__":
    exit(main())
