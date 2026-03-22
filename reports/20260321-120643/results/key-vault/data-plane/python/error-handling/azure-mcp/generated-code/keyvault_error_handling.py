"""
Azure Key Vault Error Handling Examples
Demonstrates proper error handling patterns for azure-keyvault-secrets SDK
"""

from azure.keyvault.secrets import SecretClient
from azure.core.exceptions import (
    HttpResponseError,
    ResourceNotFoundError,
    ServiceRequestError
)
from azure.identity import DefaultAzureCredential
import time


def handle_specific_status_codes(secret_client: SecretClient, secret_name: str):
    """
    Example: Handling specific HTTP status codes (403, 404, 429)
    """
    try:
        secret = secret_client.get_secret(secret_name)
        print(f"Successfully retrieved secret: {secret.name}")
        return secret.value
    
    except HttpResponseError as e:
        # Inspect the status code and error details
        status_code = e.status_code
        error_code = e.error.code if e.error else "Unknown"
        error_message = e.message
        
        print(f"HTTP Status Code: {status_code}")
        print(f"Error Code: {error_code}")
        print(f"Error Message: {error_message}")
        
        # Handle specific status codes
        if status_code == 403:
            # Access Denied - RBAC permissions issue
            print("ERROR: Access Denied (403)")
            print("Your identity doesn't have the required RBAC role.")
            print("Needed: 'Key Vault Secrets User' or 'Key Vault Secrets Officer'")
            print(f"Make sure your identity has GET permission on secret: {secret_name}")
            # Could trigger an alert, log to monitoring system, etc.
            raise
        
        elif status_code == 404:
            # Secret Not Found
            print(f"ERROR: Secret '{secret_name}' not found (404)")
            print("The secret may not exist or could be soft-deleted.")
            # Could fall back to default value, create the secret, etc.
            return None
        
        elif status_code == 429:
            # Throttling - Rate limit exceeded
            print("ERROR: Throttling (429) - Rate limit exceeded")
            print("Too many requests to Key Vault. Implementing exponential backoff...")
            
            # Get retry-after header if available
            retry_after = e.response.headers.get('Retry-After', 5)
            retry_after = int(retry_after) if isinstance(retry_after, str) else retry_after
            
            print(f"Waiting {retry_after} seconds before retry...")
            time.sleep(retry_after)
            
            # Retry the request (in production, use exponential backoff)
            return handle_specific_status_codes(secret_client, secret_name)
        
        else:
            # Other HTTP errors
            print(f"ERROR: Unexpected HTTP error (Status: {status_code})")
            raise


def handle_soft_deleted_secret(secret_client: SecretClient, secret_name: str):
    """
    Example: What happens when trying to get a soft-deleted secret
    
    When soft-delete is enabled:
    - get_secret() will raise 404 (ResourceNotFoundError)
    - The secret still exists in the deleted state
    - Use get_deleted_secret() to retrieve it
    - Use recover_deleted_secret() to restore it
    """
    try:
        secret = secret_client.get_secret(secret_name)
        print(f"Secret found: {secret.name}")
        return secret.value
    
    except ResourceNotFoundError as e:
        print(f"Secret '{secret_name}' not found with get_secret()")
        print("Checking if it's soft-deleted...")
        
        try:
            # Try to get the deleted secret
            deleted_secret = secret_client.get_deleted_secret(secret_name)
            print(f"Found soft-deleted secret: {deleted_secret.name}")
            print(f"Deleted on: {deleted_secret.deleted_date}")
            print(f"Scheduled purge date: {deleted_secret.scheduled_purge_date}")
            print(f"Recovery ID: {deleted_secret.recovery_id}")
            
            # Option 1: Recover the secret
            print("Recovering the soft-deleted secret...")
            recovered_secret = secret_client.begin_recover_deleted_secret(secret_name).result()
            print(f"Secret recovered: {recovered_secret.name}")
            
            # Now we can get it normally
            secret = secret_client.get_secret(secret_name)
            return secret.value
        
        except ResourceNotFoundError:
            print(f"Secret '{secret_name}' is not soft-deleted either.")
            print("It truly doesn't exist or was purged.")
            return None
        
        except HttpResponseError as recover_error:
            if recover_error.status_code == 403:
                print("ERROR: Cannot recover - missing permissions")
                print("Needed: 'Key Vault Secrets Officer' role for recovery")
            raise


