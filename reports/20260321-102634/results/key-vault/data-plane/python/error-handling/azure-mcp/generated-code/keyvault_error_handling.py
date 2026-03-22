"""
Azure Key Vault Error Handling Examples
Demonstrates proper exception handling patterns for azure-keyvault-secrets SDK
"""

from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential
from azure.core.exceptions import HttpResponseError, ResourceNotFoundError
import time


def handle_get_secret_with_retry(secret_client: SecretClient, secret_name: str, max_retries: int = 3):
    """
    Example: Get a secret with comprehensive error handling and retry logic for throttling
    """
    retry_count = 0
    
    while retry_count <= max_retries:
        try:
            secret = secret_client.get_secret(secret_name)
            print(f"Successfully retrieved secret: {secret.name}")
            print(f"Value: {secret.value}")
            return secret
            
        except HttpResponseError as e:
            # Check the HTTP status code
            status_code = e.status_code
            
            if status_code == 403:
                # Access Denied - RBAC permission issue
                print(f"❌ Access Denied (403): Your identity lacks the required RBAC role")
                print(f"Error message: {e.message}")
                print(f"Error code: {e.error.code if hasattr(e, 'error') and e.error else 'N/A'}")
                print("Required role: 'Key Vault Secrets User' or 'Key Vault Secrets Officer'")
                raise  # Don't retry on permission errors
                
            elif status_code == 404:
                # Secret not found
                print(f"❌ Secret Not Found (404): '{secret_name}' does not exist")
                print(f"Error message: {e.message}")
                
                # Check if it might be a soft-deleted secret
                try:
                    deleted_secret = secret_client.get_deleted_secret(secret_name)
                    print(f"⚠️  Secret '{secret_name}' exists in soft-deleted state")
                    print(f"   Deleted on: {deleted_secret.deleted_date}")
                    print(f"   Scheduled purge: {deleted_secret.scheduled_purge_date}")
                    print("   Action: Recover the secret or wait for purge, then recreate")
                except HttpResponseError as del_error:
                    if del_error.status_code == 404:
                        print("   Secret does not exist in active or deleted state")
                    else:
                        print(f"   Could not check deleted secrets: {del_error.message}")
                
                raise  # Don't retry on not found
                
            elif status_code == 429:
                # Throttling - Rate limit exceeded
                retry_count += 1
                
                # Try to get retry-after header
                retry_after = e.response.headers.get('Retry-After', 5) if hasattr(e, 'response') else 5
                try:
                    retry_after = int(retry_after)
                except (ValueError, TypeError):
                    retry_after = 5
                
                if retry_count <= max_retries:
                    print(f"⚠️  Throttled (429): Rate limit exceeded, attempt {retry_count}/{max_retries}")
                    print(f"Error message: {e.message}")
                    print(f"Waiting {retry_after} seconds before retry...")
                    time.sleep(retry_after)
                    continue  # Retry the operation
                else:
                    print(f"❌ Max retries exceeded after throttling")
                    raise
                    
            else:
                # Other HTTP errors
                print(f"❌ HTTP Error {status_code}: {e.message}")
                print(f"Response: {e.response.text() if hasattr(e, 'response') else 'N/A'}")
                raise
                
        except ResourceNotFoundError as e:
            # Alternative exception type that may be raised for 404
            print(f"❌ Resource Not Found: {e.message}")
            raise
            
        except Exception as e:
            # Catch-all for unexpected errors
            print(f"❌ Unexpected error: {type(e).__name__}: {str(e)}")
            raise


def handle_set_secret(secret_client: SecretClient, secret_name: str, secret_value: str):
    """
    Example: Set a secret with error handling
    """
    try:
        secret = secret_client.set_secret(secret_name, secret_value)
        print(f"✅ Secret '{secret_name}' created/updated successfully")
        return secret
        
    except HttpResponseError as e:
        if e.status_code == 403:
            print(f"❌ Access Denied (403): Cannot set secret - insufficient permissions")
            print(f"Required role: 'Key Vault Secrets Officer'")
            print(f"Error: {e.message}")
        elif e.status_code == 409:
            print(f"❌ Conflict (409): Secret may be in soft-deleted state")
            print(f"Error: {e.message}")
            print("Action: Recover or purge the deleted secret first")
        else:
            print(f"❌ Error setting secret: {e.status_code} - {e.message}")
        raise


