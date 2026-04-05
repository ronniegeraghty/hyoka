---
id: service-bus-dp-java-crud
properties:
  service: service-bus
  plane: data-plane
  language: java
  category: crud
  difficulty: intermediate
  description: 'Can a developer send and receive messages using Azure Service Bus queues and topics with the Java SDK?

    '
  sdk_package: azure-messaging-servicebus
  doc_url: https://learn.microsoft.com/en-us/java/api/overview/azure/messaging-servicebus-readme
  created: '2025-07-28'
  author: ronniegeraghty
tags:
- service-bus
- messaging
- queues
- topics
---

# Send and Receive Messages: Azure Service Bus (Java)

## Prompt

Write a Java program that demonstrates
messaging with Azure Service Bus:
1. Create a ServiceBusSenderClient using ServiceBusClientBuilder for a queue
2. Send a single message with ServiceBusMessage
3. Send a batch of 5 messages using ServiceBusMessageBatch
4. Create a ServiceBusReceiverClient and receive messages using receiveMessages()
5. Complete a message with complete() after processing
6. Create a ServiceBusProcessorClient for continuous processing with handlers
7. Demonstrate sending to a topic and receiving from a subscription

Show required Maven dependency (com.azure:azure-messaging-servicebus) and
proper resource cleanup with close().

## Evaluation Criteria

The generated code should include:
- `azure-messaging-servicebus` Maven dependency
- `ServiceBusClientBuilder` with connection string
- `ServiceBusSenderClient` and `ServiceBusMessage`
- `createMessageBatch()` and `tryAddMessage()`
- `ServiceBusReceiverClient.receiveMessages()` and `complete()`
- `ServiceBusProcessorClient` with `processMessage` and `processError` handlers
- Topic operations with `.topicName()` and `.subscriptionName()` on the builder

## Context

The Java Service Bus SDK uses the Azure SDK builder pattern with separate clients
for sending and receiving. This tests whether the generated code covers the full
message lifecycle including the processor client for continuous processing.
