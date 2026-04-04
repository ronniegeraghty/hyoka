---
id: event-hubs-dp-js-ts-streaming
properties:
  service: event-hubs
  plane: data-plane
  language: js-ts
  category: streaming
  difficulty: intermediate
  description: 'Can a developer send and receive events using Azure Event Hubs with the JavaScript/TypeScript SDK?

    '
  sdk_package: '@azure/event-hubs'
  doc_url: https://learn.microsoft.com/en-us/javascript/api/overview/azure/event-hubs-readme
  created: '2025-07-28'
  author: ronniegeraghty
tags:
- event-hubs
- streaming
- producer
- consumer
---

# Send and Receive Events: Azure Event Hubs (JavaScript/TypeScript)

## Prompt

Write a TypeScript program
that demonstrates sending and receiving events with Azure Event Hubs:
1. Create an EventHubProducerClient using a connection string
2. Create a batch with createBatch() and add 10 events with custom properties
3. Send the batch using sendBatch()
4. Create an EventHubConsumerClient with a BlobCheckpointStore
5. Subscribe to events using subscribe() with processEvents and processError handlers
6. Print received event bodies and update checkpoints
7. Implement graceful shutdown with close()

Show required npm packages (@azure/event-hubs and
@azure/eventhubs-checkpointstore-blob) and proper async/await patterns.

## Evaluation Criteria

The generated code should include:
- `@azure/event-hubs` and `@azure/eventhubs-checkpointstore-blob` npm packages
- `EventHubProducerClient` constructor
- `createBatch()` and `EventDataBatch.tryAdd()`
- `sendBatch()` for publishing
- `EventHubConsumerClient` with `BlobCheckpointStore`
- `subscribe()` with `SubscriptionEventHandlers` (processEvents, processError)
- `updateCheckpoint()` in the processEvents handler
- `close()` for cleanup

## Context

The JavaScript Event Hubs SDK uses a subscribe pattern with handler objects.
This tests whether the generated code covers the subscription model and proper
checkpoint integration for reliable event processing.
