#!/usr/bin/env python3
"""
Azure Key Vault Secrets Pagination Demo

This script demonstrates how to list secrets from an Azure Key Vault with hundreds of secrets
using the ItemPaged pattern and by_page() pagination method from the Azure SDK for Python.

Based on the official Azure SDK for Python documentation:
- https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme
- https://learn.microsoft.com/en-us/python/api/azure-keyvault-secrets/azure.keyvault.secrets.secretclient
- https://learn.microsoft.com/en-us/python/api/azure-core/azure.core.paging.itempaged

Requirements:
    pip install azure-keyvault-secrets azure-identity

Environment Variables:
    VAULT_URL: The URL of your Azure Key Vault (e.g., https://my-key-vault.vault.azure.net/)
"""

import os
from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient


def main():
    # Get vault URL from environment variable
    vault_url = os.environ.get("VAULT_URL")
    if not vault_url:
        print("ERROR: VAULT_URL environment variable not set")
        print("Example: export VAULT_URL='https://my-key-vault.vault.azure.net/'")
        return

    # Create credential and secret client using DefaultAzureCredential
    # DefaultAzureCredential supports multiple authentication methods:
    # - Environment variables (AZURE_CLIENT_ID, AZURE_CLIENT_SECRET, AZURE_TENANT_ID)
    # - Managed Identity
    # - Azure CLI authentication
    # - Visual Studio Code
    # - Azure PowerShell
    print(f"Connecting to Key Vault: {vault_url}")
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)

    # list_properties_of_secrets() returns an ItemPaged[SecretProperties] object
    # ItemPaged is an iterator that automatically handles pagination behind the scenes
    print("\n" + "=" * 80)
    print("METHOD 1: Simple iteration (ItemPaged handles pagination automatically)")
    print("=" * 80)
    
    secret_properties = client.list_properties_of_secrets()
    total_secrets = 0
    enabled_secrets = 0
    
    for secret in secret_properties:
        # Filter to show only enabled secrets
        if secret.enabled:
            total_secrets += 1
            enabled_secrets += 1
            print(f"\nSecret #{enabled_secrets}:")
            print(f"  Name: {secret.name}")
            print(f"  Content Type: {secret.content_type or 'Not set'}")
            print(f"  Created: {secret.created_on}")
            print(f"  Enabled: {secret.enabled}")
        else:
            total_secrets += 1
    
    print(f"\n{total_secrets} total secrets found ({enabled_secrets} enabled, {total_secrets - enabled_secrets} disabled)")

    # Using by_page() to process secrets in pages
    # This is useful for large vaults where you want to process secrets in batches
    print("\n" + "=" * 80)
    print("METHOD 2: Process secrets page by page using by_page()")
    print("=" * 80)
    
    # Get a fresh iterator for the second example
    secret_properties_paged = client.list_properties_of_secrets()
    
    # by_page() returns an iterator of pages, where each page is itself an iterator
    pages = secret_properties_paged.by_page()
    
    page_number = 0
    total_secrets_in_pages = 0
    
    for page in pages:
        page_number += 1
        secrets_in_page = 0
        enabled_in_page = 0
        
        print(f"\n--- Page {page_number} ---")
        
        # Each page is an iterator of SecretProperties objects
        for secret in page:
            secrets_in_page += 1
            total_secrets_in_pages += 1
            
            # Filter to show only enabled secrets
            if secret.enabled:
                enabled_in_page += 1
                print(f"  [{secrets_in_page}] {secret.name}")
                print(f"      Content Type: {secret.content_type or 'Not set'}")
                print(f"      Created: {secret.created_on}")
        
        print(f"Page {page_number} summary: {secrets_in_page} secrets ({enabled_in_page} enabled)")
    
    print(f"\nTotal secrets across all pages: {total_secrets_in_pages}")

    # Demonstrating continuation tokens for resuming pagination
    print("\n" + "=" * 80)
    print("METHOD 3: Using continuation tokens (for resumable pagination)")
    print("=" * 80)
    
    secret_properties_resumable = client.list_properties_of_secrets()
    
    # Get first page
    pages_iter = secret_properties_resumable.by_page()
    first_page = next(pages_iter)
    
    print("First page secrets (enabled only):")
    count = 0
    for secret in first_page:
        if secret.enabled:
            count += 1
            print(f"  {count}. {secret.name} (Created: {secret.created_on})")
    
    # Get continuation token from the iterator
    # In a real scenario, you could save this token and use it later to resume
    try:
        continuation_token = pages_iter.continuation_token
        if continuation_token:
            print(f"\nContinuation token available: {continuation_token[:50]}...")
            
            # Resume from where we left off using the continuation token
            print("\nResuming pagination from continuation token:")
            resumed_pages = client.list_properties_of_secrets().by_page(continuation_token=continuation_token)
            
            next_page = next(resumed_pages, None)
            if next_page:
                print("Next page secrets (first 5, enabled only):")
                count = 0
                for secret in next_page:
                    if secret.enabled and count < 5:
                        count += 1
                        print(f"  {count}. {secret.name}")
        else:
            print("\nNo continuation token (only one page of results)")
    except AttributeError:
        print("\nContinuation token not available in this response")

    print("\n" + "=" * 80)
    print("Pagination demonstration complete!")
    print("=" * 80)
    
    # Clean up - close the credential
    credential.close()


if __name__ == "__main__":
    main()
