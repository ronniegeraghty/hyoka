---
id: storage-mp-python-account-mgmt
service: storage
plane: management-plane
language: python
category: provisioning
difficulty: intermediate
description: >
  Can a developer create, configure, and manage Azure Storage Accounts
  using the Python management SDK?
sdk_package: azure-mgmt-storage
doc_url: https://learn.microsoft.com/en-us/python/api/overview/azure/mgmt-storage-readme
tags:
  - storage-account
  - management-plane
  - provisioning
created: 2025-07-28
author: ronniegeraghty
---

# Storage Account Management: Azure Storage (Python)

## Prompt

Write a Python script that manages
Azure Storage Accounts using the management plane SDK:
1. Authenticate using DefaultAzureCredential
2. Create a new Storage Account with Standard_LRS SKU in "eastus"
3. List all Storage Accounts in a resource group
4. Get the properties of the created Storage Account
5. Update the account to enable blob versioning
6. Delete the Storage Account

Show required pip packages and include proper error handling.

## Evaluation Criteria

The generated code should include:
- `azure-mgmt-storage` and `azure-identity` pip packages
- `StorageManagementClient` creation with credential and subscription_id
- `storage_accounts.begin_create()` with `StorageAccountCreateParameters`
- SKU and kind configuration
- `storage_accounts.list_by_resource_group()` iteration
- `storage_accounts.get_properties()` for details
- `storage_accounts.update()` for property changes
- `storage_accounts.delete()` for cleanup

## Context

Storage Account management is one of the most common management plane tasks.
This tests whether the generated code covers the full lifecycle including
the long-running create operation and model configuration for SKUs and features.
