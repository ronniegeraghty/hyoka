---
id: cosmos-db-dp-python-crud
service: cosmos-db
plane: data-plane
language: python
category: crud
difficulty: basic
description: >
  Can a developer create, read, query, and delete items in an Azure Cosmos DB
  container using the Python SDK documentation?
sdk_package: azure-cosmos
doc_url: https://learn.microsoft.com/en-us/azure/cosmos-db/nosql/quickstart-python
tags:
  - cosmos-db
  - nosql
  - crud
  - getting-started
created: 2025-07-28
author: ronniegeraghty
---

# CRUD Items: Azure Cosmos DB (Python)

## Prompt

Using only the Azure SDK for Python documentation, write a Python script that performs
CRUD operations on items in an Azure Cosmos DB NoSQL container:
1. Create a CosmosClient using endpoint and key
2. Create a database "TestDB" and container "Items" with partition key "/category"
3. Upsert an item dict with keys: id, category, name, quantity
4. Read the item back using read_item() with id and partition key
5. Query items where category equals "electronics" using parameterized query
6. Replace the item with updated quantity
7. Delete the item

Show required pip packages and handle exceptions from azure.cosmos.exceptions.

## Evaluation Criteria

The documentation should cover:
- `azure-cosmos` pip package
- `CosmosClient` creation
- `database_client.create_database_if_not_exists()`
- `database.create_container_if_not_exists()` with `PartitionKey`
- `container.create_item()`, `read_item()`, `replace_item()`, `delete_item()`
- `container.query_items()` with `enable_cross_partition_query`
- `CosmosHttpResponseError` exception handling

## Context

The Python Cosmos DB SDK uses dictionary-based items rather than typed models.
This tests whether the Python docs cover the dict-based API and the differences
in query behavior (especially cross-partition queries).
