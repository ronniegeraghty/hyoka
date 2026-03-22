#!/usr/bin/env python3
"""
Azure Key Vault Secrets CRUD Operations Demo

This script demonstrates all four CRUD operations on Azure Key Vault secrets:
1. Create - Set a new secret
2. Read - Get a secret value
3. Update - Update a secret to a new value
4. Delete - Delete and purge a secret

Required pip packages:
    pip install azure-keyvault-secrets azure-identity

Prerequisites:
    - An Azure Key Vault with soft-delete enabled
    - Appropriate authentication configured for DefaultAzureCredential
    - Required permissions: secrets/set, secrets/get, secrets/delete, secrets/purge
    - Set AZURE_KEY_VAULT_URL environment variable to your vault URL
"""

import os
import sys
from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient
from azure.core.exceptions import ResourceNotFoundError, HttpResponseError


def main():
    """Demonstrate CRUD operations on Azure Key Vault secrets."""
    
    # Get vault URL from environment variable
    vault_url = os.environ.get("AZURE_KEY_VAULT_URL")
    
    if not vault_url:
        print("Error: AZURE_KEY_VAULT_URL environment variable is not set")
        print("Example: export AZURE_KEY_VAULT_URL='https://my-vault.vault.azure.net/'")
        sys.exit(1)
    
    print(f"Using Key Vault: {vault_url}\n")
    
    # Initialize the SecretClient with DefaultAzureCredential
    try:
        credential = DefaultAzureCredential()
        client = SecretClient(vault_url=vault_url, credential=credential)
        print("✓ Successfully authenticated with DefaultAzureCredential\n")
    except Exception as e:
        print(f"✗ Authentication failed: {e}")
        sys.exit(1)
    
    secret_name = "my-secret"
    
    try:
        # ====================================================================
        # 1. CREATE - Set a new secret
        # ====================================================================
        print("=" * 60)
        print("1. CREATE - Setting a new secret")
        print("=" * 60)
        
        secret_value = "my-secret-value"
        created_secret = client.set_secret(secret_name, secret_value)
        
        print(f"✓ Secret created successfully:")
        print(f"  Name: {created_secret.name}")
        print(f"  Value: {created_secret.value}")
        print(f"  Version: {created_secret.properties.version}")
        print(f"  Created on: {created_secret.properties.created_on}")
        print()
        
        # ====================================================================
        # 2. READ - Retrieve the secret
        # ====================================================================
        print("=" * 60)
        print("2. READ - Retrieving the secret")
        print("=" * 60)
        
        retrieved_secret = client.get_secret(secret_name)
        
        print(f"✓ Secret retrieved successfully:")
        print(f"  Name: {retrieved_secret.name}")
        print(f"  Value: {retrieved_secret.value}")
        print(f"  Version: {retrieved_secret.properties.version}")
        print()
        
        # ====================================================================
        # 3. UPDATE - Update the secret to a new value
        # ====================================================================
        print("=" * 60)
        print("3. UPDATE - Updating the secret with a new value")
        print("=" * 60)
        
        new_value = "updated-value"
        updated_secret = client.set_secret(secret_name, new_value)
        
        print(f"✓ Secret updated successfully:")
        print(f"  Name: {updated_secret.name}")
        print(f"  New Value: {updated_secret.value}")
        print(f"  New Version: {updated_secret.properties.version}")
        print(f"  Updated on: {updated_secret.properties.updated_on}")
        print()
        
        # Verify the update by reading again
        verified_secret = client.get_secret(secret_name)
        print(f"✓ Verified updated value: {verified_secret.value}")
        print()
        
        # ====================================================================
        # 4. DELETE - Delete and purge the secret
        # ====================================================================
        print("=" * 60)
        print("4. DELETE - Deleting and purging the secret")
        print("=" * 60)
        
        # Begin delete operation (returns a poller for soft-delete enabled vaults)
        print("  Initiating delete operation...")
        delete_poller = client.begin_delete_secret(secret_name)
        
        # Wait for deletion to complete
        deleted_secret = delete_poller.result()
        
        print(f"✓ Secret soft-deleted successfully:")
        print(f"  Name: {deleted_secret.name}")
        print(f"  Deleted on: {deleted_secret.deleted_date}")
        print(f"  Scheduled purge date: {deleted_secret.scheduled_purge_date}")
        print(f"  Recovery ID: {deleted_secret.recovery_id}")
        print()
        
        # Purge the deleted secret (permanent deletion)
        print("  Purging deleted secret permanently...")
        client.purge_deleted_secret(secret_name)
        
        print(f"✓ Secret '{secret_name}' purged successfully (permanent deletion)")
        print()
        
        # Verify deletion
        print("  Verifying secret no longer exists...")
        try:
            client.get_secret(secret_name)
            print("✗ Unexpected: Secret still exists!")
        except ResourceNotFoundError:
            print(f"✓ Confirmed: Secret '{secret_name}' no longer exists")
        print()
        
        print("=" * 60)
        print("All CRUD operations completed successfully!")
        print("=" * 60)
        
    except ResourceNotFoundError as e:
        print(f"✗ Resource not found: {e.message}")
        sys.exit(1)
    except HttpResponseError as e:
        print(f"✗ HTTP error occurred: {e.message}")
        print(f"  Status code: {e.status_code}")
        print(f"  Error code: {e.error.code if hasattr(e, 'error') else 'N/A'}")
        sys.exit(1)
    except Exception as e:
        print(f"✗ An unexpected error occurred: {type(e).__name__}: {e}")
        sys.exit(1)
    finally:
        # Clean up - close the credential
        credential.close()


if __name__ == "__main__":
    main()
