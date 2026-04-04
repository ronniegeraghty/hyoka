---
id: resource-manager-mp-go-rg-crud
properties:
  service: resource-manager
  plane: management-plane
  language: go
  category: crud
  difficulty: basic
  description: 'Can a developer create, list, update, and delete Azure Resource Groups using the Go management SDK?

    '
  sdk_package: github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources
  doc_url: https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources
  created: '2025-07-28'
  author: ronniegeraghty
tags:
- resource-groups
- management-plane
- provisioning
- getting-started
---

# Resource Group Management: Azure Resource Manager (Go)

## Prompt

Write a Go program that manages
Azure Resource Groups using the management plane SDK:
1. Authenticate using DefaultAzureCredential from azidentity
2. Create a ResourceGroupsClient
3. Create a new resource group in "eastus" region
4. List all resource groups in the subscription
5. Get details of the created resource group
6. Update the resource group by adding a tag
7. Delete the resource group

Show required Go module imports and proper error handling.
Use the armresources package from the azure-sdk-for-go/sdk module.

## Evaluation Criteria

The generated code should include:
- `github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources` module
- `azidentity.NewDefaultAzureCredential()` for authentication
- `armresources.NewResourceGroupsClient()` with subscription ID and credential
- `BeginCreateOrUpdate()` with `armresources.ResourceGroup` and location
- `NewListPager()` for listing with pager iteration
- `Get()` for fetching resource group details
- `BeginDelete()` and poller pattern for long-running operations
- Tags as `map[string]*string`

## Context

The Go Azure SDK uses a module-per-service pattern with poller-based LROs.
This tests whether the generated code covers the armresources module and the
pager/poller patterns that are central to Go management plane operations.
