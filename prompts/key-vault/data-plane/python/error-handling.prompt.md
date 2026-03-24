---
id: key-vault-dp-python-error-handling
service: key-vault
plane: data-plane
language: python
category: error-handling
difficulty: intermediate
description: >
  Can a developer handle Key Vault errors including
  access denied (403), secret not found (404), and throttling (429) in Python?
sdk_package: azure-keyvault-secrets
doc_url: https://learn.microsoft.com/en-us/python/api/overview/azure/keyvault-secrets-readme
tags:
  - error-handling
  - exceptions
  - access-policy
  - throttling
created: 2025-07-27
author: ronniegeraghty
---

# Error Handling: Azure Key Vault Secrets (Python)

## Prompt

How do I properly handle errors when working with Azure Key Vault secrets in Python?
I need to handle access denied (403) when my app's identity doesn't have the right
RBAC role, secret not found (404), and throttling (429) when rate limits are hit.
Show me try/except patterns with the azure-keyvault-secrets SDK including
how to inspect the status_code and error message on HttpResponseError.
Also explain what happens when I try to get a soft-deleted secret.

## Evaluation Criteria

- `HttpResponseError` and `ResourceNotFoundError` exception types
- Extracting `status_code`, `error.code`, and `message`
- 403 handling: diagnosing RBAC vs. access policy issues
- 404 handling: secret not found vs. deleted-but-not-purged
- 429 throttling: Key Vault rate limits and retry-after
- `SecretClient` retry configuration via kwargs
- Soft-delete aware error handling (`begin_recover_deleted_secret`)
- Using `azure.core.exceptions` imports

## Context

Key Vault access errors (403) are the single most common support question.
Python developers need clear guidance on diagnosing whether the issue is
RBAC configuration, access policy, or network restrictions.
