#!/usr/bin/env python3
"""
Azure Key Vault Secrets CRUD Operations Demo

This script demonstrates all four CRUD operations on Azure Key Vault secrets:
1. Create - Set a new secret
2. Read - Retrieve the secret value
3. Update - Modify the secret value
4. Delete - Delete and purge the secret

Required pip packages:
    pip install azure-keyvault-secrets azure-identity

Prerequisites:
    - Azure subscription
    - Azure Key Vault with soft-delete enabled
    - Proper authentication configured (Azure CLI login, managed identity, etc.)
    - Key Vault access policies granting: secrets/set, secrets/get, secrets/delete, secrets/purge
    - Environment variable VAULT_URL set to your Key Vault URL
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
    """Perform CRUD operations on Azure Key Vault secrets."""
    
    # Get vault URL from environment variable
    vault_url = os.environ.get("VAULT_URL")
    if not vault_url:
        print("Error: VAULT_URL environment variable not set")
        print("Example: export VAULT_URL='https://my-key-vault.vault.azure.net/'")
        sys.exit(1)
    
    print(f"Connecting to Key Vault: {vault_url}")
    
    try:
        # Authenticate using DefaultAzureCredential
        credential = DefaultAzureCredential()
        
        # Create the SecretClient
        client = SecretClient(vault_url=vault_url, credential=credential)
        
        secret_name = "my-secret"
        
        # 1. CREATE - Set a new secret
        print("\n" + "="*60)
        print("1. CREATE - Setting a new secret")
        print("="*60)
        try:
            secret = client.set_secret(secret_name, "my-secret-value")
            print(f"✓ Secret created successfully")
            print(f"  Name: {secret.name}")
            print(f"  Value: {secret.value}")
            print(f"  Version: {secret.properties.version}")
            print(f"  Created: {secret.properties.created_on}")
        except HttpResponseError as e:
            print(f"✗ Failed to create secret: {e.message}")
            raise
        
        # 2. READ - Retrieve the secret
        print("\n" + "="*60)
        print("2. READ - Retrieving the secret")
        print("="*60)
        try:
            retrieved_secret = client.get_secret(secret_name)
            print(f"✓ Secret retrieved successfully")
            print(f"  Name: {retrieved_secret.name}")
            print(f"  Value: {retrieved_secret.value}")
            print(f"  Version: {retrieved_secret.properties.version}")
        except ResourceNotFoundError:
            print(f"✗ Secret '{secret_name}' not found")
            raise
        except HttpResponseError as e:
            print(f"✗ Failed to retrieve secret: {e.message}")
            raise
        
        # 3. UPDATE - Set a new value (creates a new version)
        print("\n" + "="*60)
        print("3. UPDATE - Updating the secret value")
        print("="*60)
        try:
            updated_secret = client.set_secret(secret_name, "updated-value")
            print(f"✓ Secret updated successfully")
            print(f"  Name: {updated_secret.name}")
            print(f"  New Value: {updated_secret.value}")
            print(f"  New Version: {updated_secret.properties.version}")
            print(f"  Updated: {updated_secret.properties.updated_on}")
        except HttpResponseError as e:
            print(f"✗ Failed to update secret: {e.message}")
            raise
        
        # 4. DELETE - Delete and purge the secret
        print("\n" + "="*60)
        print("4. DELETE - Deleting and purging the secret")
        print("="*60)
        try:
            # Begin delete operation (returns a poller)
            print(f"  Initiating deletion of '{secret_name}'...")
            delete_operation = client.begin_delete_secret(secret_name)
            
            # Wait for deletion to complete
            deleted_secret = delete_operation.result()
            print(f"✓ Secret deleted successfully")
            print(f"  Name: {deleted_secret.name}")
            print(f"  Deleted Date: {deleted_secret.deleted_date}")
            print(f"  Scheduled Purge Date: {deleted_secret.scheduled_purge_date}")
            print(f"  Recovery ID: {deleted_secret.recovery_id}")
            
            # Purge the deleted secret (permanent deletion)
            print(f"\n  Purging deleted secret '{secret_name}'...")
            client.purge_deleted_secret(secret_name)
            print(f"✓ Secret purged successfully (permanently deleted)")
            
        except HttpResponseError as e:
            print(f"✗ Failed to delete/purge secret: {e.message}")
            raise
        
        print("\n" + "="*60)
        print("All CRUD operations completed successfully!")
        print("="*60)
        
    except ClientAuthenticationError as e:
        print(f"\n✗ Authentication failed: {e.message}")
        print("  Make sure you're logged in via Azure CLI or have proper credentials configured")
        sys.exit(1)
    except HttpResponseError as e:
        print(f"\n✗ HTTP error occurred: {e.message}")
        print(f"  Status code: {e.status_code}")
        sys.exit(1)
    except Exception as e:
        print(f"\n✗ Unexpected error: {str(e)}")
        sys.exit(1)
    finally:
        # Close the credential to clean up resources
        credential.close()


if __name__ == "__main__":
    main()
