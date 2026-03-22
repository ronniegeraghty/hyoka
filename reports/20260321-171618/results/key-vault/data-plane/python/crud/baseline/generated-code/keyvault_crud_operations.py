#!/usr/bin/env python3
"""
Azure Key Vault CRUD Operations Script

This script demonstrates all four CRUD operations on Azure Key Vault secrets:
1. Create a new secret
2. Read the secret back
3. Update the secret to a new value
4. Delete and purge the secret (for soft-delete enabled vaults)

Requirements:
    pip install azure-keyvault-secrets azure-identity

Authentication:
    Uses DefaultAzureCredential which supports multiple authentication methods:
    - Environment variables (AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET)
    - Managed Identity
    - Azure CLI (az login)
    - Azure PowerShell
    - Interactive browser

Environment Variables:
    VAULT_URL: The URL of your Azure Key Vault (e.g., https://my-vault.vault.azure.net/)
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
    """Main function to demonstrate CRUD operations on Azure Key Vault secrets."""
    
    # Get the vault URL from environment variable
    vault_url = os.environ.get("VAULT_URL")
    if not vault_url:
        print("Error: VAULT_URL environment variable is not set")
        print("Example: export VAULT_URL='https://my-vault.vault.azure.net/'")
        sys.exit(1)
    
    print(f"Connecting to Key Vault: {vault_url}\n")
    
    # Initialize the credential and client
    try:
        credential = DefaultAzureCredential()
        client = SecretClient(vault_url=vault_url, credential=credential)
    except ClientAuthenticationError as e:
        print(f"Authentication failed: {e.message}")
        sys.exit(1)
    except Exception as e:
        print(f"Failed to create client: {str(e)}")
        sys.exit(1)
    
    secret_name = "my-secret"
    
    try:
        # ==========================================
        # 1. CREATE - Set a new secret
        # ==========================================
        print("=" * 50)
        print("1. CREATE - Setting a new secret")
        print("=" * 50)
        
        secret_value = "my-secret-value"
        secret = client.set_secret(secret_name, secret_value)
        
        print(f"✓ Secret created successfully!")
        print(f"  Name: {secret.name}")
        print(f"  Value: {secret.value}")
        print(f"  Version: {secret.properties.version}")
        print(f"  Created: {secret.properties.created_on}")
        print()
        
        # ==========================================
        # 2. READ - Retrieve the secret
        # ==========================================
        print("=" * 50)
        print("2. READ - Retrieving the secret")
        print("=" * 50)
        
        retrieved_secret = client.get_secret(secret_name)
        
        print(f"✓ Secret retrieved successfully!")
        print(f"  Name: {retrieved_secret.name}")
        print(f"  Value: {retrieved_secret.value}")
        print(f"  Version: {retrieved_secret.properties.version}")
        print()
        
        # ==========================================
        # 3. UPDATE - Update the secret value
        # ==========================================
        print("=" * 50)
        print("3. UPDATE - Updating the secret value")
        print("=" * 50)
        
        new_value = "updated-value"
        updated_secret = client.set_secret(secret_name, new_value)
        
        print(f"✓ Secret updated successfully!")
        print(f"  Name: {updated_secret.name}")
        print(f"  New Value: {updated_secret.value}")
        print(f"  New Version: {updated_secret.properties.version}")
        print(f"  Updated: {updated_secret.properties.updated_on}")
        print()
        
        # ==========================================
        # 4. DELETE - Delete and purge the secret
        # ==========================================
        print("=" * 50)
        print("4. DELETE - Deleting the secret")
        print("=" * 50)
        
        # Begin delete operation (returns a poller for soft-delete enabled vaults)
        delete_poller = client.begin_delete_secret(secret_name)
        deleted_secret = delete_poller.result()
        
        print(f"✓ Secret deletion initiated!")
        print(f"  Name: {deleted_secret.name}")
        print(f"  Deleted Date: {deleted_secret.deleted_date}")
        print(f"  Scheduled Purge Date: {deleted_secret.scheduled_purge_date}")
        print(f"  Recovery ID: {deleted_secret.recovery_id}")
        print()
        
        # Purge the deleted secret (permanent deletion for soft-delete enabled vaults)
        print("Purging the deleted secret...")
        client.purge_deleted_secret(secret_name)
        
        print(f"✓ Secret purged successfully!")
        print(f"  The secret '{secret_name}' has been permanently deleted.")
        print()
        
        print("=" * 50)
        print("All CRUD operations completed successfully!")
        print("=" * 50)
        
    except ResourceNotFoundError as e:
        print(f"Error: Secret not found - {e.message}")
        sys.exit(1)
    except HttpResponseError as e:
        print(f"HTTP Error: {e.message}")
        print(f"Status Code: {e.status_code}")
        if e.status_code == 403:
            print("\nPermission denied. Ensure you have the following permissions:")
            print("  - secrets/set (for create/update)")
            print("  - secrets/get (for read)")
            print("  - secrets/delete (for delete)")
            print("  - secrets/purge (for purge)")
        sys.exit(1)
    except Exception as e:
        print(f"An unexpected error occurred: {str(e)}")
        sys.exit(1)
    finally:
        # Clean up: close the client
        client.close()
        credential.close()


if __name__ == "__main__":
    main()
