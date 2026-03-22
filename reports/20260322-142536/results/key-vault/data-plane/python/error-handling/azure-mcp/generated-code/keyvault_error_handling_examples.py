"""
Azure Key Vault Secrets - Error Handling Patterns
Demonstrates proper exception handling for common scenarios
"""

from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential
from azure.core.exceptions import HttpResponseError, ResourceNotFoundError
import time


# Initialize the client
credential = DefaultAzureCredential()
vault_url = "https://your-vault-name.vault.azure.net/"
client = SecretClient(vault_url=vault_url, credential=credential)


# Example 1: Handling 403 Access Denied (Missing RBAC permissions)
def handle_access_denied():
    """
    403 Forbidden - Occurs when the identity lacks RBAC role
    Required roles: Key Vault Secrets User (read) or Key Vault Secrets Officer (write)
    """
    try:
        secret = client.get_secret("my-secret")
        print(f"Secret value: {secret.value}")
    
    except HttpResponseError as e:
        if e.status_code == 403:
            print("Access Denied!")
            print(f"Status Code: {e.status_code}")
            print(f"Error Code: {e.error.code if e.error else 'N/A'}")
            print(f"Error Message: {e.message}")
            print("\nTroubleshooting:")
            print("- Verify your identity has 'Key Vault Secrets User' role")
            print("- Check RBAC assignments in Azure Portal")
            print("- Ensure you're using RBAC, not access policies")
        else:
            raise


# Example 2: Handling 404 Secret Not Found
def handle_secret_not_found():
    """
    404 Not Found - Secret doesn't exist or has been deleted
    """
    try:
        secret = client.get_secret("non-existent-secret")
        print(f"Secret value: {secret.value}")
    
    except ResourceNotFoundError as e:
        # ResourceNotFoundError is a subclass of HttpResponseError for 404s
        print("Secret Not Found!")
        print(f"Status Code: {e.status_code}")
        print(f"Error Message: {e.message}")
        print("\nPossible reasons:")
        print("- Secret name is incorrect")
        print("- Secret has been deleted")
        print("- Secret is in soft-deleted state")
    
    except HttpResponseError as e:
        if e.status_code == 404:
            # Alternative way to handle 404
            print(f"Secret not found: {e.message}")
        else:
            raise


# Example 3: Handling 429 Throttling (Rate Limit)
def handle_throttling_with_retry():
    """
    429 Too Many Requests - Azure throttles requests to protect the service
    Default limits: ~2000 requests per 10 seconds per vault
    """
    max_retries = 3
    retry_count = 0
    
    while retry_count < max_retries:
        try:
            secret = client.get_secret("my-secret")
            print(f"Secret retrieved: {secret.name}")
            return secret
        
        except HttpResponseError as e:
            if e.status_code == 429:
                retry_count += 1
                
                # Extract retry-after header if available
                retry_after = e.response.headers.get('Retry-After', 5)
                retry_after = int(retry_after) if isinstance(retry_after, str) else retry_after
                
                print(f"Throttled! Status Code: {e.status_code}")
                print(f"Error Message: {e.message}")
                print(f"Retry attempt {retry_count}/{max_retries}")
                print(f"Waiting {retry_after} seconds...")
                
                if retry_count < max_retries:
                    time.sleep(retry_after)
                else:
                    print("Max retries exceeded")
                    raise
            else:
                raise


# Example 4: Handling Soft-Deleted Secrets
def handle_soft_deleted_secret():
    """
    When you try to get_secret() on a soft-deleted secret:
    - Returns 404 ResourceNotFoundError (secret is not "active")
    - The secret exists in deleted state but is not accessible via get_secret()
    
    To access soft-deleted secrets, use get_deleted_secret() instead
    """
    secret_name = "deleted-secret"
    
    # This will fail with 404 if the secret is soft-deleted
    try:
        secret = client.get_secret(secret_name)
        print(f"Active secret found: {secret.value}")
    
    except ResourceNotFoundError as e:
        print(f"Secret not found in active state: {e.message}")
        print("\nChecking if it's soft-deleted...")
        
        # Try to retrieve the deleted secret
        try:
            deleted_secret = client.get_deleted_secret(secret_name)
            print(f"Found soft-deleted secret!")
            print(f"Name: {deleted_secret.name}")
            print(f"Deleted on: {deleted_secret.deleted_date}")
            print(f"Scheduled purge: {deleted_secret.scheduled_purge_date}")
            print(f"Recovery ID: {deleted_secret.recovery_id}")
            print("\nTo recover: client.begin_recover_deleted_secret(name)")
            
        except ResourceNotFoundError:
            print("Secret does not exist (not active or deleted)")


