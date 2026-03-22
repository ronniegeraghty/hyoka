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


# Example 1: Handle specific HTTP status codes
def get_secret_with_status_handling(secret_name: str):
    """
    Demonstrates how to handle different HTTP status codes
    """
    try:
        secret = client.get_secret(secret_name)
        print(f"Secret retrieved: {secret.name}")
        return secret.value
        
    except HttpResponseError as e:
        # Access the status code
        status_code = e.status_code
        
        # Access the error message
        error_message = e.message
        
        # Access additional error details
        error_code = e.error.code if e.error else None
        
        print(f"Status Code: {status_code}")
        print(f"Error Message: {error_message}")
        print(f"Error Code: {error_code}")
        
        if status_code == 403:
            # Access Denied - RBAC permissions issue
            print("ERROR: Access denied. Your identity does not have permission.")
            print("Required RBAC role: 'Key Vault Secrets User' or 'Key Vault Secrets Officer'")
            print(f"Check permissions for identity accessing vault: {vault_url}")
            
        elif status_code == 404:
            # Secret not found
            print(f"ERROR: Secret '{secret_name}' not found in the vault.")
            print("The secret may not exist or could be soft-deleted.")
            
        elif status_code == 429:
            # Throttling - rate limit exceeded
            print("ERROR: Request throttled due to rate limiting.")
            print("Too many requests in a short time period.")
            # Implement exponential backoff
            retry_after = e.response.headers.get('Retry-After', 60)
            print(f"Retry after {retry_after} seconds")
            
        else:
            print(f"ERROR: Unexpected HTTP error: {status_code}")
        
        raise  # Re-raise if you want calling code to handle it


# Example 2: Comprehensive error handling with retry logic for throttling
def get_secret_with_retry(secret_name: str, max_retries: int = 3):
    """
    Get secret with exponential backoff for throttling
    """
    retry_count = 0
    base_delay = 1
    
    while retry_count <= max_retries:
        try:
            secret = client.get_secret(secret_name)
            return secret.value
            
        except HttpResponseError as e:
            if e.status_code == 429:
                # Handle throttling with exponential backoff
                retry_count += 1
                if retry_count > max_retries:
                    print(f"Max retries ({max_retries}) exceeded")
                    raise
                
                # Check for Retry-After header
                retry_after = e.response.headers.get('Retry-After')
                if retry_after:
                    delay = int(retry_after)
                else:
                    # Exponential backoff: 1s, 2s, 4s, 8s...
                    delay = base_delay * (2 ** (retry_count - 1))
                
                print(f"Throttled. Retry {retry_count}/{max_retries} after {delay}s")
                time.sleep(delay)
                continue
                
            elif e.status_code == 403:
                print("Access denied - check RBAC permissions")
                print(f"Identity needs 'Key Vault Secrets User' role on {vault_url}")
                raise
                
            elif e.status_code == 404:
                print(f"Secret '{secret_name}' not found")
                raise
                
            else:
                print(f"HTTP Error {e.status_code}: {e.message}")
                raise


# Example 3: Handling soft-deleted secrets
def handle_soft_deleted_secret(secret_name: str):
    """
    What happens when you try to get a soft-deleted secret
    
    When soft-delete is enabled and a secret is deleted:
    - get_secret() raises ResourceNotFoundError (404)
    - The secret exists in a deleted state but is not accessible via get_secret()
    - You must use get_deleted_secret() to retrieve its properties
    - You can recover it with begin_recover_deleted_secret()
    - Or permanently delete with begin_delete_secret() (purge)
    """
    try:
        # This will fail with 404 if the secret is soft-deleted
        secret = client.get_secret(secret_name)
        print(f"Active secret found: {secret.name}")
        return secret.value
        
    except ResourceNotFoundError as e:
        print(f"Secret '{secret_name}' not found in active secrets (404)")
        print("Checking if it's soft-deleted...")
        
        try:
            # Try to get the deleted secret
            deleted_secret = client.get_deleted_secret(secret_name)
            print(f"Secret is SOFT-DELETED:")
            print(f"  - Deleted on: {deleted_secret.deleted_date}")
            print(f"  - Scheduled purge: {deleted_secret.scheduled_purge_date}")
            print(f"  - Recovery ID: {deleted_secret.recovery_id}")
            
            # Option 1: Recover the secret
            print("\nTo recover: client.begin_recover_deleted_secret(secret_name).wait()")
            
            # Option 2: Purge permanently (if you have permission)
            print("To purge: client.begin_delete_secret(secret_name).wait()")
            
            return None
            
        except HttpResponseError as del_error:
            if del_error.status_code == 403:
                print("Access denied checking deleted secrets")
                print("Need 'Key Vault Secrets Officer' role to manage deleted secrets")
            elif del_error.status_code == 404:
                print("Secret truly does not exist (not deleted, just absent)")
            else:
                print(f"Error checking deleted secrets: {del_error.status_code}")
            raise


