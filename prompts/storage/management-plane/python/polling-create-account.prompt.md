---
id: storage-mp-python-polling
service: storage
plane: management-plane
language: python
category: polling
difficulty: intermediate
description: >
  Can a developer use the LRO polling pattern to create a Storage Account
  and wait for completion using the Python management SDK?
sdk_package: azure-mgmt-storage
tags:
  - polling
  - lro
  - long-running-operation
  - management-plane
created: 2025-07-27
author: ronniegeraghty
---

# Polling/LRO: Create Storage Account (Python)

## Prompt

Using only the Azure SDK for Python documentation, write a Python script that creates
an Azure Storage Account using the management plane SDK and properly handles the
long-running operation (LRO):
1. Start the create operation using begin_create
2. Wait for completion using result() or wait()
3. Show how to poll for status manually with status() and done()
4. Handle timeout scenarios
5. Demonstrate using the polling callback pattern

Use azure-mgmt-storage with DefaultAzureCredential. Show required pip packages
and explain the LROPoller pattern used across Azure management SDKs.

## Expected Coverage

- `StorageManagementClient.storage_accounts.begin_create()` returning `LROPoller`
- `LROPoller.result()` for blocking wait
- `LROPoller.wait()` with timeout parameter
- `LROPoller.status()` and `LROPoller.done()` for manual polling
- `LROPoller.add_done_callback()` for async notification
- `StorageAccountCreateParameters` configuration
- Error handling when the LRO fails (`HttpResponseError`)
- Timeout handling patterns

## Context

The LROPoller pattern is the foundation of all Azure management operations in
Python. Understanding it with Storage Account creation teaches a pattern that
applies to every Azure resource provisioning scenario.
