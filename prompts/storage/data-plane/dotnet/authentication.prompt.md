---
id: storage-dp-dotnet-auth
service: storage
plane: data-plane
language: dotnet
category: authentication
difficulty: basic
description: >
  Can a developer authenticate to Azure Blob Storage
  using DefaultAzureCredential in .NET?
sdk_package: Azure.Storage.Blobs
doc_url: https://learn.microsoft.com/en-us/dotnet/api/overview/azure/storage.blobs-readme
tags:
  - identity
  - default-azure-credential
  - getting-started
created: 2025-07-27
author: ronniegeraghty
---

# Authentication: Azure Blob Storage (.NET)

## Prompt

How do I authenticate to Azure Blob Storage using DefaultAzureCredential in C#?
I need to create a BlobServiceClient that uses managed identity in production
but falls back to Azure CLI credentials during local development.
Show me the complete setup including required NuGet packages.

## Evaluation Criteria

The generated code should include:
- Installing `Azure.Identity` and `Azure.Storage.Blobs` packages
- Creating a `DefaultAzureCredential` instance
- Passing the credential to `BlobServiceClient`
- Explanation of the credential chain (managed identity → CLI → env vars)
- Error handling when no credential is available

## Context

This is one of the most common first tasks for a developer using Azure Storage.
The generated code should be easy to follow without prior Azure SDK experience.
