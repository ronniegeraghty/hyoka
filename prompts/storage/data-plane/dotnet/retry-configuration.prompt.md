---
id: storage-dp-dotnet-retries
service: storage
plane: data-plane
language: dotnet
category: retries
difficulty: advanced
description: >
  Can a developer configure custom retry policies for Azure Blob Storage
  including exponential backoff and per-operation timeouts in .NET?
sdk_package: Azure.Storage.Blobs
tags:
  - retries
  - retry-policy
  - resilience
  - exponential-backoff
created: 2025-07-27
author: ronniegeraghty
---

# Retry Configuration: Azure Blob Storage (.NET)

## Prompt

How do I configure custom retry policies for Azure Blob Storage operations in .NET?
My application needs to:
1. Set a custom retry policy with 5 max retries and exponential backoff
2. Configure per-operation timeouts so a single upload doesn't hang forever
3. Handle network errors (transient) differently from auth errors (non-transient)
4. Use a custom retry policy for specific high-value operations
5. Implement circuit-breaker patterns for sustained failures

Show me how to configure BlobClientOptions with custom RetryOptions,
and explain which HTTP status codes the SDK considers retryable by default.
Use the Azure.Storage.Blobs SDK.

## Expected Coverage

- `BlobClientOptions.Retry` configuration with `RetryOptions`
- `MaxRetries`, `Delay`, `MaxDelay`, `Mode` (Exponential vs Fixed)
- `NetworkTimeout` for per-request timeouts
- Default retryable status codes (408, 429, 500, 502, 503, 504)
- Non-retryable errors (400, 401, 403, 404, 409)
- Per-operation `CancellationToken` for timeout control
- Geo-redundant retry with `GeoRedundantSecondaryUri`
- Interaction with Polly or other resilience libraries

## Context

Default retry policies work for simple scenarios, but production applications
need fine-tuned retry behavior. Developers building resilient storage pipelines
need docs that explain the full retry model and when to override defaults.
