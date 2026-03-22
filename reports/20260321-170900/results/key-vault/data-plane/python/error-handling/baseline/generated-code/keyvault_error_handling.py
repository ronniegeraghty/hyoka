"""
Azure Key Vault Secrets - Error Handling Examples

This module demonstrates proper error handling patterns when working with
Azure Key Vault secrets using the azure-keyvault-secrets SDK.
"""

from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient
from azure.core.exceptions import (
    HttpResponseError,
    ResourceNotFoundError,
    ServiceRequestError,
)
import time


def create_secret_client(vault_url: str) -> SecretClient:
    """Create and return a SecretClient instance."""
    credential = DefaultAzureCredential()
    return SecretClient(vault_url=vault_url, credential=credential)


# ==============================================================================
# EXAMPLE 1: Basic Error Handling - Inspecting HttpResponseError
# ==============================================================================

def get_secret_with_basic_error_handling(client: SecretClient, secret_name: str):
    """
    Demonstrates basic error handling with HttpResponseError inspection.
    Shows how to access status_code, error message, and other properties.
    """
    try:
        secret = client.get_secret(secret_name)
        print(f"Successfully retrieved secret: {secret.name}")
        return secret.value
    
    except HttpResponseError as e:
        # HttpResponseError has several useful properties:
        # - status_code: HTTP status code (403, 404, 429, etc.)
        # - message: Human-readable error message
        # - error: Detailed error information
        # - reason: HTTP reason phrase
        
        print(f"HTTP Error occurred:")
        print(f"  Status Code: {e.status_code}")
        print(f"  Message: {e.message}")
        print(f"  Reason: {e.reason}")
        
        # The error attribute contains structured error details
        if hasattr(e, 'error') and e.error:
            print(f"  Error Code: {e.error.code if hasattr(e.error, 'code') else 'N/A'}")
            print(f"  Error Message: {e.error.message if hasattr(e.error, 'message') else 'N/A'}")
        
        raise


# ==============================================================================
# EXAMPLE 2: Handling Specific Status Codes (403, 404, 429)
# ==============================================================================

def get_secret_with_specific_error_handling(
    client: SecretClient, 
    secret_name: str,
    max_retries: int = 3
):
    """
    Demonstrates handling specific HTTP status codes:
    - 403 Forbidden: Access denied (insufficient RBAC permissions)
    - 404 Not Found: Secret doesn't exist
    - 429 Too Many Requests: Rate limiting/throttling
    """
    retry_count = 0
    base_delay = 1  # seconds
    
    while retry_count <= max_retries:
        try:
            secret = client.get_secret(secret_name)
            print(f"Successfully retrieved secret: {secret.name}")
            return secret.value
        
        except HttpResponseError as e:
            
            # 403 Forbidden - Access Denied
            if e.status_code == 403:
                print(f"ERROR: Access denied to secret '{secret_name}'")
                print("Your application identity does not have the required permissions.")
                print("Required RBAC role: 'Key Vault Secrets User' or 'Key Vault Secrets Officer'")
                print(f"Error details: {e.message}")
                # Don't retry - this requires permission changes
                raise
            
            # 404 Not Found - Secret doesn't exist
            elif e.status_code == 404:
                print(f"ERROR: Secret '{secret_name}' not found in the Key Vault")
                print(f"Error details: {e.message}")
                # Don't retry - secret doesn't exist
                raise
            
            # 429 Too Many Requests - Throttling
            elif e.status_code == 429:
                retry_count += 1
                
                if retry_count > max_retries:
                    print(f"ERROR: Max retries ({max_retries}) exceeded due to throttling")
                    raise
                
                # Check for Retry-After header
                retry_after = None
                if hasattr(e, 'response') and e.response:
                    retry_after = e.response.headers.get('Retry-After')
                
                if retry_after:
                    # Retry-After can be in seconds or a date
                    try:
                        delay = int(retry_after)
                    except ValueError:
                        # If it's a date, use exponential backoff
                        delay = base_delay * (2 ** (retry_count - 1))
                else:
                    # Exponential backoff: 1s, 2s, 4s, 8s, etc.
                    delay = base_delay * (2 ** (retry_count - 1))
                
                print(f"Throttled (429). Retry {retry_count}/{max_retries} after {delay}s...")
                time.sleep(delay)
                continue
            
            # Other HTTP errors
            else:
                print(f"ERROR: Unexpected HTTP error (status {e.status_code})")
                print(f"Message: {e.message}")
                raise
        
        except ServiceRequestError as e:
            # Network errors, DNS failures, connection timeouts
            print(f"ERROR: Service request failed - {str(e)}")
            print("This typically indicates network connectivity issues")
            raise
        
        except Exception as e:
            # Catch any other unexpected errors
            print(f"ERROR: Unexpected error occurred - {type(e).__name__}: {str(e)}")
            raise


# ==============================================================================
# EXAMPLE 3: Handling Soft-Deleted Secrets
# ==============================================================================

