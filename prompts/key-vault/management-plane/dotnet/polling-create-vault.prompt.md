---
id: key-vault-mp-dotnet-polling
service: key-vault
plane: management-plane
language: dotnet
category: polling
difficulty: intermediate
description: >
  Can a developer use the LRO polling pattern to create a Key Vault
  and wait for completion using the .NET management SDK?
sdk_package: Azure.ResourceManager.KeyVault
tags:
  - polling
  - lro
  - long-running-operation
  - management-plane
created: 2025-07-27
author: ronniegeraghty
---

# Polling/LRO: Create Key Vault (.NET)

## Prompt

Using only the Azure SDK for .NET documentation, write a C# program that creates
an Azure Key Vault using the management plane SDK and handles the long-running
operation:
1. Authenticate using DefaultAzureCredential
2. Create a Key Vault with RBAC authorization enabled in "eastus"
3. Configure soft-delete and purge protection
4. Wait for the vault creation to complete using the ArmOperation pattern
5. Verify the vault is accessible by creating a SecretClient pointing to it

Use Azure.ResourceManager.KeyVault. Show required NuGet packages and how
to set access policies or RBAC roles during creation.

## Expected Coverage

- `KeyVaultCollection.CreateOrUpdateAsync()` returning `ArmOperation<KeyVaultResource>`
- `KeyVaultCreateOrUpdateContent` with `KeyVaultProperties`
- Configuring `EnableRbacAuthorization`, `EnableSoftDelete`, `EnablePurgeProtection`
- `VaultAccessPolicy` vs RBAC authorization model
- `ArmOperation<T>.WaitForCompletionAsync()` for completion
- `WaitUntil.Completed` vs `WaitUntil.Started`
- Tenant ID and object ID configuration
- Error handling for existing vaults and soft-deleted vaults

## Context

Key Vault creation requires configuring security-sensitive properties like
access policies and purge protection at creation time. The LRO pattern is
critical because vault DNS propagation adds latency to the operation.
