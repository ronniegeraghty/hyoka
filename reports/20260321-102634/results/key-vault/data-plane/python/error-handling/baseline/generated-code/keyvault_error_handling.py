"""
Azure Key Vault Secrets - Error Handling Patterns

This module demonstrates proper error handling when working with Azure Key Vault
secrets using the azure-keyvault-secrets SDK.
"""

from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential
from azure.core.exceptions import HttpResponseError, ResourceNotFoundError
import time


def handle_common_errors(vault_url: str, secret_name: str):
    """
    Demonstrates handling common Key Vault errors: 403, 404, and 429.
    """
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    try:
        secret = client.get_secret(secret_name)
        print(f"Successfully retrieved secret: {secret.name}")
        return secret.value
        
    except HttpResponseError as e:
        # Inspect the status code to handle different error scenarios
        status_code = e.status_code
        error_message = e.message
        
        if status_code == 403:
            # Access Denied - Missing RBAC permissions
            print(f"Access Denied (403): {error_message}")
            print("Your app identity needs 'Key Vault Secrets User' or "
                  "'Key Vault Secrets Officer' role assigned.")
            print(f"Error details: {e.error}")
            
        elif status_code == 404:
            # Secret Not Found
            print(f"Secret Not Found (404): {error_message}")
            print(f"The secret '{secret_name}' does not exist in the vault.")
            
        elif status_code == 429:
            # Throttling - Rate limit exceeded
            print(f"Rate Limited (429): {error_message}")
            print("Too many requests. Implementing retry with backoff...")
            # Check for Retry-After header
            retry_after = e.response.headers.get('Retry-After', 60)
            print(f"Retry after {retry_after} seconds")
            
        else:
            # Other HTTP errors
            print(f"HTTP Error ({status_code}): {error_message}")
            print(f"Error code: {e.error.code if e.error else 'N/A'}")
            
        # Re-raise if you want calling code to handle it
        raise


def handle_with_retry(vault_url: str, secret_name: str, max_retries: int = 3):
    """
    Demonstrates handling 429 throttling errors with exponential backoff retry.
    """
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    retry_count = 0
    base_delay = 1
    
    while retry_count < max_retries:
        try:
            secret = client.get_secret(secret_name)
            return secret.value
            
        except HttpResponseError as e:
            if e.status_code == 429:
                retry_count += 1
                if retry_count >= max_retries:
                    print(f"Max retries ({max_retries}) exceeded")
                    raise
                
                # Use Retry-After header if available, otherwise exponential backoff
                retry_after = e.response.headers.get('Retry-After')
                if retry_after:
                    delay = int(retry_after)
                else:
                    delay = base_delay * (2 ** (retry_count - 1))
                
                print(f"Rate limited. Retrying in {delay} seconds... "
                      f"(attempt {retry_count}/{max_retries})")
                time.sleep(delay)
            else:
                # For non-429 errors, don't retry
                raise


def handle_soft_deleted_secret(vault_url: str, secret_name: str):
    """
    Demonstrates what happens when accessing a soft-deleted secret.
    
    When you try to get a soft-deleted secret:
    - You'll receive a 404 (ResourceNotFoundError/HttpResponseError)
    - The error message will indicate the secret is in a deleted state
    - You cannot access the secret value until you recover it
    - You can list deleted secrets or get deleted secret properties instead
    """
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    try:
        # This will fail if the secret is soft-deleted
        secret = client.get_secret(secret_name)
        print(f"Secret found: {secret.name}")
        return secret.value
        
    except HttpResponseError as e:
        if e.status_code == 404:
            print(f"Secret not found (404): {e.message}")
            
            # Check if it's a soft-deleted secret
            try:
                deleted_secret = client.get_deleted_secret(secret_name)
                print(f"\nSecret '{secret_name}' is SOFT-DELETED:")
                print(f"  Deleted on: {deleted_secret.deleted_date}")
                print(f"  Scheduled purge: {deleted_secret.scheduled_purge_date}")
                print(f"  Recovery ID: {deleted_secret.recovery_id}")
                print("\nTo access this secret, you must first recover it:")
                print(f"  recovery_operation = client.begin_recover_deleted_secret('{secret_name}')")
                print(f"  recovered_secret = recovery_operation.result()")
                
            except HttpResponseError as deleted_err:
                if deleted_err.status_code == 404:
                    print(f"Secret truly does not exist (not soft-deleted)")
                elif deleted_err.status_code == 403:
                    print("Cannot check deleted secrets - missing permissions")
                    print("Need 'Key Vault Secrets Officer' role or higher")
        raise


def inspect_error_details(vault_url: str, secret_name: str):
    """
    Demonstrates how to extract detailed information from HttpResponseError.
    """
    credential = DefaultAzureCredential()
    client = SecretClient(vault_url=vault_url, credential=credential)
    
    try:
        secret = client.get_secret(secret_name)
        return secret.value
        
    except HttpResponseError as e:
        print("=== HttpResponseError Details ===")
        print(f"Status Code: {e.status_code}")
        print(f"Message: {e.message}")
        print(f"Reason: {e.reason}")
        
        # Error object (if available)
        if e.error:
            print(f"\nError Code: {e.error.code}")
            print(f"Error Message: {e.error.message}")
        
        # Response headers
        print(f"\nResponse Headers:")
        if e.response:
            for key, value in e.response.headers.items():
                if key.lower() in ['x-ms-request-id', 'x-ms-client-request-id', 
                                   'retry-after', 'www-authenticate']:
                    print(f"  {key}: {value}")
        
        raise


def categorize_error(e: HttpResponseError) -> str:
    """
    Helper function to categorize Key Vault errors.
    """
    status_code = e.status_code
    
    error_categories = {
        403: "PERMISSION_DENIED",
        404: "NOT_FOUND",
        429: "THROTTLED",
        401: "AUTHENTICATION_FAILED",
        409: "CONFLICT",
        500: "SERVER_ERROR",
        503: "SERVICE_UNAVAILABLE"
    }
    
    return error_categories.get(status_code, "UNKNOWN_ERROR")


# Example usage patterns
if __name__ == "__main__":
    vault_url = "https://your-keyvault.vault.azure.net/"
    
    # Example 1: Basic error handling
    try:
        value = handle_common_errors(vault_url, "my-secret")
    except HttpResponseError:
        print("Failed to retrieve secret")
    
    # Example 2: With retry logic for throttling
    try:
        value = handle_with_retry(vault_url, "my-secret", max_retries=3)
    except HttpResponseError:
        print("Failed after retries")
    
    # Example 3: Handling soft-deleted secrets
    try:
        value = handle_soft_deleted_secret(vault_url, "deleted-secret")
    except HttpResponseError:
        print("Could not access secret")
