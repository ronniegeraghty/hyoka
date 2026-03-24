---
id: event-hubs-dp-dotnet-streaming
service: event-hubs
plane: data-plane
language: dotnet
category: streaming
difficulty: intermediate
description: >
  Can a developer send and receive events using Azure Event Hubs
  with the .NET SDK?
sdk_package: Azure.Messaging.EventHubs
doc_url: https://learn.microsoft.com/en-us/dotnet/api/overview/azure/messaging.eventhubs-readme
tags:
  - event-hubs
  - streaming
  - producer
  - consumer
created: 2025-07-28
author: ronniegeraghty
---

# Send and Receive Events: Azure Event Hubs (.NET)

## Prompt

Write a C# program that demonstrates
sending and receiving events with Azure Event Hubs:
1. Create an EventHubProducerClient using a connection string
2. Create a batch of events using CreateBatchAsync()
3. Add 10 events with custom properties to the batch
4. Send the batch to the event hub
5. Create an EventProcessorClient with a BlobContainerClient for checkpointing
6. Register ProcessEventAsync and ProcessErrorAsync handlers
7. Start processing events and print received event bodies
8. Implement proper checkpointing with ProcessEventArgs.UpdateCheckpointAsync()

Show required NuGet packages (Azure.Messaging.EventHubs and
Azure.Messaging.EventHubs.Processor) and proper disposal patterns.

## Evaluation Criteria

The generated code should include:
- `Azure.Messaging.EventHubs` and `Azure.Messaging.EventHubs.Processor` NuGet packages
- `EventHubProducerClient` and `EventHubConsumerClient`
- `CreateBatchAsync()` and `EventDataBatch.TryAdd()`
- `SendAsync()` for publishing events
- `EventProcessorClient` with `BlobContainerClient` for checkpointing
- Event handler delegates and `ProcessEventArgs`
- `UpdateCheckpointAsync()` for reliable processing

## Context

Event Hubs is Azure's high-throughput event streaming service. The producer/consumer
pattern with checkpointing is the core usage model. This tests whether the generated code
covers both sides of the pipeline with proper checkpoint storage.
