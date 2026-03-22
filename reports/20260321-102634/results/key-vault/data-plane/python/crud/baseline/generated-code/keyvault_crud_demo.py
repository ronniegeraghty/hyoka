"""
Azure Key Vault Secrets CRUD Operations Demo

This script demonstrates all four CRUD operations on Azure Key Vault secrets:
1. Create a new secret
2. Read the secret back
3. Update the secret to a new value
4. Delete and purge the secret

Prerequisites:
- Azure Key Vault with soft-delete enabled
- Appropriate permissions (secrets/get, secrets/set, secrets/delete, secrets/purge)
- Azure credentials configured (via Azure CLI, environment variables, or managed identity)

Required packages:
    pip install azure-keyvault-secrets azure-identity
"""

import os
import sys
from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient
from azure.core.exceptions import ResourceNotFoundError, HttpResponseError


def main():
    # Get vault URL from environment variable
    vault_url = os.environ.get("VAULT_URL")
    
    if not vault_url:
        print("Error: VAULT_URL environment variable is not set")
        print("Example: export VAULT_URL='https://your-vault-name.vault.azure.net/'")
        sys.exit(1)
    
    print(f"Connecting to Key Vault: {vault_url}\n")
    
    # Create credential and client
    try:
        credential = DefaultAzureCredential()
        client = SecretClient(vault_url=vault_url, credential=credential)
    except Exception as e:
        print(f"Error creating Key Vault client: {e}")
        sys.exit(1)
    
    secret_name = "my-secret"
    
    # 1. CREATE - Set a new secret
    print("=" * 60)
    print("1. CREATE - Creating new secret")
    print("=" * 60)
    try:
        secret = client.set_secret(secret_name, "my-secret-value")
        print(f"✓ Secret created successfully")
        print(f"  Name: {secret.name}")
        print(f"  Value: {secret.value}")
        print(f"  Version: {secret.properties.version}")
        print()
    except HttpResponseError as e:
        print(f"✗ Error creating secret: {e.message}")
        sys.exit(1)
    
    # 2. READ - Retrieve the secret
    print("=" * 60)
    print("2. READ - Retrieving secret")
    print("=" * 60)
    try:
        retrieved_secret = client.get_secret(secret_name)
        print(f"✓ Secret retrieved successfully")
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
        print(f"✗ Error retrieving secret: {e.message}")
        sys.exit(1)
    
    # 3. UPDATE - Update the secret to a new value
    print("=" * 60)
    print("3. UPDATE - Updating secret to new value")
    print("=" * 60)
    try:
        # set_secret creates a new version with the updated value
        updated_secret = client.set_secret(secret_name, "updated-value")
        print(f"✓ Secret updated successfully")
        print(f"  Name: {updated_secret.name}")
        print(f"  Value: {updated_secret.value}")
        print(f"  New Version: {updated_secret.properties.version}")
        print()
        
        # Optionally, update secret properties (metadata) without changing the value
        # This demonstrates updating properties like content_type, enabled, tags, etc.
        print("  Updating secret metadata...")
        updated_properties = client.update_secret_properties(
            secret_name,
            content_type="text/plain",
            tags={"environment": "demo", "purpose": "crud-example"}
        )
        print(f"  ✓ Metadata updated")
        print(f"    Content Type: {updated_properties.content_type}")
        print(f"    Tags: {updated_properties.tags}")
        print()
    except HttpResponseError as e:
        print(f"✗ Error updating secret: {e.message}")
        sys.exit(1)
    
    # 4. DELETE - Delete and purge the secret
    print("=" * 60)
    print("4. DELETE - Deleting and purging secret")
    print("=" * 60)
    try:
        # begin_delete_secret returns a poller for long-running delete operation
        print(f"  Deleting secret '{secret_name}'...")
        delete_poller = client.begin_delete_secret(secret_name)
        
        # Wait for deletion to complete
        deleted_secret = delete_poller.result()
        print(f"✓ Secret deleted successfully")
        print(f"  Name: {deleted_secret.name}")
        print(f"  Deleted Date: {deleted_secret.deleted_date}")
        print(f"  Recovery ID: {deleted_secret.recovery_id}")
        print(f"  Scheduled Purge Date: {deleted_secret.scheduled_purge_date}")
        print()
        
        # Purge the secret permanently (only available with soft-delete enabled)
        print(f"  Purging secret '{secret_name}' permanently...")
        client.purge_deleted_secret(secret_name)
        print(f"✓ Secret purged successfully")
        print(f"  The secret has been permanently deleted and cannot be recovered.")
        print()
        
    except ResourceNotFoundError:
        print(f"✗ Secret '{secret_name}' not found for deletion")
        sys.exit(1)
    except HttpResponseError as e:
        print(f"✗ Error deleting/purging secret: {e.message}")
        print(f"  Note: Purge requires soft-delete to be enabled on the vault")
        sys.exit(1)
    
    print("=" * 60)
    print("All CRUD operations completed successfully!")
    print("=" * 60)


if __name__ == "__main__":
    main()
