---
id: resource-manager-mp-dotnet-rg-crud
service: resource-manager
plane: management-plane
language: dotnet
category: crud
difficulty: basic
description: >
  Can a developer create, list, update, and delete Azure Resource Groups
  using the .NET management SDK documentation?
sdk_package: Azure.ResourceManager
doc_url: https://learn.microsoft.com/en-us/dotnet/azure/sdk/resource-management
tags:
  - resource-groups
  - management-plane
  - provisioning
  - getting-started
created: 2025-07-28
author: ronniegeraghty
---

# Resource Group Management: Azure Resource Manager (.NET)

## Prompt

Using only the Azure SDK for .NET documentation, write a C# program that manages
Azure Resource Groups using the management plane SDK:
1. Authenticate using DefaultAzureCredential
2. Create a new resource group in "eastus" region
3. List all resource groups in the subscription
4. Get details of the created resource group
5. Add a tag to the resource group
6. Delete the resource group

Show required NuGet packages and proper error handling.
Use the Azure.ResourceManager SDK (not the older Microsoft.Azure.Management packages).

## Evaluation Criteria

The documentation should cover:
- `Azure.ResourceManager` NuGet package
- `ArmClient` creation with `DefaultAzureCredential`
- `GetDefaultSubscription()` and `GetResourceGroups()` collection
- `CreateOrUpdate()`, `Get()`, `GetAll()` operations
- Tag management with `SetTags()` or `AddTag()`
- `Delete()` with `WaitForCompletion()`

## Context

Resource group management is the foundation of Azure management plane operations.
This tests whether the .NET docs cover the modern Azure.ResourceManager SDK
(track 2) rather than the older Microsoft.Azure.Management packages.
