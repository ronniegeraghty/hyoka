---
id: service-bus-dp-dotnet-crud
service: service-bus
plane: data-plane
language: dotnet
category: crud
difficulty: intermediate
description: >
  Can a developer send and receive messages using Azure Service Bus
  queues and topics with the .NET SDK?
sdk_package: Azure.Messaging.ServiceBus
doc_url: https://learn.microsoft.com/en-us/dotnet/api/overview/azure/messaging.servicebus-readme
tags:
  - service-bus
  - messaging
  - queues
  - topics
created: 2025-07-28
author: ronniegeraghty
---

# Send and Receive Messages: Azure Service Bus (.NET)

## Prompt

Write a C# program that demonstrates
messaging with Azure Service Bus:
1. Create a ServiceBusClient using a connection string
2. Create a ServiceBusSender for a queue and send a single message
3. Send a batch of 5 messages using ServiceBusMessageBatch
4. Create a ServiceBusReceiver and receive messages using ReceiveMessagesAsync
5. Complete a message after processing with CompleteMessageAsync
6. Create a ServiceBusProcessor for continuous processing with handlers
7. Demonstrate sending to a topic and receiving from a subscription

Show required NuGet packages and proper disposal with await using.

## Evaluation Criteria

The generated code should include:
- `Azure.Messaging.ServiceBus` NuGet package
- `ServiceBusClient` creation with connection string or `DefaultAzureCredential`
- `ServiceBusSender` and `ServiceBusMessage` for sending
- `ServiceBusMessageBatch` and `TryAddMessage()`
- `ServiceBusReceiver` and `ReceiveMessagesAsync()`
- `CompleteMessageAsync()`, `AbandonMessageAsync()`, `DeadLetterMessageAsync()`
- `ServiceBusProcessor` with `ProcessMessageAsync` and `ProcessErrorAsync`
- Topic/subscription with `CreateSender(topicName)` and `CreateReceiver(topicName, subscriptionName)`

## Context

Service Bus is Azure's enterprise messaging service supporting queues and pub/sub topics.
This tests both pull-based receiving and the processor pattern, plus the queue vs.
topic distinction that is fundamental to Service Bus architecture.