def get_secret_handling_soft_delete(client: SecretClient, secret_name: str):
    """
    Demonstrates what happens when trying to get a soft-deleted secret.
    
    IMPORTANT: When you try to get a soft-deleted secret using get_secret(),
    you will receive a 404 (Not Found) error - the same as if the secret
    never existed. You CANNOT retrieve the value of a soft-deleted secret.
    
    To work with soft-deleted secrets, you must use:
    - get_deleted_secret() - to see deleted secret metadata
    - recover_deleted_secret() - to restore it
    - purge_deleted_secret() - to permanently delete it
    """
    try:
        # This will fail with 404 if the secret is soft-deleted
        secret = client.get_secret(secret_name)
        print(f"Secret found: {secret.name}")
        return secret.value
    
    except HttpResponseError as e:
        if e.status_code == 404:
            print(f"Secret '{secret_name}' not found.")
            print("Checking if it's soft-deleted...")
            
            try:
                # Try to get the deleted secret metadata
                deleted_secret = client.get_deleted_secret(secret_name)
                
                print(f"\nFound soft-deleted secret:")
                print(f"  Name: {deleted_secret.name}")
                print(f"  Deleted On: {deleted_secret.deleted_on}")
                print(f"  Scheduled Purge Date: {deleted_secret.scheduled_purge_date}")
                print(f"  Recovery ID: {deleted_secret.recovery_id}")
                print("\nThe secret is soft-deleted. You can:")
                print("  1. Recover it using recover_deleted_secret()")
                print("  2. Wait for automatic purge")
                print("  3. Manually purge it using purge_deleted_secret() (if you have permissions)")
                
                return None
            
            except HttpResponseError as deleted_check_error:
                if deleted_check_error.status_code == 404:
                    print("Secret does not exist (not found in active or deleted state)")
                elif deleted_check_error.status_code == 403:
                    print("Cannot check deleted secrets - insufficient permissions")
                    print("Required: 'Key Vault Secrets Officer' role or 'List' + 'Get' on deleted secrets")
                else:
                    print(f"Error checking deleted secrets: {deleted_check_error.message}")
                
                raise deleted_check_error
        else:
            raise


# ==============================================================================
# EXAMPLE 4: Complete Error Handling with Recovery
# ==============================================================================

def get_secret_with_recovery(client: SecretClient, secret_name: str):
    """
    Complete example that handles soft-deleted secrets by recovering them.
    """
    try:
        secret = client.get_secret(secret_name)
        return secret.value
    
    except HttpResponseError as e:
        if e.status_code == 404:
            # Check if soft-deleted and attempt recovery
            try:
                deleted_secret = client.get_deleted_secret(secret_name)
                print(f"Secret '{secret_name}' is soft-deleted. Attempting recovery...")
                
                # Recover the secret
                recover_operation = client.begin_recover_deleted_secret(secret_name)
                recovered_secret = recover_operation.result()
                
                print(f"Secret '{secret_name}' recovered successfully!")
                
                # Now retrieve it
                secret = client.get_secret(secret_name)
                return secret.value
            
            except HttpResponseError as recovery_error:
                if recovery_error.status_code == 403:
                    print("Cannot recover secret - insufficient permissions")
                    print("Required: 'Key Vault Secrets Officer' role or 'Recover' permission")
                else:
                    print(f"Recovery failed: {recovery_error.message}")
                raise
        
        elif e.status_code == 403:
            print(f"Access denied. Check RBAC role assignments for your identity.")
            print("Required role: 'Key Vault Secrets User' (for get) or 'Key Vault Secrets Officer' (for all operations)")
            raise
        
        elif e.status_code == 429:
            print("Rate limited. Implement retry logic with exponential backoff.")
            raise
        
        else:
            print(f"Unexpected error (status {e.status_code}): {e.message}")
            raise


# ==============================================================================
# EXAMPLE 5: Using ResourceNotFoundError (Convenience Exception)
# ==============================================================================

def get_secret_using_resource_not_found(client: SecretClient, secret_name: str):
    """
    The SDK also provides ResourceNotFoundError as a convenience exception
    that you can catch specifically for 404 errors.
    """
    try:
        secret = client.get_secret(secret_name)
        return secret.value
    
    except ResourceNotFoundError as e:
        # This is raised for 404 errors - more specific than HttpResponseError
        print(f"Secret '{secret_name}' not found")
        print(f"Status code: {e.status_code}")  # Will be 404
        print(f"Message: {e.message}")
        raise
    
    except HttpResponseError as e:
        # Handle other HTTP errors (403, 429, etc.)
        if e.status_code == 403:
            print("Access denied - check RBAC permissions")
        elif e.status_code == 429:
            print("Throttled - implement retry logic")
        else:
            print(f"HTTP error {e.status_code}: {e.message}")
        raise


# ==============================================================================
# USAGE EXAMPLES
# ==============================================================================

if __name__ == "__main__":
    # Replace with your Key Vault URL
    vault_url = "https://your-keyvault-name.vault.azure.net/"
    
    try:
        client = create_secret_client(vault_url)
        
        # Example 1: Basic error handling
        print("=" * 70)
        print("Example 1: Basic Error Handling")
        print("=" * 70)
        get_secret_with_basic_error_handling(client, "my-secret")
        
        # Example 2: Specific status code handling
        print("\n" + "=" * 70)
        print("Example 2: Handling Specific Status Codes")
        print("=" * 70)
        get_secret_with_specific_error_handling(client, "my-secret", max_retries=3)
        
        # Example 3: Soft-deleted secret handling
        print("\n" + "=" * 70)
        print("Example 3: Handling Soft-Deleted Secrets")
        print("=" * 70)
        get_secret_handling_soft_delete(client, "deleted-secret")
        
        # Example 4: With automatic recovery
        print("\n" + "=" * 70)
        print("Example 4: Error Handling with Recovery")
        print("=" * 70)
        get_secret_with_recovery(client, "my-secret")
        
        # Example 5: Using ResourceNotFoundError
        print("\n" + "=" * 70)
        print("Example 5: Using ResourceNotFoundError")
        print("=" * 70)
        get_secret_using_resource_not_found(client, "my-secret")
    
    except Exception as e:
        print(f"\nFatal error: {type(e).__name__}: {str(e)}")
