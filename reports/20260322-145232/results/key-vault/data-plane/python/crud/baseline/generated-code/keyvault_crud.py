#!/usr/bin/env python3
"""
Azure Key Vault Secrets CRUD Operations Script

This script demonstrates all four CRUD operations on Azure Key Vault secrets:
1. Create - Set a new secret
2. Read - Get a secret value
3. Update - Update secret to a new value
4. Delete - Delete and purge a secret (soft-delete enabled vault)

Required packages:
    pip install azure-keyvault-secrets azure-identity

Prerequisites:
    - Azure subscription with a Key Vault created
    - Set environment variable AZURE_KEY_VAULT_URL with your vault URL
    - Appropriate authentication configured for DefaultAzureCredential
"""

import os
import sys
from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient
from azure.core.exceptions import (
    ResourceNotFoundError,
    HttpResponseError,
    ClientAuthenticationError
)


def main():
    """Main function to demonstrate CRUD operations on Azure Key Vault secrets."""
    
    # Get the Key Vault URL from environment variable
    vault_url = os.environ.get("AZURE_KEY_VAULT_URL")
    if not vault_url:
        print("Error: AZURE_KEY_VAULT_URL environment variable is not set.")
        print("Please set it to your Key Vault URL (e.g., https://my-key-vault.vault.azure.net/)")
        sys.exit(1)
    
    print(f"Using Key Vault: {vault_url}\n")
    
    # Initialize credential and client
    try:
        credential = DefaultAzureCredential()
        secret_client = SecretClient(vault_url=vault_url, credential=credential)
        print("✓ Successfully initialized SecretClient with DefaultAzureCredential\n")
    except ClientAuthenticationError as e:
        print(f"Authentication failed: {e.message}")
        sys.exit(1)
    except Exception as e:
        print(f"Error initializing client: {e}")
        sys.exit(1)
    
    secret_name = "my-secret"
    
    # ========================================
    # 1. CREATE - Set a new secret
    # ========================================
    print("=" * 60)
    print("1. CREATE - Setting a new secret")
    print("=" * 60)
    try:
        initial_value = "my-secret-value"
        secret = secret_client.set_secret(secret_name, initial_value)
        print(f"✓ Secret created successfully")
        print(f"  Name: {secret.name}")
        print(f"  Value: {secret.value}")
        print(f"  Version: {secret.properties.version}")
        print(f"  Created: {secret.properties.created_on}\n")
    except HttpResponseError as e:
        print(f"✗ HTTP error creating secret: {e.message}")
        print(f"  Status code: {e.status_code}")
        sys.exit(1)
    except Exception as e:
        print(f"✗ Error creating secret: {e}")
        sys.exit(1)
    
    # ========================================
    # 2. READ - Get the secret value
    # ========================================
    print("=" * 60)
    print("2. READ - Retrieving the secret")
    print("=" * 60)
    try:
        retrieved_secret = secret_client.get_secret(secret_name)
        print(f"✓ Secret retrieved successfully")
        print(f"  Name: {retrieved_secret.name}")
        print(f"  Value: {retrieved_secret.value}")
        print(f"  Version: {retrieved_secret.properties.version}")
        print(f"  Enabled: {retrieved_secret.properties.enabled}\n")
    except ResourceNotFoundError:
        print(f"✗ Secret '{secret_name}' not found in the vault")
        sys.exit(1)
    except HttpResponseError as e:
        print(f"✗ HTTP error retrieving secret: {e.message}")
        sys.exit(1)
    except Exception as e:
        print(f"✗ Error retrieving secret: {e}")
        sys.exit(1)
    
    # ========================================
    # 3. UPDATE - Update the secret value
    # ========================================
    print("=" * 60)
    print("3. UPDATE - Updating the secret to a new value")
    print("=" * 60)
    try:
        new_value = "updated-value"
        updated_secret = secret_client.set_secret(secret_name, new_value)
        print(f"✓ Secret updated successfully")
        print(f"  Name: {updated_secret.name}")
        print(f"  New Value: {updated_secret.value}")
        print(f"  New Version: {updated_secret.properties.version}")
        print(f"  Updated: {updated_secret.properties.updated_on}\n")
    except HttpResponseError as e:
        print(f"✗ HTTP error updating secret: {e.message}")
        sys.exit(1)
    except Exception as e:
        print(f"✗ Error updating secret: {e}")
        sys.exit(1)
    
    # ========================================
    # 4. DELETE - Delete and purge the secret
    # ========================================
    print("=" * 60)
    print("4. DELETE - Deleting and purging the secret")
    print("=" * 60)
    try:
        # Delete the secret (soft delete)
        print("Step 1: Deleting secret (soft delete)...")
        delete_poller = secret_client.begin_delete_secret(secret_name)
        deleted_secret = delete_poller.result()
        print(f"✓ Secret deleted successfully")
        print(f"  Name: {deleted_secret.name}")
        print(f"  Deleted Date: {deleted_secret.deleted_date}")
        print(f"  Scheduled Purge Date: {deleted_secret.scheduled_purge_date}")
        print(f"  Recovery ID: {deleted_secret.recovery_id}\n")
        
        # Purge the secret (permanent deletion)
        print("Step 2: Purging secret (permanent deletion)...")
        secret_client.purge_deleted_secret(secret_name)
        print(f"✓ Secret purged successfully")
        print(f"  The secret '{secret_name}' has been permanently deleted.\n")
        
    except ResourceNotFoundError:
        print(f"✗ Secret '{secret_name}' not found for deletion")
        sys.exit(1)
    except HttpResponseError as e:
        print(f"✗ HTTP error during delete/purge: {e.message}")
        print(f"  Status code: {e.status_code}")
        if e.status_code == 403:
            print("  Note: Ensure you have 'secrets/delete' and 'secrets/purge' permissions")
        sys.exit(1)
    except Exception as e:
        print(f"✗ Error deleting/purging secret: {e}")
        sys.exit(1)
    
    print("=" * 60)
    print("All CRUD operations completed successfully!")
    print("=" * 60)


if __name__ == "__main__":
    main()
