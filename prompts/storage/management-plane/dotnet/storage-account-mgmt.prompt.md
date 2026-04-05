---
id: storage-mp-dotnet-account-mgmt
properties:
  service: storage
  plane: management-plane
  language: dotnet
  category: provisioning
  difficulty: intermediate
  description: 'Can a developer create, configure, and manage Azure Storage Accounts using the .NET management SDK?

    '
  sdk_package: Azure.ResourceManager.Storage
  doc_url: https://learn.microsoft.com/en-us/dotnet/api/overview/azure/resourcemanager.storage-readme
  created: '2025-07-28'
  author: ronniegeraghty
tags:
- storage-account
- management-plane
- provisioning
---

# Storage Account Management: Azure Storage (.NET)

## Prompt

Write a C# program that manages
Azure Storage Accounts using the management plane SDK:
1. Authenticate using DefaultAzureCredential
2. Create a new Storage Account with Standard_LRS SKU in "eastus"
3. List all Storage Accounts in a resource group
4. Get the properties of the created Storage Account
5. Update the account to enable blob versioning
6. Delete the Storage Account

Show required NuGet packages and proper error handling.
Use the Azure.ResourceManager.Storage SDK.

## Evaluation Criteria

The generated code should include:
- `Azure.ResourceManager.Storage` NuGet package
- `ArmClient` and subscription/resource group navigation
- `StorageAccountCollection.CreateOrUpdate()` with `StorageAccountCreateOrUpdateContent`
- SKU and kind configuration (`StorageSku`, `StorageKind`)
- Listing and getting storage accounts
- Updating properties via `StorageAccountPatch`
- Delete operation

## Context

Storage Account management is one of the most common management plane tasks.
This tests whether the generated code covers the full lifecycle of a Storage Account
including the more complex configuration options like SKU, kind, and feature toggles.
