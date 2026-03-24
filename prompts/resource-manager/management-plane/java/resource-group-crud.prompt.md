---
id: resource-manager-mp-java-rg-crud
service: resource-manager
plane: management-plane
language: java
category: crud
difficulty: basic
description: >
  Can a developer create, list, update, and delete Azure Resource Groups
  using the Java management SDK?
sdk_package: azure-resourcemanager-resources
doc_url: https://learn.microsoft.com/en-us/java/api/overview/azure/resourcemanager-resources-readme
tags:
  - resource-groups
  - management-plane
  - provisioning
  - getting-started
created: 2025-07-28
author: ronniegeraghty
---

# Resource Group Management: Azure Resource Manager (Java)

## Prompt

Write a Java application that manages
Azure Resource Groups using the management plane SDK:
1. Authenticate using DefaultAzureCredential
2. Create a new resource group in "eastus" region
3. List all resource groups in the subscription
4. Get details of the created resource group
5. Add a tag to the resource group
6. Delete the resource group

Show required Maven dependencies and include proper exception handling.
Use the modern azure-resourcemanager SDK.

## Evaluation Criteria

The generated code should include:
- Maven dependency for `azure-resourcemanager` and `azure-identity`
- `AzureResourceManager.authenticate()` with credential and profile
- `resourceGroups().define().withRegion().create()`
- `resourceGroups().list()` iteration
- `resourceGroups().getByName()` for details
- Tag management via `update().withTag().apply()`
- `resourceGroups().deleteByName()` for cleanup

## Context

Resource group management is the foundation of Azure management plane operations.
This tests whether the generated code covers the fluent management SDK pattern
which differs significantly from the data plane client builder pattern.
