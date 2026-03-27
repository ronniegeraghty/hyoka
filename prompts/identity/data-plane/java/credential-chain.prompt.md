---
id: identity-dp-java-credential-chain
service: identity
plane: data-plane
language: java
category: auth
difficulty: intermediate
description: >
  Can an agent build environment-specific Azure credential chains (dev, CI,
  production) using ChainedTokenCredential, with managed identity, workload
  identity, CAE support, and environment auto-detection?
sdk_package: com.azure:azure-identity
doc_url: https://learn.microsoft.com/en-us/java/api/overview/azure/identity-readme
tags:
  - identity
  - chained-credential
  - managed-identity
  - workload-identity
  - cae
  - environment-detection
  - azure-pipelines
  - async
  - reactor
created: 2026-03-25
author: JonathanGiles, samvaity
---

# Credential Chain Builder: Azure Identity (Java)

## Prompt

Create a small Java 17 Maven project that demonstrates how to correctly build Azure credential chains tailored to different deployment environments — local development, CI/CD pipelines, and production.

The project needs:

- A **credential factory class** that builds the appropriate Azure credential for each environment. For local development, it should chain together credentials that work from developer tools (CLI, IDE plugins, etc.). For CI pipelines, it should support credentials sourced from pipeline environment variables or Azure Pipelines service connections. For production, it should prefer managed identity (supporting both system-assigned and user-assigned, where the user-assigned identity's client ID comes from an environment variable), with workload identity as a fallback for Kubernetes scenarios. The factory should also support enabling Continuous Access Evaluation (CAE) on the credentials, which lets Azure revoke tokens mid-session for security events.

- An **environment detector class** that auto-detects which environment the app is running in by probing for well-known environment variables (e.g., CI pipeline workspace variables, managed identity endpoint availability). It should classify the environment as dev, CI, or production.

- A **connectivity tester class** (both sync and async versions) that verifies a credential works by requesting a token for a given Azure scope. It should print success/failure, the token's expiry time, and whether the token is CAE-enabled. It should handle and report the specific failure reason if authentication fails (expired cert, wrong tenant, no identity available, etc.) rather than just printing a generic error.

- A **Main class** that detects the current environment, builds the right credential, and runs the connectivity test against Azure Resource Manager using the sync tester first, then repeats with the async tester. Print the detected environment, the selected credential strategy, and the test results from both.

Include a complete `pom.xml` with the necessary Azure SDK dependencies.

## Evaluation Criteria

### Credential Chain Construction
- Uses `ChainedTokenCredentialBuilder` to compose multiple credentials
- Credentials added via `.addLast()` — order matters

### Environment-Specific Chains
- **Dev chain**: includes `AzureCliCredential`; may include `IntelliJCredential`, `VisualStudioCodeCredential`, `AzurePowerShellCredential`
- **CI chain**: uses `EnvironmentCredential` or `AzurePipelinesCredential` (not just `DefaultAzureCredential`)
- **Production chain**: `ManagedIdentityCredential` first (supports user-assigned via `clientId()`), `WorkloadIdentityCredential` as fallback

### CAE Support
- Enables CAE via `TokenRequestContext.setCaeEnabled(true)` or `enableCae()` on credential builders

### Environment Detection
- Detects CI (checks `CI`, `TF_BUILD`, `AZURE_PIPELINE_WORKSPACE`, or similar)
- Detects production/managed identity (checks `IDENTITY_ENDPOINT`, `MSI_ENDPOINT`, or similar)
- Falls back to dev if neither detected

### Token Request & Testing
- Creates `TokenRequestContext` with correct scope (`https://management.azure.com/.default`)
- Calls `getToken()` and prints token expiry from `AccessToken.getExpiresAt()`
- Handles failure with specific exception info

### Scenario-Specific Async
- Async tester uses reactive `getToken()` returning `Mono<AccessToken>`

### Anti-Patterns (scenario-specific)
- NOT using `DefaultAzureCredential` as the CI credential (too broad)

## Context

This goes beyond basic DefaultAzureCredential usage (covered by `default-azure-credential.prompt.md`)
to test whether the agent can build targeted credential chains for each deployment environment.
Key differentiators: ChainedTokenCredentialBuilder (not just DefaultAzureCredential),
AzurePipelinesCredential for CI, user-assigned managed identity, WorkloadIdentityCredential for
Kubernetes, and Continuous Access Evaluation (CAE) — a recent feature that many LLMs don't know
about without skill augmentation.
