"""
Azure Key Vault Error Handling Examples
Demonstrates proper error handling patterns for common scenarios
"""

from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential
from azure.core.exceptions import (
    HttpResponseError,
    ResourceNotFoundError,
    ServiceRequestError
)
import time


def handle_secret_operations():
    """
    Comprehensive error handling for Key Vault operations
    """
    # Initialize client
    vault_url = "https://your-keyvault.vault.azure.net/"
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    secret_name = "my-secret"
    
    try:
        # Attempt to get secret
        secret = client.get_secret(secret_name)
        print(f"Secret value: {secret.value}")
        
    except HttpResponseError as e:
        # HttpResponseError is the base exception for HTTP errors
        # Inspect the status_code to determine the specific error
        
        if e.status_code == 403:
            # Access Denied - Missing RBAC permissions
            print("❌ Access Denied (403)")
            print(f"Error code: {e.error.code if e.error else 'N/A'}")
            print(f"Message: {e.message}")
            print("\nTroubleshooting:")
            print("- Ensure your identity has 'Key Vault Secrets User' role")
            print("- Or use 'Key Vault Secrets Officer' for read/write")
            print(f"- Vault: {vault_url}")
            print(f"- Identity needs access to secret: {secret_name}")
            
        elif e.status_code == 404:
            # Secret Not Found
            print("❌ Secret Not Found (404)")
            print(f"Secret '{secret_name}' does not exist in the vault")
            print(f"Error: {e.message}")
            
            # Check if it might be soft-deleted
            try:
                deleted_secret = client.get_deleted_secret(secret_name)
                print("\n⚠️ Secret is SOFT-DELETED!")
                print(f"Deleted on: {deleted_secret.deleted_date}")
                print(f"Scheduled purge: {deleted_secret.scheduled_purge_date}")
                print(f"Recovery ID: {deleted_secret.recovery_id}")
                print("\nTo recover: client.begin_recover_deleted_secret(secret_name)")
            except HttpResponseError as del_err:
                if del_err.status_code == 404:
                    print("Secret has never existed or was purged")
                    
        elif e.status_code == 429:
            # Throttling - Rate limit exceeded
            print("❌ Rate Limit Exceeded (429)")
            print(f"Message: {e.message}")
            
            # Check for Retry-After header
            retry_after = e.response.headers.get('Retry-After')
            if retry_after:
                print(f"Retry after: {retry_after} seconds")
                # Implement exponential backoff
                time.sleep(int(retry_after))
            else:
                # Default backoff
                print("Applying exponential backoff...")
                time.sleep(5)
                
        elif e.status_code == 401:
            # Authentication failed
            print("❌ Authentication Failed (401)")
            print("Check your credentials and token")
            print(f"Error: {e.message}")
            
        else:
            # Other HTTP errors
            print(f"❌ HTTP Error {e.status_code}")
            print(f"Message: {e.message}")
            if e.error:
                print(f"Error code: {e.error.code}")
                print(f"Error details: {e.error.message}")
                
    except ServiceRequestError as e:
        # Network-level errors (DNS, connection failures, etc.)
        print("❌ Service Request Error")
        print(f"Failed to connect to Key Vault: {e}")
        print("Check network connectivity and vault URL")
        
    except Exception as e:
        # Catch-all for unexpected errors
        print(f"❌ Unexpected error: {type(e).__name__}")
        print(f"Details: {e}")


def get_secret_with_retry(client: SecretClient, secret_name: str, max_retries: int = 3):
    """
    Get secret with automatic retry logic for throttling
    """
    for attempt in range(max_retries):
        try:
            secret = client.get_secret(secret_name)
            return secret.value
            
        except HttpResponseError as e:
            if e.status_code == 429:
                # Rate limited - implement exponential backoff
                retry_after = int(e.response.headers.get('Retry-After', 2 ** attempt))
                print(f"Rate limited. Retrying in {retry_after}s... (attempt {attempt + 1}/{max_retries})")
                
                if attempt < max_retries - 1:
                    time.sleep(retry_after)
                    continue
                else:
                    print("Max retries reached")
                    raise
                    
            elif e.status_code == 403:
                # Don't retry permission errors
                print(f"Access denied. Check RBAC permissions for secret '{secret_name}'")
                raise
                
            elif e.status_code == 404:
                # Don't retry not found errors
                print(f"Secret '{secret_name}' not found")
                raise
            else:
                # Other errors
                raise
                
    return None


