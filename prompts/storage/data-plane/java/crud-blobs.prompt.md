---
id: storage-dp-java-crud
service: storage
plane: data-plane
language: java
category: crud
difficulty: basic
description: >
  Can a developer upload, download, list, and delete blobs in Azure Blob Storage
  using the Java SDK?
sdk_package: azure-storage-blob
doc_url: https://learn.microsoft.com/en-us/java/api/overview/azure/storage-blob-readme
tags:
  - blob
  - crud
  - getting-started
created: 2025-07-27
author: ronniegeraghty
---

# CRUD Blobs: Azure Blob Storage (Java)

## Prompt

Write a Java application that performs
CRUD operations on Azure Blob Storage:
1. Create a BlobServiceClient using DefaultAzureCredential
2. Create a container named "my-container" if it doesn't exist
3. Upload a local file "data.txt" as a blob named "uploads/data.txt"
4. List all blobs in the container and print their names and sizes
5. Download the blob back to a local file "data-downloaded.txt"
6. Delete the blob and then delete the container

Show required Maven dependencies and proper error handling with BlobStorageException.

## Evaluation Criteria

- Maven dependency for `azure-storage-blob` and `azure-identity`
- `BlobServiceClientBuilder` with `DefaultAzureCredential`
- `BlobContainerClient.create()` and `exists()` check
- `BlobClient.uploadFromFile()` and `downloadToFile()`
- `BlobContainerClient.listBlobs()` iteration
- `BlobClient.delete()` and `BlobContainerClient.delete()`
- `BlobStorageException` handling with status codes

## Context

Blob Storage is Azure's most widely used data service. Java is a top enterprise
language. CRUD operations test whether the generated code provides a complete,
runnable workflow covering the full blob lifecycle.
