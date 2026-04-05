---
id: storage-dp-java-blob-event-notifier
properties:
  service: storage
  plane: data-plane
  language: java
  category: streaming
  difficulty: intermediate
  description: 'Can an agent generate a Blob Storage event processor using Event Grid, supporting both EventGridEvent and
    CloudEvents 1.0 schemas, event routing by type, blob subject parsing, custom event publishing, and race condition handling?

    '
  sdk_package: com.azure:azure-messaging-eventgrid
  doc_url: https://learn.microsoft.com/en-us/java/api/overview/azure/messaging-eventgrid-readme
  created: '2026-03-25'
  author: JonathanGiles, samvaity
tags:
- event-grid
- blob-storage
- cloud-events
- event-routing
- async
- reactor
- multi-service
---

# Blob Event Notifier: Azure Event Grid + Blob Storage (Java)

## Prompt

Create a small Java 17 Maven project that processes Azure Blob Storage lifecycle events delivered via Event Grid.

The project needs:

- An **event receiver class** (both sync and async versions) that accepts a JSON payload (as if received from an Event Grid webhook endpoint) and deserializes it into structured event objects. It should support both Event Grid schema and CloudEvents 1.0 schema (since Event Grid supports both and the configured schema may vary). It should route events by type — blob-created events get processed one way, blob-deleted another, and unrecognized types are logged as warnings.

- A **blob event handler class** that processes individual blob events. For blob-created events, it should parse the blob's container and name from the event subject, download the blob, and print a summary (name, size, content type, and the blob's access tier). For blob-deleted events, it should just log the deletion. It should handle race conditions gracefully — the blob may have already been deleted or moved to a different tier by the time we try to read it.

- An **event publisher class** (both sync and async versions) that can publish custom events to an Event Grid topic. Given a topic endpoint and a list of custom event objects, it should send them to Event Grid. This would be used for downstream notifications (e.g., "document processed" events). It should support setting a subject hierarchy for event filtering (e.g., "/documents/invoices/processed").

- A **configuration class** that connects to Azure Blob Storage and Event Grid securely. Authentication should use managed identity — no access keys or SAS tokens.

- A **Main class** that demos both implementations: constructs a sample Event Grid JSON payload (with both CloudEvents and EventGrid-schema examples) containing mock blob-created and blob-deleted events with realistic structure, feeds them through the receiver and handler, and publishes a custom downstream event. Run the full demo with the sync implementation first, then repeat with the async implementation.

Include a complete `pom.xml` with the necessary Azure SDK dependencies.

## Evaluation Criteria

### Dependencies
- Uses `com.azure:azure-messaging-eventgrid`
- Uses `com.azure:azure-storage-blob`
- Uses `com.azure:azure-identity`
- No `com.microsoft.azure` groupId anywhere
- Specifies Java 17

### Authentication
- Uses `DefaultAzureCredential` — no access keys, connection strings, or SAS tokens
- Reads endpoints from environment variables

### Client Construction
- Uses `BlobServiceClientBuilder` for Blob Storage
- Uses `EventGridPublisherClientBuilder` for Event Grid publishing
- Both builders use `.endpoint()` and `.credential()`

### SDK Patterns
- Handles Event Grid native schema via `EventGridEvent.fromString()` deserialization
- Handles CloudEvents 1.0 schema via `CloudEvent.fromString()` deserialization
- Does NOT manually parse JSON without the SDK's deserialization helpers
- Routes events based on event type string (`Microsoft.Storage.BlobCreated`, `Microsoft.Storage.BlobDeleted`)
- Logs a warning for unrecognized event types
- Parses container name and blob name from event subject (`/blobServices/default/containers/{container}/blobs/{blob}`)
- Publishes custom events with subject hierarchy for filtering
- Retrieves and prints blob access tier from blob properties

### Error Handling
- Handles race condition: blob may no longer exist (catches `BlobStorageException` with 404 status)
- Catches Event Grid-specific exceptions for publishing errors
- Does not use bare `Exception` catches

### Async Quality
- Uses `BlobAsyncClient` and `EventGridPublisherAsyncClient`
- Uses Project Reactor types (`Mono`, `Flux`)
- Does not call `.block()` inside the async implementation

### Anti-Patterns (should NOT appear)
- `CloudStorageAccount` or `CloudBlobClient` (deprecated v8 API)
- Fabricated Event Grid classes that don't exist in the SDK
- `com.microsoft.azure.*` imports

## Context

This is a multi-service scenario testing Event Grid integration with Blob Storage. It exercises
the agent's knowledge of Event Grid's dual-schema support (EventGridEvent vs CloudEvents 1.0),
event deserialization using SDK helpers (not manual JSON parsing), event routing by type, blob
subject parsing, and the common race condition where a blob is deleted between the event firing
and the handler processing it. The event publishing side tests custom event creation with
subject hierarchies for downstream filtering.
