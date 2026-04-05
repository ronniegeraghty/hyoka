---
id: cosmos-db-dp-python-retries
properties:
  service: cosmos-db
  plane: data-plane
  language: python
  category: retries
  difficulty: advanced
  description: 'Can a developer configure custom retry policies for Azure Cosmos DB including 429 throttle handling and connection
    retry in Python?

    '
  sdk_package: azure-cosmos
  doc_url: https://learn.microsoft.com/en-us/python/api/overview/azure/cosmos-readme
  created: '2025-07-27'
  author: ronniegeraghty
tags:
- retries
- retry-policy
- throttling
- request-units
---

# Retry Configuration: Azure Cosmos DB (Python)

## Prompt

How do I configure custom retry policies for Azure Cosmos DB operations in Python?
My application frequently hits 429 (rate limited) errors during bulk inserts and
I need to:
1. Configure max retry attempts for rate-limited (429) requests
2. Set the max wait time for retries
3. Handle transient connection failures with automatic reconnection
4. Implement custom backoff logic for specific operations
5. Understand which errors the SDK retries automatically vs which require manual handling

Show me how to configure CosmosClient with custom retry settings
and explain the interaction between SDK retries and provisioned throughput.

## Evaluation Criteria

- `CosmosClient` retry configuration via `connection_policy` or constructor kwargs
- `retry_total`, `retry_backoff_factor`, `retry_backoff_max` settings
- 429 retry behavior: automatic retry with `retry-after-ms` header
- `max_retry_attempts_on_throttled_requests` and `max_retry_wait_time_in_seconds`
- Connection retry vs request retry distinction
- Session consistency retry on session-not-found (404/1002)
- Non-retryable errors (400, 403, 409)
- Bulk execution mode and its built-in retry handling

## Context

Cosmos DB's consumption model means 429 errors are expected, not exceptional.
Developers need to tune retry policies to balance throughput against cost.
The generated code should explain how SDK retries interact with provisioned RU/s.
