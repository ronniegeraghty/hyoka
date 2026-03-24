---
id: key-vault-dp-python-pagination
service: key-vault
plane: data-plane
language: python
category: pagination
difficulty: intermediate
description: >
  Can a developer paginate through a large list of Key Vault secrets
  using the Python SDK?
sdk_package: azure-keyvault-secrets
doc_url: https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme
tags:
  - pagination
  - list-secrets
  - iteration
created: 2025-07-27
author: ronniegeraghty
---

# Pagination: List Key Vault Secrets (Python)

## Prompt

Write a Python script that lists all
secrets in an Azure Key Vault that contains hundreds of secrets. The script should:
1. Use SecretClient with DefaultAzureCredential
2. Iterate through secrets using the ItemPaged pattern
3. Process secrets in pages using by_page()
4. Print the name, content type, and created date of each secret
5. Filter to show only enabled secrets

I need to understand how the azure-keyvault-secrets SDK handles pagination
for large vaults. Show required pip packages.

## Evaluation Criteria

- `SecretClient.list_properties_of_secrets()` returning `ItemPaged`
- Direct iteration via `for secret in client.list_properties_of_secrets()`
- Page-by-page iteration via `.by_page()`
- Continuation token support for resumable listing
- `SecretProperties` attributes (name, content_type, enabled, created_on)
- `max_page_size` parameter for controlling page size
- Error handling with `HttpResponseError`

## Context

Enterprise Key Vaults often have hundreds of secrets across multiple
applications. Python developers need to enumerate secrets efficiently
and understand the ItemPaged pattern used throughout the Azure SDK for Python.
