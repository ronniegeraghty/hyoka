#!/usr/bin/env python3
"""
Azure Key Vault Secrets CRUD Operations Demo

This script demonstrates all four CRUD operations on Azure Key Vault secrets:
1. Create a new secret
2. Read the secret back
3. Update the secret to a new value
4. Delete and purge the secret (for soft-delete enabled vaults)

Required environment variable:
    AZURE_KEY_VAULT_URL: The URL of your Azure Key Vault
                         (e.g., https://my-key-vault.vault.azure.net/)

Authentication uses DefaultAzureCredential, which supports multiple authentication methods:
- Environment variables (AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET)
- Azure CLI (az login)
- Azure PowerShell
- Managed Identity
- Interactive browser

Required pip packages:
    pip install azure-keyvault-secrets azure-identity
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
    """Main function demonstrating CRUD operations on Azure Key Vault secrets."""
    
    # Get the Key Vault URL from environment variable
    vault_url = os.environ.get("AZURE_KEY_VAULT_URL")
    if not vault_url:
        print("ERROR: AZURE_KEY_VAULT_URL environment variable is not set.")
        print("Example: export AZURE_KEY_VAULT_URL='https://my-key-vault.vault.azure.net/'")
        sys.exit(1)
    
    print(f"Using Key Vault: {vault_url}\n")
    
    # Initialize credential and client
    try:
        credential = DefaultAzureCredential()
        client = SecretClient(vault_url=vault_url, credential=credential)
        print("✓ Successfully initialized SecretClient with DefaultAzureCredential\n")
    except ClientAuthenticationError as e:
        print(f"ERROR: Authentication failed: {e.message}")
        print("\nPlease ensure you are authenticated using one of these methods:")
        print("  - Azure CLI: az login")
        print("  - Environment variables: AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET")
        print("  - Managed Identity (when running in Azure)")
        sys.exit(1)
    except Exception as e:
        print(f"ERROR: Failed to initialize client: {e}")
        sys.exit(1)
    
    secret_name = "my-secret"
    
    try:
        # ========================================
        # 1. CREATE - Set a new secret
        # ========================================
        print("=" * 60)
        print("1. CREATE - Creating a new secret")
        print("=" * 60)
        
        initial_value = "my-secret-value"
        secret = client.set_secret(secret_name, initial_value)
        
        print(f"✓ Secret created successfully!")
        print(f"  Name:    {secret.name}")
        print(f"  Value:   {secret.value}")
        print(f"  Version: {secret.properties.version}")
        print(f"  Created: {secret.properties.created_on}")
        print()
        
        # ========================================
        # 2. READ - Retrieve the secret
        # ========================================
        print("=" * 60)
        print("2. READ - Retrieving the secret")
        print("=" * 60)
        
        retrieved_secret = client.get_secret(secret_name)
        
        print(f"✓ Secret retrieved successfully!")
        print(f"  Name:    {retrieved_secret.name}")
        print(f"  Value:   {retrieved_secret.value}")
        print(f"  Version: {retrieved_secret.properties.version}")
        print()
        
        # ========================================
        # 3. UPDATE - Update the secret value
        # ========================================
        print("=" * 60)
        print("3. UPDATE - Updating the secret to a new value")
        print("=" * 60)
        
        updated_value = "updated-value"
        # set_secret creates a new version when the name already exists
        updated_secret = client.set_secret(secret_name, updated_value)
        
        print(f"✓ Secret updated successfully!")
        print(f"  Name:        {updated_secret.name}")
        print(f"  New Value:   {updated_secret.value}")
        print(f"  New Version: {updated_secret.properties.version}")
        print(f"  Updated:     {updated_secret.properties.updated_on}")
        print()
        
        # Verify the update
        verified_secret = client.get_secret(secret_name)
        print(f"  Verification: Current value is '{verified_secret.value}'")
        print()
        
        # ========================================
        # 4. DELETE - Delete and purge the secret
        # ========================================
        print("=" * 60)
        print("4. DELETE - Deleting and purging the secret")
        print("=" * 60)
        
        # Delete the secret (soft delete)
        print(f"Deleting secret '{secret_name}'...")
        delete_poller = client.begin_delete_secret(secret_name)
        deleted_secret = delete_poller.result()
        
        print(f"✓ Secret deleted successfully!")
        print(f"  Name:           {deleted_secret.name}")
        print(f"  Deleted Date:   {deleted_secret.deleted_date}")
        print(f"  Recovery ID:    {deleted_secret.recovery_id}")
        print(f"  Scheduled Purge: {deleted_secret.scheduled_purge_date}")
        print()
        
        # Purge the secret (permanent deletion for soft-delete enabled vaults)
        print(f"Purging secret '{secret_name}' permanently...")
        client.purge_deleted_secret(secret_name)
        
        print(f"✓ Secret purged successfully!")
        print(f"  The secret '{secret_name}' has been permanently deleted.")
        print()
        
        print("=" * 60)
        print("✓ All CRUD operations completed successfully!")
        print("=" * 60)
        
    except ResourceNotFoundError as e:
        print(f"ERROR: Resource not found: {e.message}")
    except HttpResponseError as e:
        print(f"ERROR: HTTP response error: {e.message}")
        if "Operation purge is not allowed" in str(e):
            print("\nNote: Purge operation may not be allowed if:")
            print("  - The vault does not have soft-delete enabled")
            print("  - The vault's purge protection is enabled")
            print("  - The secret's recovery level does not support purging")
    except Exception as e:
        print(f"ERROR: An unexpected error occurred: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)
    finally:
        # Clean up
        credential.close()
        client.close()


if __name__ == "__main__":
    main()