# Example 5: Comprehensive Error Handling Pattern
def comprehensive_error_handling(secret_name):
    """
    Complete pattern covering all common scenarios
    """
    try:
        secret = client.get_secret(secret_name)
        return secret.value
    
    except ResourceNotFoundError as e:
        # 404 - Secret doesn't exist
        print(f"❌ Secret '{secret_name}' not found")
        print(f"   Status: {e.status_code}")
        print(f"   Message: {e.message}")
        return None
    
    except HttpResponseError as e:
        # Inspect the error details
        status_code = e.status_code
        error_code = e.error.code if e.error else "Unknown"
        error_message = e.message
        
        if status_code == 403:
            print(f"❌ Access Denied (403)")
            print(f"   Error Code: {error_code}")
            print(f"   Message: {error_message}")
            print(f"   Fix: Grant 'Key Vault Secrets User' RBAC role")
        
        elif status_code == 429:
            print(f"❌ Rate Limit Exceeded (429)")
            print(f"   Message: {error_message}")
            retry_after = e.response.headers.get('Retry-After', 'unknown')
            print(f"   Retry after: {retry_after} seconds")
        
        elif status_code == 401:
            print(f"❌ Unauthorized (401)")
            print(f"   Message: {error_message}")
            print(f"   Fix: Check authentication credentials")
        
        else:
            print(f"❌ HTTP Error {status_code}")
            print(f"   Error Code: {error_code}")
            print(f"   Message: {error_message}")
        
        # Re-raise if you want the caller to handle it
        raise
    
    except Exception as e:
        # Catch-all for other exceptions (network errors, etc.)
        print(f"❌ Unexpected error: {type(e).__name__}")
        print(f"   Details: {str(e)}")
        raise


# Example 6: Inspecting HttpResponseError Details
def inspect_error_details():
    """
    Shows all available properties on HttpResponseError
    """
    try:
        secret = client.get_secret("problematic-secret")
    
    except HttpResponseError as e:
        print("=== HttpResponseError Details ===")
        print(f"status_code: {e.status_code}")
        print(f"reason: {e.reason}")
        print(f"message: {e.message}")
        
        # Error object (may be None)
        if e.error:
            print(f"error.code: {e.error.code}")
            print(f"error.message: {e.error.message}")
        
        # Response object
        if e.response:
            print(f"response.status_code: {e.response.status_code}")
            print(f"response.headers: {dict(e.response.headers)}")
        
        # Additional context
        print(f"model: {e.model}")
        print(f"exc_type: {type(e).__name__}")


# Example 7: Setting Secrets with Error Handling
def set_secret_with_error_handling(secret_name, secret_value):
    """
    Error handling when creating/updating secrets
    """
    try:
        secret = client.set_secret(secret_name, secret_value)
        print(f"✓ Secret '{secret_name}' saved successfully")
        return secret
    
    except HttpResponseError as e:
        if e.status_code == 403:
            print(f"❌ Cannot set secret - missing permissions")
            print(f"   Required role: 'Key Vault Secrets Officer'")
        
        elif e.status_code == 409:
            # Conflict - secret might be in deleted state
            print(f"❌ Conflict (409) - Secret may be soft-deleted")
            print(f"   Purge or recover the deleted secret first")
        
        else:
            print(f"❌ Failed to set secret: {e.message}")
        
        raise


if __name__ == "__main__":
    print("Azure Key Vault Error Handling Examples")
    print("=" * 50)
    
    # Run examples (will fail without proper credentials/vault)
    # Uncomment to test:
    
    # comprehensive_error_handling("my-secret")
    # handle_throttling_with_retry()
    # handle_soft_deleted_secret()
