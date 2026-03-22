"""
Azure Key Vault Secrets - Error Handling Examples
Demonstrates proper error handling for common scenarios including
403 (Forbidden), 404 (Not Found), 429 (Throttling), and soft-deleted secrets.
"""

from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential
from azure.core.exceptions import HttpResponseError, ResourceNotFoundError
import time


def handle_basic_errors(secret_client: SecretClient, secret_name: str):
    """
    Basic error handling pattern for Key Vault operations.
    """
    try:
        secret = secret_client.get_secret(secret_name)
        print(f"Successfully retrieved secret: {secret.name}")
        return secret.value
    
    except HttpResponseError as e:
        # Inspect the status code
        status_code = e.status_code
        error_message = e.message
        
        print(f"HTTP Error: {status_code}")
        print(f"Error Message: {error_message}")
        
        # Access additional error details
        if e.error:
            print(f"Error Code: {e.error.code}")
            print(f"Error Details: {e.error.message}")
        
        raise


def handle_specific_status_codes(secret_client: SecretClient, secret_name: str):
    """
    Handle specific HTTP status codes with different retry/recovery strategies.
    """
    try:
        secret = secret_client.get_secret(secret_name)
        return secret.value
    
    except HttpResponseError as e:
        if e.status_code == 403:
            # Access Denied - RBAC permissions missing
            print(f"Access Denied (403): The identity does not have permission to access secret '{secret_name}'")
            print(f"Required role: 'Key Vault Secrets User' or 'Key Vault Secrets Officer'")
            print(f"Error details: {e.message}")
            # Don't retry - this requires permission changes
            raise PermissionError(f"Missing RBAC role for Key Vault access") from e
        
        elif e.status_code == 404:
            # Secret not found
            print(f"Secret Not Found (404): '{secret_name}' does not exist")
            print(f"Error details: {e.message}")
            # Don't retry - secret doesn't exist
            raise
        
        elif e.status_code == 429:
            # Rate limit / Throttling
            print(f"Rate Limited (429): Too many requests to Key Vault")
            print(f"Error details: {e.message}")
            
            # Check for Retry-After header
            retry_after = e.response.headers.get('Retry-After')
            if retry_after:
                wait_seconds = int(retry_after)
                print(f"Retry after {wait_seconds} seconds")
                time.sleep(wait_seconds)
            else:
                # Default exponential backoff
                print("Applying exponential backoff...")
                time.sleep(5)
            
            # Retry the operation
            return handle_specific_status_codes(secret_client, secret_name)
        
        else:
            # Other HTTP errors
            print(f"Unexpected HTTP Error ({e.status_code}): {e.message}")
            raise


def handle_soft_deleted_secret(secret_client: SecretClient, secret_name: str):
    """
    Handle soft-deleted secrets.
    
    When you try to get a soft-deleted secret:
    - You get a 404 (Not Found) error
    - The secret is in the deleted state but can be recovered
    - You need to use get_deleted_secret() to access its metadata
    - You need 'recover' permission to restore it
    """
    try:
        # This will fail with 404 if the secret is soft-deleted
        secret = secret_client.get_secret(secret_name)
        print(f"Retrieved active secret: {secret.name}")
        return secret.value
    
    except HttpResponseError as e:
        if e.status_code == 404:
            print(f"Secret '{secret_name}' not found in active state")
            
            # Check if it's soft-deleted
            try:
                deleted_secret = secret_client.get_deleted_secret(secret_name)
                print(f"Found soft-deleted secret: {deleted_secret.name}")
                print(f"Deleted on: {deleted_secret.deleted_date}")
                print(f"Scheduled purge date: {deleted_secret.scheduled_purge_date}")
                print(f"Recovery ID: {deleted_secret.recovery_id}")
                
                # To recover the secret (requires 'recover' permission):
                print(f"\nTo recover this secret, use:")
                print(f"  recovered_secret = secret_client.begin_recover_deleted_secret('{secret_name}')")
                print(f"  recovered_secret.wait()")
                
                return None
            
            except HttpResponseError as recover_error:
                if recover_error.status_code == 404:
                    print(f"Secret '{secret_name}' does not exist (not even in deleted state)")
                elif recover_error.status_code == 403:
                    print(f"No permission to view deleted secrets (requires 'list' permission on deleted secrets)")
                raise
        else:
            raise


