"""
Azure Key Vault Secrets - Error Handling Examples

This module demonstrates proper error handling patterns when working with
Azure Key Vault secrets using the azure-keyvault-secrets SDK.
"""

from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential
from azure.core.exceptions import HttpResponseError, ResourceNotFoundError
import time


def handle_specific_errors(secret_client: SecretClient, secret_name: str):
    """
    Example: Handle specific HTTP status codes when getting a secret.
    
    Common scenarios:
    - 403 Forbidden: Missing RBAC permissions (Key Vault Secrets User role)
    - 404 Not Found: Secret doesn't exist
    - 429 Too Many Requests: Rate limiting/throttling
    """
    try:
        secret = secret_client.get_secret(secret_name)
        print(f"Successfully retrieved secret: {secret.name}")
        return secret.value
    
    except HttpResponseError as e:
        # Inspect the status code
        status_code = e.status_code
        
        # Access error message
        error_message = e.message
        
        # Get additional error details from the response
        error_code = e.error.code if hasattr(e, 'error') and e.error else None
        
        print(f"HTTP Error: Status {status_code}")
        print(f"Error Message: {error_message}")
        print(f"Error Code: {error_code}")
        
        if status_code == 403:
            # Access Denied - Missing RBAC permissions
            print("ACCESS DENIED: The application identity lacks the required RBAC role.")
            print("Required role: 'Key Vault Secrets User' or 'Key Vault Secrets Officer'")
            print("Assign permissions using: az role assignment create --role 'Key Vault Secrets User' ...")
            # Option: Return None, raise custom exception, or use default value
            return None
        
        elif status_code == 404:
            # Secret Not Found
            print(f"SECRET NOT FOUND: '{secret_name}' does not exist in the Key Vault.")
            print("The secret may have been deleted or never created.")
            # Option: Create the secret, return default, or raise
            return None
        
        elif status_code == 429:
            # Rate Limiting / Throttling
            print("THROTTLING: Too many requests to Key Vault.")
            
            # Check for Retry-After header
            retry_after = e.response.headers.get('Retry-After')
            if retry_after:
                wait_time = int(retry_after)
                print(f"Retry after {wait_time} seconds (from Retry-After header)")
            else:
                # Use exponential backoff if no Retry-After header
                wait_time = 5
                print(f"Using default backoff: {wait_time} seconds")
            
            time.sleep(wait_time)
            # Retry the operation
            return handle_specific_errors(secret_client, secret_name)
        
        else:
            # Other HTTP errors
            print(f"Unexpected HTTP error: {status_code}")
            raise


def get_secret_with_retry(secret_client: SecretClient, secret_name: str, max_retries: int = 3):
    """
    Example: Implement retry logic with exponential backoff for transient errors.
    """
    for attempt in range(max_retries):
        try:
            secret = secret_client.get_secret(secret_name)
            return secret.value
        
        except HttpResponseError as e:
            if e.status_code == 429:
                # Throttling - apply backoff
                retry_after = e.response.headers.get('Retry-After')
                wait_time = int(retry_after) if retry_after else (2 ** attempt)
                
                print(f"Throttled. Retry {attempt + 1}/{max_retries} after {wait_time}s")
                
                if attempt < max_retries - 1:
                    time.sleep(wait_time)
                    continue
                else:
                    print("Max retries reached for throttling.")
                    raise
            
            elif e.status_code in [500, 502, 503, 504]:
                # Server errors - retry with backoff
                wait_time = 2 ** attempt
                print(f"Server error {e.status_code}. Retry {attempt + 1}/{max_retries} after {wait_time}s")
                
                if attempt < max_retries - 1:
                    time.sleep(wait_time)
                    continue
                else:
                    print("Max retries reached for server errors.")
                    raise
            
            else:
                # Non-retryable errors (403, 404, etc.)
                raise


