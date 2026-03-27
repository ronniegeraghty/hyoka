---
id: storage-dp-java-blob-manager
service: storage
plane: data-plane
language: java
category: crud
difficulty: advanced
description: >
  Can an agent generate a complete, production-ready Azure Blob Storage management
  utility with sync and async implementations, covering upload (large files, index tags),
  download, list, delete, concurrency prevention, retry configuration, and HTTP logging?
sdk_package: com.azure:azure-storage-blob
doc_url: https://learn.microsoft.com/en-us/java/api/overview/azure/storage-blob-readme
tags:
  - identity
  - default-azure-credential
  - blob-storage
  - async
  - reactor
  - retry
  - lease
  - parallel-upload
  - index-tags
created: 2026-03-19
author: JonathanGiles, samvaity
---

# Blob Storage Manager: Azure Blob Storage (Java)

## Prompt

Create a small Java 17 Maven project that provides a reusable Azure Blob Storage management utility.

The project needs:

- A **service class** (both sync and async versions) that wraps blob operations: upload (with optional metadata and blob index tags for later querying), download, list blobs in a container, and delete. The upload method should handle large files efficiently so that uploading a multi-gigabyte file doesn't load the entire thing into memory or fail on slow connections. The service should also prevent concurrent writers from overwriting each other's changes when updating the same blob.

- A **configuration class** that connects to Azure securely using the storage account endpoint (from an environment variable). No connection strings or account keys should be used — the app will run in Azure with managed identity. The configuration should set up a custom retry policy (exponential backoff, configurable max retries and delay) and a per-request timeout, so the app behaves predictably under transient failures. It should also enable HTTP request/response logging at a configurable level for debugging.

- A **Main class** that wires everything together and demos each operation using the sync implementation first, then repeats the same operations using the async implementation: uploads a sample file with some index tags, lists blobs, downloads the file back, acquires a lease and overwrites it, and finally deletes it. Print status at each step.

Include a complete `pom.xml` with the necessary Azure SDK dependencies.

## Evaluation Criteria

### Scenario-Specific Patterns
- Configures custom retry policy (exponential backoff, max retries, delay)
- Sets per-request or per-operation timeout
- Enables HTTP logging (`HttpLogOptions`)
- Implements blob lease acquisition before overwrite (lease-specific API)
- Implements parallel/block upload for large files (`ParallelTransferOptions`, not manual chunking)
- Sets blob index tags on upload (not just metadata) — `Map<String, String>` via upload options
- Properly composes reactive chains in the demo

## Context

This is the most common Azure Storage scenario: a reusable CRUD wrapper. It tests whether
the agent knows the modern v12 Azure SDK patterns (builder pattern, DefaultAzureCredential,
Reactor async, ParallelTransferOptions) vs the deprecated v8 patterns that LLMs frequently
generate. The prompt is intentionally business-level — it says "handle large files efficiently"
not "use ParallelTransferOptions" — so skills must teach the agent the right SDK approach.

Cross-cutting concerns tested: authentication, retry/timeout configuration, HTTP pipeline
logging, async/reactive patterns, blob leasing for concurrency, and blob index tags.