# Example 4: Pattern for handling all common errors in one place
def robust_secret_operation(secret_name: str):
    """
    Production-ready error handling pattern
    """
    try:
        secret = client.get_secret(secret_name)
        return secret.value
        
    except ResourceNotFoundError as e:
        # Specific exception for 404
        print(f"Secret not found: {secret_name}")
        print("Check if the secret name is correct or if it's been deleted")
        return None
        
    except HttpResponseError as e:
        # General HTTP errors
        status = e.status_code
        
        error_handlers = {
            403: lambda: print(
                "Permission denied. Required role: 'Key Vault Secrets User'\n"
                f"Grant access: az role assignment create --role 'Key Vault Secrets User' "
                f"--assignee <identity> --scope <vault-resource-id>"
            ),
            429: lambda: print(
                "Rate limit exceeded. Implement retry with exponential backoff.\n"
                f"Retry-After header: {e.response.headers.get('Retry-After', 'not specified')}"
            ),
            503: lambda: print("Service unavailable. Azure Key Vault may be experiencing issues."),
        }
        
        handler = error_handlers.get(status, lambda: print(f"HTTP {status}: {e.message}"))
        handler()
        
        # Log additional details for debugging
        if e.error:
            print(f"Error code: {e.error.code}")
            print(f"Error details: {e.error.message}")
        
        raise
        
    except Exception as e:
        # Catch any other unexpected errors
        print(f"Unexpected error: {type(e).__name__}: {str(e)}")
        raise


# Example 5: Inspecting all HttpResponseError properties
def detailed_error_inspection(secret_name: str):
    """
    Shows all properties you can access on HttpResponseError
    """
    try:
        secret = client.get_secret(secret_name)
        return secret.value
        
    except HttpResponseError as e:
        print("=== HttpResponseError Details ===")
        
        # HTTP Status Code
        print(f"status_code: {e.status_code}")
        
        # Error message
        print(f"message: {e.message}")
        
        # Reason phrase
        print(f"reason: {e.reason}")
        
        # The response object
        print(f"response: {e.response}")
        
        # Response headers
        if e.response:
            print(f"headers: {dict(e.response.headers)}")
            print(f"Retry-After: {e.response.headers.get('Retry-After', 'N/A')}")
        
        # Error details (if available)
        if e.error:
            print(f"error.code: {e.error.code}")
            print(f"error.message: {e.error.message}")
            print(f"error.innererror: {getattr(e.error, 'innererror', None)}")
        
        # Model (deserialized error body)
        print(f"model: {e.model}")
        
        raise


# Example usage patterns
if __name__ == "__main__":
    secret_name = "my-secret"
    
    print("Example 1: Basic status code handling")
    try:
        get_secret_with_status_handling(secret_name)
    except HttpResponseError:
        pass
    
    print("\n" + "="*50 + "\n")
    
    print("Example 2: With retry logic")
    try:
        value = get_secret_with_retry(secret_name)
        print(f"Secret value: {value}")
    except HttpResponseError as e:
        print(f"Failed after retries: {e.status_code}")
    
    print("\n" + "="*50 + "\n")
    
    print("Example 3: Soft-deleted secret handling")
    handle_soft_deleted_secret(secret_name)
    
    print("\n" + "="*50 + "\n")
    
    print("Example 4: Robust production pattern")
    robust_secret_operation(secret_name)
