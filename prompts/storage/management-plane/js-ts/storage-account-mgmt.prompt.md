---
id: storage-mp-js-ts-account-mgmt
service: storage
plane: management-plane
language: js-ts
category: provisioning
difficulty: intermediate
description: >
  Can a developer create, configure, and manage Azure Storage Accounts
  using the JavaScript/TypeScript management SDK?
sdk_package: "@azure/arm-storage"
doc_url: https://learn.microsoft.com/en-us/javascript/api/overview/azure/arm-storage-readme
tags:
  - storage-account
  - management-plane
  - provisioning
created: 2025-07-28
author: ronniegeraghty
---

# Storage Account Management: Azure Storage (JavaScript/TypeScript)

## Prompt

Write a TypeScript program
that manages Azure Storage Accounts using the management plane SDK:
1. Authenticate using DefaultAzureCredential from @azure/identity
2. Create a StorageManagementClient with the credential and subscription ID
3. Create a new Storage Account with Standard_LRS SKU in "eastus"
4. List all Storage Accounts in a resource group using async iteration
5. Get the properties of the created Storage Account
6. Update the account to enable blob versioning
7. Delete the Storage Account

Show required npm packages (@azure/arm-storage) and proper async/await patterns.

## Evaluation Criteria

The generated code should include:
- `@azure/arm-storage` and `@azure/identity` npm packages
- `StorageManagementClient` constructor with credential and subscriptionId
- `storageAccounts.beginCreateAndWait()` with `StorageAccountCreateParameters`
- SKU and kind configuration in create parameters
- `storageAccounts.listByResourceGroup()` with async iteration
- `storageAccounts.getProperties()` for details
- `storageAccounts.update()` for modifying properties
- `storageAccounts.delete()` for removal

## Context

The JavaScript storage management SDK uses the ARM client pattern with beginAndWait
for LROs. This tests whether the generated code covers the full storage account lifecycle
including the async creation pattern.
