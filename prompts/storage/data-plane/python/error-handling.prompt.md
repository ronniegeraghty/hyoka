---
id: storage-dp-python-error-handling
service: storage
plane: data-plane
language: python
category: error-handling
difficulty: intermediate
description: >
  Can the docs help a developer handle common Azure Blob Storage errors
  including 404, 403, and 429 responses in Python?
sdk_package: azure-storage-blob
tags:
  - error-handling
  - exceptions
  - retry
created: 2025-07-27
author: ronniegeraghty
---

# Error Handling: Azure Blob Storage (Python)

## Prompt

How do I properly handle errors when working with Azure Blob Storage in Python?
I need to understand what exceptions are raised for common failure scenarios:
blob not found (404), access denied (403), and resource conflict (409).
Show me idiomatic try/except patterns with the azure-storage-blob SDK including
how to inspect the status_code and error_code on HttpResponseError.
Also explain the difference between HttpResponseError and ResourceNotFoundError.

## Evaluation Criteria

- `HttpResponseError` as the base exception type
- Specific exceptions: `ResourceNotFoundError`, `ResourceExistsError`
- Extracting `status_code`, `error_code`, and `message` from exceptions
- Handling 404, 403, 409, and 429 status codes
- Retry policy configuration via `kwargs` or `RetryPolicy`
- Using `azure.core.exceptions` imports
- Logging configuration for debugging

## Context

Python developers expect clear exception hierarchies. Azure Storage uses
azure-core exceptions that differ from typical Python patterns. Docs must
clearly explain the exception tree so developers catch the right types.
