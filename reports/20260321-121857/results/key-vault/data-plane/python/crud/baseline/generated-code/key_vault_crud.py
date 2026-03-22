"""
Azure Key Vault Secrets CRUD Operations Demo

This script demonstrates all four CRUD operations on Azure Key Vault secrets:
1. Create a new secret
2. Read the secret back
3. Update the secret to a new value
4. Delete and purge the secret

Requirements:
    pip install azure-keyvault-secrets azure-identity

Authentication:
    Uses DefaultAzureCredential which supports multiple authentication methods:
    - Environment variables (AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET)
    - Managed Identity (when running on Azure)
    - Azure CLI (az login)
    - Visual Studio Code
    - Azure PowerShell
    - Interactive browser

Prerequisites:
    - An Azure Key Vault with the URL set in the VAULT_URL environment variable
    - Appropriate permissions: secrets/set, secrets/get, secrets/delete, secrets/purge
    - Soft-delete must be enabled on the vault for purge operation
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
    
    # Get the vault URL from environment variable
    vault_url = os.environ.get("VAULT_URL")
    if not vault_url:
        print("Error: VAULT_URL environment variable is not set.")
        print("Example: export VAULT_URL='https://your-vault-name.vault.azure.net/'")
        sys.exit(1)
    
    print(f"Connecting to Key Vault: {vault_url}\n")
    
    # Initialize the credential and client
    try:
        credential = DefaultAzureCredential()
        client = SecretClient(vault_url=vault_url, credential=credential)
    except ClientAuthenticationError as e:
        print(f"Authentication Error: {e.message}")
        print("\nPlease ensure you are authenticated using one of:")
        print("  - Azure CLI: az login")
        print("  - Environment variables: AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET")
        print("  - Managed Identity (when running on Azure)")
        sys.exit(1)
    except Exception as e:
        print(f"Error initializing client: {e}")
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
        print(f"Setting secret '{secret_name}' with value '{initial_value}'...")
        
        secret = client.set_secret(secret_name, initial_value)
        
        print(f"✓ Secret created successfully!")
        print(f"  Name: {secret.name}")
        print(f"  Value: {secret.value}")
        print(f"  Version: {secret.properties.version}")
        print(f"  Created: {secret.properties.created_on}")
        print()
        
        # ========================================
        # 2. READ - Retrieve the secret
        # ========================================
        print("=" * 60)
        print("2. READ - Retrieving the secret")
        print("=" * 60)
        
        print(f"Getting secret '{secret_name}'...")
        
        retrieved_secret = client.get_secret(secret_name)
        
        print(f"✓ Secret retrieved successfully!")
        print(f"  Name: {retrieved_secret.name}")
        print(f"  Value: {retrieved_secret.value}")
        print(f"  Version: {retrieved_secret.properties.version}")
        print(f"  Enabled: {retrieved_secret.properties.enabled}")
        print()
        
        # ========================================
        # 3. UPDATE - Update the secret value
        # ========================================
        print("=" * 60)
        print("3. UPDATE - Updating the secret value")
        print("=" * 60)
        
        updated_value = "updated-value"
        print(f"Updating secret '{secret_name}' to new value '{updated_value}'...")
        
        # set_secret creates a new version when the secret already exists
        updated_secret = client.set_secret(secret_name, updated_value)
        
        print(f"✓ Secret updated successfully!")
        print(f"  Name: {updated_secret.name}")
        print(f"  Value: {updated_secret.value}")
        print(f"  New Version: {updated_secret.properties.version}")
        print(f"  Updated: {updated_secret.properties.updated_on}")
        print()
        
        # Optional: Update secret properties (metadata) without changing the value
        print(f"Updating secret properties (metadata)...")
        
        updated_properties = client.update_secret_properties(
            secret_name,
            content_type="text/plain",
            enabled=True
        )
        
        print(f"✓ Secret properties updated!")
        print(f"  Content Type: {updated_properties.content_type}")
        print(f"  Enabled: {updated_properties.enabled}")
        print()
        
        # ========================================
        # 4. DELETE - Delete and purge the secret
        # ========================================
        print("=" * 60)
        print("4. DELETE - Deleting and purging the secret")
        print("=" * 60)
        
        print(f"Deleting secret '{secret_name}'...")
        
        # begin_delete_secret returns a poller for long-running operations
        delete_poller = client.begin_delete_secret(secret_name)
        
        # Wait for the deletion to complete
        deleted_secret = delete_poller.result()
        
        print(f"✓ Secret deleted successfully!")
        print(f"  Name: {deleted_secret.name}")
        print(f"  Deleted On: {deleted_secret.deleted_date}")
        print(f"  Scheduled Purge Date: {deleted_secret.scheduled_purge_date}")
        print(f"  Recovery ID: {deleted_secret.recovery_id}")
        print()
        
        # Purge the deleted secret (permanent deletion)
        # This is only possible in vaults with soft-delete enabled
        print(f"Purging deleted secret '{secret_name}' (permanent deletion)...")
        
        client.purge_deleted_secret(secret_name)
        
        print(f"✓ Secret purged successfully!")
        print(f"  The secret has been permanently deleted and cannot be recovered.")
        print()
        
        # ========================================
        # Verification
        # ========================================
        print("=" * 60)
        print("Verification - Attempting to retrieve the deleted secret")
        print("=" * 60)
        
        try:
            client.get_secret(secret_name)
            print("⚠ Warning: Secret still exists (unexpected)")
        except ResourceNotFoundError:
            print(f"✓ Confirmed: Secret '{secret_name}' no longer exists")
        print()
        
        print("=" * 60)
        print("CRUD Operations Completed Successfully!")
        print("=" * 60)
        
    except HttpResponseError as e:
        print(f"\n❌ HTTP Error: {e.message}")
        print(f"   Status Code: {e.status_code}")
        print(f"   Error Code: {e.error.code if hasattr(e, 'error') else 'N/A'}")
        
        if e.status_code == 403:
            print("\n   This usually means insufficient permissions.")
            print("   Required permissions: secrets/set, secrets/get, secrets/delete, secrets/purge")
        
        sys.exit(1)
        
    except ResourceNotFoundError as e:
        print(f"\n❌ Resource Not Found: {e.message}")
        sys.exit(1)
        
    except Exception as e:
        print(f"\n❌ Unexpected Error: {type(e).__name__}: {e}")
        sys.exit(1)
        
    finally:
        # Clean up - close the credential
        credential.close()


if __name__ == "__main__":
    main()
