"""
Azure Key Vault Secrets Error Handling Examples
Demonstrates proper exception handling for common scenarios:
- 403 Forbidden (Access Denied / RBAC)
- 404 Not Found (Secret doesn't exist)
- 429 Too Many Requests (Throttling)
- Soft-deleted secrets
"""

from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient
from azure.core.exceptions import (
    ResourceNotFoundError,
    HttpResponseError,
    ServiceRequestError
)
import time


def create_secret_client(vault_url: str) -> SecretClient:
    """Initialize the Key Vault secret client."""
    credential = DefaultAzureCredential()
    return SecretClient(vault_url=vault_url, credential=credential)


def handle_get_secret_with_specific_errors(client: SecretClient, secret_name: str):
    """
    Example 1: Handle specific HTTP status codes when getting a secret.
    """
    try:
        secret = client.get_secret(secret_name)
        print(f"✓ Retrieved secret: {secret.name}")
        return secret.value
    
    except ResourceNotFoundError as e:
        # 404 - Secret not found (or soft-deleted and you're trying to access it normally)
        print(f"✗ Secret '{secret_name}' not found (404)")
        print(f"  Error message: {e.message}")
        print(f"  Status code: {e.status_code}")
        # Note: If the secret is soft-deleted, you need to use get_deleted_secret()
        return None
    
    except HttpResponseError as e:
        # Check the specific status code for other errors
        if e.status_code == 403:
            # Forbidden - Missing RBAC permissions
            print(f"✗ Access denied to secret '{secret_name}' (403)")
            print(f"  Error message: {e.message}")
            print(f"  Your app identity needs 'Key Vault Secrets User' role")
            print(f"  or 'Key Vault Reader' + data plane permissions")
            
        elif e.status_code == 429:
            # Too Many Requests - Rate limiting/throttling
            print(f"✗ Rate limit exceeded (429)")
            print(f"  Error message: {e.message}")
            
            # Check for Retry-After header
            retry_after = e.response.headers.get('Retry-After', 'unknown')
            print(f"  Retry after: {retry_after} seconds")
            
            # Implement exponential backoff
            wait_time = int(retry_after) if retry_after.isdigit() else 5
            print(f"  Waiting {wait_time} seconds before retry...")
            time.sleep(wait_time)
            
        else:
            # Other HTTP errors
            print(f"✗ HTTP error occurred (status: {e.status_code})")
            print(f"  Error message: {e.message}")
            print(f"  Error code: {e.error.code if e.error else 'N/A'}")
        
        return None
    
    except ServiceRequestError as e:
        # Network-level errors (DNS, connection, etc.)
        print(f"✗ Network error: {str(e)}")
        return None
    
    except Exception as e:
        # Catch-all for unexpected errors
        print(f"✗ Unexpected error: {type(e).__name__}: {str(e)}")
        return None


def handle_get_secret_with_retry(client: SecretClient, secret_name: str, max_retries: int = 3):
    """
    Example 2: Implement retry logic with exponential backoff for throttling.
    """
    retry_count = 0
    base_delay = 1  # seconds
    
    while retry_count < max_retries:
        try:
            secret = client.get_secret(secret_name)
            print(f"✓ Retrieved secret: {secret.name}")
            return secret.value
        
        except HttpResponseError as e:
            if e.status_code == 429:
                retry_count += 1
                if retry_count >= max_retries:
                    print(f"✗ Max retries ({max_retries}) reached for throttling")
                    raise
                
                # Exponential backoff
                retry_after = e.response.headers.get('Retry-After')
                if retry_after and retry_after.isdigit():
                    wait_time = int(retry_after)
                else:
                    wait_time = base_delay * (2 ** retry_count)
                
                print(f"⟳ Throttled (429), retry {retry_count}/{max_retries} in {wait_time}s")
                time.sleep(wait_time)
            else:
                # For non-throttling errors, don't retry
                raise
        
        except ResourceNotFoundError:
            # Don't retry on 404
            raise


def check_secret_exists(client: SecretClient, secret_name: str) -> bool:
    """
    Example 3: Check if a secret exists without raising exceptions.
    """
    try:
        client.get_secret(secret_name)
        return True
    except ResourceNotFoundError:
        return False
    except HttpResponseError as e:
        if e.status_code == 403:
            # We can't determine if it exists due to permissions
            print(f"⚠ Cannot check existence: Access denied (403)")
            return None
        raise


