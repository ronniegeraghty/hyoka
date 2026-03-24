---
id: storage-mp-dotnet-polling
service: storage
plane: management-plane
language: dotnet
category: polling
difficulty: intermediate
description: >
  Can a developer use the LRO polling pattern to create a Storage Account
  and wait for completion using the .NET management SDK?
sdk_package: Azure.ResourceManager.Storage
doc_url: https://learn.microsoft.com/en-us/dotnet/api/overview/azure/resourcemanager.storage-readme
tags:
  - polling
  - lro
  - long-running-operation
  - management-plane
created: 2025-07-27
author: ronniegeraghty
---

# Polling/LRO: Create Storage Account (.NET)

## Prompt

Write a C# program that creates
an Azure Storage Account using the management plane SDK and properly handles the
long-running operation (LRO):
1. Start the create operation using CreateOrUpdateAsync
2. Poll for completion using WaitForCompletionAsync
3. Show how to check the operation status while it's in progress
4. Handle timeout scenarios where the operation takes too long
5. Demonstrate the difference between WaitForCompletion and manual polling

Use Azure.ResourceManager.Storage with DefaultAzureCredential. Show required
NuGet packages and explain the ArmOperation<T> pattern.

## Evaluation Criteria

- `StorageAccountCollection.CreateOrUpdateAsync()` returning `ArmOperation<StorageAccountResource>`
- `ArmOperation<T>.WaitForCompletionAsync()` for simple completion
- `ArmOperation<T>.HasCompleted` and `UpdateStatusAsync()` for manual polling
- `ArmOperation<T>.Value` to get the result after completion
- Timeout handling with `CancellationToken`
- `WaitUntil.Completed` vs `WaitUntil.Started` parameter
- Error handling when the LRO fails

## Context

Storage Account creation is an LRO that typically takes 10-30 seconds. The
ArmOperation pattern is used across all Azure management SDKs, so understanding
it here teaches a transferable skill for all resource provisioning.
