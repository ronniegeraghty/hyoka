#!/usr/bin/env python3
"""
Azure Key Vault Secrets Pagination Example

This script demonstrates how to list secrets from an Azure Key Vault with
hundreds of secrets using the ItemPaged pagination pattern. It processes
secrets in pages and filters for enabled secrets only.

Requirements:
    pip install azure-keyvault-secrets azure-identity

Environment Variables:
    VAULT_URL: The URL of your Azure Key Vault (e.g., https://my-vault.vault.azure.net/)
"""

import os
from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient


def main():
    # Get the vault URL from environment variable
    vault_url = os.environ.get("VAULT_URL")
    if not vault_url:
        raise ValueError("VAULT_URL environment variable is not set")

    # Create a SecretClient using DefaultAzureCredential
    # DefaultAzureCredential automatically uses available authentication methods:
    # - Environment variables
    # - Managed Identity
    # - Azure CLI credentials
    # - Azure PowerShell credentials
    # - Interactive browser authentication
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)

    print(f"Connecting to Key Vault: {vault_url}")
    print("=" * 80)

    # list_properties_of_secrets() returns an ItemPaged[SecretProperties] object
    # This doesn't fetch all secrets at once - it uses pagination internally
    secret_properties = client.list_properties_of_secrets()

    # Process secrets page by page using by_page()
    # This is more efficient for large vaults as it processes secrets in batches
    print("\nProcessing secrets by page:\n")

    page_number = 0
    total_secrets = 0
    enabled_secrets = 0

    # by_page() returns an iterator of pages, where each page is an iterator of items
    for page in secret_properties.by_page():
        page_number += 1
        secrets_in_page = 0

        print(f"--- Page {page_number} ---")

        # Each page is an iterator of SecretProperties objects
        for secret_property in page:
            secrets_in_page += 1
            total_secrets += 1

            # Filter to show only enabled secrets
            if secret_property.enabled:
                enabled_secrets += 1

                # Print secret details
                # Note: list_properties_of_secrets() does NOT return secret values
                # Use client.get_secret(name) to retrieve the actual secret value
                print(f"\n  Secret Name: {secret_property.name}")

                # Content type is an optional field that can be used to describe the secret
                content_type = secret_property.content_type if secret_property.content_type else "Not specified"
                print(f"  Content Type: {content_type}")

                # Created date in UTC
                if secret_property.created_on:
                    created_date = secret_property.created_on.strftime("%Y-%m-%d %H:%M:%S UTC")
                    print(f"  Created On: {created_date}")
                else:
                    print(f"  Created On: Unknown")

                # Additional useful properties (commented out to keep output clean):
                # print(f"  Version: {secret_property.version}")
                # print(f"  Updated On: {secret_property.updated_on}")
                # print(f"  Enabled: {secret_property.enabled}")
                # if secret_property.expires_on:
                #     print(f"  Expires On: {secret_property.expires_on}")
                # if secret_property.tags:
                #     print(f"  Tags: {secret_property.tags}")

        print(f"\nSecrets in this page: {secrets_in_page}")
        print()

    # Summary
    print("=" * 80)
    print(f"\nSummary:")
    print(f"  Total pages processed: {page_number}")
    print(f"  Total secrets found: {total_secrets}")
    print(f"  Enabled secrets: {enabled_secrets}")
    print(f"  Disabled secrets: {total_secrets - enabled_secrets}")

    # Alternative: Iterate directly without by_page()
    # This still uses pagination internally but abstracts it away
    print("\n" + "=" * 80)
    print("\nAlternative approach - Direct iteration (also uses pagination internally):\n")

    enabled_count = 0
    for secret_property in client.list_properties_of_secrets():
        if secret_property.enabled:
            enabled_count += 1
            # Process each secret...
            # (simplified example, just counting)

    print(f"Enabled secrets (direct iteration): {enabled_count}")


if __name__ == "__main__":
    try:
        main()
    except Exception as e:
        print(f"Error: {e}")
        exit(1)
