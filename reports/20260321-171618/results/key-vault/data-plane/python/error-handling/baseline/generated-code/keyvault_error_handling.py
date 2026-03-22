"""
Azure Key Vault Secrets Error Handling Examples
Demonstrates proper error handling patterns for common scenarios
"""

from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient
from azure.core.exceptions import (
    HttpResponseError,
    ResourceNotFoundError,
    ServiceRequestError,
    ClientAuthenticationError
)
import time

# Initialize the client
vault_url = "https://your-keyvault-name.vault.azure.net"
credential = DefaultAzureCredential()
client = SecretClient(vault_url=vault_url, credential=credential)


# Example 1: Handle 403 Access Denied (RBAC permissions issue)
def get_secret_with_rbac_handling(secret_name: str):
    """Handle case where identity lacks Key Vault Secrets User role"""
    try:
        secret = client.get_secret(secret_name)
        print(f"Secret value: {secret.value}")
        return secret
    except HttpResponseError as e:
        if e.status_code == 403:
            print(f"Access Denied (403): {e.message}")
            print(f"Error code: {e.error.code if e.error else 'N/A'}")
            print("Ensure your identity has 'Key Vault Secrets User' role assigned")
            # Additional details from the error
            print(f"Full error: {e.error.message if e.error else str(e)}")
        else:
            raise
    except ClientAuthenticationError as e:
        print(f"Authentication failed: {str(e)}")
        print("Check your credential configuration")


# Example 2: Handle 404 Secret Not Found
def get_secret_with_not_found_handling(secret_name: str):
    """Handle case where secret doesn't exist"""
    try:
        secret = client.get_secret(secret_name)
        return secret
    except ResourceNotFoundError as e:
        # ResourceNotFoundError is a specialized HttpResponseError for 404
        print(f"Secret '{secret_name}' not found (404)")
        print(f"Status code: {e.status_code}")
        print(f"Error message: {e.message}")
        return None
    except HttpResponseError as e:
        if e.status_code == 404:
            print(f"Secret not found: {e.message}")
            return None
        else:
            raise


# Example 3: Handle 429 Throttling with Retry Logic
def get_secret_with_throttling_handling(secret_name: str, max_retries: int = 3):
    """Handle rate limiting with exponential backoff"""
    retry_count = 0
    base_delay = 1  # seconds
    
    while retry_count < max_retries:
        try:
            secret = client.get_secret(secret_name)
            return secret
        except HttpResponseError as e:
            if e.status_code == 429:
                retry_count += 1
                # Check for Retry-After header
                retry_after = e.response.headers.get('Retry-After')
                
                if retry_after:
                    wait_time = int(retry_after)
                    print(f"Throttled (429). Retry-After header: {wait_time}s")
                else:
                    # Exponential backoff
                    wait_time = base_delay * (2 ** (retry_count - 1))
                    print(f"Throttled (429). Backing off for {wait_time}s")
                
                if retry_count < max_retries:
                    print(f"Retry {retry_count}/{max_retries}")
                    time.sleep(wait_time)
                else:
                    print("Max retries reached")
                    raise
            else:
                raise


# Example 4: Comprehensive Error Handling with Status Code Inspection
def get_secret_comprehensive(secret_name: str):
    """Complete error handling demonstrating HttpResponseError inspection"""
    try:
        secret = client.get_secret(secret_name)
        print(f"Successfully retrieved secret: {secret.name}")
        return secret
        
    except HttpResponseError as e:
        # Inspect the status code
        status_code = e.status_code
        
        # Inspect the error message
        error_message = e.message
        
        # Inspect error details if available
        error_code = e.error.code if e.error else None
        error_detail = e.error.message if e.error else None
        
        print(f"HTTP Error occurred:")
        print(f"  Status Code: {status_code}")
        print(f"  Message: {error_message}")
        print(f"  Error Code: {error_code}")
        print(f"  Error Detail: {error_detail}")
        
        # Handle specific status codes
        if status_code == 403:
            print("\n[ACTION REQUIRED] Access Denied:")
            print("  - Verify your identity has the 'Key Vault Secrets User' role")
            print("  - Check Key Vault's RBAC settings in Azure Portal")
            print("  - Ensure no network restrictions are blocking access")
            
        elif status_code == 404:
            print(f"\n[INFO] Secret '{secret_name}' does not exist")
            print("  - Check the secret name for typos")
            print("  - Verify you're connecting to the correct Key Vault")
            print("  - The secret may have been deleted")
            
        elif status_code == 429:
            print("\n[WARNING] Rate limit exceeded:")
            print("  - Too many requests in a short time period")
            print("  - Implement retry logic with exponential backoff")
            print("  - Consider caching frequently accessed secrets")
            
        elif status_code >= 500:
            print("\n[ERROR] Server error - Azure service issue")
            print("  - Retry the operation after a delay")
            print("  - Check Azure status page for service health")
            
        raise
        
    except ClientAuthenticationError as e:
        print(f"Authentication Error: {str(e)}")
        print("  - Check DefaultAzureCredential configuration")
        print("  - Verify environment variables or managed identity setup")
        raise
        
    except ServiceRequestError as e:
        print(f"Network/Request Error: {str(e)}")
        print("  - Check network connectivity")
        print("  - Verify Key Vault URL is correct")
        raise


