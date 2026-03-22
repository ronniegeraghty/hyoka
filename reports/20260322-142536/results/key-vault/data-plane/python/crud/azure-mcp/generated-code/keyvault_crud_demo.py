#!/usr/bin/env python3
"""
Azure Key Vault CRUD Operations Demo
Demonstrates Create, Read, Update, and Delete operations for secrets
using the Azure SDK for Python.

Required packages:
- azure-keyvault-secrets
- azure-identity

Install with: pip install azure-keyvault-secrets azure-identity
"""

import os
import sys
from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient
from azure.core.exceptions import ResourceNotFoundError, HttpResponseError


def main():
    # Get vault URL from environment variable
    vault_url = os.environ.get("AZURE_KEY_VAULT_URL")
    
    if not vault_url:
        print("Error: AZURE_KEY_VAULT_URL environment variable not set")
        print("Example: export AZURE_KEY_VAULT_URL='https://my-key-vault.vault.azure.net/'")
        sys.exit(1)
    
    print(f"Connecting to Azure Key Vault: {vault_url}")
    
    try:
        # Authenticate using DefaultAzureCredential
        # This will try multiple authentication methods in order:
        # 1. Environment variables
        # 2. Managed identity
        # 3. Visual Studio Code
        # 4. Azure CLI
        # 5. Azure PowerShell
        credential = DefaultAzureCredential()
        
        # Create the SecretClient
        secret_client = SecretClient(vault_url=vault_url, credential=credential)
        
        secret_name = "my-secret"
        
        # ============================================================
        # CREATE - Set a new secret
        # ============================================================
        print("\n1. CREATE - Setting a new secret...")
        try:
            secret = secret_client.set_secret(secret_name, "my-secret-value")
            print(f"   ✓ Secret created successfully")
            print(f"   - Name: {secret.name}")
            print(f"   - Value: {secret.value}")
            print(f"   - Version: {secret.properties.version}")
        except HttpResponseError as e:
            print(f"   ✗ Failed to create secret: {e.message}")
            sys.exit(1)
        
        # ============================================================
        # READ - Retrieve the secret
        # ============================================================
        print("\n2. READ - Retrieving the secret...")
        try:
            retrieved_secret = secret_client.get_secret(secret_name)
            print(f"   ✓ Secret retrieved successfully")
            print(f"   - Name: {retrieved_secret.name}")
            print(f"   - Value: {retrieved_secret.value}")
            print(f"   - Version: {retrieved_secret.properties.version}")
        except ResourceNotFoundError:
            print(f"   ✗ Secret '{secret_name}' not found")
            sys.exit(1)
        except HttpResponseError as e:
            print(f"   ✗ Failed to retrieve secret: {e.message}")
            sys.exit(1)
        
        # ============================================================
        # UPDATE - Update the secret value
        # ============================================================
        print("\n3. UPDATE - Updating the secret to a new value...")
        try:
            # set_secret creates a new version when called on an existing secret
            updated_secret = secret_client.set_secret(secret_name, "updated-value")
            print(f"   ✓ Secret updated successfully")
            print(f"   - Name: {updated_secret.name}")
            print(f"   - New Value: {updated_secret.value}")
            print(f"   - New Version: {updated_secret.properties.version}")
        except HttpResponseError as e:
            print(f"   ✗ Failed to update secret: {e.message}")
            sys.exit(1)
        
        # ============================================================
        # DELETE - Delete and purge the secret
        # ============================================================
        print("\n4. DELETE - Deleting the secret...")
        try:
            # begin_delete_secret returns a poller for long-running operations
            delete_poller = secret_client.begin_delete_secret(secret_name)
            deleted_secret = delete_poller.result()
            print(f"   ✓ Secret deleted successfully")
            print(f"   - Name: {deleted_secret.name}")
            print(f"   - Deleted Date: {deleted_secret.deleted_date}")
            print(f"   - Scheduled Purge Date: {deleted_secret.scheduled_purge_date}")
        except ResourceNotFoundError:
            print(f"   ✗ Secret '{secret_name}' not found")
            sys.exit(1)
        except HttpResponseError as e:
            print(f"   ✗ Failed to delete secret: {e.message}")
            sys.exit(1)
        
        # ============================================================
        # PURGE - Permanently delete the secret (soft-delete vaults)
        # ============================================================
        print("\n5. PURGE - Permanently deleting the secret...")
        try:
            secret_client.purge_deleted_secret(secret_name)
            print(f"   ✓ Secret purged successfully")
            print(f"   - The secret '{secret_name}' has been permanently deleted")
        except ResourceNotFoundError:
            print(f"   ✗ Deleted secret '{secret_name}' not found")
        except HttpResponseError as e:
            # Purge may fail if vault doesn't have soft-delete or purge protection
            print(f"   ⚠ Purge operation note: {e.message}")
        
        print("\n" + "="*60)
        print("CRUD operations completed successfully!")
        print("="*60)
        
    except Exception as e:
        print(f"\n✗ Unexpected error: {type(e).__name__}: {str(e)}")
        sys.exit(1)
    finally:
        # Close the credential
        credential.close()


if __name__ == "__main__":
    main()
