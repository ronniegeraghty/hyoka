---
id: storage-dp-python-batch
properties:
  service: storage
  plane: data-plane
  language: python
  category: batch
  difficulty: advanced
  description: 'Can a developer perform batch blob operations including bulk delete and bulk set-tier using the Python SDK?

    '
  sdk_package: azure-storage-blob
  doc_url: https://learn.microsoft.com/en-us/python/api/overview/azure/storage-blob-readme
  created: '2025-07-27'
  author: ronniegeraghty
tags:
- batch
- bulk-operations
- delete
- set-tier
---

# Batch Operations: Azure Blob Storage (Python)

## Prompt

How do I perform batch operations on Azure Blob Storage in Python?
I have a container with thousands of blobs and need to:
1. Bulk delete 500 blobs in a single batched request
2. Bulk set the access tier of blobs from Hot to Cool
3. Handle partial failures where some operations succeed and others fail
4. Understand the batch size limits and how to chunk large batches

Show me how to use the azure-storage-blob Python SDK's batch delete
and batch set tier capabilities. Include proper error handling for
partial batch failures and explain the difference between batch
operations and sequential operations.

## Evaluation Criteria

- `ContainerClient.delete_blobs()` for bulk deletion
- `ContainerClient.set_standard_blob_tier_blobs()` for bulk tier changes
- Passing multiple blob names or BlobProperties to batch methods
- Partial failure handling with `PartialBatchErrorException`
- Iterating over `PartialBatchErrorException.parts` for individual results
- Batch size limits (256 operations per batch)
- Chunking patterns for operations exceeding the batch limit
- `raise_on_any_failure=False` option for lenient error handling

## Context

Python data engineers frequently need to manage thousands of blobs — archiving
old data to Cool/Archive tier or cleaning up processed files. Batch operations
are critical for performance and cost but have subtle partial-failure semantics.
