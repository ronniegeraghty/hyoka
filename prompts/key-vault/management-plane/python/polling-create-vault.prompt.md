---
id: key-vault-mp-python-polling
service: key-vault
plane: management-plane
language: python
category: polling
difficulty: intermediate
description: >
  Can a developer use the LRO polling pattern to create a Key Vault
  and wait for completion using the Python management SDK?
sdk_package: azure-mgmt-keyvault
tags:
  - polling
  - lro
  - long-running-operation
  - management-plane
created: 2025-07-27
author: ronniegeraghty
---

# Polling/LRO: Create Key Vault (Python)

## Prompt

Using only the Azure SDK for Python documentation, write a Python script that creates
an Azure Key Vault using the management plane SDK and handles the long-running
operation:
1. Authenticate using DefaultAzureCredential
2. Create a Key Vault with RBAC authorization enabled
3. Configure soft-delete retention (90 days) and purge protection
4. Wait for vault creation to complete using the LROPoller pattern
5. Verify the vault by connecting with SecretClient from azure-keyvault-secrets

Use azure-mgmt-keyvault with DefaultAzureCredential. Show required pip packages
and explain how to configure access policies vs RBAC during creation.

## Expected Coverage

- `KeyVaultManagementClient.vaults.begin_create_or_update()` returning `LROPoller`
- `VaultCreateOrUpdateParameters` with `VaultProperties`
- Configuring `enable_rbac_authorization`, `enable_soft_delete`, `enable_purge_protection`
- `soft_delete_retention_in_days` configuration
- `AccessPolicyEntry` vs RBAC model
- `LROPoller.result()` and `LROPoller.wait()` for completion
- Tenant ID and object ID configuration
- Handling soft-deleted vault recovery conflicts

## Context

Key Vault provisioning via management SDK is a common infrastructure-as-code
pattern. Python developers need to understand how to configure security-critical
properties like purge protection and choose between access policies and RBAC.
