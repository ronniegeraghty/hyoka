"""
Azure Key Vault Secrets - Error Handling Patterns

This module demonstrates proper error handling when working with Azure Key Vault
secrets, including handling 403 (access denied), 404 (not found), 429 (throttling),
and soft-deleted secrets.
"""

from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential
from azure.core.exceptions import HttpResponseError, ResourceNotFoundError
import time


def handle_keyvault_errors_basic(vault_url: str, secret_name: str):
    """
    Basic error handling pattern with status code inspection.
    
    Demonstrates how to:
    - Catch HttpResponseError
    - Inspect status_code property
    - Access error message details
    """
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    try:
        secret = client.get_secret(secret_name)
        print(f"Successfully retrieved secret: {secret.name}")
        return secret.value
        
    except HttpResponseError as e:
        # Inspect the status code to determine the error type
        status_code = e.status_code
        error_message = e.message
        
        if status_code == 403:
            # Access Denied - RBAC permissions issue
            print(f"Access Denied (403): {error_message}")
            print("Your identity lacks the required RBAC role.")
            print("Required role: 'Key Vault Secrets User' or 'Key Vault Secrets Officer'")
            
        elif status_code == 404:
            # Secret Not Found
            print(f"Secret Not Found (404): {error_message}")
            print(f"The secret '{secret_name}' does not exist in the vault.")
            
        elif status_code == 429:
            # Rate Limiting / Throttling
            print(f"Throttling (429): {error_message}")
            print("Too many requests. Service is rate limiting your requests.")
            # Check for Retry-After header
            if hasattr(e, 'response') and e.response:
                retry_after = e.response.headers.get('Retry-After')
                if retry_after:
                    print(f"Retry after {retry_after} seconds")
                    
        else:
            # Other HTTP errors
            print(f"HTTP Error ({status_code}): {error_message}")
            
        # Re-raise if you want calling code to handle it
        raise


def handle_specific_errors_separately(vault_url: str, secret_name: str):
    """
    Pattern with separate exception handlers for different scenarios.
    
    Uses ResourceNotFoundError for 404s (more specific than HttpResponseError).
    """
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    try:
        secret = client.get_secret(secret_name)
        return secret.value
        
    except ResourceNotFoundError as e:
        # More specific exception for 404 errors
        print(f"Secret '{secret_name}' not found: {e.message}")
        return None
        
    except HttpResponseError as e:
        if e.status_code == 403:
            print(f"Access denied. Check RBAC permissions: {e.message}")
            # Log additional details
            if hasattr(e, 'error'):
                print(f"Error code: {e.error.code}")
            raise
            
        elif e.status_code == 429:
            print(f"Rate limited: {e.message}")
            # Implement retry logic with exponential backoff
            raise
            
        else:
            print(f"Unexpected error ({e.status_code}): {e.message}")
            raise


def handle_with_retry_on_throttling(vault_url: str, secret_name: str, max_retries: int = 3):
    """
    Error handling with automatic retry for throttling (429).
    
    Implements exponential backoff when encountering rate limits.
    """
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    retry_count = 0
    base_delay = 1  # seconds
    
    while retry_count <= max_retries:
        try:
            secret = client.get_secret(secret_name)
            return secret.value
            
        except HttpResponseError as e:
            if e.status_code == 429 and retry_count < max_retries:
                # Calculate retry delay (exponential backoff)
                retry_after = base_delay * (2 ** retry_count)
                
                # Check if server provided Retry-After header
                if hasattr(e, 'response') and e.response:
                    server_retry_after = e.response.headers.get('Retry-After')
                    if server_retry_after:
                        retry_after = int(server_retry_after)
                
                print(f"Throttled (429). Retrying in {retry_after} seconds... (attempt {retry_count + 1}/{max_retries})")
                time.sleep(retry_after)
                retry_count += 1
                
            elif e.status_code == 403:
                print(f"Access Denied (403): {e.message}")
                print("RBAC role required: Key Vault Secrets User or Key Vault Secrets Officer")
                raise
                
            elif e.status_code == 404:
                print(f"Secret not found (404): {e.message}")
                return None
                
            else:
                # For other errors or exceeded retries, raise
                print(f"Error ({e.status_code}): {e.message}")
                raise
        
        except Exception as e:
            print(f"Unexpected error: {type(e).__name__}: {str(e)}")
            raise
    
    # If we exhausted all retries
    raise Exception(f"Failed to retrieve secret after {max_retries} retries due to throttling")


