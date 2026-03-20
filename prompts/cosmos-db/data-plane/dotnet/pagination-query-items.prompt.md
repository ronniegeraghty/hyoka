---
id: cosmos-db-dp-dotnet-pagination
service: cosmos-db
plane: data-plane
language: dotnet
category: pagination
difficulty: intermediate
description: >
  Can a developer paginate through large Cosmos DB query results
  using continuation tokens in .NET?
sdk_package: Microsoft.Azure.Cosmos
tags:
  - pagination
  - query
  - continuation-token
  - feed-iterator
created: 2025-07-27
author: ronniegeraghty
---

# Pagination: Query Items in Azure Cosmos DB (.NET)

## Prompt

Using only the Azure SDK for .NET documentation, write a C# program that queries
items in a Cosmos DB container with proper pagination:
1. Execute a SQL query "SELECT * FROM c WHERE c.category = 'electronics'" against a container
2. Process results page-by-page using FeedIterator, limiting each page to 50 items
3. Print the continuation token after each page
4. Implement a "resume from token" pattern where the query can restart from a saved token
5. Track total RU consumption across all pages

Use the Microsoft.Azure.Cosmos SDK v3. Show how to configure MaxItemCount
and explain the difference between FeedIterator and LINQ-based queries.

## Expected Coverage

- `Container.GetItemQueryIterator<T>()` with `QueryDefinition`
- `QueryRequestOptions.MaxItemCount` for page size control
- `FeedIterator<T>.HasMoreResults` and `ReadNextAsync()` loop pattern
- `FeedResponse<T>.ContinuationToken` for resumable pagination
- Passing continuation token to resume a query
- `FeedResponse<T>.RequestCharge` for RU tracking
- Cross-partition query considerations
- LINQ alternative via `GetItemLinqQueryable<T>()`

## Context

Cosmos DB pagination via FeedIterator is fundamentally different from
traditional database cursors. Developers must understand continuation tokens
and RU consumption per page to build efficient data access patterns.
