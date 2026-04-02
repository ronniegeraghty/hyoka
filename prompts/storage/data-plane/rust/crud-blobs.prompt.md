---
id: storage-dp-rust-crud
service: storage
plane: data-plane
language: rust
category: crud
difficulty: basic
description: >
  Can a developer upload, download, list, and delete blobs in Azure Blob Storage
  using the Rust SDK?
sdk_package: azure_storage_blob
doc_url: https://github.com/Azure/azure-sdk-for-rust/tree/main/sdk/storage/azure_storage_blobs
tags:
  - blobs
  - crud
  - getting-started
created: 2026-03-26
author: larryo
---

# CRUD Blobs: Azure Blob Storage (Rust)

## Prompt

Write a Rust program that performs the following blob storage operations:

1. Create a blob container called "my-container" if it does not already exist
2. Upload a local file "example.txt" as a block blob named "uploads/example.txt"
3. List all blobs in the container and print their names and sizes
4. Download the blob back to a local file "downloaded-example.txt"
5. Delete the blob and then delete the container

Include proper error handling for cases like container already existing
or blob not found.

Before the task is considered complete, ensure that the generated code compiles and runs successfully, demonstrating the CRUD operations on Azure Blob Storage using the Rust SDK. Always use Token Credentials always instead of SAS credentials.

In order to verify correct functionality, you may need to create an Azure Blob Storage instance and you will have to configure authentication. Make sure to have the necessary permissions to perform blob operations in the storage account you are using for testing.

## Evaluation Criteria

The generated code should include:

- cargo setup for `azure_storage_blob` and `azure_identity`
- The code generated should correctly compile and execute
- The code generated should be idiomatic Rust code.
- The code generated should use the latest published version of the Azure SDK for Rust.
- The code should demonstrate a blob service client with the storage account
  url and credential
- Creating a container for container creation
- Uploading from a file path
- Enumerate blobs with paging
- Download from a blob to a local file
- Deleting blob files and containers for cleanup
- Error handling

## Context

Blob storage is the most common Azure Storage service.

Testing CRUD coverage validates that the generated code covers both the build system
integration and the full lifecycle of blob operations for Rust developers.
