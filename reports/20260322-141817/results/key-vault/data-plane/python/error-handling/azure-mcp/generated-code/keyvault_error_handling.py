"""
Azure Key Vault Secrets - Error Handling Patterns

This module demonstrates proper error handling when working with 
Azure Key Vault secrets using the azure-keyvault-secrets SDK.
"""

from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential
from azure.core.exceptions import HttpResponseError, ResourceNotFoundError
import time


def example_basic_error_handling(vault_url: str, secret_name: str):
    """
    Basic error handling pattern - catch all HttpResponseError
    """
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    try:
        secret = client.get_secret(secret_name)
        print(f"Secret value: {secret.value}")
    except HttpResponseError as e:
        # Access the status code
        print(f"HTTP Status Code: {e.status_code}")
        
        # Access the error message
        print(f"Error Message: {e.message}")
        
        # Access the full error response
        if e.error:
            print(f"Error Code: {e.error.code}")
            print(f"Error Details: {e.error.message}")


def example_specific_error_handling(vault_url: str, secret_name: str):
    """
    Handle specific error scenarios based on status codes
    """
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    try:
        secret = client.get_secret(secret_name)
        return secret.value
        
    except HttpResponseError as e:
        if e.status_code == 403:
            # Access Denied - insufficient RBAC permissions
            print("Access Denied (403): The application identity lacks proper RBAC role.")
            print(f"Required roles: 'Key Vault Secrets User' or 'Key Vault Secrets Officer'")
            print(f"Error details: {e.message}")
            # You might want to log this to monitoring system
            raise PermissionError(f"Insufficient permissions to access secret '{secret_name}'") from e
            
        elif e.status_code == 404:
            # Secret not found
            print(f"Secret Not Found (404): '{secret_name}' does not exist in the vault")
            print(f"Error details: {e.message}")
            # Could return None or default value instead of raising
            return None
            
        elif e.status_code == 429:
            # Rate limiting / throttling
            print("Rate Limited (429): Too many requests to Key Vault")
            
            # Check for Retry-After header
            retry_after = e.response.headers.get('Retry-After')
            if retry_after:
                wait_time = int(retry_after)
                print(f"Retry after {wait_time} seconds")
            else:
                wait_time = 5  # Default backoff
                
            print(f"Error details: {e.message}")
            # In production, implement exponential backoff
            raise
            
        else:
            # Other HTTP errors
            print(f"Unexpected error ({e.status_code}): {e.message}")
            raise


def example_with_retry_logic(vault_url: str, secret_name: str, max_retries: int = 3):
    """
    Handle 429 (throttling) with retry logic
    """
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    retry_count = 0
    base_delay = 1  # Start with 1 second
    
    while retry_count < max_retries:
        try:
            secret = client.get_secret(secret_name)
            return secret.value
            
        except HttpResponseError as e:
            if e.status_code == 429:
                retry_count += 1
                
                if retry_count >= max_retries:
                    print(f"Max retries ({max_retries}) reached for throttling")
                    raise
                
                # Check for Retry-After header
                retry_after = e.response.headers.get('Retry-After')
                if retry_after:
                    wait_time = int(retry_after)
                else:
                    # Exponential backoff: 1s, 2s, 4s, 8s...
                    wait_time = base_delay * (2 ** (retry_count - 1))
                
                print(f"Throttled (429). Retry {retry_count}/{max_retries} after {wait_time}s")
                time.sleep(wait_time)
                continue
                
            else:
                # Not a throttling error, re-raise
                raise


def example_soft_deleted_secret(vault_url: str, secret_name: str):
    """
    Handling soft-deleted secrets
    
    When you try to get a soft-deleted secret:
    - The secret is in a "deleted" state but still exists in the vault
    - get_secret() will raise HttpResponseError with 404 status
    - To access it, you need to either:
      1. Recover the deleted secret using recover_deleted_secret()
      2. Access it via get_deleted_secret() if you just need to read it
      3. Purge it permanently using purge_deleted_secret() (if soft-delete allows)
    """
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    try:
        # This will fail with 404 if the secret is soft-deleted
        secret = client.get_secret(secret_name)
        print(f"Secret value: {secret.value}")
        
    except HttpResponseError as e:
        if e.status_code == 404:
            print(f"Secret '{secret_name}' not found. Checking if it's soft-deleted...")
            
            try:
                # Try to get the deleted secret
                deleted_secret = client.get_deleted_secret(secret_name)
                print(f"Secret is soft-deleted!")
                print(f"Deleted on: {deleted_secret.deleted_date}")
                print(f"Scheduled purge date: {deleted_secret.scheduled_purge_date}")
                print(f"Recovery ID: {deleted_secret.recovery_id}")
                
                # Option 1: Recover the deleted secret
                print("Recovering the secret...")
                recovered_secret = client.begin_recover_deleted_secret(secret_name).result()
                print(f"Secret recovered: {recovered_secret.name}")
                
                # Now you can get it normally
                secret = client.get_secret(secret_name)
                return secret.value
                
            except HttpResponseError as e2:
                if e2.status_code == 404:
                    print(f"Secret '{secret_name}' does not exist (not even in deleted state)")
                else:
                    print(f"Error checking deleted secret: {e2.message}")
                raise


