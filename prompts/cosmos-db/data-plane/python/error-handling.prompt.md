---
id: cosmos-db-dp-python-error-handling
service: cosmos-db
plane: data-plane
language: python
category: error-handling
difficulty: intermediate
description: >
  Can a developer handle Cosmos DB errors including
  throttling (429), conflicts (409), and not-found (404) in Python?
sdk_package: azure-cosmos
doc_url: https://learn.microsoft.com/en-us/python/api/overview/azure/cosmos-readme
tags:
  - error-handling
  - exceptions
  - throttling
  - request-units
created: 2025-07-27
author: ronniegeraghty
---

# Error Handling: Azure Cosmos DB (Python)

## Prompt

How do I properly handle errors when working with Azure Cosmos DB in Python?
My application occasionally hits 429 (rate limited) errors when the provisioned
throughput is exceeded. Show me how to catch CosmosHttpResponseError, inspect
the status_code and sub_status, and use the retry_after header for backoff.
Also cover 404 (item not found) and 409 (conflict on create) scenarios.
Use the azure-cosmos Python SDK.

## Evaluation Criteria

- `CosmosHttpResponseError` as the primary exception type
- Extracting `status_code`, `sub_status`, and `message`
- `retry_after` property for 429 backoff timing
- Handling 404 for missing databases, containers, or items
- Handling 409 for duplicate item creation
- `CosmosClient` retry configuration via connection policy
- Request charge tracking via response headers
- Using `azure.cosmos.exceptions` imports

## Context

Python Cosmos DB developers frequently encounter rate limiting under load.
The generated code needs to clearly demonstrate the exception model and how to implement
proper retry logic, especially the distinction between transient 429s
and permanent 404/403 errors.
