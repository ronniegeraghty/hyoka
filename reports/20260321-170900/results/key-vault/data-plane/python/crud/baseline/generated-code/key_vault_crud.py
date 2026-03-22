"""
Azure Key Vault Secret CRUD Operations

This script demonstrates Create, Read, Update, and Delete operations
on Azure Key Vault secrets using the Azure SDK for Python.

Required packages:
    pip install azure-keyvault-secrets azure-identity
"""

from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential
from azure.core.exceptions import ResourceNotFoundError, HttpResponseError
import sys


def main():
    # Replace with your Key Vault URL
    # Format: https://<your-key-vault-name>.vault.azure.net/
    key_vault_url = input("Enter your Key Vault URL: ").strip()
    
    if not key_vault_url:
        print("Error: Key Vault URL is required")
        sys.exit(1)
    
    try:
        # Authenticate using DefaultAzureCredential
        credential = DefaultAzureCredential()
        
        # Create a SecretClient
        client = SecretClient(vault_url=key_vault_url, credential=credential)
        
        print(f"\nConnected to Key Vault: {key_vault_url}")
        print("-" * 60)
        
        # 1. CREATE: Create a new secret
        print("\n1. CREATE: Creating secret 'my-secret'...")
        secret_name = "my-secret"
        secret_value = "my-secret-value"
        
        created_secret = client.set_secret(secret_name, secret_value)
        print(f"   ✓ Secret created: {created_secret.name}")
        print(f"   - Value: {created_secret.value}")
        print(f"   - Version: {created_secret.properties.version}")
        
        # 2. READ: Read the secret back
        print("\n2. READ: Reading secret 'my-secret'...")
        retrieved_secret = client.get_secret(secret_name)
        print(f"   ✓ Secret retrieved: {retrieved_secret.name}")
        print(f"   - Value: {retrieved_secret.value}")
        print(f"   - Version: {retrieved_secret.properties.version}")
        
        # 3. UPDATE: Update the secret to a new value
        print("\n3. UPDATE: Updating secret 'my-secret'...")
        new_value = "updated-value"
        updated_secret = client.set_secret(secret_name, new_value)
        print(f"   ✓ Secret updated: {updated_secret.name}")
        print(f"   - New Value: {updated_secret.value}")
        print(f"   - New Version: {updated_secret.properties.version}")
        
        # Verify the update
        verification = client.get_secret(secret_name)
        print(f"   - Verified Value: {verification.value}")
        
        # 4. DELETE: Delete and purge the secret
        print("\n4. DELETE: Deleting secret 'my-secret'...")
        poller = client.begin_delete_secret(secret_name)
        deleted_secret = poller.result()
        print(f"   ✓ Secret deleted: {deleted_secret.name}")
        print(f"   - Scheduled purge date: {deleted_secret.scheduled_purge_date}")
        print(f"   - Deleted date: {deleted_secret.deleted_date}")
        
        # Purge the deleted secret (for soft-delete enabled vaults)
        print("\n   Purging deleted secret...")
        client.purge_deleted_secret(secret_name)
        print(f"   ✓ Secret purged permanently")
        
        print("\n" + "-" * 60)
        print("All CRUD operations completed successfully!")
        
    except ResourceNotFoundError as e:
        print(f"\n✗ Error: Resource not found - {e.message}")
        sys.exit(1)
    except HttpResponseError as e:
        print(f"\n✗ HTTP Error: {e.status_code} - {e.message}")
        if e.status_code == 401:
            print("  Hint: Check your authentication credentials")
        elif e.status_code == 403:
            print("  Hint: Check your Key Vault access permissions")
        sys.exit(1)
    except Exception as e:
        print(f"\n✗ Unexpected error: {type(e).__name__} - {str(e)}")
        sys.exit(1)
    finally:
        # Clean up credential
        if 'credential' in locals():
            credential.close()


if __name__ == "__main__":
    main()
