---
id: storage-dp-python-pagination
service: storage
plane: data-plane
language: python
category: pagination
difficulty: intermediate
description: >
  Can a developer correctly paginate through a large list of blobs in
  Azure Storage using the Python SDK?
sdk_package: azure-storage-blob
doc_url: https://learn.microsoft.com/en-us/python/api/overview/azure/storage-blob-readme
tags:
  - blob
  - pagination
  - list-blobs
  - continuation-token
created: 2025-07-27
author: ronniegeraghty
---

# Pagination: List Blobs in Azure Storage (Python)

## Prompt

Write a script that lists all blobs
in a container that has over 10,000 blobs. The script should:
1. Use page-by-page iteration (not loading all results into memory)
2. Process blobs in pages of 100
3. Print the total count and names of the first 5 blobs on each page
4. Handle the case where the container might not exist

Use DefaultAzureCredential for authentication. Show all required pip packages.

## Evaluation Criteria

The generated code should include:
- Using `list_blobs()` with `results_per_page` parameter
- Iterating by page using `.by_page()` or async iteration
- Continuation token usage for resumable listing
- Memory-efficient processing of large result sets
- Error handling for missing containers (`ResourceNotFoundError`)

## Context

Pagination is a core SDK pattern that many developers get wrong by loading
all results into a list. This tests whether the generated code demonstrates developers toward
the efficient page-by-page approach for large-scale blob listing.
