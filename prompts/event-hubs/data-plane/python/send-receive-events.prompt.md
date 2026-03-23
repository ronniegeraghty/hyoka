---
id: event-hubs-dp-python-streaming
service: event-hubs
plane: data-plane
language: python
category: streaming
difficulty: intermediate
description: >
  Can a developer send and receive events using Azure Event Hubs
  with the Python SDK documentation?
sdk_package: azure-eventhub
doc_url: https://learn.microsoft.com/en-us/azure/event-hubs/event-hubs-python-get-started-send
tags:
  - event-hubs
  - streaming
  - producer
  - consumer
created: 2025-07-28
author: ronniegeraghty
---

# Send and Receive Events: Azure Event Hubs (Python)

## Prompt

Using only the Azure SDK for Python documentation, write a Python script that
demonstrates sending and receiving events with Azure Event Hubs:
1. Create an EventHubProducerClient using from_connection_string()
2. Create an EventDataBatch and add 10 events with custom properties
3. Send the batch to the event hub
4. Create an EventHubConsumerClient for receiving
5. Create a BlobCheckpointStore for checkpointing
6. Define an on_event callback that prints event body and updates checkpoint
7. Start receiving with receive() or receive_batch() using the callback
8. Handle errors with an on_error callback

Show required pip packages (azure-eventhub and
azure-eventhub-checkpointstoreblob-aio) and async patterns.

## Evaluation Criteria

The documentation should cover:
- `azure-eventhub` and `azure-eventhub-checkpointstoreblob-aio` pip packages
- `EventHubProducerClient.from_connection_string()`
- `create_batch()` and `EventDataBatch.add()`
- `send_batch()` for publishing
- `EventHubConsumerClient` with `BlobCheckpointStore`
- `receive()` with `on_event` and `on_error` callbacks
- Async variants with `aio` module
- Context manager (async with) patterns

## Context

The Python Event Hubs SDK supports both sync and async patterns. This tests
whether the Python docs cover the async-first approach with proper checkpoint
store integration and the callback-based consumer model.
