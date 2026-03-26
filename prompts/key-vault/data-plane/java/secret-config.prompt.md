---
id: key-vault-dp-java-secret-config
service: key-vault
plane: data-plane
language: java
category: crud
difficulty: intermediate
description: >
  Can an agent generate a Key Vault-backed configuration provider with secret
  versioning, expiry inspection, in-memory caching with bulk-load, and safe
  secret rotation using long-running delete operations (SyncPoller/PollerFlux)?
sdk_package: com.azure:azure-security-keyvault-secrets
doc_url: https://learn.microsoft.com/en-us/java/api/overview/azure/security-keyvault-secrets-readme
tags:
  - key-vault
  - secrets
  - caching
  - secret-rotation
  - lro
  - sync-poller
  - poller-flux
  - versioning
  - expiry
  - async
  - reactor
created: 2026-03-25
author: JonathanGiles, samvaity
---

# Secret Config Provider: Azure Key Vault (Java)

## Prompt

Create a small Java 17 Maven project that implements an application configuration provider backed by Azure Key Vault.

The project needs:

- A **secret provider class** (both sync and async versions) that retrieves secrets from Key Vault by name, with graceful handling when a secret doesn't exist (return a default value instead of crashing). It should also be able to retrieve a specific version of a secret (not just the latest), and inspect a secret's expiry date so the caller can tell if a secret is about to expire.

- A **caching layer** on top of the provider that stores secret values in memory after first retrieval. It should support bulk-loading a predefined set of required config keys at startup, on-demand refresh of individual keys, and automatic re-fetch of any secret whose expiry date is within a configurable warning window (e.g., 7 days out).

- A **configuration/factory class** that connects securely to the Key Vault using the vault URL from an environment variable. The application runs in Azure and should authenticate using managed identity — no client secrets or certificates in code.

- A **secret rotation helper** that safely rotates a secret: delete the old secret, ensure the deletion is fully complete, then create the new secret with an updated value and expiry date. The rotation must be safe — don't assume deletion is instantaneous, since Key Vault's soft-delete feature means the secret may not be immediately gone.

- A **Main class** that demos both implementations: loading several config keys at startup, reading them from cache, refreshing one, printing a warning if any secret is near expiry, and performing a secret rotation (delete old, wait for completion, create new). Run the full demo with the sync implementation first, then repeat with the async implementation.

Include a complete `pom.xml` with the necessary Azure SDK dependencies.

## Evaluation Criteria

### Dependencies
- Uses `com.azure:azure-security-keyvault-secrets` (not `com.microsoft.azure:azure-keyvault`)
- Uses `com.azure:azure-identity`
- No `com.microsoft.azure` groupId anywhere
- Specifies Java 17

### Authentication
- Uses `DefaultAzureCredential` — no client secrets, certificates, or tenant IDs in code
- Reads Key Vault URL from environment variable

### Client Construction
- Uses `SecretClientBuilder` (sync) / `SecretAsyncClient` builder (async)
- Builder chain includes `.vaultUrl()` and `.credential()`

### SDK Patterns
- Secret versioning: retrieves specific version via `getSecret(name, version)`
- Secret expiry: accesses `properties().getExpiresOn()` on `SecretProperties`
- Configurable warning window for near-expiry detection
- In-memory caching (e.g., `ConcurrentHashMap`) with bulk-load and single-key refresh
- Secret rotation uses `beginDeleteSecret()` as a long-running operation
- Sync uses `SyncPoller` to wait for delete completion
- Async uses `PollerFlux` to wait for delete completion
- Creates new secret only after delete completes (not concurrently)

### Error Handling
- Catches `ResourceNotFoundException` or `HttpResponseException` with 404 for missing secrets
- Returns a default value when secret is not found (does not crash)
- Does not use bare `Exception` catches

### Async Quality
- Uses `SecretAsyncClient` (not sync on background thread)
- Uses Project Reactor types (`Mono`, `Flux`)
- LRO uses `PollerFlux` (not `SyncPoller`)
- Does not call `.block()` inside the async implementation

### Anti-Patterns (should NOT appear)
- `KeyVaultClient` (old v7 API)
- `ServiceClientCredentials` or `AuthenticationCallback`
- `com.microsoft.azure.*` imports
- Fire-and-forget `deleteSecret()` without waiting for completion

## Context

This goes beyond basic secret CRUD (covered by `crud-secrets.prompt.md`) to test production
patterns: secret versioning, expiry-aware caching, and safe secret rotation using long-running
operations. The rotation pattern is critical — Key Vault uses soft-delete, so `beginDeleteSecret()`
returns a `SyncPoller`/`PollerFlux` that must be polled to completion before the new secret
can be created. LLMs frequently generate a simple `deleteSecret()` call without waiting, which
fails in production. The caching layer tests whether the agent can build a practical config
provider on top of the raw Key Vault client.
