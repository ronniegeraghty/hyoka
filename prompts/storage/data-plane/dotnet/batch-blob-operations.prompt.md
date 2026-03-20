---
id: storage-dp-dotnet-batch
service: storage
plane: data-plane
language: dotnet
category: batch
difficulty: advanced
description: >
  Can a developer perform batch blob operations including bulk delete
  and bulk set-tier using the .NET SDK?
sdk_package: Azure.Storage.Blobs
tags:
  - batch
  - bulk-operations
  - delete
  - set-tier
created: 2025-07-27
author: ronniegeraghty
---

# Batch Operations: Azure Blob Storage (.NET)

## Prompt

How do I perform batch operations on Azure Blob Storage in .NET?
I have a container with thousands of blobs and need to:
1. Bulk delete 500 blobs in a single batched HTTP request
2. Bulk set the access tier of 200 blobs from Hot to Cool
3. Handle partial failures where some operations in the batch succeed and others fail
4. Understand the limits — max operations per batch and size restrictions

Show me how to use BlobBatchClient to submit batch requests with
the Azure.Storage.Blobs.Batch package. Include proper error handling
for partial batch failures.

## Expected Coverage

- `BlobBatchClient` from `Azure.Storage.Blobs.Batch` package
- `BlobBatchClient.DeleteBlobsAsync()` for bulk delete
- `BlobBatchClient.SetBlobsAccessTierAsync()` for bulk tier changes
- Custom batch via `BlobBatchClient.CreateBatch()` and `SubmitBatchAsync()`
- Batch size limits (256 operations per batch)
- Partial failure handling: `AggregateException` with per-operation status
- `RequestFailedException` for individual operation failures within a batch
- Authentication scopes for batch operations

## Context

Batch operations are essential for cost optimization (changing storage tiers)
and cleanup (bulk deletion). Without batch support, developers resort to
sequential API calls that are slow and consume excessive transaction units.
