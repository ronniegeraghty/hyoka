---
id: cosmos-db-dp-java-crud
properties:
  service: cosmos-db
  plane: data-plane
  language: java
  category: crud
  difficulty: basic
  description: 'Can a developer create, read, query, and delete items in an Azure Cosmos DB container using the Java SDK?

    '
  sdk_package: azure-cosmos
  doc_url: https://learn.microsoft.com/en-us/java/api/overview/azure/cosmos-readme
  created: '2025-07-28'
  author: ronniegeraghty
tags:
- cosmos-db
- nosql
- crud
- getting-started
---

# CRUD Items: Azure Cosmos DB (Java)

## Prompt

Write a Java program that performs
CRUD operations on items in an Azure Cosmos DB NoSQL container:
1. Create a CosmosClient using endpoint and key with CosmosClientBuilder
2. Create a database "TestDB" and container "Items" with partition key "/category"
3. Insert a POJO item with properties: id, category, name, quantity
4. Read the item back by id and partition key value
5. Query items where category equals "electronics" using parameterized SQL
6. Replace the item with updated quantity
7. Delete the item by id and partition key

Show required Maven dependency and handle CosmosException appropriately.

## Evaluation Criteria

The generated code should include:
- `azure-cosmos` Maven dependency (com.azure:azure-cosmos)
- `CosmosClientBuilder` and `CosmosClient`
- `CosmosDatabase` and `CosmosContainer` creation
- `CosmosContainer.createItem()`, `readItem()`, `replaceItem()`, `deleteItem()`
- `CosmosQueryRequestOptions` and `CosmosPagedIterable`
- `SqlQuerySpec` with parameters
- `CosmosException` error handling

## Context

Cosmos DB Java SDK uses a builder pattern and typed POJO-based operations. This tests
whether the generated code properly covers the fluent API and parameterized queries,
which differ significantly from the .NET SDK approach.
