"""
Azure Key Vault Secrets CRUD Operations Demo

This script demonstrates all four CRUD operations on Azure Key Vault secrets:
- Create a secret
- Read a secret
- Update a secret
- Delete and purge a secret

Required pip packages:
    pip install azure-keyvault-secrets azure-identity

Prerequisites:
    - Azure subscription
    - Azure Key Vault with soft-delete enabled
    - Set environment variable VAULT_URL with your Key Vault URL
      Example: export VAULT_URL="https://my-key-vault.vault.azure.net/"
    - Appropriate authentication configured for DefaultAzureCredential
"""

import os
from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient
from azure.core.exceptions import ResourceNotFoundError, HttpResponseError


def main():
    """Main function demonstrating CRUD operations on Azure Key Vault secrets."""
    
    # Get vault URL from environment variable
    vault_url = os.environ.get("VAULT_URL")
    if not vault_url:
        print("ERROR: VAULT_URL environment variable is not set")
        print("Please set it to your Key Vault URL, e.g.:")
        print('export VAULT_URL="https://my-key-vault.vault.azure.net/"')
        return
    
    print(f"Connecting to Key Vault: {vault_url}")
    print("-" * 60)
    
    # Authenticate and create the secret client
    try:
        credential = DefaultAzureCredential()
        client = SecretClient(vault_url=vault_url, credential=credential)
        print("✓ Successfully authenticated with DefaultAzureCredential\n")
    except Exception as e:
        print(f"ERROR: Failed to create SecretClient: {e}")
        return
    
    secret_name = "my-secret"
    
    # 1. CREATE - Set a new secret
    print("1. CREATE - Setting a new secret")
    print("-" * 60)
    try:
        secret = client.set_secret(secret_name, "my-secret-value")
        print(f"✓ Created secret: {secret.name}")
        print(f"  Value: {secret.value}")
        print(f"  Version: {secret.properties.version}")
        print(f"  Created on: {secret.properties.created_on}\n")
    except HttpResponseError as e:
        print(f"ERROR: Failed to create secret: {e.message}")
        return
    except Exception as e:
        print(f"ERROR: Unexpected error creating secret: {e}")
        return
    
    # 2. READ - Retrieve the secret
    print("2. READ - Retrieving the secret")
    print("-" * 60)
    try:
        retrieved_secret = client.get_secret(secret_name)
        print(f"✓ Retrieved secret: {retrieved_secret.name}")
        print(f"  Value: {retrieved_secret.value}")
        print(f"  Version: {retrieved_secret.properties.version}\n")
    except ResourceNotFoundError:
        print(f"ERROR: Secret '{secret_name}' not found")
        return
    except Exception as e:
        print(f"ERROR: Failed to retrieve secret: {e}")
        return
    
    # 3. UPDATE - Update the secret with a new value
    print("3. UPDATE - Updating the secret to a new value")
    print("-" * 60)
    try:
        # Setting a secret with an existing name creates a new version
        updated_secret = client.set_secret(secret_name, "updated-value")
        print(f"✓ Updated secret: {updated_secret.name}")
        print(f"  New value: {updated_secret.value}")
        print(f"  New version: {updated_secret.properties.version}")
        print(f"  Updated on: {updated_secret.properties.updated_on}\n")
    except HttpResponseError as e:
        print(f"ERROR: Failed to update secret: {e.message}")
        return
    except Exception as e:
        print(f"ERROR: Unexpected error updating secret: {e}")
        return
    
    # 4. DELETE - Delete and purge the secret
    print("4. DELETE - Deleting and purging the secret")
    print("-" * 60)
    try:
        # Begin delete operation (returns a poller)
        delete_operation = client.begin_delete_secret(secret_name)
        deleted_secret = delete_operation.result()
        print(f"✓ Deleted secret: {deleted_secret.name}")
        print(f"  Deleted on: {deleted_secret.deleted_date}")
        print(f"  Scheduled purge date: {deleted_secret.scheduled_purge_date}")
        
        # Purge the secret permanently (for soft-delete enabled vaults)
        client.purge_deleted_secret(secret_name)
        print(f"✓ Purged secret '{secret_name}' permanently\n")
        
    except ResourceNotFoundError:
        print(f"ERROR: Secret '{secret_name}' not found for deletion")
        return
    except HttpResponseError as e:
        print(f"ERROR: Failed to delete/purge secret: {e.message}")
        print("Note: Ensure your vault has soft-delete enabled and you have purge permissions")
        return
    except Exception as e:
        print(f"ERROR: Unexpected error during delete/purge: {e}")
        return
    
    print("=" * 60)
    print("All CRUD operations completed successfully!")
    print("=" * 60)


if __name__ == "__main__":
    main()
