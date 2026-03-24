---
id: cosmos-db-dp-dotnet-error-handling
service: cosmos-db
plane: data-plane
language: dotnet
category: error-handling
difficulty: intermediate
description: >
  Can a developer handle Cosmos DB errors including
  throttling (429), conflicts (409), and not-found (404) in .NET?
sdk_package: Microsoft.Azure.Cosmos
doc_url: https://learn.microsoft.com/en-us/dotnet/api/overview/azure/microsoft.azure.cosmos-readme
tags:
  - error-handling
  - exceptions
  - throttling
  - request-units
created: 2025-07-27
author: ronniegeraghty
---

# Error Handling: Azure Cosmos DB (.NET)

## Prompt

How do I properly handle errors when working with Azure Cosmos DB in .NET?
I'm particularly concerned about throttling — when my application exceeds
the provisioned Request Units (RU/s), I get 429 errors. Show me how to
catch CosmosException, extract the status code and retry-after header,
and implement proper retry logic. Also cover handling 404 (item not found)
and 409 (conflict) responses. Use the Microsoft.Azure.Cosmos SDK v3.

## Evaluation Criteria

- `CosmosException` as the primary exception type
- Extracting `StatusCode`, `SubStatusCode`, and `RetryAfter` properties
- Handling 429 (TooManyRequests) with retry-after backoff
- Handling 404 (NotFound) for missing items or containers
- Handling 409 (Conflict) for duplicate items
- `CosmosClientOptions.MaxRetryAttemptsOnRateLimitedRequests` configuration
- RU consumption tracking via `RequestCharge` on responses
- Diagnostics string for troubleshooting

## Context

Cosmos DB's Request Unit model means 429 throttling is expected behavior
under load, not an error to panic about. The generated code needs to clearly
explain how to handle rate limiting gracefully with proper retry logic.