def handle_delete_secret(secret_client: SecretClient, secret_name: str):
    """
    Example: Delete a secret with error handling
    """
    try:
        poller = secret_client.begin_delete_secret(secret_name)
        deleted_secret = poller.result()
        print(f"✅ Secret '{secret_name}' deleted (soft-delete)")
        print(f"   Recovery ID: {deleted_secret.recovery_id}")
        print(f"   Can be recovered until: {deleted_secret.scheduled_purge_date}")
        return deleted_secret
        
    except HttpResponseError as e:
        if e.status_code == 403:
            print(f"❌ Access Denied (403): Cannot delete secret")
            print(f"Required role: 'Key Vault Secrets Officer'")
        elif e.status_code == 404:
            print(f"❌ Secret '{secret_name}' not found - may already be deleted")
        else:
            print(f"❌ Error deleting secret: {e.status_code} - {e.message}")
        raise


def handle_recover_deleted_secret(secret_client: SecretClient, secret_name: str):
    """
    Example: Recover a soft-deleted secret
    """
    try:
        poller = secret_client.begin_recover_deleted_secret(secret_name)
        recovered_secret = poller.result()
        print(f"✅ Secret '{secret_name}' recovered successfully")
        return recovered_secret
        
    except HttpResponseError as e:
        if e.status_code == 403:
            print(f"❌ Access Denied (403): Cannot recover secret")
            print(f"Required role: 'Key Vault Secrets Officer'")
        elif e.status_code == 404:
            print(f"❌ Secret '{secret_name}' not found in deleted state")
        else:
            print(f"❌ Error recovering secret: {e.status_code} - {e.message}")
        raise


def comprehensive_error_inspection(secret_client: SecretClient, secret_name: str):
    """
    Example: Detailed error inspection showing all available properties
    """
    try:
        secret = secret_client.get_secret(secret_name)
        return secret
        
    except HttpResponseError as e:
        print("=== Detailed Error Information ===")
        print(f"Status Code: {e.status_code}")
        print(f"Reason: {e.reason}")
        print(f"Message: {e.message}")
        
        # Error object (if available)
        if hasattr(e, 'error') and e.error:
            print(f"Error Code: {e.error.code}")
            print(f"Error Message: {e.error.message}")
        
        # Response object (if available)
        if hasattr(e, 'response'):
            print(f"Response Headers: {dict(e.response.headers)}")
            print(f"Response Text: {e.response.text()}")
        
        # Additional context
        print(f"Exception Type: {type(e).__name__}")
        print(f"Exception String: {str(e)}")
        
        raise


# Main demonstration
if __name__ == "__main__":
    # Initialize client
    vault_url = "https://your-keyvault-name.vault.azure.net/"
    
    try:
        credential = DefaultAzureCredential()
        client = SecretClient(vault_url=vault_url, credential=credential)
        
        print("=== Example 1: Handling 404 (Not Found) ===")
        try:
            handle_get_secret_with_retry(client, "non-existent-secret")
        except HttpResponseError:
            pass
        
        print("\n=== Example 2: Handling 403 (Access Denied) ===")
        print("(Simulated - would occur with insufficient RBAC permissions)")
        
        print("\n=== Example 3: Handling 429 (Throttling) ===")
        print("(Automatic retry with exponential backoff)")
        
        print("\n=== Example 4: Working with Soft-Deleted Secrets ===")
        secret_name = "test-secret"
        
        # Set a secret
        handle_set_secret(client, secret_name, "test-value")
        
        # Delete it (soft-delete)
        handle_delete_secret(client, secret_name)
        
        # Try to get it - will get 404
        print("\nAttempting to get soft-deleted secret:")
        try:
            handle_get_secret_with_retry(client, secret_name)
        except HttpResponseError:
            pass
        
        # Recover it
        print("\nRecovering the secret:")
        handle_recover_deleted_secret(client, secret_name)
        
        # Now we can get it again
        handle_get_secret_with_retry(client, secret_name)
        
    except Exception as e:
        print(f"\n❌ Failed to initialize: {e}")


"""
KEY TAKEAWAYS:

1. ALWAYS catch HttpResponseError from azure.core.exceptions
2. Check e.status_code to handle specific HTTP errors
3. Use e.message for human-readable error description
4. For 429 (throttling), implement retry logic with Retry-After header
5. For 404, check if secret is soft-deleted using get_deleted_secret()
6. For 403, verify RBAC role assignments (Key Vault Secrets User/Officer)
7. Soft-deleted secrets cannot be accessed until recovered or purged
8. Use ResourceNotFoundError as alternative for 404 handling

RBAC Roles Required:
- Read secrets: 'Key Vault Secrets User'
- Create/Update/Delete secrets: 'Key Vault Secrets Officer'
- Purge/Recover secrets: 'Key Vault Secrets Officer'
"""
