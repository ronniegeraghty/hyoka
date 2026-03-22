"""
Azure Key Vault Secrets - Error Handling Examples
Demonstrates proper exception handling for common scenarios:
- 403 Forbidden (access denied)
- 404 Not Found (secret doesn't exist)
- 429 Too Many Requests (throttling)
- Soft-deleted secrets
"""

from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential
from azure.core.exceptions import HttpResponseError, ResourceNotFoundError
import time


def basic_error_handling_example(vault_url: str, secret_name: str):
    """
    Basic pattern: Catch HttpResponseError and inspect status_code
    """
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    try:
        secret = client.get_secret(secret_name)
        print(f"Secret value: {secret.value}")
        
    except HttpResponseError as e:
        # Access the status code
        status_code = e.status_code
        
        # Access the error message
        error_message = e.message
        
        # Access additional error details from the response
        if e.error:
            error_code = e.error.code  # e.g., "Forbidden", "SecretNotFound"
            error_details = e.error.message
            
        print(f"Status Code: {status_code}")
        print(f"Error Message: {error_message}")
        
        # Handle specific status codes
        if status_code == 403:
            print("Access denied! Check RBAC permissions.")
            print("Required role: 'Key Vault Secrets User' or 'Key Vault Secrets Officer'")
            
        elif status_code == 404:
            print(f"Secret '{secret_name}' not found in vault.")
            
        elif status_code == 429:
            print("Rate limit exceeded. Implement retry with backoff.")
            
        else:
            print(f"Unexpected error: {error_message}")


def handle_specific_errors_separately(vault_url: str, secret_name: str):
    """
    Pattern: Handle specific HTTP status codes with separate except blocks
    """
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    try:
        secret = client.get_secret(secret_name)
        return secret.value
        
    except HttpResponseError as e:
        if e.status_code == 403:
            # Access denied - insufficient permissions
            print(f"❌ Access Denied (403)")
            print(f"   Error Code: {e.error.code if e.error else 'N/A'}")
            print(f"   Message: {e.message}")
            print(f"   This identity needs 'Key Vault Secrets User' role assignment")
            raise PermissionError("Insufficient permissions to access Key Vault") from e
            
        elif e.status_code == 404:
            # Secret not found
            print(f"❌ Secret Not Found (404)")
            print(f"   Secret Name: {secret_name}")
            print(f"   Vault: {vault_url}")
            # Note: Could be deleted or never existed
            raise KeyError(f"Secret '{secret_name}' does not exist") from e
            
        elif e.status_code == 429:
            # Throttling - too many requests
            print(f"⚠️  Rate Limited (429)")
            print(f"   Message: {e.message}")
            
            # Check for Retry-After header
            retry_after = e.response.headers.get('Retry-After')
            if retry_after:
                print(f"   Retry after: {retry_after} seconds")
            
            raise RuntimeError("Key Vault rate limit exceeded") from e
            
        else:
            # Other HTTP errors
            print(f"❌ Unexpected Error ({e.status_code})")
            print(f"   Message: {e.message}")
            raise


