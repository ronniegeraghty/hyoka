"""
Azure Key Vault Secret Error Handling Examples

This module demonstrates proper error handling patterns when working with
Azure Key Vault secrets using the azure-keyvault-secrets SDK.
"""

from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential
from azure.core.exceptions import (
    HttpResponseError,
    ResourceNotFoundError,
    ServiceRequestError
)
import time


# Initialize the client
vault_url = "https://your-keyvault-name.vault.azure.net/"
credential = DefaultAzureCredential()
client = SecretClient(vault_url=vault_url, credential=credential)


def handle_get_secret_basic(secret_name: str):
    """
    Basic error handling pattern for getting a secret.
    Handles common HTTP errors by checking status codes.
    """
    try:
        secret = client.get_secret(secret_name)
        print(f"Successfully retrieved secret: {secret.name}")
        return secret.value
    
    except HttpResponseError as e:
        # HttpResponseError contains status_code and error details
        status_code = e.status_code
        error_message = e.message
        
        if status_code == 403:
            # Access Denied - Missing RBAC permissions
            print(f"Access Denied (403): Your identity lacks permissions.")
            print(f"Error: {error_message}")
            print(f"Required role: 'Key Vault Secrets User' or 'Key Vault Secrets Officer'")
            
        elif status_code == 404:
            # Secret Not Found
            print(f"Secret Not Found (404): '{secret_name}' does not exist.")
            print(f"Error: {error_message}")
            
        elif status_code == 429:
            # Rate Limit / Throttling
            print(f"Throttling (429): Rate limit exceeded.")
            print(f"Error: {error_message}")
            # Check for Retry-After header if available
            retry_after = e.response.headers.get('Retry-After')
            if retry_after:
                print(f"Retry after {retry_after} seconds")
        
        else:
            # Other HTTP errors
            print(f"HTTP Error ({status_code}): {error_message}")
        
        # Re-raise or return None based on your needs
        raise


def handle_get_secret_with_specific_exceptions(secret_name: str):
    """
    Using specific exception types for cleaner error handling.
    ResourceNotFoundError is a subclass of HttpResponseError for 404s.
    """
    try:
        secret = client.get_secret(secret_name)
        return secret.value
    
    except ResourceNotFoundError as e:
        # This catches 404 specifically
        print(f"Secret '{secret_name}' not found (404)")
        print(f"Details: {e.message}")
        return None
    
    except HttpResponseError as e:
        if e.status_code == 403:
            print(f"Access denied: Check your RBAC role assignment")
            print(f"Error details: {e.message}")
        elif e.status_code == 429:
            print(f"Rate limited: {e.message}")
        else:
            print(f"HTTP error {e.status_code}: {e.message}")
        raise


def handle_get_secret_with_retry(secret_name: str, max_retries: int = 3):
    """
    Handling throttling (429) with exponential backoff retry logic.
    """
    retry_count = 0
    backoff = 1  # Initial backoff in seconds
    
    while retry_count < max_retries:
        try:
            secret = client.get_secret(secret_name)
            return secret.value
        
        except HttpResponseError as e:
            if e.status_code == 429:
                retry_count += 1
                if retry_count >= max_retries:
                    print(f"Max retries ({max_retries}) exceeded")
                    raise
                
                # Check for Retry-After header
                retry_after = e.response.headers.get('Retry-After')
                if retry_after:
                    wait_time = int(retry_after)
                else:
                    wait_time = backoff * (2 ** retry_count)  # Exponential backoff
                
                print(f"Rate limited. Retrying in {wait_time} seconds...")
                time.sleep(wait_time)
            
            elif e.status_code == 403:
                print(f"Access denied. Cannot retry.")
                raise
            
            elif e.status_code == 404:
                print(f"Secret not found. Cannot retry.")
                raise
            
            else:
                raise


def handle_soft_deleted_secret(secret_name: str):
    """
    Understanding soft-deleted secrets behavior.
    
    When a secret is deleted with soft-delete enabled (default in most vaults):
    - The secret moves to a "deleted" state
    - Attempting get_secret() raises ResourceNotFoundError (404)
    - The name is reserved and cannot be reused until purged
    - You must use get_deleted_secret() to access it
    - Recovery is possible using recover_deleted_secret()
    """
    try:
        # This will fail with 404 if the secret is soft-deleted
        secret = client.get_secret(secret_name)
        print(f"Secret is active: {secret.name}")
        return secret.value
    
    except ResourceNotFoundError as e:
        print(f"Secret not found (404). Checking if it's soft-deleted...")
        
        # Try to get the deleted secret
        try:
            deleted_secret = client.get_deleted_secret(secret_name)
            print(f"Secret is soft-deleted!")
            print(f"  Deleted on: {deleted_secret.deleted_date}")
            print(f"  Scheduled purge: {deleted_secret.scheduled_purge_date}")
            print(f"  Recovery ID: {deleted_secret.recovery_id}")
            print(f"  Can be recovered using: client.begin_recover_deleted_secret()")
            
            # To recover (requires 'Key Vault Secrets Officer' role):
            # poller = client.begin_recover_deleted_secret(secret_name)
            # recovered_secret = poller.result()
            
            return None
        
        except HttpResponseError as deleted_error:
            if deleted_error.status_code == 404:
                print(f"Secret truly does not exist (not even in deleted state)")
            elif deleted_error.status_code == 403:
                print(f"No permission to view deleted secrets")
            else:
                print(f"Error checking deleted secrets: {deleted_error.message}")
            return None


