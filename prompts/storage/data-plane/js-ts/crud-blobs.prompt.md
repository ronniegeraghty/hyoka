---
id: storage-dp-js-ts-crud
properties:
  service: storage
  plane: data-plane
  language: js-ts
  category: crud
  difficulty: basic
  description: 'Can a developer upload, download, list, and delete blobs in Azure Blob Storage using the JavaScript/TypeScript
    SDK?

    '
  sdk_package: '@azure/storage-blob'
  doc_url: https://learn.microsoft.com/en-us/javascript/api/overview/azure/storage-blob-readme
  created: '2025-07-27'
  author: ronniegeraghty
tags:
- blob
- crud
- getting-started
---

# CRUD Blobs: Azure Blob Storage (JavaScript/TypeScript)

## Prompt

Write a TypeScript program that
performs CRUD operations on Azure Blob Storage:
1. Create a BlobServiceClient using DefaultAzureCredential
2. Create a container named "my-container" if it doesn't exist
3. Upload a string "Hello Azure!" as a block blob named "greeting.txt"
4. List all blobs in the container and log their names
5. Download the blob and print its content as a string
6. Delete the blob and then delete the container

Show required npm packages and proper error handling with RestError.
Use async/await throughout.

## Evaluation Criteria

- Installing `@azure/storage-blob` and `@azure/identity` packages
- `BlobServiceClient` construction with `DefaultAzureCredential`
- `ContainerClient.createIfNotExists()`
- `BlockBlobClient.upload()` or `uploadData()` for string content
- `ContainerClient.listBlobsFlat()` async iteration
- `BlobClient.download()` and reading the response stream
- `BlobClient.delete()` and `ContainerClient.delete()`
- `RestError` handling with `statusCode`

## Context

JavaScript/TypeScript is the most used language on npm and critical for
full-stack Azure developers. This tests whether the generated code provides clear
guidance on blob CRUD including the async streaming download pattern.
