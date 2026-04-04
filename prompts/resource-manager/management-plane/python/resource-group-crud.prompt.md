---
id: resource-manager-mp-python-rg-crud
properties:
  service: resource-manager
  plane: management-plane
  language: python
  category: crud
  difficulty: basic
  description: 'Can a developer create, list, update, and delete Azure Resource Groups using the Python management SDK?

    '
  sdk_package: azure-mgmt-resource
  doc_url: https://learn.microsoft.com/en-us/python/api/overview/azure/mgmt-resource-readme
  created: '2025-07-28'
  author: ronniegeraghty
tags:
- resource-groups
- management-plane
- provisioning
- getting-started
---

# Resource Group Management: Azure Resource Manager (Python)

## Prompt

Write a Python script that manages
Azure Resource Groups using the management plane SDK:
1. Authenticate using DefaultAzureCredential
2. Create a new resource group in "eastus" region
3. List all resource groups in the subscription
4. Get details of the created resource group
5. Add a tag to the resource group
6. Delete the resource group

Show required pip packages and include proper error handling.

## Evaluation Criteria

The generated code should include:
- `azure-mgmt-resource` and `azure-identity` pip packages
- `ResourceManagementClient` creation with credential and subscription_id
- `resource_groups.create_or_update()` with `ResourceGroup` model
- `resource_groups.list()` iteration
- `resource_groups.get()` for details
- Tag updates via `resource_groups.update()`
- `resource_groups.begin_delete()` with poller

## Context

Resource group management is the foundation of Azure management plane operations.
This tests whether the generated code clearly demonstrates the management client pattern
including subscription ID requirements and the long-running delete operation.
