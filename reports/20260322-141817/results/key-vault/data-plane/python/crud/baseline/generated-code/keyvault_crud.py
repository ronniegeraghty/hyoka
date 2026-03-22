"""
Azure Key Vault CRUD Operations Demo
Demonstrates Create, Read, Update, and Delete operations for secrets.

Required packages:
    pip install azure-keyvault-secrets azure-identity
"""

import os
import sys
from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient
from azure.core.exceptions import ResourceNotFoundError, HttpResponseError


def main():
    # Get the Key Vault URL from environment variable
    vault_url = os.environ.get("VAULT_URL")
    
    if not vault_url:
        print("Error: VAULT_URL environment variable is not set")
        print("Please set it to your Key Vault URL (e.g., https://my-vault.vault.azure.net/)")
        sys.exit(1)
    
    try:
        # Authenticate using DefaultAzureCredential
        credential = DefaultAzureCredential()
        
        # Create the SecretClient
        client = SecretClient(vault_url=vault_url, credential=credential)
        
        print(f"Connected to Key Vault: {vault_url}\n")
        
        # 1. CREATE - Set a new secret
        print("1. CREATE: Creating secret 'my-secret'...")
        try:
            secret = client.set_secret("my-secret", "my-secret-value")
            print(f"   ✓ Secret created successfully")
            print(f"   - Name: {secret.name}")
            print(f"   - Version: {secret.properties.version}")
            print()
        except HttpResponseError as e:
            print(f"   ✗ Failed to create secret: {e.message}")
            sys.exit(1)
        
        # 2. READ - Retrieve the secret
        print("2. READ: Reading secret 'my-secret'...")
        try:
            retrieved_secret = client.get_secret("my-secret")
            print(f"   ✓ Secret retrieved successfully")
            print(f"   - Name: {retrieved_secret.name}")
            print(f"   - Value: {retrieved_secret.value}")
            print(f"   - Version: {retrieved_secret.properties.version}")
            print()
        except ResourceNotFoundError:
            print(f"   ✗ Secret 'my-secret' not found")
            sys.exit(1)
        except HttpResponseError as e:
            print(f"   ✗ Failed to retrieve secret: {e.message}")
            sys.exit(1)
        
        # 3. UPDATE - Update the secret value
        print("3. UPDATE: Updating secret 'my-secret' to new value...")
        try:
            updated_secret = client.set_secret("my-secret", "updated-value")
            print(f"   ✓ Secret updated successfully")
            print(f"   - Name: {updated_secret.name}")
            print(f"   - New Value: {updated_secret.value}")
            print(f"   - New Version: {updated_secret.properties.version}")
            print()
        except HttpResponseError as e:
            print(f"   ✗ Failed to update secret: {e.message}")
            sys.exit(1)
        
        # 4. DELETE - Delete and purge the secret
        print("4. DELETE: Deleting secret 'my-secret'...")
        try:
            # Begin delete operation (returns a poller)
            delete_poller = client.begin_delete_secret("my-secret")
            deleted_secret = delete_poller.result()  # Wait for deletion to complete
            print(f"   ✓ Secret deleted successfully")
            print(f"   - Name: {deleted_secret.name}")
            print(f"   - Deleted Date: {deleted_secret.deleted_date}")
            print(f"   - Scheduled Purge Date: {deleted_secret.scheduled_purge_date}")
            print()
        except ResourceNotFoundError:
            print(f"   ✗ Secret 'my-secret' not found")
            sys.exit(1)
        except HttpResponseError as e:
            print(f"   ✗ Failed to delete secret: {e.message}")
            sys.exit(1)
        
        # PURGE - Permanently delete the secret (soft-delete enabled vaults)
        print("   PURGE: Permanently purging deleted secret 'my-secret'...")
        try:
            client.purge_deleted_secret("my-secret")
            print(f"   ✓ Secret purged successfully (permanently deleted)")
            print()
        except ResourceNotFoundError:
            print(f"   ✗ Deleted secret 'my-secret' not found for purging")
            sys.exit(1)
        except HttpResponseError as e:
            print(f"   ✗ Failed to purge secret: {e.message}")
            print(f"   Note: Purging may fail if soft-delete is not enabled or vault doesn't support purge")
            sys.exit(1)
        
        print("All CRUD operations completed successfully!")
        
    except Exception as e:
        print(f"Unexpected error: {str(e)}")
        sys.exit(1)


if __name__ == "__main__":
    main()