def handle_throttling_with_retry(vault_url: str, secret_name: str, max_retries: int = 3):
    """
    Pattern: Implement exponential backoff for throttling (429)
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
            if e.status_code == 429:
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
                
                print(f"Rate limited. Retry {retry_count}/{max_retries} after {delay}s")
                time.sleep(delay)
                
            else:
                # Not a throttling error, re-raise
                raise


def handle_soft_deleted_secret(vault_url: str, secret_name: str):
    """
    Pattern: Handle soft-deleted secrets
    
    When a secret is deleted (with soft-delete enabled), it enters a "deleted" state.
    Attempting to get_secret() on a soft-deleted secret returns 404 Not Found.
    
    To work with soft-deleted secrets, use get_deleted_secret() instead.
    """
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    try:
        # This will fail with 404 if the secret is soft-deleted
        secret = client.get_secret(secret_name)
        print(f"✅ Active secret found: {secret.name}")
        return secret.value
        
    except HttpResponseError as e:
        if e.status_code == 404:
            print(f"Secret not found. Checking if it's soft-deleted...")
            
            # Check if the secret is in soft-deleted state
            try:
                deleted_secret = client.get_deleted_secret(secret_name)
                print(f"⚠️  Secret '{secret_name}' is soft-deleted")
                print(f"   Deleted on: {deleted_secret.deleted_date}")
                print(f"   Scheduled purge: {deleted_secret.scheduled_purge_date}")
                print(f"   Recovery ID: {deleted_secret.recovery_id}")
                print(f"\n   To access this secret:")
                print(f"   1. Recover it: client.begin_recover_deleted_secret('{secret_name}')")
                print(f"   2. Or permanently delete: client.purge_deleted_secret('{secret_name}')")
                
                return None  # Cannot get value of soft-deleted secret
                
            except HttpResponseError as deleted_error:
                if deleted_error.status_code == 404:
                    print(f"❌ Secret '{secret_name}' does not exist (not active or deleted)")
                else:
                    print(f"Error checking deleted secrets: {deleted_error.message}")
                return None
        else:
            raise


def comprehensive_error_handling(vault_url: str, secret_name: str):
    """
    Comprehensive pattern: Handle all common scenarios
    """
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    try:
        secret = client.get_secret(secret_name)
        print(f"✅ Successfully retrieved secret: {secret.name}")
        return secret.value
        
    except HttpResponseError as e:
        # Extract error details
        status_code = e.status_code
        error_message = e.message
        error_code = e.error.code if e.error else "Unknown"
        
        print(f"\n{'='*60}")
        print(f"HTTP Response Error Details:")
        print(f"{'='*60}")
        print(f"Status Code: {status_code}")
        print(f"Error Code:  {error_code}")
        print(f"Message:     {error_message}")
        
        # Additional response details if available
        if hasattr(e, 'response') and e.response:
            print(f"Response Headers: {dict(e.response.headers)}")
        
        print(f"{'='*60}\n")
        
        # Handle specific cases
        if status_code == 403:
            print("🔒 PERMISSION DENIED")
            print("   Your identity lacks the required RBAC role.")
            print("   Required: 'Key Vault Secrets User' (read) or")
            print("            'Key Vault Secrets Officer' (read/write)")
            print(f"   Vault: {vault_url}")
            print("   Solution: Assign RBAC role to your identity (user/service principal/managed identity)")
            
        elif status_code == 404:
            print("🔍 SECRET NOT FOUND")
            print(f"   The secret '{secret_name}' doesn't exist or is soft-deleted.")
            print("   Checking soft-deleted state...")
            # Check for soft-deleted (implementation from previous example)
            
        elif status_code == 429:
            print("⏱️  RATE LIMIT EXCEEDED")
            print("   Too many requests to Key Vault.")
            retry_after = e.response.headers.get('Retry-After') if hasattr(e, 'response') else None
            if retry_after:
                print(f"   Retry after: {retry_after} seconds")
            print("   Solution: Implement exponential backoff retry logic")
            
        elif status_code == 401:
            print("🔐 AUTHENTICATION FAILED")
            print("   The credential failed to authenticate.")
            print("   Check: Azure CLI login, managed identity configuration, or service principal credentials")
            
        else:
            print(f"❌ UNEXPECTED ERROR ({status_code})")
            print(f"   {error_message}")
        
        raise
        
    except Exception as e:
        # Catch non-HTTP errors (network issues, authentication problems, etc.)
        print(f"❌ Non-HTTP Error: {type(e).__name__}")
        print(f"   Message: {str(e)}")
        raise


def check_secret_exists_pattern(vault_url: str, secret_name: str) -> bool:
    """
    Pattern: Check if a secret exists without raising exceptions
    """
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    try:
        client.get_secret(secret_name)
        return True
        
    except HttpResponseError as e:
        if e.status_code == 404:
            return False
        else:
            # Other errors (403, 429, etc.) should be raised
            raise


# Example usage (commented out to avoid execution)
if __name__ == "__main__":
    VAULT_URL = "https://your-key-vault.vault.azure.net/"
    SECRET_NAME = "my-secret"
    
    # Example 1: Basic error handling
    # basic_error_handling_example(VAULT_URL, SECRET_NAME)
    
    # Example 2: Specific error handling
    # handle_specific_errors_separately(VAULT_URL, SECRET_NAME)
    
    # Example 3: Throttling with retry
    # handle_throttling_with_retry(VAULT_URL, SECRET_NAME)
    
    # Example 4: Soft-deleted secrets
    # handle_soft_deleted_secret(VAULT_URL, SECRET_NAME)
    
    # Example 5: Comprehensive handling
    # comprehensive_error_handling(VAULT_URL, SECRET_NAME)
    
    # Example 6: Check existence
    # exists = check_secret_exists_pattern(VAULT_URL, SECRET_NAME)
    # print(f"Secret exists: {exists}")
    
    print("Examples ready. Update VAULT_URL and SECRET_NAME to test.")
