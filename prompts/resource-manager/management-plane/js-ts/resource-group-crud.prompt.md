---
id: resource-manager-mp-js-ts-rg-crud
service: resource-manager
plane: management-plane
language: js-ts
category: crud
difficulty: basic
description: >
  Can a developer create, list, update, and delete Azure Resource Groups
  using the JavaScript/TypeScript management SDK documentation?
sdk_package: "@azure/arm-resources"
doc_url: https://learn.microsoft.com/en-us/javascript/api/overview/azure/arm-resources-readme
tags:
  - resource-groups
  - management-plane
  - provisioning
  - getting-started
created: 2025-07-28
author: ronniegeraghty
---

# Resource Group Management: Azure Resource Manager (JavaScript/TypeScript)

## Prompt

Using only the Azure SDK for JavaScript documentation, write a TypeScript program
that manages Azure Resource Groups using the management plane SDK:
1. Authenticate using DefaultAzureCredential from @azure/identity
2. Create a ResourceManagementClient with the credential and subscription ID
3. Create a new resource group in "eastus" region
4. List all resource groups in the subscription using iteration
5. Get details of the created resource group
6. Update the resource group by adding a tag
7. Delete the resource group using beginDeleteAndWait

Show required npm packages and proper async/await patterns.
Use the @azure/arm-resources package.

## Evaluation Criteria

The documentation should cover:
- `@azure/arm-resources` and `@azure/identity` npm packages
- `DefaultAzureCredential` for authentication
- `ResourceManagementClient` constructor with credential and subscriptionId
- `resourceGroups.createOrUpdate()` with resource group name and parameters
- `resourceGroups.list()` with async iteration (`for await...of`)
- `resourceGroups.get()` for fetching details
- `resourceGroups.beginDeleteAndWait()` for long-running delete
- Tag updates via `resourceGroups.update()` with tags parameter

## Context

The JavaScript management SDK uses a client-per-service pattern. This tests
whether the JS/TS docs cover the ARM client creation, async iteration for
listing operations, and the beginAndWait pattern for LROs.
