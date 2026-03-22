"""
QUICK REFERENCE: Azure Key Vault Secrets Pagination

Based on Official Azure SDK for Python Documentation
====================================================

1. INSTALLATION
   pip install azure-keyvault-secrets azure-identity

2. BASIC SETUP
   from azure.keyvault.secrets import SecretClient
   from azure.identity import DefaultAzureCredential
   
   credential = DefaultAzureCredential()
   client = SecretClient(vault_url=vault_url, credential=credential)

3. PAGINATION PATTERN (Recommended for large vaults)
   
   # Get ItemPaged iterator
   secrets_paged = client.list_properties_of_secrets()
   
   # Iterate page by page
   for page in secrets_paged.by_page():
       for secret_props in page:
           print(secret_props.name)
           print(secret_props.content_type)
           print(secret_props.created_on)
           print(secret_props.enabled)

4. SIMPLE ITERATION (For small vaults)
   
   for secret_props in client.list_properties_of_secrets():
       print(secret_props.name)

5. KEY PROPERTIES AVAILABLE (No secret values!)
   
   secret_props.name           # str - Secret name
   secret_props.content_type   # str - Content type hint
   secret_props.created_on     # datetime - Creation time (UTC)
   secret_props.updated_on     # datetime - Last update time
   secret_props.enabled        # bool - Is secret enabled?
   secret_props.expires_on     # datetime - Expiration time
   secret_props.tags           # dict - Custom metadata
   secret_props.version        # str - Version identifier

6. GET SECRET VALUE (Separate call required)
   
   secret = client.get_secret("secret-name")
   value = secret.value  # Now you have the actual secret value

7. FILTERING
   
   # Filter at application level
   for secret_props in client.list_properties_of_secrets():
       if secret_props.enabled and secret_props.content_type == "text/plain":
           print(secret_props.name)

8. CONTINUATION TOKENS (For resuming)
   
   paged = client.list_properties_of_secrets()
   pages = paged.by_page()
   
   # Process first page
   first_page = next(pages)
   
   # Save token for later
   token = pages.continuation_token
   
   # Later, resume from token
   resumed = client.list_properties_of_secrets().by_page(token)

9. ENVIRONMENT SETUP
   
   export VAULT_URL="https://your-vault.vault.azure.net/"
   
   # For service principal auth:
   export AZURE_CLIENT_ID="..."
   export AZURE_CLIENT_SECRET="..."
   export AZURE_TENANT_ID="..."
   
   # Or use Azure CLI:
   az login

10. REQUIRED PERMISSIONS
    
    secrets/list - To list secret properties

WHY USE by_page()?
==================
✓ Memory efficient for large vaults
✓ Better performance (processes in chunks)
✓ Supports continuation tokens
✓ Follows Azure SDK best practices
✓ Matches how the Azure API works internally

IMPORTANT NOTES
===============
• list_properties_of_secrets() does NOT return secret VALUES
• Use get_secret(name) to retrieve actual secret values
• SecretProperties only contains metadata
• Pagination happens automatically server-side
• Page size is determined by Azure (typically 25 items)
"""
