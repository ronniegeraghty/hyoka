---
id: storage-dp-go-crud
service: storage
plane: data-plane
language: go
category: crud
difficulty: basic
description: >
  Can a developer upload, download, list, and delete blobs in Azure Blob Storage
  using the Go SDK?
sdk_package: azblob
doc_url: https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/storage/azblob
tags:
  - blob
  - crud
  - getting-started
created: 2025-07-27
author: ronniegeraghty
---

# CRUD Blobs: Azure Blob Storage (Go)

## Prompt

Write a Go program that performs
CRUD operations on Azure Blob Storage:
1. Create a service client using DefaultAzureCredential from azidentity
2. Create a container named "my-container"
3. Upload a byte slice with content "Hello from Go" as a blob named "hello.txt"
4. List all blobs in the container using a pager
5. Download the blob and print its content
6. Delete the blob and then delete the container

Show required Go modules and proper error handling using ResponseError.

## Evaluation Criteria

- `go get` for `github.com/Azure/azure-sdk-for-go/sdk/storage/azblob` and `azidentity`
- `azblob.NewClient()` with `azidentity.NewDefaultAzureCredential()`
- `Client.CreateContainer()` and container operations
- `Client.UploadBuffer()` or `UploadStream()`
- `Client.NewListBlobsFlatPager()` with pager iteration pattern
- `Client.DownloadStream()` and reading the response body
- `Client.DeleteBlob()` and `Client.DeleteContainer()`
- `*azcore.ResponseError` type assertion for error details

## Context

Go is increasingly popular for cloud-native Azure applications. The Go SDK
uses a different pattern (pagers, streaming) than other languages, making it
important to test whether the generated code clearly demonstrates Go-idiomatic blob operations.
