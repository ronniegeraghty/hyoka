---
id: service-bus-dp-java-order-processor
service: service-bus
plane: data-plane
language: java
category: streaming
difficulty: intermediate
description: >
  Can an agent generate a Service Bus order processing system with batch sending,
  scheduled delivery, dead-letter queue handling, session-aware processing, and
  proper error categorization?
sdk_package: com.azure:azure-messaging-servicebus
doc_url: https://learn.microsoft.com/en-us/java/api/overview/azure/messaging-servicebus-readme
tags:
  - service-bus
  - batch-sending
  - scheduled-delivery
  - dead-letter-queue
  - session-aware
  - correlation
  - async
  - reactor
created: 2026-03-25
author: JonathanGiles, samvaity
---

# Order Processor: Azure Service Bus (Java)

## Prompt

Create a small Java 17 Maven project that implements an order processing system using Azure Service Bus.

The project needs:

- A **model class** for an Order with fields for order ID, customer name, product, quantity, total price, and status (pending/processing/completed/failed). It should be serializable to and from JSON.

- A **sender class** (both sync and async versions) that publishes order messages to a Service Bus queue. It should support sending individual orders and sending a batch of orders efficiently (respecting the maximum batch size so messages aren't rejected). Each message should carry the order ID as a correlation property, and orders above a certain dollar threshold should be sent as high-priority with a scheduled delivery delay of 30 seconds (to allow for fraud review before processing).

- A **processor class** (both sync and async versions) that receives and processes orders from the queue. It should handle messages as they arrive, deserialize them, and log the results. If processing fails (e.g., a deserialization error), the message should be sent to the dead-letter queue with a reason string rather than being silently abandoned. The processor should also be able to read from the dead-letter queue so failed messages can be inspected and reprocessed. It should guarantee that orders from the same customer are processed in sequence, not interleaved with other customers' orders.

- A **Main class** that demos both implementations: connects to the Service Bus namespace (from an environment variable) with managed identity, runs the full send/receive/dead-letter cycle using the sync implementation first, then repeats with the async implementation.

Include a complete `pom.xml` with the necessary Azure SDK dependencies.

## Evaluation Criteria

### Scenario-Specific Client Construction
- Sender uses `.sender().queueName().buildClient()` chain (or async equivalent)
- Processor uses `.processor().queueName().processMessage().processError()` chain

### Scenario-Specific Patterns
- Batch sending: creates `ServiceBusMessageBatch`, checks `tryAddMessage()` return value
- Handles the case where a message doesn't fit in the current batch
- Scheduled delivery: uses `scheduleMessage()` or `setScheduledEnqueueTime()` (~30s delay)
- Correlation: sets order ID as correlation property via `setCorrelationId()` or application properties
- Dead-letter: explicitly dead-letters failed messages with `deadLetter()` and a reason string
- Dead-letter queue reading: uses `SubQueue.DEAD_LETTER_QUEUE` or `$deadletterqueue` path
- Session-aware processing: uses `.sessionProcessor()` or session-enabled receiver
- Session ID keyed by customer name for ordered processing

### Scenario-Specific Error Handling
- Error handler in processor logs entity path and error source
- Distinguishes transient vs non-transient errors via `isTransient()` or `getReason()`

## Context

This goes beyond basic send/receive (covered by `send-receive-messages.prompt.md`) to test
production messaging patterns: batch sending with size-check guards, scheduled delivery for
fraud review workflows, dead-letter queue handling with reason strings for message forensics,
and session-aware processing to guarantee ordered delivery per customer. The session-aware
pattern is particularly important — without it, concurrent message processing can interleave
orders from the same customer, leading to race conditions in order fulfillment.