def comprehensive_error_handling(secret_client: SecretClient, secret_name: str):
    """
    Example: Comprehensive error handling with multiple exception types
    """
    max_retries = 3
    retry_count = 0
    base_delay = 2  # seconds
    
    while retry_count < max_retries:
        try:
            secret = secret_client.get_secret(secret_name)
            print(f"✓ Successfully retrieved secret: {secret.name}")
            return secret.value
        
        except ResourceNotFoundError as e:
            # 404 - Secret not found
            print(f"✗ Secret '{secret_name}' not found (404)")
            # Check if soft-deleted
            try:
                deleted = secret_client.get_deleted_secret(secret_name)
                print(f"  → Found in soft-deleted state. Deleted: {deleted.deleted_date}")
                return None  # Or trigger recovery workflow
            except:
                print(f"  → Secret doesn't exist (not even in deleted state)")
                return None
        
        except HttpResponseError as e:
            status_code = e.status_code
            error_code = e.error.code if e.error else "Unknown"
            
            print(f"✗ HTTP Error: {status_code} - {error_code}")
            print(f"  Message: {e.message}")
            
            if status_code == 403:
                # Permission denied - don't retry
                print(f"  → Access denied. Check RBAC permissions.")
                print(f"  → Required role: 'Key Vault Secrets User'")
                raise  # Don't retry permission errors
            
            elif status_code == 429:
                # Throttling - retry with backoff
                retry_after = int(e.response.headers.get('Retry-After', base_delay * (2 ** retry_count)))
                print(f"  → Throttled. Retry after {retry_after}s (attempt {retry_count + 1}/{max_retries})")
                time.sleep(retry_after)
                retry_count += 1
                continue
            
            elif status_code == 503:
                # Service unavailable - retry with backoff
                delay = base_delay * (2 ** retry_count)
                print(f"  → Service unavailable. Retry in {delay}s (attempt {retry_count + 1}/{max_retries})")
                time.sleep(delay)
                retry_count += 1
                continue
            
            else:
                # Other errors - don't retry
                print(f"  → Unexpected error. Not retrying.")
                raise
        
        except ServiceRequestError as e:
            # Network errors, DNS resolution failures, etc.
            print(f"✗ Service request error: {e}")
            print(f"  → Network or connection issue. Retry {retry_count + 1}/{max_retries}")
            delay = base_delay * (2 ** retry_count)
            time.sleep(delay)
            retry_count += 1
            continue
        
        except Exception as e:
            # Catch-all for unexpected errors
            print(f"✗ Unexpected error: {type(e).__name__}: {e}")
            raise
    
    print(f"✗ Failed after {max_retries} retries")
    return None


def batch_get_with_error_handling(secret_client: SecretClient, secret_names: list):
    """
    Example: Batch operations with per-secret error handling
    """
    results = {}
    errors = {}
    
    for secret_name in secret_names:
        try:
            secret = secret_client.get_secret(secret_name)
            results[secret_name] = secret.value
            print(f"✓ {secret_name}: retrieved")
        
        except HttpResponseError as e:
            error_info = {
                'status_code': e.status_code,
                'error_code': e.error.code if e.error else None,
                'message': e.message
            }
            errors[secret_name] = error_info
            print(f"✗ {secret_name}: {e.status_code} - {error_info['error_code']}")
        
        except Exception as e:
            errors[secret_name] = {'error': str(e)}
            print(f"✗ {secret_name}: {type(e).__name__}: {e}")
    
    return results, errors


def inspect_error_details(secret_client: SecretClient, secret_name: str):
    """
    Example: Detailed error inspection showing all available properties
    """
    try:
        secret = secret_client.get_secret(secret_name)
        return secret.value
    
    except HttpResponseError as e:
        print("=" * 60)
        print("HttpResponseError Details:")
        print("=" * 60)
        
        # Status code
        print(f"status_code: {e.status_code}")
        
        # Error object (if available)
        if e.error:
            print(f"error.code: {e.error.code}")
            print(f"error.message: {e.error.message}")
        
        # Message
        print(f"message: {e.message}")
        
        # Response object (if available)
        if e.response:
            print(f"response.status_code: {e.response.status_code}")
            print(f"response.reason: {e.response.reason}")
            print(f"response.headers: {dict(e.response.headers)}")
        
        # Additional attributes
        print(f"reason: {e.reason if hasattr(e, 'reason') else 'N/A'}")
        
        print("=" * 60)
        raise


# Example usage
if __name__ == "__main__":
    # Initialize client
    vault_url = "https://your-vault-name.vault.azure.net/"
    credential = DefaultAzureCredential()
    secret_client = SecretClient(vault_url=vault_url, credential=credential)
    
    # Example 1: Handle specific status codes
    print("\n--- Example 1: Specific Status Code Handling ---")
    try:
        value = handle_specific_status_codes(secret_client, "my-secret")
    except Exception as e:
        print(f"Failed: {e}")
    
    # Example 2: Soft-deleted secret handling
    print("\n--- Example 2: Soft-Deleted Secret Handling ---")
    try:
        value = handle_soft_deleted_secret(secret_client, "deleted-secret")
    except Exception as e:
        print(f"Failed: {e}")
    
    # Example 3: Comprehensive error handling with retries
    print("\n--- Example 3: Comprehensive Error Handling ---")
    value = comprehensive_error_handling(secret_client, "my-secret")
    
    # Example 4: Batch operations
    print("\n--- Example 4: Batch Operations ---")
    secret_names = ["secret1", "secret2", "secret3"]
    results, errors = batch_get_with_error_handling(secret_client, secret_names)
    print(f"Retrieved: {len(results)}, Failed: {len(errors)}")
    
    # Example 5: Detailed error inspection
    print("\n--- Example 5: Detailed Error Inspection ---")
    try:
        inspect_error_details(secret_client, "nonexistent-secret")
    except Exception:
        pass
