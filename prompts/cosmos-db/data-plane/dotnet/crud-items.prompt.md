---
id: cosmos-db-dp-dotnet-crud
service: cosmos-db
plane: data-plane
language: dotnet
category: crud
difficulty: basic
description: >
  Can a developer create, read, query, and delete items in an Azure Cosmos DB
  container using the .NET SDK documentation?
sdk_package: Microsoft.Azure.Cosmos
doc_url: https://learn.microsoft.com/en-us/azure/cosmos-db/nosql/quickstart-dotnet
tags:
  - cosmos-db
  - nosql
  - crud
  - getting-started
created: 2025-07-28
author: ronniegeraghty
---

# CRUD Items: Azure Cosmos DB (.NET)

## Prompt

Using only the Azure SDK for .NET documentation, write a C# program that performs
CRUD operations on items in an Azure Cosmos DB NoSQL container:
1. Create a CosmosClient using a connection string
2. Create a database named "TestDB" and a container named "Items" with partition key "/category"
3. Insert a JSON item with properties: id, category, name, and quantity
4. Read the item back by id and partition key
5. Query items where category equals "electronics" using SQL-like syntax
6. Replace the item with updated quantity
7. Delete the item

Show required NuGet packages and proper error handling with CosmosException.

## Evaluation Criteria

The documentation should cover:
- `Microsoft.Azure.Cosmos` NuGet package
- `CosmosClient` creation and configuration
- `Database.CreateDatabaseIfNotExistsAsync()`
- `Container.CreateContainerIfNotExistsAsync()` with partition key
- `Container.CreateItemAsync<T>()`, `ReadItemAsync<T>()`, `ReplaceItemAsync<T>()`, `DeleteItemAsync<T>()`
- `Container.GetItemQueryIterator<T>()` with `QueryDefinition`
- `CosmosException` handling with status codes

## Context

Cosmos DB is one of the most popular Azure data services. CRUD operations test
whether the .NET docs cover the full item lifecycle including partitioning
and SQL-like query syntax in the NoSQL API.
