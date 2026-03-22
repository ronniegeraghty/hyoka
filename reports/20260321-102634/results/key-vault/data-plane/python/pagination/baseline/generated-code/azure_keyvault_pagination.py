#!/usr/bin/env python3
"""
Azure Key Vault Secrets Pagination Demo

This script demonstrates how to handle pagination when listing secrets
in an Azure Key Vault that contains hundreds of secrets.

Required packages:
    pip install azure-keyvault-secrets azure-identity

Prerequisites:
    - An Azure Key Vault with secrets
    - Appropriate authentication configured for DefaultAzureCredential
    - Environment variable VAULT_URL set to your Key Vault URL
      (e.g., https://my-vault.vault.azure.net/)
"""

import os
from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient


def list_secrets_basic_iteration():
    """
    Basic iteration over secrets using the ItemPaged pattern.
    The SDK handles pagination automatically behind the scenes.
    """
    print("=" * 80)
    print("Method 1: Basic Iteration (Automatic Pagination)")
    print("=" * 80)
    
    vault_url = os.environ.get("VAULT_URL")
    if not vault_url:
        print("ERROR: VAULT_URL environment variable not set")
        return
    
    # Create a SecretClient using DefaultAzureCredential
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    # list_properties_of_secrets() returns an ItemPaged[SecretProperties] object
    # This allows iteration over all secrets, with automatic pagination
    secret_properties = client.list_properties_of_secrets()
    
    secret_count = 0
    enabled_count = 0
    
    # Iterate over all secrets - pagination is handled automatically
    for secret_property in secret_properties:
        # Filter to show only enabled secrets
        if secret_property.enabled:
            enabled_count += 1
            print(f"\nSecret #{enabled_count}:")
            print(f"  Name: {secret_property.name}")
            print(f"  Content Type: {secret_property.content_type or 'Not set'}")
            print(f"  Created On: {secret_property.created_on}")
            print(f"  Enabled: {secret_property.enabled}")
        
        secret_count += 1
    
    print(f"\n{'-' * 80}")
    print(f"Total secrets: {secret_count}")
    print(f"Enabled secrets: {enabled_count}")
    print(f"Disabled secrets: {secret_count - enabled_count}")


def list_secrets_by_page():
    """
    Process secrets page by page using the by_page() method.
    This approach gives you explicit control over pagination.
    """
    print("\n" + "=" * 80)
    print("Method 2: Page-by-Page Iteration (Explicit Pagination)")
    print("=" * 80)
    
    vault_url = os.environ.get("VAULT_URL")
    if not vault_url:
        print("ERROR: VAULT_URL environment variable not set")
        return
    
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    # Get an ItemPaged object
    secret_properties = client.list_properties_of_secrets()
    
    # Use by_page() to iterate page by page
    # This returns an iterator of pages, where each page is itself an iterator
    page_iterator = secret_properties.by_page()
    
    page_number = 0
    total_secrets = 0
    total_enabled = 0
    
    for page in page_iterator:
        page_number += 1
        page_secrets = list(page)
        page_enabled = sum(1 for s in page_secrets if s.enabled)
        
        print(f"\n--- Page {page_number} ---")
        print(f"Secrets in this page: {len(page_secrets)}")
        print(f"Enabled secrets in this page: {page_enabled}")
        
        # Process each secret in the page
        for secret_property in page_secrets:
            if secret_property.enabled:
                print(f"  - {secret_property.name} "
                      f"(Content Type: {secret_property.content_type or 'Not set'}, "
                      f"Created: {secret_property.created_on})")
        
        total_secrets += len(page_secrets)
        total_enabled += page_enabled
    
    print(f"\n{'-' * 80}")
    print(f"Total pages processed: {page_number}")
    print(f"Total secrets: {total_secrets}")
    print(f"Total enabled secrets: {total_enabled}")


