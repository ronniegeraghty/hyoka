#!/usr/bin/env python3
"""
Azure Key Vault Secrets CRUD Operations Demo

This script demonstrates all four CRUD operations on Azure Key Vault secrets:
1. Create a new secret
2. Read the secret back
3. Update the secret to a new value
4. Delete and purge the secret

Requirements:
    pip install azure-keyvault-secrets azure-identity

Prerequisites:
    - Azure Key Vault with soft-delete enabled
    - Set VAULT_URL environment variable (e.g., https://my-vault.vault.azure.net/)
    - Configure authentication (Azure CLI, managed identity, etc. for DefaultAzureCredential)
    - Required permissions: secrets/set, secrets/get, secrets/delete, secrets/purge
"""

import os
import sys
import time
from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient
from azure.core.exceptions import (
    ResourceNotFoundError,
    HttpResponseError,
    ClientAuthenticationError
)


def main():
    """Main function demonstrating CRUD operations on Azure Key Vault secrets."""
    
    # Get vault URL from environment variable
    vault_url = os.environ.get("VAULT_URL")
    if not vault_url:
        print("ERROR: VAULT_URL environment variable not set")
        print("Example: export VAULT_URL='https://my-vault.vault.azure.net/'")
        sys.exit(1)
    
    print(f"Connecting to Azure Key Vault: {vault_url}\n")
    
    # Initialize the credential and secret client
    try:
        credential = DefaultAzureCredential()
        client = SecretClient(vault_url=vault_url, credential=credential)
        print("✓ Successfully created SecretClient with DefaultAzureCredential\n")
    except ClientAuthenticationError as e:
        print(f"ERROR: Authentication failed: {e.message}")
        sys.exit(1)
    except Exception as e:
        print(f"ERROR: Failed to create SecretClient: {e}")
        sys.exit(1)
    
    secret_name = "my-secret"
    
    try:
        # ===================================================================
        # 1. CREATE - Set a new secret
        # ===================================================================
        print("=" * 70)
        print("1. CREATE - Setting a new secret")
        print("=" * 70)
        
        secret_value = "my-secret-value"
        print(f"Creating secret '{secret_name}' with value '{secret_value}'...")
        
        secret = client.set_secret(secret_name, secret_value)
        
        print(f"✓ Secret created successfully!")
        print(f"  Name: {secret.name}")
        print(f"  Value: {secret.value}")
        print(f"  Version: {secret.properties.version}")
        print(f"  Created: {secret.properties.created_on}")
        print()
        
        # ===================================================================
        # 2. READ - Retrieve the secret
        # ===================================================================
        print("=" * 70)
        print("2. READ - Retrieving the secret")
        print("=" * 70)
        
        print(f"Reading secret '{secret_name}'...")
        
        retrieved_secret = client.get_secret(secret_name)
        
        print(f"✓ Secret retrieved successfully!")
        print(f"  Name: {retrieved_secret.name}")
        print(f"  Value: {retrieved_secret.value}")
        print(f"  Version: {retrieved_secret.properties.version}")
        print(f"  Content Type: {retrieved_secret.properties.content_type}")
        print(f"  Enabled: {retrieved_secret.properties.enabled}")
        print()
        
        # ===================================================================
        # 3. UPDATE - Update the secret to a new value
        # ===================================================================
        print("=" * 70)
        print("3. UPDATE - Updating the secret value")
        print("=" * 70)
        
        new_value = "updated-value"
        print(f"Updating secret '{secret_name}' to new value '{new_value}'...")
        
        # set_secret creates a new version when the secret already exists
        updated_secret = client.set_secret(secret_name, new_value)
        
        print(f"✓ Secret updated successfully!")
        print(f"  Name: {updated_secret.name}")
        print(f"  New Value: {updated_secret.value}")
        print(f"  New Version: {updated_secret.properties.version}")
        print(f"  Updated: {updated_secret.properties.updated_on}")
        print()
        
        # Optionally, update secret properties (metadata) without changing the value
        print(f"Updating secret properties (content type)...")
        updated_properties = client.update_secret_properties(
            secret_name,
            content_type="text/plain"
        )
        print(f"✓ Secret properties updated!")
        print(f"  Content Type: {updated_properties.content_type}")
        print()
        
        # ===================================================================
        # 4. DELETE - Delete and purge the secret
        # ===================================================================
        print("=" * 70)
        print("4. DELETE - Deleting and purging the secret")
        print("=" * 70)
        
        print(f"Deleting secret '{secret_name}'...")
        
        # begin_delete_secret returns a poller for long-running operations
        delete_poller = client.begin_delete_secret(secret_name)
        
        # Wait for the deletion to complete
        deleted_secret = delete_poller.result()
        
        print(f"✓ Secret deleted successfully!")
        print(f"  Name: {deleted_secret.name}")
        print(f"  Deleted Date: {deleted_secret.deleted_date}")
        print(f"  Scheduled Purge Date: {deleted_secret.scheduled_purge_date}")
        print(f"  Recovery ID: {deleted_secret.recovery_id}")
        print()
        
        # Wait a moment to ensure deletion is complete before purging
        print("Waiting for deletion to propagate...")
        time.sleep(2)
        
        # Purge the deleted secret (permanent deletion)
        print(f"Purging deleted secret '{secret_name}'...")
        client.purge_deleted_secret(secret_name)
        
        print(f"✓ Secret purged successfully!")
        print(f"  The secret '{secret_name}' has been permanently deleted.")
        print()
        
        # Verify the secret no longer exists
        print("Verifying secret is completely removed...")
        try:
            client.get_secret(secret_name)
            print("WARNING: Secret still exists!")
        except ResourceNotFoundError:
            print("✓ Confirmed: Secret no longer exists in the vault.")
        print()
        
    except ResourceNotFoundError as e:
        print(f"ERROR: Resource not found: {e.message}")
        sys.exit(1)
    except HttpResponseError as e:
        print(f"ERROR: HTTP error occurred: {e.message}")
        print(f"Status code: {e.status_code}")
        sys.exit(1)
    except Exception as e:
        print(f"ERROR: An unexpected error occurred: {e}")
        sys.exit(1)
    finally:
        # Clean up - close the credential
        credential.close()
    
    print("=" * 70)
    print("All CRUD operations completed successfully!")
    print("=" * 70)


if __name__ == "__main__":
    main()
