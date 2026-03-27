---
id: storage-auth-dotnet
service: storage
plane: data-plane
language: dotnet
category: authentication
difficulty: beginner
description: "Authenticate to Azure Blob Storage using DefaultAzureCredential"
sdk_package: Azure.Storage.Blobs
api_version: "2023-11-03"
doc_url: https://learn.microsoft.com/en-us/dotnet/api/azure.storage.blobs
tags:
  - authentication
  - blob
  - identity
created: "2024-01-15"
author: test-author
expected_packages:
  - Azure.Storage.Blobs
  - Azure.Identity
---

# Storage Authentication (.NET)

## Prompt

Write a C# console application that authenticates to Azure Blob Storage
using DefaultAzureCredential and lists all containers in the storage account.

The application should:
1. Use Azure.Identity for authentication
2. Use Azure.Storage.Blobs for storage operations
3. Accept the storage account URL as a command-line argument
4. List all containers and print their names and last modified dates
5. Handle authentication errors gracefully

## Notes

This is a beginner-level prompt for testing SDK evaluation tooling.
