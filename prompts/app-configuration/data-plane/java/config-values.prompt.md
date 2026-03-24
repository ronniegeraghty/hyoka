---
id: app-configuration-dp-java-crud
service: app-configuration
plane: data-plane
language: java
category: crud
difficulty: basic
description: >
  Can a developer read and write configuration values and feature flags
  in Azure App Configuration using the Java SDK?
sdk_package: azure-data-appconfiguration
doc_url: https://learn.microsoft.com/en-us/java/api/overview/azure/data-appconfiguration-readme
tags:
  - app-configuration
  - configuration
  - feature-flags
  - crud
created: 2025-07-28
author: ronniegeraghty
---

# Configuration Values: Azure App Configuration (Java)

## Prompt

Write a Java program that manages
configuration settings in Azure App Configuration:
1. Create a ConfigurationClient using ConfigurationClientBuilder with a connection string
2. Set a configuration setting with key "app:Settings:FontSize" and value "24"
3. Set a setting with label "Production"
4. Get the setting by key and print its value
5. List settings with a key filter "app:Settings:*" using listConfigurationSettings
6. Create a feature flag configuration setting for "BetaFeature"
7. Delete the setting by key

Show required Maven dependency (com.azure:azure-data-appconfiguration) and
proper error handling with HttpResponseException.

## Evaluation Criteria

The generated code should include:
- `azure-data-appconfiguration` Maven dependency
- `ConfigurationClientBuilder` and `ConfigurationClient`
- `setConfigurationSetting()` with key, value, label
- `getConfigurationSetting()` by key and label
- `listConfigurationSettings()` with `SettingSelector`
- `FeatureFlagConfigurationSetting` for feature flags
- `deleteConfigurationSetting()` and exception handling

## Context

The Java App Configuration SDK uses the Azure SDK builder pattern. This tests
whether the generated code covers the builder setup, labeled settings, and the
feature flag model using FeatureFlagConfigurationSetting.
