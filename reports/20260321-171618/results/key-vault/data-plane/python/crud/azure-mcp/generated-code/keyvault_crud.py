"""
Azure Key Vault Secrets CRUD Operations Demo

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
    key_vault_url = "https://<your-key-vault-name>.vault.azure.net/"
    
    # Check if Key Vault URL is provided as command line argument
    if len(sys.argv) > 1:
        key_vault_url = sys.argv[1]
    
    if "<your-key-vault-name>" in key_vault_url:
        print("Error: Please provide a valid Key Vault URL")
        print("Usage: python keyvault_crud.py https://<your-key-vault-name>.vault.azure.net/")
        sys.exit(1)
    
    try:
        # Authenticate using DefaultAzureCredential
        credential = DefaultAzureCredential()
        
        # Create SecretClient
        client = SecretClient(vault_url=key_vault_url, credential=credential)
        
        print(f"Connected to Key Vault: {key_vault_url}\n")
        
        # 1. CREATE - Set a new secret
        print("=" * 60)
        print("1. CREATE - Setting secret 'my-secret'")
        print("=" * 60)
        secret_name = "my-secret"
        secret_value = "my-secret-value"
        
        secret = client.set_secret(secret_name, secret_value)
        print(f"✓ Secret created successfully")
        print(f"  Name: {secret.name}")
        print(f"  Value: {secret.value}")
        print(f"  Version: {secret.properties.version}")
        print()
        
        # 2. READ - Retrieve the secret
        print("=" * 60)
        print("2. READ - Getting secret 'my-secret'")
        print("=" * 60)
        
        retrieved_secret = client.get_secret(secret_name)
        print(f"✓ Secret retrieved successfully")
        print(f"  Name: {retrieved_secret.name}")
        print(f"  Value: {retrieved_secret.value}")
        print(f"  Version: {retrieved_secret.properties.version}")
        print(f"  Created: {retrieved_secret.properties.created_on}")
        print()
        
        # 3. UPDATE - Update the secret value
        print("=" * 60)
        print("3. UPDATE - Updating secret 'my-secret'")
        print("=" * 60)
        new_value = "updated-value"
        
        updated_secret = client.set_secret(secret_name, new_value)
        print(f"✓ Secret updated successfully")
        print(f"  Name: {updated_secret.name}")
        print(f"  Value: {updated_secret.value}")
        print(f"  New Version: {updated_secret.properties.version}")
        print()
        
        # 4. DELETE - Delete and purge the secret
        print("=" * 60)
        print("4. DELETE - Deleting secret 'my-secret'")
        print("=" * 60)
        
        # Begin delete operation (soft delete)
        deleted_secret = client.begin_delete_secret(secret_name).result()
        print(f"✓ Secret deleted (soft-deleted)")
        print(f"  Name: {deleted_secret.name}")
        print(f"  Deleted on: {deleted_secret.deleted_date}")
        print(f"  Scheduled purge: {deleted_secret.scheduled_purge_date}")
        print()
        
        # Purge the deleted secret permanently
        print("Purging secret permanently...")
        client.purge_deleted_secret(secret_name)
        print(f"✓ Secret purged successfully")
        print()
        
        print("=" * 60)
        print("All CRUD operations completed successfully!")
        print("=" * 60)
        
    except ResourceNotFoundError as e:
        print(f"Error: Resource not found - {e.message}")
        sys.exit(1)
    except HttpResponseError as e:
        print(f"Error: HTTP response error - {e.message}")
        print(f"Status code: {e.status_code}")
        sys.exit(1)
    except Exception as e:
        print(f"Error: An unexpected error occurred - {str(e)}")
        sys.exit(1)
    finally:
        # Close the credential
        credential.close()


if __name__ == "__main__":
    main()