def list_secrets_with_continuation_token():
    """
    Demonstrates using continuation tokens for pagination.
    This is useful for resuming iteration or implementing custom pagination.
    """
    print("\n" + "=" * 80)
    print("Method 3: Using Continuation Tokens")
    print("=" * 80)
    
    vault_url = os.environ.get("VAULT_URL")
    if not vault_url:
        print("ERROR: VAULT_URL environment variable not set")
        return
    
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    secret_properties = client.list_properties_of_secrets()
    
    # Start iteration with no continuation token
    page_iterator = secret_properties.by_page()
    
    page_number = 0
    total_secrets = 0
    
    for page in page_iterator:
        page_number += 1
        page_secrets = list(page)
        enabled_in_page = [s for s in page_secrets if s.enabled]
        
        print(f"\n--- Page {page_number} ---")
        print(f"Total secrets in page: {len(page_secrets)}")
        print(f"Enabled secrets in page: {len(enabled_in_page)}")
        
        # Show first 3 enabled secrets from this page
        for i, secret_property in enumerate(enabled_in_page[:3], 1):
            print(f"  {i}. Name: {secret_property.name}")
            print(f"     Content Type: {secret_property.content_type or 'Not set'}")
            print(f"     Created On: {secret_property.created_on}")
        
        if len(enabled_in_page) > 3:
            print(f"  ... and {len(enabled_in_page) - 3} more enabled secrets")
        
        total_secrets += len(page_secrets)
        
        # Get the continuation token for this page
        # Note: continuation_token is available on the page iterator
        try:
            continuation_token = page_iterator.continuation_token
            if continuation_token:
                print(f"\nContinuation token available for next page")
        except AttributeError:
            # continuation_token may not always be available
            pass
    
    print(f"\n{'-' * 80}")
    print(f"Total pages: {page_number}")
    print(f"Total secrets: {total_secrets}")


def demonstrate_secret_properties():
    """
    Show all available properties of SecretProperties objects.
    """
    print("\n" + "=" * 80)
    print("Method 4: Detailed Secret Properties")
    print("=" * 80)
    
    vault_url = os.environ.get("VAULT_URL")
    if not vault_url:
        print("ERROR: VAULT_URL environment variable not set")
        return
    
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    secret_properties = client.list_properties_of_secrets()
    
    # Get first enabled secret for detailed display
    for secret_property in secret_properties:
        if secret_property.enabled:
            print(f"\nDetailed properties for secret: {secret_property.name}")
            print(f"{'Property':<20} {'Value'}")
            print("-" * 80)
            print(f"{'Name':<20} {secret_property.name}")
            print(f"{'Enabled':<20} {secret_property.enabled}")
            print(f"{'Content Type':<20} {secret_property.content_type or 'Not set'}")
            print(f"{'Created On':<20} {secret_property.created_on}")
            print(f"{'Updated On':<20} {secret_property.updated_on}")
            print(f"{'Expires On':<20} {secret_property.expires_on or 'No expiration'}")
            print(f"{'Not Before':<20} {secret_property.not_before or 'No restriction'}")
            print(f"{'Version':<20} {secret_property.version}")
            print(f"{'Vault URL':<20} {secret_property.vault_url}")
            print(f"{'Managed':<20} {secret_property.managed}")
            print(f"{'Recovery Level':<20} {secret_property.recovery_level}")
            print(f"{'Recoverable Days':<20} {secret_property.recoverable_days}")
            print(f"{'Tags':<20} {secret_property.tags or 'No tags'}")
            break
    else:
        print("No enabled secrets found in vault")


if __name__ == "__main__":
    print("\nAzure Key Vault Secrets - Pagination Demonstration")
    print("Using SecretClient with DefaultAzureCredential")
    print()
    
    try:
        # Demonstrate different pagination approaches
        list_secrets_basic_iteration()
        list_secrets_by_page()
        list_secrets_with_continuation_token()
        demonstrate_secret_properties()
        
        print("\n" + "=" * 80)
        print("Demonstration Complete!")
        print("=" * 80)
        
    except Exception as e:
        print(f"\nError: {e}")
        print("\nMake sure:")
        print("1. VAULT_URL environment variable is set")
        print("2. You have appropriate authentication configured")
        print("3. You have 'secrets/list' permission on the vault")
