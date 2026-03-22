#!/usr/bin/env python3
"""
Azure Key Vault Secrets CRUD Operations Demo

This script demonstrates all four CRUD operations on Azure Key Vault secrets:
1. Create - Set a new secret
2. Read - Retrieve a secret value
3. Update - Update a secret to a new value
4. Delete - Delete and purge a secret (for soft-delete enabled vaults)

Requirements:
    pip install azure-keyvault-secrets azure-identity

Environment Variables:
    AZURE_KEY_VAULT_URL - Your Key Vault URL (e.g., https://my-vault.vault.azure.net/)

Authentication:
    Uses DefaultAzureCredential which supports multiple authentication methods:
    - Azure CLI (az login)
    - Managed Identity (in Azure)
    - Environment variables (AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET)
    - And more...
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
    """Main function demonstrating CRUD operations on Key Vault secrets."""
    
    # Get Key Vault URL from environment variable
    vault_url = os.environ.get("AZURE_KEY_VAULT_URL")
    if not vault_url:
        print("Error: AZURE_KEY_VAULT_URL environment variable is not set")
        print("Example: export AZURE_KEY_VAULT_URL='https://my-vault.vault.azure.net/'")
        sys.exit(1)
    
    print(f"Connecting to Key Vault: {vault_url}")
    print("-" * 80)
    
    try:
        # Initialize the credential and client
        credential = DefaultAzureCredential()
        client = SecretClient(vault_url=vault_url, credential=credential)
        
        # 1. CREATE - Set a new secret
        print("\n1. CREATE - Setting a new secret 'my-secret'")
        secret_name = "my-secret"
        secret_value = "my-secret-value"
        
        try:
            created_secret = client.set_secret(secret_name, secret_value)
            print(f"   ✓ Secret created successfully")
            print(f"   Name: {created_secret.name}")
            print(f"   Value: {created_secret.value}")
            print(f"   Version: {created_secret.properties.version}")
            print(f"   Created on: {created_secret.properties.created_on}")
        except HttpResponseError as e:
            print(f"   ✗ Failed to create secret: {e.message}")
            sys.exit(1)
        
        # 2. READ - Retrieve the secret
        print("\n2. READ - Retrieving the secret 'my-secret'")
        try:
            retrieved_secret = client.get_secret(secret_name)
            print(f"   ✓ Secret retrieved successfully")
            print(f"   Name: {retrieved_secret.name}")
            print(f"   Value: {retrieved_secret.value}")
            print(f"   Version: {retrieved_secret.properties.version}")
        except ResourceNotFoundError:
            print(f"   ✗ Secret '{secret_name}' not found")
            sys.exit(1)
        except HttpResponseError as e:
            print(f"   ✗ Failed to retrieve secret: {e.message}")
            sys.exit(1)
        
        # 3. UPDATE - Update the secret to a new value
        print("\n3. UPDATE - Updating secret to a new value")
        new_value = "updated-value"
        
        try:
            updated_secret = client.set_secret(secret_name, new_value)
            print(f"   ✓ Secret updated successfully")
            print(f"   Name: {updated_secret.name}")
            print(f"   New Value: {updated_secret.value}")
            print(f"   New Version: {updated_secret.properties.version}")
            print(f"   Updated on: {updated_secret.properties.updated_on}")
        except HttpResponseError as e:
            print(f"   ✗ Failed to update secret: {e.message}")
            sys.exit(1)
        
        # 4. DELETE - Delete and purge the secret
        print("\n4. DELETE - Deleting and purging the secret")
        
        # Step 4a: Begin delete operation
        try:
            print("   4a. Starting deletion...")
            delete_poller = client.begin_delete_secret(secret_name)
            deleted_secret = delete_poller.result()
            print(f"   ✓ Secret deleted successfully")
            print(f"   Name: {deleted_secret.name}")
            print(f"   Deleted on: {deleted_secret.deleted_date}")
            print(f"   Scheduled purge date: {deleted_secret.scheduled_purge_date}")
            print(f"   Recovery ID: {deleted_secret.recovery_id}")
        except ResourceNotFoundError:
            print(f"   ✗ Secret '{secret_name}' not found for deletion")
            sys.exit(1)
        except HttpResponseError as e:
            print(f"   ✗ Failed to delete secret: {e.message}")
            sys.exit(1)
        
        # Step 4b: Purge the deleted secret (permanent deletion)
        try:
            print("   4b. Purging deleted secret (permanent deletion)...")
            client.purge_deleted_secret(secret_name)
            print(f"   ✓ Secret purged successfully (permanently deleted)")
        except ResourceNotFoundError:
            print(f"   ⚠ Secret '{secret_name}' not found in deleted secrets")
            print("      (May have been already purged or vault may not have soft-delete enabled)")
        except HttpResponseError as e:
            if "not enabled" in str(e).lower() or "soft delete" in str(e).lower():
                print(f"   ⚠ Purge not needed - vault may not have soft-delete enabled")
                print(f"      Secret was permanently deleted in the delete step")
            else:
                print(f"   ✗ Failed to purge secret: {e.message}")
                sys.exit(1)
        
        print("\n" + "=" * 80)
        print("All CRUD operations completed successfully!")
        print("=" * 80)
        
    except ClientAuthenticationError as e:
        print(f"\nAuthentication Error: {e.message}")
        print("\nPlease ensure you are authenticated. Try one of:")
        print("  - Run 'az login' (Azure CLI)")
        print("  - Set AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET")
        print("  - Use Managed Identity (if running in Azure)")
        sys.exit(1)
    except Exception as e:
        print(f"\nUnexpected error: {type(e).__name__}: {str(e)}")
        sys.exit(1)
    finally:
        # Clean up - close the credential
        try:
            credential.close()
        except:
            pass


if __name__ == "__main__":
    main()
