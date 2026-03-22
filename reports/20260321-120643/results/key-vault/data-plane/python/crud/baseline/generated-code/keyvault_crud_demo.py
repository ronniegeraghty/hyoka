"""
Azure Key Vault Secrets CRUD Operations Demo

This script demonstrates all four CRUD operations on Azure Key Vault secrets:
1. Create a new secret
2. Read the secret back
3. Update the secret to a new value
4. Delete and purge the secret

Required pip packages:
    pip install azure-keyvault-secrets azure-identity

Prerequisites:
- An Azure Key Vault instance
- Set the VAULT_URL environment variable to your vault URL
  (e.g., https://my-key-vault.vault.azure.net/)
- Appropriate Azure credentials configured for DefaultAzureCredential
- Required permissions: secrets/set, secrets/get, secrets/delete, secrets/purge
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
        print("Example: export VAULT_URL='https://my-key-vault.vault.azure.net/'")
        sys.exit(1)

    # Authenticate using DefaultAzureCredential
    try:
        credential = DefaultAzureCredential()
        client = SecretClient(vault_url=vault_url, credential=credential)
        print(f"Connected to Key Vault: {vault_url}\n")
    except Exception as e:
        print(f"Error: Failed to authenticate or create client: {e}")
        sys.exit(1)

    secret_name = "my-secret"

    # ========================================================================
    # 1. CREATE: Set a new secret
    # ========================================================================
    print("=" * 70)
    print("1. CREATE: Setting a new secret")
    print("=" * 70)
    try:
        secret = client.set_secret(secret_name, "my-secret-value")
        print(f"✓ Secret created successfully!")
        print(f"  Name: {secret.name}")
        print(f"  Value: {secret.value}")
        print(f"  Version: {secret.properties.version}")
        print()
    except HttpResponseError as e:
        print(f"✗ Failed to create secret: {e.message}")
        sys.exit(1)
    except Exception as e:
        print(f"✗ Unexpected error during secret creation: {e}")
        sys.exit(1)

    # ========================================================================
    # 2. READ: Get the secret back
    # ========================================================================
    print("=" * 70)
    print("2. READ: Retrieving the secret")
    print("=" * 70)
    try:
        retrieved_secret = client.get_secret(secret_name)
        print(f"✓ Secret retrieved successfully!")
        print(f"  Name: {retrieved_secret.name}")
        print(f"  Value: {retrieved_secret.value}")
        print(f"  Version: {retrieved_secret.properties.version}")
        print()
    except ResourceNotFoundError:
        print(f"✗ Secret '{secret_name}' not found")
        sys.exit(1)
    except HttpResponseError as e:
        print(f"✗ Failed to retrieve secret: {e.message}")
        sys.exit(1)
    except Exception as e:
        print(f"✗ Unexpected error during secret retrieval: {e}")
        sys.exit(1)

    # ========================================================================
    # 3. UPDATE: Set a new value for the secret
    # ========================================================================
    print("=" * 70)
    print("3. UPDATE: Updating the secret to a new value")
    print("=" * 70)
    try:
        # set_secret creates a new version when the secret name already exists
        updated_secret = client.set_secret(secret_name, "updated-value")
        print(f"✓ Secret updated successfully!")
        print(f"  Name: {updated_secret.name}")
        print(f"  Value: {updated_secret.value}")
        print(f"  New Version: {updated_secret.properties.version}")
        print()
    except HttpResponseError as e:
        print(f"✗ Failed to update secret: {e.message}")
        sys.exit(1)
    except Exception as e:
        print(f"✗ Unexpected error during secret update: {e}")
        sys.exit(1)

    # ========================================================================
    # 4. DELETE: Delete and purge the secret
    # ========================================================================
    print("=" * 70)
    print("4. DELETE: Deleting and purging the secret")
    print("=" * 70)
    try:
        # Delete the secret (soft-delete)
        print(f"Deleting secret '{secret_name}'...")
        delete_poller = client.begin_delete_secret(secret_name)
        deleted_secret = delete_poller.result()
        print(f"✓ Secret soft-deleted successfully!")
        print(f"  Name: {deleted_secret.name}")
        print(f"  Deleted Date: {deleted_secret.deleted_date}")
        print(f"  Scheduled Purge Date: {deleted_secret.scheduled_purge_date}")
        print()

        # Purge the secret (permanent deletion)
        print(f"Purging secret '{secret_name}' permanently...")
        client.purge_deleted_secret(secret_name)
        print(f"✓ Secret purged successfully!")
        print(f"  The secret '{secret_name}' has been permanently deleted.")
        print()
    except HttpResponseError as e:
        print(f"✗ Failed to delete/purge secret: {e.message}")
        sys.exit(1)
    except Exception as e:
        print(f"✗ Unexpected error during secret deletion/purge: {e}")
        sys.exit(1)

    print("=" * 70)
    print("All CRUD operations completed successfully!")
    print("=" * 70)


if __name__ == "__main__":
    main()
