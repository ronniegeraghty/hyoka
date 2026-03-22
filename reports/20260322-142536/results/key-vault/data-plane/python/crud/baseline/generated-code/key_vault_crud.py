#!/usr/bin/env python3
"""
Azure Key Vault Secrets CRUD Operations Demo

This script demonstrates all four CRUD operations on Azure Key Vault secrets:
- Create a secret
- Read a secret
- Update a secret
- Delete and purge a secret

Required pip packages:
    pip install azure-keyvault-secrets azure-identity
"""

import os
import sys
from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential
from azure.core.exceptions import ResourceNotFoundError, HttpResponseError


def main():
    # Get Key Vault URL from environment variable
    vault_url = os.getenv("AZURE_KEY_VAULT_URL")
    
    if not vault_url:
        print("Error: AZURE_KEY_VAULT_URL environment variable is not set", file=sys.stderr)
        print("Example: export AZURE_KEY_VAULT_URL='https://your-vault-name.vault.azure.net/'", file=sys.stderr)
        sys.exit(1)
    
    try:
        # Authenticate using DefaultAzureCredential
        credential = DefaultAzureCredential()
        
        # Create a SecretClient
        client = SecretClient(vault_url=vault_url, credential=credential)
        
        secret_name = "my-secret"
        
        # 1. CREATE: Create a new secret
        print(f"Creating secret '{secret_name}'...")
        try:
            created_secret = client.set_secret(secret_name, "my-secret-value")
            print(f"✓ Secret created successfully")
            print(f"  Name: {created_secret.name}")
            print(f"  Value: {created_secret.value}")
            print(f"  Version: {created_secret.properties.version}")
            print()
        except HttpResponseError as e:
            print(f"✗ Failed to create secret: {e.message}", file=sys.stderr)
            sys.exit(1)
        
        # 2. READ: Read the secret back
        print(f"Reading secret '{secret_name}'...")
        try:
            retrieved_secret = client.get_secret(secret_name)
            print(f"✓ Secret retrieved successfully")
            print(f"  Name: {retrieved_secret.name}")
            print(f"  Value: {retrieved_secret.value}")
            print(f"  Version: {retrieved_secret.properties.version}")
            print()
        except ResourceNotFoundError:
            print(f"✗ Secret '{secret_name}' not found", file=sys.stderr)
            sys.exit(1)
        except HttpResponseError as e:
            print(f"✗ Failed to read secret: {e.message}", file=sys.stderr)
            sys.exit(1)
        
        # 3. UPDATE: Update the secret to a new value
        print(f"Updating secret '{secret_name}'...")
        try:
            updated_secret = client.set_secret(secret_name, "updated-value")
            print(f"✓ Secret updated successfully")
            print(f"  Name: {updated_secret.name}")
            print(f"  Value: {updated_secret.value}")
            print(f"  Version: {updated_secret.properties.version}")
            print()
        except HttpResponseError as e:
            print(f"✗ Failed to update secret: {e.message}", file=sys.stderr)
            sys.exit(1)
        
        # 4. DELETE: Delete the secret
        print(f"Deleting secret '{secret_name}'...")
        try:
            deleted_secret = client.begin_delete_secret(secret_name)
            # Wait for the delete operation to complete
            deleted_secret_result = deleted_secret.result()
            print(f"✓ Secret deleted successfully")
            print(f"  Name: {deleted_secret_result.name}")
            print(f"  Deleted on: {deleted_secret_result.deleted_date}")
            print(f"  Scheduled purge: {deleted_secret_result.scheduled_purge_date}")
            print()
        except ResourceNotFoundError:
            print(f"✗ Secret '{secret_name}' not found", file=sys.stderr)
            sys.exit(1)
        except HttpResponseError as e:
            print(f"✗ Failed to delete secret: {e.message}", file=sys.stderr)
            sys.exit(1)
        
        # PURGE: Purge the deleted secret (for soft-delete enabled vaults)
        print(f"Purging secret '{secret_name}'...")
        try:
            client.purge_deleted_secret(secret_name)
            print(f"✓ Secret purged successfully")
            print(f"  The secret '{secret_name}' has been permanently deleted")
        except ResourceNotFoundError:
            print(f"✗ Deleted secret '{secret_name}' not found", file=sys.stderr)
            sys.exit(1)
        except HttpResponseError as e:
            print(f"✗ Failed to purge secret: {e.message}", file=sys.stderr)
            print(f"  Note: Purging may require additional permissions or the vault may not have soft-delete enabled", file=sys.stderr)
            sys.exit(1)
    
    except Exception as e:
        print(f"✗ Unexpected error: {str(e)}", file=sys.stderr)
        sys.exit(1)
    finally:
        # Close the credential
        credential.close()
    
    print("\n✓ All CRUD operations completed successfully!")


if __name__ == "__main__":
    main()
