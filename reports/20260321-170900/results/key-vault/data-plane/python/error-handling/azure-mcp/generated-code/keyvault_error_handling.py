"""
Azure Key Vault Error Handling Examples
Demonstrates proper exception handling with azure-keyvault-secrets SDK
"""

from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential
from azure.core.exceptions import (
    HttpResponseError,
    ResourceNotFoundError,
    ServiceRequestError
)
import time


def get_secret_with_basic_error_handling(client: SecretClient, secret_name: str):
    """
    Basic error handling pattern - catches specific status codes
    """
    try:
        secret = client.get_secret(secret_name)
        print(f"✓ Retrieved secret: {secret.name}")
        return secret.value
    
    except HttpResponseError as e:
        # Access the status code directly from the error
        status_code = e.status_code
        error_message = e.message
        
        if status_code == 403:
            # Access Denied - missing RBAC role (Key Vault Secrets User or similar)
            print(f"❌ Access Denied (403): {error_message}")
            print("   Your identity lacks the required RBAC role.")
            print("   Needed: 'Key Vault Secrets User' or 'Key Vault Secrets Officer'")
        
        elif status_code == 404:
            # Secret Not Found
            print(f"❌ Secret Not Found (404): {secret_name}")
            print(f"   Error: {error_message}")
        
        elif status_code == 429:
            # Throttling - too many requests
            print(f"❌ Throttled (429): {error_message}")
            print("   Rate limit exceeded. Implement exponential backoff.")
        
        else:
            # Other HTTP errors
            print(f"❌ HTTP Error {status_code}: {error_message}")
        
        # Re-raise if you want calling code to handle it
        raise
    
    except ServiceRequestError as e:
        # Network connectivity issues
        print(f"❌ Network Error: {e}")
        raise


def get_secret_with_retry_logic(client: SecretClient, secret_name: str, max_retries: int = 3):
    """
    Advanced pattern with exponential backoff for throttling (429)
    """
    for attempt in range(max_retries):
        try:
            secret = client.get_secret(secret_name)
            return secret.value
        
        except HttpResponseError as e:
            if e.status_code == 429:
                # Throttled - implement exponential backoff
                if attempt < max_retries - 1:
                    wait_time = 2 ** attempt  # 1s, 2s, 4s, etc.
                    
                    # Check if Retry-After header is present
                    retry_after = e.response.headers.get('Retry-After')
                    if retry_after:
                        wait_time = int(retry_after)
                    
                    print(f"⚠️  Throttled. Retrying in {wait_time}s (attempt {attempt + 1}/{max_retries})")
                    time.sleep(wait_time)
                    continue
                else:
                    print(f"❌ Max retries exceeded for throttling")
                    raise
            
            elif e.status_code == 404:
                # Don't retry on 404 - secret doesn't exist
                print(f"❌ Secret '{secret_name}' not found")
                raise
            
            elif e.status_code == 403:
                # Don't retry on 403 - permission issue won't resolve
                print(f"❌ Access denied. Check RBAC roles.")
                raise
            
            else:
                # Other errors - re-raise immediately
                raise


def handle_soft_deleted_secret(client: SecretClient, secret_name: str):
    """
    Demonstrates what happens when accessing a soft-deleted secret
    
    When you try to get a soft-deleted secret:
    - You get a 404 (ResourceNotFoundError) - the secret is not "active"
    - To retrieve it, you must use get_deleted_secret() instead
    - Or recover it with begin_recover_deleted_secret()
    """
    try:
        secret = client.get_secret(secret_name)
        print(f"✓ Secret is active: {secret.name}")
        return secret.value
    
    except HttpResponseError as e:
        if e.status_code == 404:
            print(f"⚠️  Secret not found in active secrets (404)")
            print(f"   Checking if '{secret_name}' is soft-deleted...")
            
            try:
                # Try to get the deleted secret
                deleted_secret = client.get_deleted_secret(secret_name)
                print(f"✓ Found soft-deleted secret: {deleted_secret.name}")
                print(f"   Deleted on: {deleted_secret.deleted_date}")
                print(f"   Scheduled purge: {deleted_secret.scheduled_purge_date}")
                print(f"   To use it, recover with: client.begin_recover_deleted_secret('{secret_name}')")
                return None
            
            except HttpResponseError as deleted_err:
                if deleted_err.status_code == 404:
                    print(f"❌ Secret '{secret_name}' does not exist (not active, not deleted)")
                else:
                    print(f"❌ Error checking deleted secrets: {deleted_err.message}")
                return None
        else:
            raise


