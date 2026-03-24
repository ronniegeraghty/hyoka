---
id: storage-dp-cpp-crud
service: storage
plane: data-plane
language: cpp
category: crud
difficulty: basic
description: >
  Can a developer upload, download, list, and delete blobs in Azure Blob Storage
  using the C++ SDK?
sdk_package: azure-storage-blobs-cpp
doc_url: https://github.com/Azure/azure-sdk-for-cpp/tree/main/sdk/storage/azure-storage-blobs
tags:
  - blobs
  - crud
  - getting-started
created: 2025-07-28
author: ronniegeraghty
---

# CRUD Blobs: Azure Blob Storage (C++)

## Prompt

Write a C++ program that performs
the following blob storage operations:
1. Create a blob container called "my-container" if it does not already exist
2. Upload a local file "example.txt" as a block blob named "uploads/example.txt"
3. List all blobs in the container and print their names and sizes
4. Download the blob back to a local file "downloaded-example.txt"
5. Delete the blob and then delete the container

Use DefaultAzureCredential from Azure Identity for authentication.
Show the required CMake configuration and vcpkg dependencies.
Include proper exception handling for cases like container already existing
or blob not found.

## Evaluation Criteria

The generated code should include:
- vcpkg/CMake setup for `azure-storage-blobs-cpp` and `azure-identity-cpp`
- Creating a `BlobServiceClient` with the storage account URL and credential
- `BlobContainerClient::CreateIfNotExists()` for container creation
- `BlockBlobClient::UploadFrom()` for uploading from a file path
- `BlobContainerClient::ListBlobs()` to enumerate blobs with paging
- `BlobClient::DownloadTo()` for downloading to a local file
- `BlobClient::Delete()` and `BlobContainerClient::Delete()` for cleanup
- Exception handling with `Azure::Core::RequestFailedException`

## Context

Blob storage is the most common Azure Storage service. The C++ SDK requires
CMake and vcpkg setup that differs significantly from other language SDKs.
Testing CRUD coverage validates that the generated code covers both the build system
integration and the full lifecycle of blob operations for C++ developers.
