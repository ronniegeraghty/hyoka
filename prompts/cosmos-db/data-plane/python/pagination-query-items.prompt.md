---
id: cosmos-db-dp-python-pagination
service: cosmos-db
plane: data-plane
language: python
category: pagination
difficulty: intermediate
description: >
  Can a developer paginate through large Cosmos DB query results
  using continuation tokens in Python?
sdk_package: azure-cosmos
doc_url: https://learn.microsoft.com/en-us/python/api/overview/azure/cosmos-readme
tags:
  - pagination
  - query
  - continuation-token
created: 2025-07-27
author: ronniegeraghty
---

# Pagination: Query Items in Azure Cosmos DB (Python)

## Prompt

Write a Python script that queries
items in a Cosmos DB container with proper pagination:
1. Execute a SQL query "SELECT * FROM c WHERE c.status = 'active'" against a container
2. Process results in pages of 25 items using max_item_count
3. Capture and print the continuation token after each page
4. Implement a resume pattern that restarts the query from a saved continuation token
5. Track the total request charge (RU/s) across all pages

Use the azure-cosmos Python SDK. Show how to use query_items with
enable_cross_partition_query when the query doesn't filter by partition key.

## Evaluation Criteria

- `ContainerProxy.query_items()` with SQL query string
- `max_item_count` parameter for page size
- `.by_page()` for page-by-page iteration
- Continuation token via page response headers
- Resuming queries with continuation token parameter
- `enable_cross_partition_query=True` for cross-partition queries
- Request charge tracking via response headers
- Parameterized queries to prevent injection

## Context

Cosmos DB's Python SDK pagination differs from the typical ItemPaged pattern
used in other Azure SDKs. Developers need to understand the continuation token
model and cross-partition query behavior for efficient data access.
