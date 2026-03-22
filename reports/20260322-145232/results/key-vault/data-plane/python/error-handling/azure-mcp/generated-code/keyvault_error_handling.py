"""
Azure Key Vault Secrets Error Handling Examples

This module demonstrates proper error handling patterns when working with
Azure Key Vault secrets, including handling 403 (access denied), 404 (not found),
429 (throttling), and soft-deleted secrets.
"""

from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential
from azure.core.exceptions import (
    HttpResponseError,
    ResourceNotFoundError,
    ServiceRequestError,
)
import time


def get_secret_with_basic_error_handling(client: SecretClient, secret_name: str):
    """
    Basic error handling pattern using HttpResponseError.
    
    This demonstrates how to catch and inspect the status code and error message.
    """
    try:
        secret = client.get_secret(secret_name)
        print(f"✓ Successfully retrieved secret: {secret.name}")
        return secret.value
    
    except HttpResponseError as e:
        # The status_code property contains the HTTP status code
        status_code = e.status_code
        
        # The message property contains the error description
        error_message = e.message
        
        # The error property contains more detailed error information
        error_code = e.error.code if hasattr(e, 'error') and e.error else None
        
        print(f"✗ HTTP Error {status_code}: {error_message}")
        if error_code:
            print(f"  Error Code: {error_code}")
        
        # Handle specific status codes
        if status_code == 403:
            print("  → Access Denied: Check RBAC permissions (Key Vault Secrets User role)")
        elif status_code == 404:
            print("  → Secret Not Found: The secret does not exist or may be soft-deleted")
        elif status_code == 429:
            print("  → Rate Limit Exceeded: Too many requests, implement retry with backoff")
        
        raise  # Re-raise after logging


def get_secret_with_specific_handling(client: SecretClient, secret_name: str):
    """
    Handle specific error scenarios with tailored responses.
    """
    try:
        secret = client.get_secret(secret_name)
        return secret.value
    
    except HttpResponseError as e:
        if e.status_code == 403:
            # Access denied - identity lacks required RBAC role
            print(f"Access denied to secret '{secret_name}'")
            print("Required RBAC role: 'Key Vault Secrets User' or 'Key Vault Secrets Officer'")
            print(f"Error details: {e.message}")
            return None
        
        elif e.status_code == 404:
            # Secret not found - may not exist or is soft-deleted
            print(f"Secret '{secret_name}' not found")
            print("The secret may be soft-deleted. Use get_deleted_secret() to check.")
            return None
        
        elif e.status_code == 429:
            # Throttling - rate limit exceeded
            retry_after = e.response.headers.get('Retry-After', 60)
            print(f"Rate limit exceeded. Retry after {retry_after} seconds")
            raise
        
        else:
            # Other HTTP errors
            print(f"Unexpected error (HTTP {e.status_code}): {e.message}")
            raise


def get_secret_with_retry_on_throttle(client: SecretClient, secret_name: str, max_retries: int = 3):
    """
    Implement retry logic for throttling (429) errors.
    """
    for attempt in range(max_retries):
        try:
            secret = client.get_secret(secret_name)
            return secret.value
        
        except HttpResponseError as e:
            if e.status_code == 429:
                if attempt < max_retries - 1:
                    # Get retry-after header or use exponential backoff
                    retry_after = int(e.response.headers.get('Retry-After', 2 ** attempt))
                    print(f"Rate limited. Retrying in {retry_after} seconds (attempt {attempt + 1}/{max_retries})")
                    time.sleep(retry_after)
                    continue
                else:
                    print(f"Max retries exceeded for secret '{secret_name}'")
                    raise
            else:
                # Non-throttling errors - don't retry
                raise
    
    return None


def handle_soft_deleted_secret(client: SecretClient, secret_name: str):
    """
    Demonstrate handling of soft-deleted secrets.
    
    When a secret is deleted (soft delete enabled), it enters a soft-deleted state.
    - get_secret() will return 404 (Not Found)
    - get_deleted_secret() can retrieve information about the deleted secret
    - You must purge or recover the secret before creating a new one with the same name
    """
    try:
        # Try to get the secret normally
        secret = client.get_secret(secret_name)
        print(f"✓ Secret '{secret_name}' exists and is active")
        return secret.value
    
    except HttpResponseError as e:
        if e.status_code == 404:
            print(f"Secret '{secret_name}' not found in active secrets")
            
            # Check if it's soft-deleted
            try:
                deleted_secret = client.get_deleted_secret(secret_name)
                print(f"✓ Found soft-deleted secret: {deleted_secret.name}")
                print(f"  Deleted on: {deleted_secret.deleted_date}")
                print(f"  Scheduled purge date: {deleted_secret.scheduled_purge_date}")
                print(f"  Recovery ID: {deleted_secret.recovery_id}")
                print("\nOptions:")
                print("  1. Recover: client.begin_recover_deleted_secret(secret_name)")
                print("  2. Purge: client.purge_deleted_secret(secret_name)")
                return None
            
            except HttpResponseError as deleted_error:
                if deleted_error.status_code == 404:
                    print(f"✗ Secret '{secret_name}' does not exist (not active or deleted)")
                else:
                    print(f"Error checking deleted secrets: {deleted_error.message}")
                return None
        else:
            raise