def handle_soft_deleted_secrets(vault_url: str, secret_name: str):
    """
    Demonstrates behavior when attempting to access soft-deleted secrets.
    
    Key Points:
    1. get_secret() on a soft-deleted secret raises ResourceNotFoundError (404)
    2. The secret exists in the deleted state but cannot be retrieved via get_secret()
    3. You must use get_deleted_secret() to access properties of deleted secrets
    4. You cannot retrieve the VALUE of a deleted secret
    5. To use the secret name again, either purge or recover it first
    """
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    print(f"\n=== Attempting to get soft-deleted secret '{secret_name}' ===")
    
    try:
        # This will fail with 404 if the secret is soft-deleted
        secret = client.get_secret(secret_name)
        print(f"Secret retrieved successfully: {secret.name}")
        return secret.value
        
    except ResourceNotFoundError as e:
        print(f"Secret not found (404): {e.message}")
        print("\nChecking if secret is in deleted state...")
        
        try:
            # Try to get the deleted secret
            deleted_secret = client.get_deleted_secret(secret_name)
            
            print(f"\n✓ Secret '{secret_name}' exists but is SOFT-DELETED")
            print(f"  Deleted on: {deleted_secret.deleted_date}")
            print(f"  Scheduled purge date: {deleted_secret.scheduled_purge_date}")
            print(f"  Recovery ID: {deleted_secret.recovery_id}")
            
            print("\nTo use this secret:")
            print("  Option 1: Recover it using client.begin_recover_deleted_secret()")
            print("  Option 2: Purge it using client.purge_deleted_secret() (requires purge permission)")
            print("\nNote: You CANNOT retrieve the value of a deleted secret.")
            
            return None
            
        except ResourceNotFoundError:
            print(f"\n✗ Secret '{secret_name}' does not exist (not active or deleted)")
            return None
            
        except HttpResponseError as deleted_error:
            if deleted_error.status_code == 403:
                print("\n✗ Cannot check deleted secrets - missing 'List' permission")
                print("  Required: 'Key Vault Secrets User' role or 'List' data action")
            else:
                print(f"\nError checking deleted secrets: {deleted_error.message}")
            raise
    
    except HttpResponseError as e:
        if e.status_code == 403:
            print(f"Access Denied (403): {e.message}")
            print("Check RBAC role assignment on the Key Vault")
        elif e.status_code == 429:
            print(f"Throttled (429): {e.message}")
        else:
            print(f"HTTP Error ({e.status_code}): {e.message}")
        raise


def inspect_error_details(vault_url: str, secret_name: str):
    """
    Demonstrates how to inspect all available properties on HttpResponseError.
    
    Useful properties:
    - status_code: HTTP status code (int)
    - message: Error message (str)
    - error: Structured error object (if available)
    - response: Raw HTTP response (if available)
    """
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    try:
        secret = client.get_secret(secret_name)
        return secret.value
        
    except HttpResponseError as e:
        print("=== HttpResponseError Details ===")
        print(f"Status Code: {e.status_code}")
        print(f"Message: {e.message}")
        
        # Check for structured error object
        if hasattr(e, 'error') and e.error:
            print(f"Error Code: {e.error.code if hasattr(e.error, 'code') else 'N/A'}")
            print(f"Error Message: {e.error.message if hasattr(e.error, 'message') else 'N/A'}")
        
        # Check response headers
        if hasattr(e, 'response') and e.response:
            print("\nResponse Headers:")
            for header, value in e.response.headers.items():
                if header.lower() in ['retry-after', 'x-ms-request-id', 'x-ms-keyvault-service-version']:
                    print(f"  {header}: {value}")
        
        # Additional context
        print(f"\nException Type: {type(e).__name__}")
        print(f"String Representation: {str(e)}")
        
        raise


def main():
    """
    Example usage demonstrating different error handling patterns.
    """
    vault_url = "https://your-keyvault-name.vault.azure.net/"
    
    print("Example 1: Basic error handling with status code inspection")
    print("-" * 60)
    try:
        handle_keyvault_errors_basic(vault_url, "my-secret")
    except HttpResponseError:
        pass
    
    print("\n\nExample 2: Handling throttling with retry logic")
    print("-" * 60)
    try:
        handle_with_retry_on_throttling(vault_url, "my-secret", max_retries=3)
    except Exception:
        pass
    
    print("\n\nExample 3: Soft-deleted secret handling")
    print("-" * 60)
    try:
        handle_soft_deleted_secrets(vault_url, "deleted-secret")
    except HttpResponseError:
        pass
    
    print("\n\nExample 4: Detailed error inspection")
    print("-" * 60)
    try:
        inspect_error_details(vault_url, "my-secret")
    except HttpResponseError:
        pass


if __name__ == "__main__":
    main()
