---
id: storage-mp-java-account-mgmt
service: storage
plane: management-plane
language: java
category: provisioning
difficulty: intermediate
description: >
  Can a developer create, configure, and manage Azure Storage Accounts
  using the Java management SDK documentation?
sdk_package: azure-resourcemanager-storage
doc_url: https://learn.microsoft.com/en-us/java/api/overview/azure/resourcemanager-storage-readme
tags:
  - storage-account
  - management-plane
  - provisioning
created: 2025-07-28
author: ronniegeraghty
---

# Storage Account Management: Azure Storage (Java)

## Prompt

Using only the Azure SDK for Java documentation, write a Java program that manages
Azure Storage Accounts using the management plane SDK:
1. Authenticate using DefaultAzureCredential
2. Create a StorageManager instance with the credential and subscription
3. Create a new Storage Account with Standard_LRS SKU in "eastus"
4. List all Storage Accounts in a resource group
5. Get the properties of the created Storage Account
6. Update the account to enable blob versioning
7. Delete the Storage Account

Show required Maven dependency (com.azure.resourcemanager:azure-resourcemanager-storage)
and proper error handling.

## Evaluation Criteria

The documentation should cover:
- `azure-resourcemanager-storage` Maven dependency
- `StorageManager.authenticate()` with credential and profile
- `storageAccounts().define().withRegion().withExistingResourceGroup().withSku().create()`
- Fluent builder pattern for account creation
- `storageAccounts().listByResourceGroup()` for listing
- `storageAccounts().getByResourceGroup()` for details
- `update().withBlobAccessTier()` or service properties update
- `storageAccounts().deleteByResourceGroup()`

## Context

The Java management SDK uses a fluent builder pattern (define/create/update) that
differs significantly from other languages. This tests whether the Java docs cover
the fluent API for storage account lifecycle management.