def comprehensive_error_handling_example(client: SecretClient, secret_name: str):
    """
    Comprehensive error handling combining all scenarios.
    """
    try:
        secret = client.get_secret(secret_name)
        print(f"✓ Retrieved secret: {secret.name}")
        return secret.value
    
    except HttpResponseError as e:
        # Extract error details
        status_code = e.status_code
        message = e.message
        error_code = getattr(e.error, 'code', 'Unknown') if hasattr(e, 'error') and e.error else 'Unknown'
        
        print(f"\n{'='*60}")
        print(f"HTTP Response Error Details:")
        print(f"{'='*60}")
        print(f"Status Code:  {status_code}")
        print(f"Error Code:   {error_code}")
        print(f"Message:      {message}")
        print(f"{'='*60}\n")
        
        # Handle each scenario
        if status_code == 403:
            print("🔒 ACCESS DENIED")
            print("Cause: Your application's identity lacks the required RBAC role")
            print("Solution:")
            print("  1. Ensure managed identity is enabled (if using Azure services)")
            print("  2. Assign 'Key Vault Secrets User' role to the identity")
            print("  3. Command: az role assignment create \\")
            print("       --role 'Key Vault Secrets User' \\")
            print("       --assignee <principal-id> \\")
            print("       --scope <key-vault-resource-id>")
            
        elif status_code == 404:
            print("🔍 SECRET NOT FOUND")
            print("Checking if secret is soft-deleted...")
            handle_soft_deleted_secret(client, secret_name)
            
        elif status_code == 429:
            print("⏱️  RATE LIMIT EXCEEDED")
            retry_after = e.response.headers.get('Retry-After', 'unknown')
            print(f"Retry-After header: {retry_after} seconds")
            print("Solution:")
            print("  1. Implement exponential backoff retry logic")
            print("  2. Reduce request frequency")
            print("  3. Consider caching secrets if appropriate")
            
        elif status_code >= 500:
            print("🔧 SERVER ERROR")
            print("This is a transient error from Azure Key Vault")
            print("Implement retry logic with exponential backoff")
        
        return None
    
    except ServiceRequestError as e:
        print(f"Network error: {e}")
        print("Check network connectivity and firewall rules")
        return None
    
    except Exception as e:
        print(f"Unexpected error: {type(e).__name__}: {e}")
        raise


def main():
    """
    Example usage of error handling patterns.
    """
    # Initialize the client
    # vault_url = "https://your-keyvault.vault.azure.net/"
    # credential = DefaultAzureCredential()
    # client = SecretClient(vault_url=vault_url, credential=credential)
    
    print("""
Azure Key Vault Secrets - Error Handling Patterns
================================================

This file demonstrates how to properly handle errors when working with
Azure Key Vault secrets in Python.

Key Error Scenarios:
-------------------
1. **403 Forbidden (Access Denied)**
   - Cause: Identity lacks RBAC permissions
   - Required role: 'Key Vault Secrets User' or 'Key Vault Secrets Officer'
   - Properties: e.status_code == 403, e.message contains details

2. **404 Not Found**
   - Cause: Secret doesn't exist or is soft-deleted
   - Check: Use get_deleted_secret() to check soft-deleted secrets
   - Properties: e.status_code == 404

3. **429 Too Many Requests (Throttling)**
   - Cause: Rate limit exceeded
   - Solution: Implement retry with exponential backoff
   - Check 'Retry-After' header: e.response.headers.get('Retry-After')

4. **Soft-Deleted Secrets**
   - get_secret() returns 404
   - Use get_deleted_secret() to retrieve deleted secret info
   - Options: recover (begin_recover_deleted_secret) or purge (purge_deleted_secret)

HttpResponseError Properties:
----------------------------
- e.status_code: HTTP status code (int)
- e.message: Error message (str)
- e.error: Detailed error object (may have .code attribute)
- e.response: Full HTTP response (has .headers, .status_code, etc.)

Example Usage:
-------------
    from azure.keyvault.secrets import SecretClient
    from azure.identity import DefaultAzureCredential
    
    vault_url = "https://your-keyvault.vault.azure.net/"
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    # Basic error handling
    secret_value = get_secret_with_basic_error_handling(client, "my-secret")
    
    # Comprehensive handling
    secret_value = comprehensive_error_handling_example(client, "my-secret")
    
    # With retry on throttling
    secret_value = get_secret_with_retry_on_throttle(client, "my-secret")
    
    # Check soft-deleted
    handle_soft_deleted_secret(client, "deleted-secret")
""")


if __name__ == "__main__":
    main()
