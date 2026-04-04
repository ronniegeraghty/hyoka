---
id: storage-dp-dotnet-error-handling
properties:
  service: storage
  plane: data-plane
  language: dotnet
  category: error-handling
  difficulty: intermediate
  description: 'Can a developer handle common Azure Blob Storage errors including 404, 403, and 429 responses in .NET?

    '
  sdk_package: Azure.Storage.Blobs
  doc_url: https://learn.microsoft.com/en-us/dotnet/api/overview/azure/storage.blobs-readme
  created: '2025-07-27'
  author: ronniegeraghty
tags:
- error-handling
- exceptions
- retry
---

# Error Handling: Azure Blob Storage (.NET)

## Prompt

How do I properly handle errors when working with Azure Blob Storage in .NET?
I need to understand what exceptions are thrown for common failure scenarios:
container not found (404), access denied (403), and throttling (429).
Show me idiomatic try/catch patterns with the Azure.Storage.Blobs SDK
including how to extract the error code and HTTP status from RequestFailedException.

## Evaluation Criteria

- `RequestFailedException` as the primary exception type
- Extracting `Status` and `ErrorCode` from the exception
- Handling specific HTTP status codes (404, 403, 409, 429)
- Retry policy configuration via `BlobClientOptions`
- Conditional request failures (ETags, leases)
- Logging and diagnostics for troubleshooting

## Context

Error handling is critical for production applications. Developers frequently
struggle with Azure-specific exception types and need clear guidance on
which exceptions to catch and how to extract actionable information from them.
