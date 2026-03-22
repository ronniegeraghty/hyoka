#!/usr/bin/env python3
"""
Azure Key Vault Secrets CRUD Operations Demo

This script demonstrates all four CRUD operations on Azure Key Vault secrets:
1. Create - Create a new secret
2. Read - Retrieve the secret value
3. Update - Update the secret to a new value
4. Delete - Delete and purge the secret

Prerequisites:
- pip install azure-keyvault-secrets azure-identity
- An Azure Key Vault with soft-delete enabled
- Proper authentication configured for DefaultAzureCredential
- Required permissions: secrets/set, secrets/get, secrets/delete, secrets/purge

Environment Variables:
- AZURE_KEY_VAULT_URL: Your Key Vault URL (e.g., https://my-vault.vault.azure.net/)
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
    """Main function to demonstrate Key Vault CRUD operations."""
    
    # Get Key Vault URL from environment variable
    vault_url = os.environ.get("AZURE_KEY_VAULT_URL")
    if not vault_url:
        print("Error: AZURE_KEY_VAULT_URL environment variable is not set")
        print("Example: export AZURE_KEY_VAULT_URL='https://my-vault.vault.azure.net/'")
        sys.exit(1)
    
    print(f"Connecting to Key Vault: {vault_url}\n")
    
    try:
        # Initialize credential and client
        credential = DefaultAzureCredential()
        client = SecretClient(vault_url=vault_url, credential=credential)
        
        secret_name = "my-secret"
        
        # ================================================================
        # 1. CREATE - Create a new secret
        # ================================================================
        print("=" * 60)
        print("1. CREATE OPERATION")
        print("=" * 60)
        
        try:
            secret_value = "my-secret-value"
            print(f"Creating secret '{secret_name}' with value '{secret_value}'...")
            
            created_secret = client.set_secret(secret_name, secret_value)
            
            print(f"✓ Secret created successfully!")
            print(f"  Name: {created_secret.name}")
            print(f"  Value: {created_secret.value}")
            print(f"  Version: {created_secret.properties.version}")
            print(f"  Created: {created_secret.properties.created_on}")
            print()
            
        except HttpResponseError as e:
            print(f"✗ Failed to create secret: {e.message}")
            sys.exit(1)
        
        # ================================================================
        # 2. READ - Retrieve the secret
        # ================================================================
        print("=" * 60)
        print("2. READ OPERATION")
        print("=" * 60)
        
        try:
            print(f"Reading secret '{secret_name}'...")
            
            retrieved_secret = client.get_secret(secret_name)
            
            print(f"✓ Secret retrieved successfully!")
            print(f"  Name: {retrieved_secret.name}")
            print(f"  Value: {retrieved_secret.value}")
            print(f"  Version: {retrieved_secret.properties.version}")
            print(f"  Content Type: {retrieved_secret.properties.content_type}")
            print(f"  Enabled: {retrieved_secret.properties.enabled}")
            print()
            
        except ResourceNotFoundError:
            print(f"✗ Secret '{secret_name}' not found")
            sys.exit(1)
        except HttpResponseError as e:
            print(f"✗ Failed to read secret: {e.message}")
            sys.exit(1)
        
        # ================================================================
        # 3. UPDATE - Update the secret value
        # ================================================================
        print("=" * 60)
        print("3. UPDATE OPERATION")
        print("=" * 60)
        
        try:
            new_value = "updated-value"
            print(f"Updating secret '{secret_name}' to new value '{new_value}'...")
            
            # Note: set_secret creates a new version when updating
            updated_secret = client.set_secret(secret_name, new_value)
            
            print(f"✓ Secret updated successfully!")
            print(f"  Name: {updated_secret.name}")
            print(f"  New Value: {updated_secret.value}")
            print(f"  New Version: {updated_secret.properties.version}")
            print(f"  Updated: {updated_secret.properties.updated_on}")
            print()
            
            # Verify the update by reading it back
            print(f"Verifying update by reading secret again...")
            verified_secret = client.get_secret(secret_name)
            print(f"✓ Verified value: {verified_secret.value}")
            print()
            
        except HttpResponseError as e:
            print(f"✗ Failed to update secret: {e.message}")
            sys.exit(1)
        
        # ================================================================
        # 4. DELETE - Delete and purge the secret
        # ================================================================
        print("=" * 60)
        print("4. DELETE OPERATION")
        print("=" * 60)
        
        try:
            print(f"Deleting secret '{secret_name}'...")
            
            # Begin delete operation (returns a poller for soft-delete vaults)
            delete_poller = client.begin_delete_secret(secret_name)
            
            # Wait for deletion to complete
            deleted_secret = delete_poller.result()
            
            print(f"✓ Secret deleted successfully!")
            print(f"  Name: {deleted_secret.name}")
            print(f"  Deleted Date: {deleted_secret.deleted_date}")
            print(f"  Scheduled Purge Date: {deleted_secret.scheduled_purge_date}")
            print(f"  Recovery ID: {deleted_secret.recovery_id}")
            print()
            
        except HttpResponseError as e:
            print(f"✗ Failed to delete secret: {e.message}")
            sys.exit(1)
        
        # ================================================================
        # PURGE - Permanently delete the secret (soft-delete enabled)
        # ================================================================
        print("=" * 60)
        print("5. PURGE OPERATION (Permanent Deletion)")
        print("=" * 60)
        
        try:
            print(f"Purging deleted secret '{secret_name}' permanently...")
            
            # Purge the deleted secret (permanent deletion)
            client.purge_deleted_secret(secret_name)
            
            print(f"✓ Secret purged successfully!")
            print(f"  The secret '{secret_name}' has been permanently deleted.")
            print(f"  It cannot be recovered.")
            print()
            
        except HttpResponseError as e:
            # If purge is not supported (e.g., vault doesn't have soft-delete enabled)
            print(f"Note: Purge operation failed or not needed: {e.message}")
            print(f"This may happen if soft-delete is not enabled on the vault.")
            print()
        
        print("=" * 60)
        print("All CRUD operations completed successfully!")
        print("=" * 60)
        
    except ClientAuthenticationError as e:
        print(f"\n✗ Authentication failed: {e.message}")
        print("\nTroubleshooting:")
        print("- Ensure you are logged in with Azure CLI: az login")
        print("- Or set up other authentication methods for DefaultAzureCredential")
        print("- Check that your account has proper permissions on the Key Vault")
        sys.exit(1)
        
    except Exception as e:
        print(f"\n✗ Unexpected error: {type(e).__name__}: {str(e)}")
        sys.exit(1)
    
    finally:
        # Close the client connection
        client.close()
        credential.close()


if __name__ == "__main__":
    main()
