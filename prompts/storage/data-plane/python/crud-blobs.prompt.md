---
id: storage-dp-python-crud
service: storage
plane: data-plane
language: python
category: crud
difficulty: basic
description: >
  Can a developer upload, download, list, and delete blobs in Azure Blob Storage
  using the Python SDK?
sdk_package: azure-storage-blob
doc_url: https://learn.microsoft.com/en-us/python/api/overview/azure/storage-blob-readme
tags:
  - blob
  - crud
  - getting-started
created: 2025-07-27
author: ronniegeraghty
---

# CRUD Blobs: Azure Blob Storage (Python)

## Prompt

Write a Python script that performs
CRUD operations on Azure Blob Storage:
1. Create a BlobServiceClient using DefaultAzureCredential
2. Create a container named "my-container" if it doesn't exist
3. Upload a local file "report.csv" as a blob named "reports/report.csv"
4. List all blobs in the container and print each blob's name and content length
5. Download the blob and save it to "report-downloaded.csv"
6. Delete the blob and then delete the container

Show required pip packages and proper error handling with HttpResponseError.

## Evaluation Criteria

- Installing `azure-storage-blob` and `azure-identity` packages
- `BlobServiceClient` with `DefaultAzureCredential`
- `BlobServiceClient.create_container()` or `ContainerClient.create_container()`
- `BlobClient.upload_blob()` with `overwrite` parameter
- `ContainerClient.list_blobs()` iteration
- `BlobClient.download_blob()` and `readall()` or `readinto()`
- `BlobClient.delete_blob()` and `ContainerClient.delete_container()`
- `HttpResponseError` and `ResourceExistsError` handling

## Context

Python is the most popular language for data engineering and ML workflows
that use Azure Blob Storage. This tests whether the generated code provides a
complete blob lifecycle tutorial that a data engineer can follow end-to-end.
