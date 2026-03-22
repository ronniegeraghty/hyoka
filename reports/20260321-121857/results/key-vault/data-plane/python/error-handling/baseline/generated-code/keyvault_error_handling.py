#!/usr/bin/env python3
"""
Azure Key Vault Secrets - Comprehensive Error Handling Guide

This script demonstrates proper error handling patterns for common scenarios
when working with Azure Key Vault secrets using azure-keyvault-secrets SDK.
"""

from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient
from azure.core.exceptions import (
    HttpResponseError,
    ResourceNotFoundError,
    ServiceRequestError
)
import time


def handle_specific_errors(secret_client: SecretClient, secret_name: str):
    """
    Demonstrates handling specific HTTP status codes (403, 404, 429)
    by inspecting the HttpResponseError exception.
    """
    try:
        secret = secret_client.get_secret(secret_name)
        print(f"Successfully retrieved secret: {secret.name}")
        return secret
    
    except HttpResponseError as e:
        # Extract status code and error message
        status_code = e.status_code
        error_message = e.message
        
        # Handle specific error codes
        if status_code == 403:
            # Access Denied - RBAC permission issue
            print(f"ERROR 403 - Access Denied")
            print(f"Message: {error_message}")
            print(f"Your identity lacks the required RBAC role.")
            print(f"Required role: 'Key Vault Secrets User' or 'Key Vault Secrets Officer'")
            print(f"Check Azure Portal → Key Vault → Access control (IAM)")
            # In production, you might want to log this and return None or raise a custom exception
            
        elif status_code == 404:
            # Secret Not Found
            print(f"ERROR 404 - Secret Not Found")
            print(f"Message: {error_message}")
            print(f"Secret '{secret_name}' does not exist in this Key Vault")
            # Could be that the secret was never created or has been purged
            
        elif status_code == 429:
            # Rate Limit / Throttling
            print(f"ERROR 429 - Too Many Requests (Throttling)")
            print(f"Message: {error_message}")
            
            # Extract retry-after header if available
            retry_after = None
            if hasattr(e, 'response') and e.response:
                retry_after = e.response.headers.get('Retry-After')
            
            if retry_after:
                print(f"Retry after: {retry_after} seconds")
                # In production, implement exponential backoff
                time.sleep(int(retry_after))
            else:
                # Default backoff if no Retry-After header
                print("No Retry-After header, using exponential backoff")
                time.sleep(5)
            
        else:
            # Other HTTP errors
            print(f"ERROR {status_code}")
            print(f"Message: {error_message}")
        
        # Always available attributes on HttpResponseError
        print(f"\nDetailed error info:")
        print(f"  Status Code: {e.status_code}")
        print(f"  Error Code: {e.error.code if hasattr(e, 'error') and e.error else 'N/A'}")
        print(f"  Message: {e.message}")
        
        return None


def handle_soft_deleted_secret(secret_client: SecretClient, secret_name: str):
    """
    Demonstrates what happens when trying to get a soft-deleted secret.
    
    When a secret is soft-deleted:
    - get_secret() will raise HttpResponseError with status_code 404
    - The error message will indicate the secret is deleted
    - You can recover it using begin_recover_deleted_secret()
    - Or permanently delete it using purge_deleted_secret()
    """
    try:
        secret = secret_client.get_secret(secret_name)
        print(f"Secret retrieved: {secret.name}")
        return secret
    
    except HttpResponseError as e:
        if e.status_code == 404:
            print(f"Secret '{secret_name}' not found (404)")
            
            # Check if it's soft-deleted by trying to get deleted secret
            try:
                deleted_secret = secret_client.get_deleted_secret(secret_name)
                print(f"\nSecret is SOFT-DELETED:")
                print(f"  Name: {deleted_secret.name}")
                print(f"  Deleted on: {deleted_secret.deleted_date}")
                print(f"  Scheduled purge: {deleted_secret.scheduled_purge_date}")
                print(f"  Recovery ID: {deleted_secret.recovery_id}")
                
                print(f"\nTo recover: secret_client.begin_recover_deleted_secret('{secret_name}')")
                print(f"To purge: secret_client.purge_deleted_secret('{secret_name}')")
                
                return None
                
            except HttpResponseError as delete_error:
                if delete_error.status_code == 404:
                    print(f"Secret '{secret_name}' does not exist and is not soft-deleted")
                elif delete_error.status_code == 403:
                    print(f"Cannot check deleted secrets - missing 'list' or 'get' permission on deleted secrets")
                else:
                    print(f"Error checking deleted secret: {delete_error.message}")
            
        return None


