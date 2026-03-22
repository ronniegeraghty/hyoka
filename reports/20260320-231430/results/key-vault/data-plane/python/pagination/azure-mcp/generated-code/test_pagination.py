#!/usr/bin/env python3
"""
Simple test to verify Azure Key Vault pagination behavior.

This script provides a minimal example to test the pagination concepts
without all the detailed output of the main script.
"""

import os
from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient


def test_pagination():
    """Quick test of Key Vault pagination."""
    
    vault_url = os.environ.get("AZURE_KEY_VAULT_URL")
    if not vault_url:
        print("Error: AZURE_KEY_VAULT_URL environment variable not set")
        return
    
    try:
        # Setup client
        credential = DefaultAzureCredential()
        client = SecretClient(vault_url=vault_url, credential=credential)
        
        print(f"Testing pagination for: {vault_url}\n")
        
        # Get secret properties (returns ItemPaged object)
        secret_properties = client.list_properties_of_secrets()
        
        # Method 1: Using by_page() for explicit page control
        print("Method 1: Using by_page()")
        print("-" * 40)
        
        page_num = 0
        total_count = 0
        
        for page in secret_properties.by_page():
            page_num += 1
            page_count = 0
            
            for secret_prop in page:
                page_count += 1
                total_count += 1
            
            print(f"Page {page_num}: {page_count} secrets")
        
        print(f"Total secrets: {total_count}\n")
        
        # Method 2: Direct iteration (pagination automatic)
        print("Method 2: Direct iteration")
        print("-" * 40)
        
        secret_properties_2 = client.list_properties_of_secrets()
        
        count = 0
        enabled_count = 0
        
        for secret_prop in secret_properties_2:
            count += 1
            if secret_prop.enabled:
                enabled_count += 1
                
                # Show first 5 as sample
                if enabled_count <= 5:
                    print(f"  {secret_prop.name} (enabled)")
        
        print(f"\nTotal: {count} secrets ({enabled_count} enabled)")
        
        print("\n✓ Pagination test completed successfully")
        
    except Exception as e:
        print(f"Error: {e}")
        print("\nTroubleshooting:")
        print("1. Verify AZURE_KEY_VAULT_URL is correct")
        print("2. Check authentication (try: az login)")
        print("3. Verify Key Vault permissions (list, get secrets)")


if __name__ == "__main__":
    test_pagination()
