"""
Azure Key Vault Secrets CRUD Operations

Required packages:
    pip install azure-keyvault-secrets azure-identity

Environment variables required:
    AZURE_KEY_VAULT_URL - Your Key Vault URL (e.g., https://your-vault.vault.azure.net/)
"""

import os
import sys
from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential
from azure.core.exceptions import ResourceNotFoundError, HttpResponseError


def main():
    # Get Key Vault URL from environment variable
    vault_url = os.environ.get("AZURE_KEY_VAULT_URL")
    
    if not vault_url:
        print("Error: AZURE_KEY_VAULT_URL environment variable is not set")
        print("Example: export AZURE_KEY_VAULT_URL='https://your-vault.vault.azure.net/'")
        sys.exit(1)
    
    print(f"Connecting to Key Vault: {vault_url}\n")
    
    try:
        # Create a SecretClient using DefaultAzureCredential
        credential = DefaultAzureCredential()
        client = SecretClient(vault_url=vault_url, credential=credential)
        
        secret_name = "my-secret"
        
        # 1. CREATE - Set a new secret
        print("=" * 60)
        print("1. CREATE - Setting a new secret")
        print("=" * 60)
        try:
            secret = client.set_secret(secret_name, "my-secret-value")
            print(f"✓ Secret '{secret_name}' created successfully")
            print(f"  Version: {secret.properties.version}")
            print(f"  Created on: {secret.properties.created_on}")
        except HttpResponseError as e:
            print(f"✗ Failed to create secret: {e.message}")
            sys.exit(1)
        
        # 2. READ - Retrieve the secret
        print("\n" + "=" * 60)
        print("2. READ - Retrieving the secret")
        print("=" * 60)
        try:
            retrieved_secret = client.get_secret(secret_name)
            print(f"✓ Secret '{secret_name}' retrieved successfully")
            print(f"  Value: {retrieved_secret.value}")
            print(f"  Version: {retrieved_secret.properties.version}")
        except ResourceNotFoundError:
            print(f"✗ Secret '{secret_name}' not found")
        except HttpResponseError as e:
            print(f"✗ Failed to retrieve secret: {e.message}")
        
        # 3. UPDATE - Update the secret with a new value
        print("\n" + "=" * 60)
        print("3. UPDATE - Updating the secret with a new value")
        print("=" * 60)
        try:
            updated_secret = client.set_secret(secret_name, "updated-value")
            print(f"✓ Secret '{secret_name}' updated successfully")
            print(f"  New value: {updated_secret.value}")
            print(f"  New version: {updated_secret.properties.version}")
        except HttpResponseError as e:
            print(f"✗ Failed to update secret: {e.message}")
        
        # 4. DELETE - Delete and purge the secret
        print("\n" + "=" * 60)
        print("4. DELETE - Deleting and purging the secret")
        print("=" * 60)
        try:
            # Begin delete operation (soft delete)
            poller = client.begin_delete_secret(secret_name)
            deleted_secret = poller.result()
            print(f"✓ Secret '{secret_name}' deleted successfully (soft-delete)")
            print(f"  Scheduled purge date: {deleted_secret.properties.scheduled_purge_date}")
            print(f"  Deleted date: {deleted_secret.properties.deleted_date}")
            
            # Purge the deleted secret (permanent deletion)
            print(f"\n  Purging secret '{secret_name}'...")
            client.purge_deleted_secret(secret_name)
            print(f"✓ Secret '{secret_name}' purged successfully (permanent deletion)")
            
        except ResourceNotFoundError:
            print(f"✗ Secret '{secret_name}' not found for deletion")
        except HttpResponseError as e:
            print(f"✗ Failed to delete/purge secret: {e.message}")
        
        print("\n" + "=" * 60)
        print("All CRUD operations completed successfully!")
        print("=" * 60)
        
    except Exception as e:
        print(f"\n✗ An unexpected error occurred: {str(e)}")
        sys.exit(1)
    
    finally:
        # Close the credential
        credential.close()


if __name__ == "__main__":
    main()