def robust_get_secret_with_retry(secret_client: SecretClient, secret_name: str, 
                                  max_retries: int = 3):
    """
    Production-ready pattern with exponential backoff for throttling.
    """
    retry_count = 0
    base_delay = 1
    
    while retry_count < max_retries:
        try:
            secret = secret_client.get_secret(secret_name)
            return secret
        
        except HttpResponseError as e:
            if e.status_code == 429:
                # Throttling - implement exponential backoff
                retry_count += 1
                
                # Check for Retry-After header
                retry_after = None
                if hasattr(e, 'response') and e.response:
                    retry_after = e.response.headers.get('Retry-After')
                
                if retry_after:
                    delay = int(retry_after)
                else:
                    # Exponential backoff: 1s, 2s, 4s, 8s...
                    delay = base_delay * (2 ** (retry_count - 1))
                
                print(f"Throttled (429). Retry {retry_count}/{max_retries} after {delay}s...")
                time.sleep(delay)
                
            elif e.status_code == 403:
                # Permission error - don't retry
                print(f"Access denied (403): {e.message}")
                raise  # Re-raise, as retrying won't help
            
            elif e.status_code == 404:
                # Not found - don't retry
                print(f"Secret not found (404): {e.message}")
                return None
            
            else:
                # Other errors - log and potentially retry
                retry_count += 1
                if retry_count < max_retries:
                    delay = base_delay * (2 ** (retry_count - 1))
                    print(f"Error {e.status_code}. Retry {retry_count}/{max_retries} after {delay}s...")
                    time.sleep(delay)
                else:
                    raise
    
    raise Exception(f"Failed to retrieve secret after {max_retries} retries")


def comprehensive_error_handling_example():
    """
    Complete example showing all error handling patterns together.
    """
    # Initialize client
    vault_url = "https://your-keyvault-name.vault.azure.net/"
    credential = DefaultAzureCredential()
    secret_client = SecretClient(vault_url=vault_url, credential=credential)
    
    secret_name = "my-secret"
    
    try:
        # Attempt to get secret with comprehensive error handling
        secret = secret_client.get_secret(secret_name)
        
        # Success - use the secret value
        print(f"Secret Value: {secret.value}")
        print(f"Secret Version: {secret.properties.version}")
        print(f"Enabled: {secret.properties.enabled}")
        
    except ResourceNotFoundError as e:
        # More specific exception for 404 errors
        print(f"Secret not found: {e.message}")
        # Check if soft-deleted
        handle_soft_deleted_secret(secret_client, secret_name)
        
    except HttpResponseError as e:
        status_code = e.status_code
        
        if status_code == 403:
            print(f"Access Denied (403)")
            print(f"Error: {e.message}")
            print(f"\nRequired Azure RBAC roles:")
            print(f"  - Key Vault Secrets User (read-only)")
            print(f"  - Key Vault Secrets Officer (read/write)")
            print(f"\nGrant access via:")
            print(f"  az role assignment create \\")
            print(f"    --role 'Key Vault Secrets User' \\")
            print(f"    --assignee <your-identity> \\")
            print(f"    --scope /subscriptions/<sub-id>/resourceGroups/<rg>/providers/Microsoft.KeyVault/vaults/<kv-name>")
            
        elif status_code == 429:
            print(f"Rate Limit Exceeded (429)")
            print(f"Key Vault has request limits - implement exponential backoff")
            # Use the robust retry function
            secret = robust_get_secret_with_retry(secret_client, secret_name)
            
        else:
            print(f"HTTP Error {status_code}: {e.message}")
            
    except ServiceRequestError as e:
        # Network-related errors (DNS, connection, etc.)
        print(f"Network/Service Error: {e.message}")
        print(f"Check network connectivity and Key Vault URL")
        
    except Exception as e:
        # Catch-all for unexpected errors
        print(f"Unexpected error: {type(e).__name__}: {str(e)}")


def inspect_error_details():
    """
    Shows how to extract detailed information from HttpResponseError.
    """
    vault_url = "https://your-keyvault-name.vault.azure.net/"
    credential = DefaultAzureCredential()
    secret_client = SecretClient(vault_url=vault_url, credential=credential)
    
    try:
        secret = secret_client.get_secret("non-existent-secret")
    
    except HttpResponseError as e:
        print("=== HttpResponseError Details ===\n")
        
        # Status code (most important)
        print(f"Status Code: {e.status_code}")
        
        # Error message (human-readable)
        print(f"Message: {e.message}")
        
        # Error code (Azure-specific error code)
        if hasattr(e, 'error') and e.error:
            print(f"Error Code: {e.error.code}")
            print(f"Error Message: {e.error.message}")
        
        # Response object (for headers, etc.)
        if hasattr(e, 'response') and e.response:
            print(f"\nResponse Headers:")
            for key, value in e.response.headers.items():
                print(f"  {key}: {value}")
        
        # Additional attributes
        print(f"\nReason: {e.reason if hasattr(e, 'reason') else 'N/A'}")
        print(f"Exception Type: {type(e).__name__}")


if __name__ == "__main__":
    print("Azure Key Vault Secrets - Error Handling Patterns\n")
    print("This script demonstrates error handling patterns.")
    print("Update 'vault_url' with your Key Vault URL before running.\n")
    
    # Uncomment the example you want to run:
    # comprehensive_error_handling_example()
    # inspect_error_details()
