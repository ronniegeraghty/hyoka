---
id: event-hubs-dp-java-streaming
properties:
  service: event-hubs
  plane: data-plane
  language: java
  category: streaming
  difficulty: intermediate
  description: 'Can a developer send and receive events using Azure Event Hubs with the Java SDK?

    '
  sdk_package: azure-messaging-eventhubs
  doc_url: https://learn.microsoft.com/en-us/java/api/overview/azure/messaging-eventhubs-readme
  created: '2025-07-28'
  author: ronniegeraghty
tags:
- event-hubs
- streaming
- producer
- consumer
---

# Send and Receive Events: Azure Event Hubs (Java)

## Prompt

Write a Java program that demonstrates
sending and receiving events with Azure Event Hubs:
1. Create an EventHubProducerClient using EventHubClientBuilder with a connection string
2. Create an EventDataBatch and add 10 events with custom properties
3. Send the batch to the event hub
4. Create an EventProcessorClient with BlobCheckpointStore for checkpointing
5. Register processEvent and processError consumer functions
6. Start the processor, wait for events, and print received event bodies
7. Implement checkpointing with EventContext.updateCheckpoint()

Show required Maven dependencies (azure-messaging-eventhubs and
azure-messaging-eventhubs-checkpointstore-blob) and proper resource cleanup.

## Evaluation Criteria

The generated code should include:
- `azure-messaging-eventhubs` and `azure-messaging-eventhubs-checkpointstore-blob` Maven deps
- `EventHubClientBuilder` and `EventHubProducerClient`
- `createBatch()` and `EventDataBatch.tryAdd()`
- `send()` for publishing events
- `EventProcessorClientBuilder` with `BlobCheckpointStore`
- `processEvent` and `processError` consumer functions
- `EventContext.updateCheckpoint()` for reliable processing

## Context

The Java Event Hubs SDK uses a builder pattern and functional-style event handlers.
This tests whether the generated code covers the producer/consumer pattern with checkpoint
store integration using Azure Blob Storage.
