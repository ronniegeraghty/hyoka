---
id: cosmos-db-dp-java-todo-repository
service: cosmos-db
plane: data-plane
language: java
category: crud
difficulty: intermediate
description: >
  Can an agent generate a Cosmos DB CRUD repository with optimistic concurrency
  (ETags), parameterized queries, page-by-page pagination, TTL configuration,
  custom indexing policy, and RU cost logging?
sdk_package: com.azure:azure-cosmos
doc_url: https://learn.microsoft.com/en-us/java/api/overview/azure/cosmos-readme
tags:
  - cosmos-db
  - etag
  - optimistic-concurrency
  - pagination
  - parameterized-query
  - ttl
  - indexing-policy
  - request-charge
  - async
  - reactor
created: 2026-03-25
author: JonathanGiles, samvaity
---

# ToDo Repository: Azure Cosmos DB (Java)

## Prompt

Create a small Java 17 Maven project that implements a ToDo item CRUD repository backed by Azure Cosmos DB (NoSQL API).

The project needs:

- A **model class** (shared by both implementations) for a ToDo item with fields for id, title, description, completed status, created timestamp, and category (where category is the partition key).

- A **synchronous repository class** that performs CRUD operations against Cosmos DB. It should support create, read, update, delete, and a query-by-category method. Each operation should log the request charge (RU cost consumed). The update operation should prevent lost updates — if another process modified the item since it was last read, the update should fail with a clear conflict error rather than silently overwriting the other process's changes. The query method should use safe, parameterized queries and must handle large result sets properly — paginate through results rather than loading everything into memory at once, and log progress as each page is retrieved.

- An **asynchronous repository class** that provides the same CRUD operations. The query method should return results as a stream of pages, and the caller should be able to process each page as it arrives.

- A **configuration/factory class** that connects to the Cosmos DB account using its endpoint from an environment variable. Authentication must use managed identity (no master keys). It should create the database and container if they don't already exist, setting a default TTL (time-to-live) of 90 days on the container and configuring the indexing policy to exclude the `description` field from indexing (since it's never queried on).

- A **Main class** that demos both implementations: runs the full CRUD cycle using the sync repository first (including paginated query output showing page-by-page results), then runs the same operations using the async repository. Print RU costs and results to the console.

Include a complete `pom.xml` with the necessary Azure SDK dependencies.

## Evaluation Criteria

### Dependencies
- Uses `com.azure:azure-cosmos` (not `com.microsoft.azure:azure-documentdb` or `com.microsoft.azure:azure-cosmosdb`)
- Uses `com.azure:azure-identity`
- No `com.microsoft.azure` groupId anywhere
- Specifies Java 17

### Authentication
- Uses `DefaultAzureCredential` — no master keys or connection strings
- Reads Cosmos DB endpoint from environment variable

### Client Construction
- Uses `CosmosClientBuilder` with `.endpoint()` and `.credential()`
- Uses `CosmosClient` (sync) and `CosmosAsyncClient` (async)

### SDK Patterns
- Correct partition key usage: `/category` path, `PartitionKey` in all point operations
- ETag-based optimistic concurrency: captures ETag from read, passes `ifMatchETag` on update
- Handles 412 Precondition Failed as a specific error case for conflicts
- Parameterized queries using `SqlQuerySpec` with `SqlParameter` (no string concatenation)
- Page-by-page iteration using `iterableByPage()` or `CosmosPagedIterable`
- Configurable page size via `QueryRequestOptions.setMaxItemCount`
- Logs continuation token and item count per page
- Async query uses `CosmosPagedFlux` returning pages as a stream
- TTL configured at 90 days (7776000 seconds) via `ContainerProperties.setDefaultTimeToLiveInSeconds()`
- Indexing policy excludes `/description` path
- RU cost extracted from response via `getRequestCharge()` and logged per operation

### Error Handling
- Catches `CosmosException` with status code checks (404, 409, 412)
- Handles 412 separately for ETag conflicts
- Does not use bare `Exception` catches

### Async Quality
- Uses `CosmosAsyncClient` / `CosmosAsyncDatabase` / `CosmosAsyncContainer`
- Uses Project Reactor types (`Mono`, `Flux`)
- Does not call `.block()` inside the async implementation

### Anti-Patterns (should NOT appear)
- `DocumentClient` or `DocumentClientException` (old v2 API)
- Connection strings with `AccountKey=...`
- `com.microsoft.azure.*` imports
- Flattened query results (`.stream()` / `.forEach()` without page iteration)

## Context

This goes beyond basic Cosmos DB CRUD (covered by `crud-items.prompt.md`) to test production
patterns: optimistic concurrency with ETags to prevent lost updates, parameterized SQL queries
to avoid injection, page-by-page pagination with continuation tokens for large result sets,
TTL configuration for automatic document expiry, and custom indexing policy to optimize storage
and write cost. The RU cost logging tests whether the agent knows how to extract request charges
from the Cosmos DB response — critical for cost monitoring in production.
