---
id: storage-mp-go-account-mgmt
service: storage
plane: management-plane
language: go
category: provisioning
difficulty: intermediate
description: >
  Can a developer create, configure, and manage Azure Storage Accounts
  using the Go management SDK documentation?
sdk_package: github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage
doc_url: https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage
tags:
  - storage-account
  - management-plane
  - provisioning
created: 2025-07-28
author: ronniegeraghty
---

# Storage Account Management: Azure Storage (Go)

## Prompt

Using only the Azure SDK for Go documentation, write a Go program that manages
Azure Storage Accounts using the management plane SDK:
1. Authenticate using DefaultAzureCredential from azidentity
2. Create an AccountsClient from the armstorage package
3. Create a new Storage Account with Standard_LRS SKU in "eastus" using BeginCreate
4. Poll the LRO to completion
5. List all Storage Accounts in a resource group using NewListByResourceGroupPager
6. Get the properties of the created Storage Account
7. Update the account properties
8. Delete the Storage Account

Show required Go module imports and proper error handling with poller patterns.

## Evaluation Criteria

The documentation should cover:
- `github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage` module
- `azidentity.NewDefaultAzureCredential()` for authentication
- `armstorage.NewAccountsClient()` with subscription ID and credential
- `BeginCreate()` with `armstorage.AccountCreateParameters` (SKU, Kind, Location)
- `PollUntilDone()` for waiting on the LRO
- `NewListByResourceGroupPager()` with pager iteration pattern
- `GetProperties()` for account details
- `Update()` with `AccountUpdateParameters`
- `Delete()` for removal

## Context

The Go storage management SDK uses the standard LRO poller pattern and pager
iteration. This tests whether the Go docs cover the armstorage module and the
creation parameters that require SKU and Kind configuration.