def inspect_error_details():
    """
    Demonstrate how to extract detailed error information
    """
    vault_url = "https://your-keyvault.vault.azure.net/"
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    try:
        secret = client.get_secret("nonexistent-secret")
        
    except HttpResponseError as e:
        print("=== Detailed Error Inspection ===")
        
        # Status code
        print(f"Status Code: {e.status_code}")
        
        # Error message
        print(f"Message: {e.message}")
        
        # Response object (if available)
        if e.response:
            print(f"Response headers: {dict(e.response.headers)}")
            
        # Structured error object
        if e.error:
            print(f"Error code: {e.error.code}")
            print(f"Error message: {e.error.message}")
            
        # Reason phrase
        print(f"Reason: {e.reason}")
        
        # Full exception details
        print(f"\nFull exception: {repr(e)}")


def handle_soft_deleted_secret():
    """
    Working with soft-deleted secrets
    
    When soft-delete is enabled (default), deleted secrets are retained
    for the retention period (default 90 days) and can be recovered.
    """
    vault_url = "https://your-keyvault.vault.azure.net/"
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    secret_name = "my-secret"
    
    print("Scenario: Attempting to get a soft-deleted secret\n")
    
    try:
        # This will raise 404 if the secret is deleted (even soft-deleted)
        secret = client.get_secret(secret_name)
        print(f"✅ Secret is active: {secret.name}")
        
    except HttpResponseError as e:
        if e.status_code == 404:
            print("❌ get_secret() failed with 404")
            print("Secret is either deleted or never existed\n")
            
            # Check if it's soft-deleted
            try:
                deleted_secret = client.get_deleted_secret(secret_name)
                
                print("✅ Secret found in soft-deleted state!")
                print(f"Name: {deleted_secret.name}")
                print(f"Deleted on: {deleted_secret.deleted_date}")
                print(f"Scheduled purge: {deleted_secret.scheduled_purge_date}")
                print(f"Can recover: {deleted_secret.recovery_id is not None}")
                
                # To recover the secret:
                print("\nTo recover:")
                print("  poller = client.begin_recover_deleted_secret(secret_name)")
                print("  recovered_secret = poller.result()")
                
                # Note: You cannot create a new secret with the same name
                # until the deleted secret is purged or recovered
                
            except HttpResponseError as del_err:
                if del_err.status_code == 404:
                    print("Secret does not exist (never created or purged)")
                elif del_err.status_code == 403:
                    print("Access denied to view deleted secrets")
                    print("Need 'Key Vault Secrets Officer' role or 'list' + 'recover' permissions")


def bulk_operations_with_error_handling():
    """
    Handle errors when performing bulk operations
    """
    vault_url = "https://your-keyvault.vault.azure.net/"
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    secret_names = ["secret1", "secret2", "secret3", "secret4"]
    results = {}
    
    for name in secret_names:
        try:
            secret = client.get_secret(name)
            results[name] = {
                "status": "success",
                "value": secret.value
            }
            
        except HttpResponseError as e:
            results[name] = {
                "status": "error",
                "status_code": e.status_code,
                "error": e.message
            }
            
            # Continue processing other secrets
            if e.status_code == 429:
                # Rate limited - add delay before next request
                time.sleep(2)
                
    # Report results
    print("=== Bulk Operation Results ===")
    for name, result in results.items():
        if result["status"] == "success":
            print(f"✅ {name}: Retrieved")
        else:
            print(f"❌ {name}: {result['status_code']} - {result['error']}")


if __name__ == "__main__":
    print("Azure Key Vault Error Handling Examples")
    print("=" * 50)
    
    # Note: These examples will fail without proper setup
    # They demonstrate the error handling patterns
    
    try:
        handle_secret_operations()
    except Exception as e:
        print(f"Demo completed with expected errors: {type(e).__name__}")