def handle_soft_deleted_secret(client: SecretClient, secret_name: str):
    """
    Example 4: Working with soft-deleted secrets.
    
    When soft-delete is enabled on a Key Vault:
    - Deleted secrets are retained for the recovery period (default 90 days)
    - get_secret() on a soft-deleted secret raises ResourceNotFoundError (404)
    - You must use get_deleted_secret() to access soft-deleted secrets
    - You can recover or purge soft-deleted secrets
    """
    print(f"\n--- Handling Soft-Deleted Secret: {secret_name} ---")
    
    # Try to get the secret normally
    try:
        secret = client.get_secret(secret_name)
        print(f"✓ Secret is active: {secret.name}")
        return secret.value
    
    except ResourceNotFoundError:
        print(f"✗ Secret not found via get_secret() (404)")
        print(f"  It might be soft-deleted. Checking deleted secrets...")
        
        # Check if it's soft-deleted
        try:
            deleted_secret = client.get_deleted_secret(secret_name)
            print(f"✓ Found soft-deleted secret: {deleted_secret.name}")
            print(f"  Deleted on: {deleted_secret.deleted_date}")
            print(f"  Scheduled purge: {deleted_secret.scheduled_purge_date}")
            print(f"  Recovery ID: {deleted_secret.recovery_id}")
            
            # You can recover it
            print(f"\n  To recover: client.begin_recover_deleted_secret('{secret_name}')")
            # recovery_poller = client.begin_recover_deleted_secret(secret_name)
            # recovered_secret = recovery_poller.result()
            
            return None
        
        except ResourceNotFoundError:
            print(f"✗ Secret doesn't exist (not active and not soft-deleted)")
            return None
        
        except HttpResponseError as e:
            if e.status_code == 403:
                print(f"✗ Access denied to deleted secrets (403)")
                print(f"  Need 'Key Vault Secrets Officer' or equivalent permissions")
            else:
                print(f"✗ Error checking deleted secrets: {e.status_code} - {e.message}")
            return None


def handle_set_secret_errors(client: SecretClient, secret_name: str, secret_value: str):
    """
    Example 5: Handle errors when setting/creating secrets.
    """
    try:
        secret = client.set_secret(secret_name, secret_value)
        print(f"✓ Secret '{secret.name}' created/updated")
        return secret
    
    except HttpResponseError as e:
        if e.status_code == 403:
            print(f"✗ Access denied to set secret (403)")
            print(f"  Need 'Key Vault Secrets Officer' role")
            print(f"  or 'Key Vault Contributor' + data plane permissions")
        
        elif e.status_code == 409:
            # Conflict - might occur with certain policies
            print(f"✗ Conflict when setting secret (409)")
            print(f"  Error: {e.message}")
        
        elif e.status_code == 429:
            print(f"✗ Throttled when setting secret (429)")
            retry_after = e.response.headers.get('Retry-After', '5')
            print(f"  Retry after: {retry_after} seconds")
        
        else:
            print(f"✗ Error setting secret: {e.status_code} - {e.message}")
        
        return None


def inspect_http_response_error(error: HttpResponseError):
    """
    Example 6: Detailed inspection of HttpResponseError.
    """
    print("\n--- HttpResponseError Details ---")
    print(f"Status code: {error.status_code}")
    print(f"Reason: {error.reason}")
    print(f"Message: {error.message}")
    
    # Error object (if available)
    if error.error:
        print(f"Error code: {error.error.code}")
        print(f"Error message: {error.error.message}")
    
    # Response headers
    if error.response:
        print(f"\nResponse headers:")
        for key, value in error.response.headers.items():
            print(f"  {key}: {value}")
    
    # Inner error details (sometimes nested)
    if hasattr(error, 'inner_exception'):
        print(f"\nInner exception: {error.inner_exception}")


def main():
    """
    Main demonstration of error handling patterns.
    """
    # Replace with your Key Vault URL
    vault_url = "https://your-keyvault-name.vault.azure.net/"
    
    print("Azure Key Vault Error Handling Examples")
    print("=" * 50)
    
    # Create client
    try:
        client = create_secret_client(vault_url)
        print(f"✓ Connected to Key Vault: {vault_url}\n")
    except Exception as e:
        print(f"✗ Failed to create client: {e}")
        return
    
    # Example 1: Basic error handling
    print("\n1. Basic Error Handling")
    print("-" * 50)
    handle_get_secret_with_specific_errors(client, "my-secret")
    
    # Example 2: Retry logic
    print("\n2. Retry Logic with Exponential Backoff")
    print("-" * 50)
    try:
        handle_get_secret_with_retry(client, "my-secret", max_retries=3)
    except Exception as e:
        print(f"Failed after retries: {e}")
    
    # Example 3: Check existence
    print("\n3. Check Secret Existence")
    print("-" * 50)
    exists = check_secret_exists(client, "my-secret")
    print(f"Secret exists: {exists}")
    
    # Example 4: Soft-deleted secrets
    print("\n4. Soft-Deleted Secrets")
    print("-" * 50)
    handle_soft_deleted_secret(client, "deleted-secret")
    
    # Example 5: Set secret errors
    print("\n5. Setting Secrets")
    print("-" * 50)
    handle_set_secret_errors(client, "new-secret", "my-value")


if __name__ == "__main__":
    main()
