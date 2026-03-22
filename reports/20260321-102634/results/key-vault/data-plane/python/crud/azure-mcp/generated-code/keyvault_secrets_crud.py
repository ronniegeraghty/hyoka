#!/usr/bin/env python3
"""
Azure Key Vault Secrets CRUD Operations Demo

This script demonstrates all four CRUD operations on Azure Key Vault secrets:
1. Create - Set a new secret
2. Read - Get the secret value
3. Update - Change the secret value
4. Delete - Delete and purge the secret

Prerequisites:
- Azure Key Vault with soft-delete enabled
- Proper authentication configured (Azure CLI login, managed identity, etc.)
- Required permissions: secrets/set, secrets/get, secrets/delete, secrets/purge

Environment Variables Required:
- VAULT_URL: Your Key Vault URL (e.g., https://my-key-vault.vault.azure.net/)
"""

import os
import sys
import time
from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient
from azure.core.exceptions import (
    ResourceNotFoundError,
    HttpResponseError,
    ServiceRequestError
)


def main():
    """Main function demonstrating CRUD operations on Azure Key Vault secrets."""
    
    # Get vault URL from environment variable
    vault_url = os.environ.get("VAULT_URL")
    if not vault_url:
        print("Error: VAULT_URL environment variable is not set.")
        print("Example: export VAULT_URL='https://my-key-vault.vault.azure.net/'")
        sys.exit(1)
    
    print(f"Using Key Vault: {vault_url}\n")
    
    # Initialize credential and client
    try:
        credential = DefaultAzureCredential()
        client = SecretClient(vault_url=vault_url, credential=credential)
        print("✓ Successfully initialized SecretClient with DefaultAzureCredential\n")
    except Exception as e:
        print(f"Error initializing client: {e}")
        sys.exit(1)
    
    secret_name = "my-secret"
    
    try:
        # ==================== CREATE ====================
        print("=" * 60)
        print("1. CREATE - Setting a new secret")
        print("=" * 60)
        
        initial_value = "my-secret-value"
        try:
            secret = client.set_secret(secret_name, initial_value)
            print(f"✓ Secret created successfully!")
            print(f"  Name: {secret.name}")
            print(f"  Value: {secret.value}")
            print(f"  Version: {secret.properties.version}")
            print(f"  Created on: {secret.properties.created_on}")
            print()
        except HttpResponseError as e:
            print(f"✗ Failed to create secret: {e.message}")
            raise
        
        # ==================== READ ====================
        print("=" * 60)
        print("2. READ - Retrieving the secret")
        print("=" * 60)
        
        try:
            retrieved_secret = client.get_secret(secret_name)
            print(f"✓ Secret retrieved successfully!")
            print(f"  Name: {retrieved_secret.name}")
            print(f"  Value: {retrieved_secret.value}")
            print(f"  Version: {retrieved_secret.properties.version}")
            print(f"  Content Type: {retrieved_secret.properties.content_type}")
            print(f"  Enabled: {retrieved_secret.properties.enabled}")
            print()
        except ResourceNotFoundError:
            print(f"✗ Secret '{secret_name}' not found in the vault")
            raise
        except HttpResponseError as e:
            print(f"✗ Failed to retrieve secret: {e.message}")
            raise
        
        # ==================== UPDATE ====================
        print("=" * 60)
        print("3. UPDATE - Updating the secret value")
        print("=" * 60)
        
        new_value = "updated-value"
        try:
            # Setting a secret with an existing name creates a new version
            updated_secret = client.set_secret(secret_name, new_value)
            print(f"✓ Secret updated successfully!")
            print(f"  Name: {updated_secret.name}")
            print(f"  New Value: {updated_secret.value}")
            print(f"  New Version: {updated_secret.properties.version}")
            print(f"  Updated on: {updated_secret.properties.updated_on}")
            print()
            
            # Verify the update by reading it back
            verify_secret = client.get_secret(secret_name)
            print(f"✓ Verified updated value: {verify_secret.value}")
            print()
        except HttpResponseError as e:
            print(f"✗ Failed to update secret: {e.message}")
            raise
        
        # ==================== DELETE & PURGE ====================
        print("=" * 60)
        print("4. DELETE - Deleting and purging the secret")
        print("=" * 60)
        
        try:
            # Begin delete operation (returns a poller)
            print(f"Initiating delete operation for '{secret_name}'...")
            delete_poller = client.begin_delete_secret(secret_name)
            
            # Wait for deletion to complete
            deleted_secret = delete_poller.result()
            print(f"✓ Secret deleted successfully!")
            print(f"  Name: {deleted_secret.name}")
            print(f"  Deleted on: {deleted_secret.deleted_date}")
            print(f"  Scheduled purge date: {deleted_secret.scheduled_purge_date}")
            print(f"  Recovery ID: {deleted_secret.recovery_id}")
            print()
            
            # Purge the deleted secret (permanent deletion)
            print(f"Purging deleted secret '{secret_name}'...")
            client.purge_deleted_secret(secret_name)
            print(f"✓ Secret purged successfully!")
            print(f"  The secret has been permanently deleted and cannot be recovered.")
            print()
            
            # Wait a moment for purge to complete
            time.sleep(2)
            
            # Verify the secret is gone
            print("Verifying secret deletion...")
            try:
                client.get_secret(secret_name)
                print("✗ Warning: Secret still exists (unexpected)")
            except ResourceNotFoundError:
                print("✓ Confirmed: Secret no longer exists in the vault")
            
        except HttpResponseError as e:
            if "not currently in a deleted state" in str(e):
                print(f"Note: Secret may already be deleted. Error: {e.message}")
            else:
                print(f"✗ Failed to delete/purge secret: {e.message}")
                raise
        
        print()
        print("=" * 60)
        print("✓ All CRUD operations completed successfully!")
        print("=" * 60)
        
    except ServiceRequestError as e:
        print(f"\n✗ Service request error: {e}")
        print("Check your network connection and vault URL.")
        sys.exit(1)
    except Exception as e:
        print(f"\n✗ Unexpected error: {e}")
        sys.exit(1)
    finally:
        # Clean up resources
        client.close()
        credential.close()


if __name__ == "__main__":
    main()
