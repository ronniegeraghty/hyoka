#!/usr/bin/env python3
"""
Azure Key Vault Secrets Pagination Example

This script demonstrates how to list secrets from an Azure Key Vault with
hundreds of secrets using the ItemPaged pattern and by_page() method for
efficient pagination.

Required packages:
    pip install azure-keyvault-secrets azure-identity

Environment variables required:
    AZURE_KEY_VAULT_URL - The URL of your Azure Key Vault
                         (e.g., https://my-vault.vault.azure.net/)
"""

import os
from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient


def list_secrets_with_pagination():
    """
    List all enabled secrets from Azure Key Vault using pagination.
    
    Demonstrates:
    - Using SecretClient with DefaultAzureCredential
    - Iterating through ItemPaged results with by_page()
    - Processing secrets in pages
    - Filtering enabled secrets
    - Accessing secret properties (name, content_type, created_on)
    """
    
    # Get the Key Vault URL from environment variable
    vault_url = os.environ.get("AZURE_KEY_VAULT_URL")
    if not vault_url:
        raise ValueError(
            "AZURE_KEY_VAULT_URL environment variable must be set.\n"
            "Example: export AZURE_KEY_VAULT_URL='https://my-vault.vault.azure.net/'"
        )
    
    # Create credential and client
    # DefaultAzureCredential will try multiple authentication methods:
    # - Environment variables
    # - Managed Identity
    # - Azure CLI
    # - Interactive browser
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    print(f"Connecting to Key Vault: {vault_url}")
    print("=" * 80)
    print()
    
    # list_properties_of_secrets() returns an ItemPaged[SecretProperties] object
    # This doesn't include the secret values, only metadata
    secret_properties = client.list_properties_of_secrets()
    
    # Use by_page() to iterate through secrets page by page
    # This is efficient for large vaults with hundreds of secrets
    page_count = 0
    total_secrets = 0
    enabled_secrets = 0
    
    print("Processing secrets by page...\n")
    
    # by_page() returns an iterator of pages, where each page is itself an iterator
    for page in secret_properties.by_page():
        page_count += 1
        secrets_in_page = 0
        
        print(f"--- Page {page_count} ---")
        
        # Iterate through secrets in this page
        for secret_property in page:
            secrets_in_page += 1
            total_secrets += 1
            
            # Filter: only process enabled secrets
            if secret_property.enabled:
                enabled_secrets += 1
                
                # Extract the properties requested
                name = secret_property.name
                content_type = secret_property.content_type or "Not set"
                created_on = secret_property.created_on
                
                # Format created_on as a readable string
                created_date_str = (
                    created_on.strftime("%Y-%m-%d %H:%M:%S UTC")
                    if created_on
                    else "Unknown"
                )
                
                # Print secret details
                print(f"  Secret: {name}")
                print(f"    Content Type: {content_type}")
                print(f"    Created On:   {created_date_str}")
                print(f"    Enabled:      {secret_property.enabled}")
                print()
        
        print(f"Secrets in this page: {secrets_in_page}")
        print()
    
    # Summary
    print("=" * 80)
    print("Summary:")
    print(f"  Total pages processed:    {page_count}")
    print(f"  Total secrets found:      {total_secrets}")
    print(f"  Enabled secrets shown:    {enabled_secrets}")
    print(f"  Disabled secrets skipped: {total_secrets - enabled_secrets}")


def demonstrate_direct_iteration():
    """
    Alternative: Direct iteration without explicit pagination.
    
    This shows the simpler approach where ItemPaged handles pagination
    automatically behind the scenes. Use this when you don't need
    page-level control.
    """
    
    vault_url = os.environ.get("AZURE_KEY_VAULT_URL")
    if not vault_url:
        raise ValueError("AZURE_KEY_VAULT_URL environment variable must be set.")
    
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    print("\n" + "=" * 80)
    print("Alternative: Direct iteration (pagination handled automatically)")
    print("=" * 80)
    print()
    
    # ItemPaged can be iterated directly without calling by_page()
    # Pagination happens automatically behind the scenes
    secret_properties = client.list_properties_of_secrets()
    
    count = 0
    for secret_property in secret_properties:
        if secret_property.enabled:
            count += 1
            print(f"{count}. {secret_property.name} - Created: {secret_property.created_on}")
            
            # For large vaults, you might want to limit output
            if count >= 10:
                print(f"   ... (showing first 10 enabled secrets)")
                break


if __name__ == "__main__":
    try:
        # Main demonstration with by_page()
        list_secrets_with_pagination()
        
        # Show alternative approach
        demonstrate_direct_iteration()
        
    except Exception as e:
        print(f"Error: {e}")
        print("\nMake sure you have:")
        print("1. Set AZURE_KEY_VAULT_URL environment variable")
        print("2. Authenticated with Azure (az login)")
        print("3. Have appropriate permissions on the Key Vault")
