#!/usr/bin/env python3
"""
Practical Example: Export Key Vault Secrets Inventory

This script demonstrates a real-world use case for pagination:
exporting an inventory of all secrets in a Key Vault to a CSV file.
"""

import os
import csv
from datetime import datetime
from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient


def export_secrets_inventory(vault_url, output_file="secrets_inventory.csv"):
    """
    Export all secrets metadata to a CSV file using pagination.
    
    Args:
        vault_url: The Key Vault URL
        output_file: Output CSV filename
    """
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    print(f"Connecting to: {vault_url}")
    print(f"Exporting to: {output_file}\n")
    
    # Open CSV file for writing
    with open(output_file, 'w', newline='', encoding='utf-8') as csvfile:
        fieldnames = [
            'name', 
            'enabled', 
            'content_type', 
            'created_on', 
            'updated_on',
            'expires_on',
            'tags'
        ]
        writer = csv.DictWriter(csvfile, fieldnames=fieldnames)
        writer.writeheader()
        
        # Process secrets page by page
        pages = client.list_properties_of_secrets().by_page()
        
        total_secrets = 0
        enabled_count = 0
        page_count = 0
        
        for page in pages:
            page_count += 1
            page_secrets = list(page)
            
            print(f"Processing page {page_count}: {len(page_secrets)} secrets")
            
            for secret in page_secrets:
                total_secrets += 1
                
                if secret.enabled:
                    enabled_count += 1
                
                # Write to CSV
                writer.writerow({
                    'name': secret.name,
                    'enabled': secret.enabled,
                    'content_type': secret.content_type or '',
                    'created_on': secret.created_on.isoformat() if secret.created_on else '',
                    'updated_on': secret.updated_on.isoformat() if secret.updated_on else '',
                    'expires_on': secret.expires_on.isoformat() if secret.expires_on else '',
                    'tags': str(secret.tags) if secret.tags else ''
                })
    
    print(f"\n{'='*60}")
    print(f"Export Complete!")
    print(f"{'='*60}")
    print(f"Total secrets:    {total_secrets}")
    print(f"Enabled:          {enabled_count}")
    print(f"Disabled:         {total_secrets - enabled_count}")
    print(f"Pages processed:  {page_count}")
    print(f"Output file:      {output_file}")
    print(f"{'='*60}\n")


def find_expiring_secrets(vault_url, days=30):
    """
    Find secrets that will expire within the specified number of days.
    Demonstrates filtering during pagination.
    
    Args:
        vault_url: The Key Vault URL
        days: Number of days to look ahead for expiring secrets
    """
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    from datetime import timedelta
    cutoff_date = datetime.now(datetime.now().astimezone().tzinfo) + timedelta(days=days)
    
    print(f"Finding secrets expiring before: {cutoff_date.date()}\n")
    
    expiring_secrets = []
    
    # Iterate through all secrets
    for secret in client.list_properties_of_secrets():
        if secret.enabled and secret.expires_on:
            if secret.expires_on <= cutoff_date:
                expiring_secrets.append({
                    'name': secret.name,
                    'expires_on': secret.expires_on,
                    'days_until_expiry': (secret.expires_on - datetime.now(secret.expires_on.tzinfo)).days
                })
    
    if expiring_secrets:
        print(f"Found {len(expiring_secrets)} secret(s) expiring within {days} days:\n")
        for secret in sorted(expiring_secrets, key=lambda x: x['expires_on']):
            print(f"⚠️  {secret['name']}")
            print(f"   Expires: {secret['expires_on'].date()}")
            print(f"   Days remaining: {secret['days_until_expiry']}\n")
    else:
        print(f"✓ No secrets expiring within {days} days")


if __name__ == "__main__":
    vault_url = os.environ.get("VAULT_URL")
    
    if not vault_url:
        print("Error: VAULT_URL environment variable is not set")
        print("\nUsage:")
        print("  export VAULT_URL='https://your-vault.vault.azure.net/'")
        print("  python practical_example.py")
        exit(1)
    
    try:
        # Example 1: Export inventory to CSV
        print("Example 1: Export Secrets Inventory")
        print("="*60 + "\n")
        export_secrets_inventory(vault_url)
        
        print("\n")
        
        # Example 2: Find expiring secrets
        print("Example 2: Find Expiring Secrets")
        print("="*60 + "\n")
        find_expiring_secrets(vault_url, days=90)
        
    except Exception as e:
        print(f"\nError: {e}")
        raise
