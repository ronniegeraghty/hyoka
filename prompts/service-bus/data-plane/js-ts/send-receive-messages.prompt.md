---
id: service-bus-dp-js-ts-crud
service: service-bus
plane: data-plane
language: js-ts
category: crud
difficulty: intermediate
description: >
  Can a developer send and receive messages using Azure Service Bus
  queues and topics with the JavaScript/TypeScript SDK?
sdk_package: "@azure/service-bus"
doc_url: https://learn.microsoft.com/en-us/javascript/api/overview/azure/service-bus-readme
tags:
  - service-bus
  - messaging
  - queues
  - topics
created: 2025-07-28
author: ronniegeraghty
---

# Send and Receive Messages: Azure Service Bus (JavaScript/TypeScript)

## Prompt

Write a TypeScript program
that demonstrates messaging with Azure Service Bus:
1. Create a ServiceBusClient using a connection string
2. Create a sender for a queue and send a single message
3. Send a batch of 5 messages using createMessageBatch() and tryAddMessage()
4. Create a receiver and receive messages using receiveMessages()
5. Complete a message with completeMessage() after processing
6. Subscribe to messages using subscribe() with processMessage and processError handlers
7. Demonstrate sending to a topic and receiving from a subscription

Show required npm package (@azure/service-bus) and proper close() cleanup.

## Evaluation Criteria

The generated code should include:
- `@azure/service-bus` npm package
- `ServiceBusClient` constructor with connection string
- `createSender()` for queue or topic
- `ServiceBusMessageBatch` with `tryAddMessage()`
- `createReceiver()` for queue or subscription
- `receiveMessages()` for batch receive and `subscribe()` for streaming
- `completeMessage()`, `abandonMessage()`, `deadLetterMessage()`
- `close()` on sender, receiver, and client for cleanup

## Context

The JavaScript Service Bus SDK supports both pull-based (receiveMessages) and
push-based (subscribe) receiving patterns. This tests whether the generated code
covers both patterns and proper resource cleanup.
