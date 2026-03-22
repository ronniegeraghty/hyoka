#!/usr/bin/env python3
"""
Azure Key Vault Secrets CRUD Operations Demo

This script demonstrates all four CRUD operations on Azure Key Vault secrets:
1. CREATE - Set a new secret
2. READ - Retrieve the secret
3. UPDATE - Update the secret value
4. DELETE - Delete and purge the secret

Requirements:
    pip install azure-keyvault-secrets azure-identity

Environment Variables:
    VAULT_URL - Your Azure Key Vault URL (e.g., https://my-vault.vault.azure.net/)

Authentication:
    Uses DefaultAzureCredential which attempts authentication via:
    - Environment variables
    - Managed Identity
    - Azure CLI
    - Azure PowerShell
    - Interactive browser
"""

import os
import sys
from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient
from azure.core.exceptions import (
    ResourceNotFoundError,
    HttpResponseError,
    AzureError
)


def main():
    """Main function demonstrating CRUD operations on Azure Key Vault secrets."""
    
    # Get vault URL from environment variable
    vault_url = os.environ.get("VAULT_URL")
    if not vault_url:
        print("Error: VAULT_URL environment variable is not set")
        print("Example: export VAULT_URL='https://my-vault.vault.azure.net/'")
        sys.exit(1)
    
    print(f"Connecting to Key Vault: {vault_url}")
    print("-" * 80)
    
    try:
        # Initialize credential and client
        credential = DefaultAzureCredential()
        client = SecretClient(vault_url=vault_url, credential=credential)
        
        secret_name = "my-secret"
        
        # ============================================================
        # 1. CREATE - Set a new secret
        # ============================================================
        print("\n1. CREATE - Setting a new secret")
        print("-" * 80)
        try:
            secret = client.set_secret(secret_name, "my-secret-value")
            print(f"✓ Secret created successfully!")
            print(f"  Name: {secret.name}")
            print(f"  Value: {secret.value}")
            print(f"  Version: {secret.properties.version}")
            print(f"  Created: {secret.properties.created_on}")
        except HttpResponseError as e:
            print(f"✗ Failed to create secret: {e.message}")
            raise
        
        # ============================================================
        # 2. READ - Retrieve the secret
        # ============================================================
        print("\n2. READ - Retrieving the secret")
        print("-" * 80)
        try:
            retrieved_secret = client.get_secret(secret_name)
            print(f"✓ Secret retrieved successfully!")
            print(f"  Name: {retrieved_secret.name}")
            print(f"  Value: {retrieved_secret.value}")
            print(f"  Version: {retrieved_secret.properties.version}")
            print(f"  Content Type: {retrieved_secret.properties.content_type}")
            print(f"  Enabled: {retrieved_secret.properties.enabled}")
        except ResourceNotFoundError:
            print(f"✗ Secret '{secret_name}' not found")
            raise
        except HttpResponseError as e:
            print(f"✗ Failed to retrieve secret: {e.message}")
            raise
        
        # ============================================================
        # 3. UPDATE - Update the secret to a new value
        # ============================================================
        print("\n3. UPDATE - Updating the secret value")
        print("-" * 80)
        try:
            # Update the secret value (creates a new version)
            updated_secret = client.set_secret(secret_name, "updated-value")
            print(f"✓ Secret value updated successfully!")
            print(f"  Name: {updated_secret.name}")
            print(f"  New Value: {updated_secret.value}")
            print(f"  New Version: {updated_secret.properties.version}")
            print(f"  Updated: {updated_secret.properties.updated_on}")
            
            # Optional: Update secret properties (metadata)
            # This updates metadata without changing the value
            print("\n  Updating secret properties...")
            updated_properties = client.update_secret_properties(
                secret_name,
                content_type="text/plain",
                enabled=True
            )
            print(f"✓ Secret properties updated!")
            print(f"  Content Type: {updated_properties.content_type}")
            print(f"  Enabled: {updated_properties.enabled}")
            
        except HttpResponseError as e:
            print(f"✗ Failed to update secret: {e.message}")
            raise
        
        # ============================================================
        # 4. DELETE - Delete and purge the secret
        # ============================================================
        print("\n4. DELETE - Deleting and purging the secret")
        print("-" * 80)
        try:
            # Delete the secret (soft delete)
            print(f"  Deleting secret '{secret_name}'...")
            delete_operation = client.begin_delete_secret(secret_name)
            deleted_secret = delete_operation.result()
            print(f"✓ Secret deleted successfully!")
            print(f"  Name: {deleted_secret.name}")
            print(f"  Deleted Date: {deleted_secret.deleted_date}")
            print(f"  Scheduled Purge Date: {deleted_secret.scheduled_purge_date}")
            print(f"  Recovery ID: {deleted_secret.recovery_id}")
            
            # Purge the secret permanently (only works with soft-delete enabled)
            print(f"\n  Purging deleted secret '{secret_name}'...")
            client.purge_deleted_secret(secret_name)
            print(f"✓ Secret purged successfully!")
            print(f"  The secret has been permanently deleted and cannot be recovered.")
            
        except ResourceNotFoundError:
            print(f"✗ Secret '{secret_name}' not found for deletion")
            raise
        except HttpResponseError as e:
            if "not currently in a deleted state" in str(e):
                print(f"✗ Secret is not in deleted state (may already be purged)")
            else:
                print(f"✗ Failed to delete/purge secret: {e.message}")
            raise
        
        print("\n" + "=" * 80)
        print("All CRUD operations completed successfully!")
        print("=" * 80)
        
    except AzureError as e:
        print(f"\n✗ Azure error occurred: {e}")
        sys.exit(1)
    except Exception as e:
        print(f"\n✗ Unexpected error occurred: {e}")
        sys.exit(1)
    finally:
        # Clean up
        try:
            credential.close()
        except:
            pass


if __name__ == "__main__":
    main()