def handle_soft_deleted_secret(secret_client: SecretClient, secret_name: str):
    """
    Example: Handle soft-deleted secrets.
    
    When you try to get a soft-deleted secret:
    - If the secret is soft-deleted (not purged), you'll get a 404 Not Found error
    - You CANNOT retrieve a soft-deleted secret using get_secret()
    - You must either:
      1. Recover it first using recover_deleted_secret()
      2. Or access it via get_deleted_secret() to view properties only (not the value)
    
    Soft delete is enabled by default on Key Vaults with purge protection.
    Deleted secrets remain in soft-deleted state for the retention period (default 90 days).
    """
    try:
        # This will fail with 404 if the secret is soft-deleted
        secret = secret_client.get_secret(secret_name)
        print(f"Secret retrieved: {secret.name}")
        return secret.value
    
    except HttpResponseError as e:
        if e.status_code == 404:
            print(f"Secret '{secret_name}' not found (404).")
            print("Checking if it's in soft-deleted state...")
            
            try:
                # Try to get the deleted secret metadata
                deleted_secret = secret_client.get_deleted_secret(secret_name)
                
                print(f"SOFT-DELETED SECRET FOUND:")
                print(f"  Name: {deleted_secret.name}")
                print(f"  Deleted On: {deleted_secret.deleted_date}")
                print(f"  Scheduled Purge Date: {deleted_secret.scheduled_purge_date}")
                print(f"  Recovery ID: {deleted_secret.recovery_id}")
                
                # NOTE: deleted_secret.value is None - you cannot get the actual secret value
                # You must recover it first to access the value
                
                print("\nTo recover this secret, use:")
                print("  recover_operation = secret_client.begin_recover_deleted_secret(secret_name)")
                print("  recovered_secret = recover_operation.result()")
                
                return None
            
            except HttpResponseError as deleted_error:
                if deleted_error.status_code == 404:
                    print(f"Secret '{secret_name}' is permanently purged or never existed.")
                else:
                    print(f"Error checking deleted secrets: {deleted_error.message}")
                return None
        else:
            raise


def recover_deleted_secret_example(secret_client: SecretClient, secret_name: str):
    """
    Example: Recover a soft-deleted secret.
    """
    try:
        # Check if secret is soft-deleted
        deleted_secret = secret_client.get_deleted_secret(secret_name)
        print(f"Found soft-deleted secret: {deleted_secret.name}")
        
        # Begin recovery (this is a long-running operation)
        print("Starting recovery operation...")
        recover_operation = secret_client.begin_recover_deleted_secret(secret_name)
        
        # Wait for recovery to complete
        recovered_secret = recover_operation.result()
        print(f"Secret recovered: {recovered_secret.name}")
        
        # Now you can get the secret value
        secret = secret_client.get_secret(secret_name)
        return secret.value
    
    except HttpResponseError as e:
        if e.status_code == 404:
            print(f"Secret '{secret_name}' is not in deleted state or doesn't exist.")
        elif e.status_code == 403:
            print("Access denied. Need 'Key Vault Secrets Officer' role to recover secrets.")
        else:
            print(f"Error during recovery: {e.message}")
        raise


def comprehensive_error_handling(secret_client: SecretClient, secret_name: str):
    """
    Example: Comprehensive error handling covering all common scenarios.
    """
    try:
        secret = secret_client.get_secret(secret_name)
        return secret.value
    
    except HttpResponseError as e:
        # Extract error details
        status_code = e.status_code
        error_message = e.message
        
        # Log the full error details
        print(f"Error accessing secret '{secret_name}':")
        print(f"  Status Code: {status_code}")
        print(f"  Message: {error_message}")
        
        # Additional details if available
        if hasattr(e, 'error') and e.error:
            print(f"  Error Code: {e.error.code}")
            if hasattr(e.error, 'message'):
                print(f"  Detailed Message: {e.error.message}")
        
        # Handle based on status code
        error_handlers = {
            403: lambda: handle_403_forbidden(secret_name),
            404: lambda: handle_404_not_found(secret_client, secret_name),
            429: lambda: handle_429_throttling(e),
            500: lambda: handle_5xx_server_error(status_code),
            502: lambda: handle_5xx_server_error(status_code),
            503: lambda: handle_5xx_server_error(status_code),
            504: lambda: handle_5xx_server_error(status_code),
        }
        
        handler = error_handlers.get(status_code)
        if handler:
            return handler()
        else:
            print(f"Unhandled HTTP error: {status_code}")
            raise
    
    except Exception as e:
        # Catch other exceptions (network errors, authentication failures, etc.)
        print(f"Unexpected error: {type(e).__name__}: {str(e)}")
        raise