# Example 5: Handling Soft-Deleted Secrets
def handle_soft_deleted_secret(secret_name: str):
    """
    Demonstrate what happens when trying to get a soft-deleted secret.
    
    When soft-delete is enabled and a secret is deleted:
    - get_secret() will raise ResourceNotFoundError (404)
    - The secret exists in deleted state but can't be accessed via get_secret()
    - Use get_deleted_secret() to retrieve deleted secret metadata
    - Use recover_deleted_secret() to restore it
    """
    try:
        # This will fail if secret is soft-deleted
        secret = client.get_secret(secret_name)
        print(f"Secret is active: {secret.name}")
        return secret
        
    except ResourceNotFoundError as e:
        print(f"Secret '{secret_name}' not found via get_secret()")
        print("Checking if it's in deleted state...")
        
        # Try to get the deleted secret
        try:
            deleted_secret = client.get_deleted_secret(secret_name)
            print(f"\n[INFO] Secret is SOFT-DELETED:")
            print(f"  Name: {deleted_secret.name}")
            print(f"  Deleted On: {deleted_secret.deleted_date}")
            print(f"  Scheduled Purge: {deleted_secret.scheduled_purge_date}")
            print(f"  Recovery ID: {deleted_secret.recovery_id}")
            print("\nTo restore, use: client.recover_deleted_secret(secret_name)")
            
            # Optionally recover the secret
            # recovery_poller = client.begin_recover_deleted_secret(secret_name)
            # recovered_secret = recovery_poller.result()
            # print(f"Secret recovered: {recovered_secret.name}")
            
            return None
            
        except ResourceNotFoundError:
            print(f"\n[INFO] Secret '{secret_name}' does not exist (not active or deleted)")
            return None
        except HttpResponseError as del_err:
            if del_err.status_code == 403:
                print("\n[WARNING] No permission to view deleted secrets")
                print("  - Requires 'Key Vault Secrets User' or higher permissions")
            raise


# Example 6: Multiple Operations with Error Context
def batch_get_secrets_with_error_tracking(secret_names: list[str]):
    """Get multiple secrets and track which ones failed"""
    results = {}
    errors = {}
    
    for secret_name in secret_names:
        try:
            secret = client.get_secret(secret_name)
            results[secret_name] = secret.value
            
        except HttpResponseError as e:
            error_info = {
                'status_code': e.status_code,
                'message': e.message,
                'error_code': e.error.code if e.error else None
            }
            errors[secret_name] = error_info
            
            # Log but continue processing other secrets
            print(f"Failed to get '{secret_name}': [{e.status_code}] {e.message}")
    
    # Summary
    print(f"\nSuccessfully retrieved: {len(results)}/{len(secret_names)} secrets")
    if errors:
        print("\nFailed secrets:")
        for name, error in errors.items():
            print(f"  - {name}: {error['status_code']} - {error['message']}")
    
    return results, errors


# Example 7: Using Response Headers
def get_secret_with_header_inspection(secret_name: str):
    """Demonstrate inspecting response headers from errors"""
    try:
        secret = client.get_secret(secret_name)
        return secret
    except HttpResponseError as e:
        print(f"Error Status: {e.status_code}")
        print(f"Error Message: {e.message}")
        
        # Inspect response headers if available
        if e.response:
            print("\nResponse Headers:")
            headers_of_interest = [
                'x-ms-keyvault-service-version',
                'x-ms-request-id',
                'Retry-After',
                'x-ms-keyvault-region'
            ]
            for header in headers_of_interest:
                value = e.response.headers.get(header)
                if value:
                    print(f"  {header}: {value}")
        
        raise


# Usage Examples
if __name__ == "__main__":
    # Example usage (uncomment to run)
    
    # Test with a secret that doesn't exist
    # get_secret_with_not_found_handling("non-existent-secret")
    
    # Test comprehensive error handling
    # get_secret_comprehensive("my-secret")
    
    # Test soft-deleted secret handling
    # handle_soft_deleted_secret("deleted-secret")
    
    # Batch operation with error tracking
    # secrets_to_fetch = ["secret1", "secret2", "secret3"]
    # results, errors = batch_get_secrets_with_error_tracking(secrets_to_fetch)
    
    pass