def comprehensive_error_handling(secret_client: SecretClient, secret_name: str, max_retries: int = 3):
    """
    Production-ready error handling with retries and exponential backoff.
    """
    retry_count = 0
    base_delay = 2
    
    while retry_count < max_retries:
        try:
            secret = secret_client.get_secret(secret_name)
            print(f"Successfully retrieved secret: {secret.name}")
            return secret.value
        
        except HttpResponseError as e:
            status_code = e.status_code
            error_code = e.error.code if e.error else "Unknown"
            
            print(f"\nAttempt {retry_count + 1}/{max_retries} failed")
            print(f"Status Code: {status_code}")
            print(f"Error Code: {error_code}")
            print(f"Error Message: {e.message}")
            
            # Non-retryable errors
            if status_code in [403, 404]:
                if status_code == 403:
                    print("\n❌ Access Denied - Check RBAC role assignments")
                    print("   Required: 'Key Vault Secrets User' or higher")
                elif status_code == 404:
                    print("\n❌ Secret not found - Check if it exists or is soft-deleted")
                raise
            
            # Retryable errors (429, 5xx)
            elif status_code in [429, 500, 502, 503, 504]:
                retry_count += 1
                
                if retry_count >= max_retries:
                    print(f"\n❌ Max retries ({max_retries}) exceeded")
                    raise
                
                # Calculate backoff with jitter
                if status_code == 429:
                    retry_after = e.response.headers.get('Retry-After')
                    wait_time = int(retry_after) if retry_after else base_delay * (2 ** retry_count)
                else:
                    wait_time = base_delay * (2 ** retry_count)
                
                print(f"⏳ Retrying in {wait_time} seconds...")
                time.sleep(wait_time)
            
            else:
                # Unknown error
                print(f"\n❌ Unexpected error: {status_code}")
                raise
        
        except Exception as e:
            # Non-HTTP errors (network issues, etc.)
            print(f"\n❌ Non-HTTP error: {type(e).__name__}: {str(e)}")
            raise


def example_usage():
    """
    Example usage of error handling patterns.
    """
    # Initialize the client
    vault_url = "https://your-keyvault-name.vault.azure.net/"
    credential = DefaultAzureCredential()
    secret_client = SecretClient(vault_url=vault_url, credential=credential)
    
    print("=" * 70)
    print("Example 1: Basic Error Handling")
    print("=" * 70)
    try:
        handle_basic_errors(secret_client, "my-secret")
    except HttpResponseError as e:
        print(f"Caught error with status code: {e.status_code}\n")
    
    print("\n" + "=" * 70)
    print("Example 2: Specific Status Code Handling")
    print("=" * 70)
    try:
        handle_specific_status_codes(secret_client, "my-secret")
    except Exception as e:
        print(f"Caught: {type(e).__name__}\n")
    
    print("\n" + "=" * 70)
    print("Example 3: Soft-Deleted Secret Handling")
    print("=" * 70)
    try:
        handle_soft_deleted_secret(secret_client, "deleted-secret")
    except Exception as e:
        print(f"Caught: {type(e).__name__}\n")
    
    print("\n" + "=" * 70)
    print("Example 4: Production-Ready Error Handling")
    print("=" * 70)
    try:
        value = comprehensive_error_handling(secret_client, "my-secret", max_retries=3)
        print(f"Secret value: {value}")
    except Exception as e:
        print(f"Final error: {type(e).__name__}: {str(e)}\n")


if __name__ == "__main__":
    example_usage()
