---
id: cosmos-db-dp-js-ts-crud
service: cosmos-db
plane: data-plane
language: js-ts
category: crud
difficulty: basic
description: >
  Can a developer create, read, query, and delete items in an Azure Cosmos DB
  container using the JavaScript/TypeScript SDK documentation?
sdk_package: "@azure/cosmos"
doc_url: https://learn.microsoft.com/en-us/azure/cosmos-db/nosql/quickstart-nodejs
tags:
  - cosmos-db
  - nosql
  - crud
  - getting-started
created: 2025-07-28
author: ronniegeraghty
---

# CRUD Items: Azure Cosmos DB (JavaScript/TypeScript)

## Prompt

Using only the Azure SDK for JavaScript documentation, write a TypeScript program
that performs CRUD operations on items in an Azure Cosmos DB NoSQL container:
1. Create a CosmosClient using endpoint and key
2. Create a database "TestDB" and container "Items" with partition key "/category"
3. Create an item with properties: id, category, name, quantity
4. Read the item back using item().read()
5. Query items where category equals "electronics" using parameterized query
6. Replace the item with updated quantity using item().replace()
7. Delete the item using item().delete()

Show required npm package and handle errors with appropriate status code checks.

## Evaluation Criteria

The documentation should cover:
- `@azure/cosmos` npm package
- `CosmosClient` constructor with endpoint and key
- `client.databases.createIfNotExists()` and `database.containers.createIfNotExists()`
- `container.items.create()`, `container.item(id, partitionKey).read()`
- `container.items.query()` with `SqlQuerySpec`
- `container.item(id, partitionKey).replace()` and `.delete()`
- `FeedResponse` iteration and error status codes

## Context

The JavaScript Cosmos DB SDK uses a fluent chain pattern (container.item().read()).
This tests whether the JS/TS docs cover the chained resource model and the
FeedResponse pattern for query results.