def example_comprehensive_error_handling(vault_url: str, secret_name: str):
    """
    Production-ready comprehensive error handling
    """
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    try:
        secret = client.get_secret(secret_name)
        return secret.value
        
    except ResourceNotFoundError as e:
        # ResourceNotFoundError is a specific subclass of HttpResponseError for 404s
        print(f"Secret '{secret_name}' not found")
        
        # Check if it might be soft-deleted
        try:
            deleted_secret = client.get_deleted_secret(secret_name)
            print(f"Secret exists in soft-deleted state. Deleted on: {deleted_secret.deleted_date}")
            print("Call recover_deleted_secret() if you want to restore it")
        except HttpResponseError:
            print("Secret does not exist in vault (not even soft-deleted)")
        
        return None
        
    except HttpResponseError as e:
        status = e.status_code
        message = e.message
        
        # Build detailed error context
        error_context = {
            "status_code": status,
            "message": message,
            "secret_name": secret_name,
            "vault_url": vault_url
        }
        
        if e.error:
            error_context["error_code"] = e.error.code
            error_context["error_message"] = e.error.message
        
        if status == 403:
            error_context["issue"] = "Insufficient RBAC permissions"
            error_context["resolution"] = (
                "Grant 'Key Vault Secrets User' role to the application's managed identity"
            )
            
        elif status == 429:
            retry_after = e.response.headers.get('Retry-After', 'unknown')
            error_context["issue"] = "Rate limit exceeded"
            error_context["retry_after"] = retry_after
            error_context["resolution"] = "Implement exponential backoff and retry logic"
            
        elif status == 401:
            error_context["issue"] = "Authentication failed"
            error_context["resolution"] = "Check that DefaultAzureCredential can authenticate"
        
        # Log the error context (in production, use proper logging)
        print(f"Key Vault Error: {error_context}")
        raise


def example_batch_operations_with_error_handling(vault_url: str, secret_names: list):
    """
    Handle errors when retrieving multiple secrets
    """
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    results = {}
    errors = {}
    
    for secret_name in secret_names:
        try:
            secret = client.get_secret(secret_name)
            results[secret_name] = secret.value
            
        except HttpResponseError as e:
            # Track which secrets failed and why
            errors[secret_name] = {
                "status_code": e.status_code,
                "message": e.message
            }
            
            # Continue processing other secrets even if one fails
            if e.status_code == 403:
                print(f"⚠️  Access denied for '{secret_name}'")
            elif e.status_code == 404:
                print(f"⚠️  Secret '{secret_name}' not found")
            elif e.status_code == 429:
                print(f"⚠️  Rate limited - consider adding delay between requests")
                # In production, you might want to stop and retry the entire batch
                break
    
    return results, errors


# Example usage patterns
if __name__ == "__main__":
    vault_url = "https://your-keyvault-name.vault.azure.net/"
    
    # Example 1: Basic error inspection
    try:
        example_basic_error_handling(vault_url, "my-secret")
    except Exception as e:
        print(f"Caught: {e}")
    
    # Example 2: Specific error handling
    value = example_specific_error_handling(vault_url, "my-secret")
    
    # Example 3: With retry logic for throttling
    value = example_with_retry_logic(vault_url, "my-secret", max_retries=3)
    
    # Example 4: Handling soft-deleted secrets
    value = example_soft_deleted_secret(vault_url, "deleted-secret")
    
    # Example 5: Comprehensive production pattern
    value = example_comprehensive_error_handling(vault_url, "my-secret")
    
    # Example 6: Batch operations
    secret_names = ["secret1", "secret2", "secret3"]
    results, errors = example_batch_operations_with_error_handling(vault_url, secret_names)
    print(f"Retrieved {len(results)} secrets successfully")
    print(f"Failed to retrieve {len(errors)} secrets")
