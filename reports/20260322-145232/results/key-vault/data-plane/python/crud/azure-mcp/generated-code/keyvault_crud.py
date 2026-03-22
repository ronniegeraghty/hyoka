#!/usr/bin/env python3
"""
Azure Key Vault CRUD Operations Demo

This script demonstrates all four CRUD operations on Azure Key Vault secrets:
1. Create - Set a new secret
2. Read - Get and display the secret value
3. Update - Modify the secret value
4. Delete - Delete and purge the secret

Required environment variable:
    VAULT_URL: The URL of your Azure Key Vault (e.g., https://my-vault.vault.azure.net/)

Required pip packages:
    pip install azure-keyvault-secrets azure-identity
"""

import os
import sys
from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient
from azure.core.exceptions import ResourceNotFoundError, HttpResponseError


def main():
    """Perform CRUD operations on Azure Key Vault secrets."""
    
    # Get vault URL from environment variable
    vault_url = os.environ.get("VAULT_URL")
    if not vault_url:
        print("Error: VAULT_URL environment variable is not set")
        print("Example: export VAULT_URL='https://my-vault.vault.azure.net/'")
        sys.exit(1)
    
    print(f"Connecting to Key Vault: {vault_url}")
    
    # Initialize the credential and client
    try:
        credential = DefaultAzureCredential()
        client = SecretClient(vault_url=vault_url, credential=credential)
        print("✓ Successfully authenticated\n")
    except Exception as e:
        print(f"Error: Failed to authenticate: {e}")
        sys.exit(1)
    
    secret_name = "my-secret"
    
    # ========================================
    # CREATE: Set a new secret
    # ========================================
    try:
        print("=" * 50)
        print("1. CREATE - Setting secret")
        print("=" * 50)
        
        secret_value = "my-secret-value"
        secret = client.set_secret(secret_name, secret_value)
        
        print(f"✓ Secret created successfully")
        print(f"  Name: {secret.name}")
        print(f"  Value: {secret.value}")
        print(f"  Version: {secret.properties.version}")
        print()
        
    except HttpResponseError as e:
        print(f"✗ Error creating secret: {e.message}")
        sys.exit(1)
    except Exception as e:
        print(f"✗ Unexpected error: {e}")
        sys.exit(1)
    
    # ========================================
    # READ: Get the secret value
    # ========================================
    try:
        print("=" * 50)
        print("2. READ - Retrieving secret")
        print("=" * 50)
        
        retrieved_secret = client.get_secret(secret_name)
        
        print(f"✓ Secret retrieved successfully")
        print(f"  Name: {retrieved_secret.name}")
        print(f"  Value: {retrieved_secret.value}")
        print(f"  Content Type: {retrieved_secret.properties.content_type}")
        print(f"  Enabled: {retrieved_secret.properties.enabled}")
        print()
        
    except ResourceNotFoundError:
        print(f"✗ Error: Secret '{secret_name}' not found")
        sys.exit(1)
    except HttpResponseError as e:
        print(f"✗ Error retrieving secret: {e.message}")
        sys.exit(1)
    except Exception as e:
        print(f"✗ Unexpected error: {e}")
        sys.exit(1)
    
    # ========================================
    # UPDATE: Change the secret value
    # ========================================
    try:
        print("=" * 50)
        print("3. UPDATE - Updating secret value")
        print("=" * 50)
        
        new_value = "updated-value"
        updated_secret = client.set_secret(secret_name, new_value)
        
        print(f"✓ Secret updated successfully")
        print(f"  Name: {updated_secret.name}")
        print(f"  New Value: {updated_secret.value}")
        print(f"  New Version: {updated_secret.properties.version}")
        print()
        
    except HttpResponseError as e:
        print(f"✗ Error updating secret: {e.message}")
        sys.exit(1)
    except Exception as e:
        print(f"✗ Unexpected error: {e}")
        sys.exit(1)
    
    # ========================================
    # DELETE: Delete and purge the secret
    # ========================================
    try:
        print("=" * 50)
        print("4. DELETE - Deleting secret")
        print("=" * 50)
        
        # Begin delete operation (returns a poller)
        delete_poller = client.begin_delete_secret(secret_name)
        
        # Wait for the deletion to complete
        deleted_secret = delete_poller.result()
        
        print(f"✓ Secret deleted successfully")
        print(f"  Name: {deleted_secret.name}")
        print(f"  Deleted Date: {deleted_secret.deleted_date}")
        print(f"  Scheduled Purge Date: {deleted_secret.scheduled_purge_date}")
        print(f"  Recovery ID: {deleted_secret.recovery_id}")
        print()
        
    except HttpResponseError as e:
        print(f"✗ Error deleting secret: {e.message}")
        sys.exit(1)
    except Exception as e:
        print(f"✗ Unexpected error: {e}")
        sys.exit(1)
    
    # ========================================
    # PURGE: Permanently delete the secret
    # ========================================
    try:
        print("=" * 50)
        print("5. PURGE - Permanently deleting secret")
        print("=" * 50)
        
        # Purge the deleted secret (permanent deletion)
        client.purge_deleted_secret(secret_name)
        
        print(f"✓ Secret purged successfully")
        print(f"  The secret '{secret_name}' has been permanently deleted")
        print()
        
    except HttpResponseError as e:
        print(f"✗ Error purging secret: {e.message}")
        print("  Note: Vault must have soft-delete enabled for purge operation")
        sys.exit(1)
    except Exception as e:
        print(f"✗ Unexpected error: {e}")
        sys.exit(1)
    
    print("=" * 50)
    print("✓ All CRUD operations completed successfully!")
    print("=" * 50)


if __name__ == "__main__":
    main()