def handle_403_forbidden(secret_name: str):
    """Handle 403 Forbidden errors."""
    print("\n=== ACCESS DENIED (403) ===")
    print(f"Your application identity doesn't have permission to read '{secret_name}'.")
    print("\nRequired Azure RBAC role:")
    print("  - 'Key Vault Secrets User' (read secrets)")
    print("  - 'Key Vault Secrets Officer' (read, write, delete secrets)")
    print("\nTo grant access:")
    print("  az role assignment create \\")
    print("    --role 'Key Vault Secrets User' \\")
    print("    --assignee <app-id-or-principal-id> \\")
    print("    --scope <key-vault-resource-id>")
    return None


def handle_404_not_found(secret_client: SecretClient, secret_name: str):
    """Handle 404 Not Found errors."""
    print("\n=== SECRET NOT FOUND (404) ===")
    print(f"Secret '{secret_name}' doesn't exist in the Key Vault.")
    print("Possible reasons:")
    print("  1. Secret was never created")
    print("  2. Secret is soft-deleted (check with get_deleted_secret)")
    print("  3. Typo in secret name")
    
    # Check if soft-deleted
    try:
        deleted = secret_client.get_deleted_secret(secret_name)
        print(f"\n✓ Secret is SOFT-DELETED (deleted on {deleted.deleted_date})")
        print(f"  Purge date: {deleted.scheduled_purge_date}")
        print("  Use begin_recover_deleted_secret() to restore it.")
    except:
        print("\n✗ Secret is not in soft-deleted state.")
    
    return None


def handle_429_throttling(error: HttpResponseError):
    """Handle 429 Too Many Requests errors."""
    print("\n=== THROTTLING (429) ===")
    print("Too many requests to Key Vault. Rate limits exceeded.")
    
    retry_after = error.response.headers.get('Retry-After')
    if retry_after:
        print(f"Retry after: {retry_after} seconds")
    else:
        print("No Retry-After header. Use exponential backoff.")
    
    print("\nKey Vault rate limits:")
    print("  - GET secrets: 2000 requests per 10 seconds")
    print("  - All operations: 2000 requests per 10 seconds per vault")
    print("\nConsider:")
    print("  - Caching secret values")
    print("  - Implementing exponential backoff")
    print("  - Reducing request frequency")
    
    raise


def handle_5xx_server_error(status_code: int):
    """Handle 5xx server errors."""
    print(f"\n=== SERVER ERROR ({status_code}) ===")
    print("Azure Key Vault service is experiencing issues.")
    print("These are typically transient. Implement retry with exponential backoff.")
    raise


# Example usage
if __name__ == "__main__":
    # Initialize the client
    vault_url = "https://your-vault-name.vault.azure.net/"
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    # Example 1: Handle specific errors
    print("=== Example 1: Handle Specific Errors ===")
    handle_specific_errors(client, "my-secret")
    
    # Example 2: Retry logic
    print("\n=== Example 2: Retry with Backoff ===")
    get_secret_with_retry(client, "my-secret", max_retries=3)
    
    # Example 3: Soft-deleted secrets
    print("\n=== Example 3: Soft-Deleted Secrets ===")
    handle_soft_deleted_secret(client, "deleted-secret")
    
    # Example 4: Comprehensive error handling
    print("\n=== Example 4: Comprehensive Handling ===")
    comprehensive_error_handling(client, "my-secret")