def comprehensive_error_handling(secret_name: str):
    """
    Comprehensive error handling covering all scenarios.
    """
    try:
        secret = client.get_secret(secret_name)
        print(f"✓ Successfully retrieved: {secret.name}")
        return secret.value
    
    except ResourceNotFoundError as e:
        # HTTP 404 - Secret not found
        print(f"✗ Secret not found (404)")
        print(f"  Message: {e.message}")
        print(f"  Error code: {e.error.code if hasattr(e, 'error') else 'N/A'}")
        print(f"  Tip: Check if secret is soft-deleted using get_deleted_secret()")
        return None
    
    except HttpResponseError as e:
        # Other HTTP errors
        status = e.status_code
        
        if status == 403:
            print(f"✗ Access Denied (403)")
            print(f"  Message: {e.message}")
            print(f"  Required RBAC roles:")
            print(f"    - Key Vault Secrets User (read-only)")
            print(f"    - Key Vault Secrets Officer (read/write)")
            print(f"  Identity: Check your DefaultAzureCredential configuration")
        
        elif status == 429:
            print(f"✗ Rate Limited (429)")
            print(f"  Message: {e.message}")
            retry_after = e.response.headers.get('Retry-After', 'Not specified')
            print(f"  Retry after: {retry_after}")
            print(f"  Tip: Implement exponential backoff retry logic")
        
        elif status == 401:
            print(f"✗ Authentication Failed (401)")
            print(f"  Message: {e.message}")
            print(f"  Tip: Check your credential configuration and token validity")
        
        else:
            print(f"✗ HTTP Error ({status})")
            print(f"  Message: {e.message}")
            # Access additional error details if available
            if hasattr(e, 'error') and e.error:
                print(f"  Error code: {e.error.code}")
                print(f"  Error details: {e.error.message}")
        
        raise
    
    except ServiceRequestError as e:
        # Network-level errors (DNS, connection failures)
        print(f"✗ Service Request Error: {e}")
        print(f"  Tip: Check network connectivity and vault URL")
        raise
    
    except Exception as e:
        # Catch-all for unexpected errors
        print(f"✗ Unexpected error: {type(e).__name__}: {e}")
        raise


def inspect_error_details(secret_name: str):
    """
    Demonstrates how to inspect all available error details from HttpResponseError.
    """
    try:
        secret = client.get_secret(secret_name)
        return secret.value
    
    except HttpResponseError as e:
        print("=== HttpResponseError Details ===")
        print(f"Status Code: {e.status_code}")
        print(f"Reason: {e.reason}")
        print(f"Message: {e.message}")
        
        # Response object details
        if e.response:
            print(f"\n=== Response Headers ===")
            for key, value in e.response.headers.items():
                print(f"{key}: {value}")
        
        # Error object (if available)
        if hasattr(e, 'error') and e.error:
            print(f"\n=== Error Object ===")
            print(f"Code: {e.error.code}")
            print(f"Message: {e.error.message}")
            if hasattr(e.error, 'innererror'):
                print(f"Inner Error: {e.error.innererror}")
        
        raise


# Example usage patterns
if __name__ == "__main__":
    secret_name = "my-secret"
    
    print("Example 1: Basic error handling")
    print("-" * 50)
    try:
        handle_get_secret_basic(secret_name)
    except Exception:
        pass
    
    print("\n\nExample 2: Specific exception handling")
    print("-" * 50)
    try:
        handle_get_secret_with_specific_exceptions(secret_name)
    except Exception:
        pass
    
    print("\n\nExample 3: Handling with retry logic")
    print("-" * 50)
    try:
        handle_get_secret_with_retry(secret_name)
    except Exception:
        pass
    
    print("\n\nExample 4: Soft-deleted secret handling")
    print("-" * 50)
    handle_soft_deleted_secret(secret_name)
    
    print("\n\nExample 5: Comprehensive error handling")
    print("-" * 50)
    try:
        comprehensive_error_handling(secret_name)
    except Exception:
        pass
