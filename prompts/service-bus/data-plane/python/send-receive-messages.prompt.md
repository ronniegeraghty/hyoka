---
id: service-bus-dp-python-crud
service: service-bus
plane: data-plane
language: python
category: crud
difficulty: intermediate
description: >
  Can a developer send and receive messages using Azure Service Bus
  queues and topics with the Python SDK documentation?
sdk_package: azure-servicebus
doc_url: https://learn.microsoft.com/en-us/azure/service-bus-messaging/service-bus-python-how-to-use-queues
tags:
  - service-bus
  - messaging
  - queues
  - topics
created: 2025-07-28
author: ronniegeraghty
---

# Send and Receive Messages: Azure Service Bus (Python)

## Prompt

Using only the Azure SDK for Python documentation, write a Python script that
demonstrates messaging with Azure Service Bus:
1. Create a ServiceBusClient using from_connection_string()
2. Get a sender for a queue and send a single ServiceBusMessage
3. Send a batch of 5 messages using a ServiceBusMessageBatch
4. Get a receiver for the queue and receive messages
5. Complete a message with receiver.complete_message() after processing
6. Demonstrate the async pattern using aio module for higher throughput
7. Send to a topic and receive from a subscription

Show required pip packages and proper context manager patterns (with statements).

## Evaluation Criteria

The documentation should cover:
- `azure-servicebus` pip package
- `ServiceBusClient.from_connection_string()`
- `ServiceBusSender` via `get_queue_sender()` or `get_topic_sender()`
- `ServiceBusMessage` and `ServiceBusMessageBatch`
- `ServiceBusReceiver` via `get_queue_receiver()` or `get_subscription_receiver()`
- `complete_message()`, `abandon_message()`, `dead_letter_message()`
- Context manager pattern (`with` statements) for resource cleanup
- Async variants in `azure.servicebus.aio`

## Context

The Python Service Bus SDK supports both sync and async patterns with context managers.
This tests whether the Python docs cover both patterns and the queue vs.
topic/subscription distinction.