def comprehensive_error_handling(client: SecretClient, secret_name: str):
    """
    Production-ready error handling with detailed inspection
    """
    try:
        secret = client.get_secret(secret_name)
        return secret.value
    
    except HttpResponseError as e:
        # Detailed error inspection
        print(f"\n{'='*60}")
        print(f"HttpResponseError Details:")
        print(f"{'='*60}")
        print(f"Status Code: {e.status_code}")
        print(f"Reason: {e.reason}")
        print(f"Message: {e.message}")
        print(f"Error Code: {e.error.code if e.error else 'N/A'}")
        
        # Access response headers if needed
        if hasattr(e, 'response') and e.response:
            print(f"Request ID: {e.response.headers.get('x-ms-request-id', 'N/A')}")
            print(f"Retry-After: {e.response.headers.get('Retry-After', 'N/A')}")
        
        # Handle specific cases
        if e.status_code == 403:
            print(f"\n🔒 RBAC Permission Required:")
            print(f"   Grant your identity one of these roles on the Key Vault:")
            print(f"   - Key Vault Secrets User (read-only)")
            print(f"   - Key Vault Secrets Officer (read/write)")
            print(f"   - Key Vault Administrator (full access)")
        
        elif e.status_code == 404:
            print(f"\n🔍 Secret Not Found:")
            print(f"   - Verify the secret name: '{secret_name}'")
            print(f"   - Check if it's been deleted (soft-delete may be enabled)")
            print(f"   - Verify you're using the correct Key Vault URL")
        
        elif e.status_code == 429:
            print(f"\n⏱️  Rate Limit Exceeded:")
            print(f"   - Implement exponential backoff")
            print(f"   - Check 'Retry-After' header for wait time")
            print(f"   - Consider caching secrets locally")
        
        print(f"{'='*60}\n")
        raise


# Usage Examples
if __name__ == "__main__":
    # Initialize client
    vault_url = "https://your-keyvault-name.vault.azure.net/"
    
    try:
        credential = DefaultAzureCredential()
        client = SecretClient(vault_url=vault_url, credential=credential)
        
        # Example 1: Basic error handling
        print("Example 1: Basic Error Handling")
        print("-" * 40)
        try:
            value = get_secret_with_basic_error_handling(client, "my-secret")
            print(f"Secret value: {value}")
        except HttpResponseError as e:
            print(f"Failed to retrieve secret: {e.status_code}")
        
        print("\n")
        
        # Example 2: Retry logic for throttling
        print("Example 2: Retry Logic with Backoff")
        print("-" * 40)
        try:
            value = get_secret_with_retry_logic(client, "my-secret", max_retries=3)
            print(f"Secret value: {value}")
        except HttpResponseError as e:
            print(f"Failed after retries: {e.status_code}")
        
        print("\n")
        
        # Example 3: Soft-deleted secret handling
        print("Example 3: Soft-Deleted Secret Handling")
        print("-" * 40)
        value = handle_soft_deleted_secret(client, "deleted-secret")
        
        print("\n")
        
        # Example 4: Comprehensive error inspection
        print("Example 4: Comprehensive Error Details")
        print("-" * 40)
        try:
            value = comprehensive_error_handling(client, "test-secret")
        except HttpResponseError:
            pass
    
    except Exception as e:
        print(f"Failed to initialize client: {e}")


# Quick Reference: Common Status Codes
"""
Status Code | Meaning | Common Cause | Retry?
------------|---------|--------------|--------
403         | Forbidden | Missing RBAC role | No - Fix permissions
404         | Not Found | Secret doesn't exist or is soft-deleted | No - Check name
429         | Too Many Requests | Rate limit exceeded | Yes - With backoff
401         | Unauthorized | Authentication failed | No - Fix credentials
500         | Server Error | Azure service issue | Yes - Transient
503         | Service Unavailable | Service temporarily down | Yes - Transient
"""
