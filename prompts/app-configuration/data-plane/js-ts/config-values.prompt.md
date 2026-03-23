---
id: app-configuration-dp-js-ts-crud
service: app-configuration
plane: data-plane
language: js-ts
category: crud
difficulty: basic
description: >
  Can a developer read and write configuration values and feature flags
  in Azure App Configuration using the JavaScript/TypeScript SDK documentation?
sdk_package: "@azure/app-configuration"
doc_url: https://learn.microsoft.com/en-us/azure/azure-app-configuration/quickstart-javascript
tags:
  - app-configuration
  - configuration
  - feature-flags
  - crud
created: 2025-07-28
author: ronniegeraghty
---

# Configuration Values: Azure App Configuration (JavaScript/TypeScript)

## Prompt

Using only the Azure SDK for JavaScript documentation, write a TypeScript program
that manages configuration settings in Azure App Configuration:
1. Create an AppConfigurationClient using a connection string
2. Set a configuration setting with key "app:Settings:FontSize" and value "24"
3. Set a setting with label "Production"
4. Get the setting by key and print its value
5. List all settings matching the key filter "app:Settings:*" using listConfigurationSettings
6. Create a feature flag configuration setting for "BetaFeature"
7. Delete the setting by key

Show required npm package (@azure/app-configuration) and
proper error handling with RestError.

## Evaluation Criteria

The documentation should cover:
- `@azure/app-configuration` npm package
- `AppConfigurationClient` constructor with connection string
- `setConfigurationSetting()` with key, value, label
- `getConfigurationSetting()` by key and label
- `listConfigurationSettings()` with `ListConfigurationSettingOptions`
- Feature flag settings with `featureFlagContentType`
- `deleteConfigurationSetting()` and `RestError` handling
- Async iteration with `for await...of` pattern

## Context

The JavaScript App Configuration SDK uses async iterators for listing.
This tests whether the JS/TS docs cover the async iteration pattern and
the feature flag content type approach.
