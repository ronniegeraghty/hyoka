---
id: app-configuration-dp-dotnet-crud
service: app-configuration
plane: data-plane
language: dotnet
category: crud
difficulty: basic
description: >
  Can a developer read and write configuration values and feature flags
  in Azure App Configuration using the .NET SDK documentation?
sdk_package: Azure.Data.AppConfiguration
doc_url: https://learn.microsoft.com/en-us/azure/azure-app-configuration/quickstart-dotnet-core-app
tags:
  - app-configuration
  - configuration
  - feature-flags
  - crud
created: 2025-07-28
author: ronniegeraghty
---

# Configuration Values: Azure App Configuration (.NET)

## Prompt

Using only the Azure SDK for .NET documentation, write a C# program that manages
configuration settings in Azure App Configuration:
1. Create a ConfigurationClient using a connection string
2. Set a configuration setting with key "app:Settings:FontSize" and value "24"
3. Set a configuration setting with a label "Production"
4. Get the setting by key and print its value
5. List all settings with the prefix "app:Settings:" using GetConfigurationSettings
6. Create a feature flag setting for "BetaFeature" that is enabled
7. Delete the setting

Show required NuGet packages and proper error handling with RequestFailedException.

## Evaluation Criteria

The documentation should cover:
- `Azure.Data.AppConfiguration` NuGet package
- `ConfigurationClient` creation with connection string or `DefaultAzureCredential`
- `SetConfigurationSetting()` with key, value, and optional label
- `GetConfigurationSetting()` by key and label
- `GetConfigurationSettings()` with `SettingSelector` for filtering
- Feature flag configuration settings with `FeatureFlagConfigurationSetting`
- `DeleteConfigurationSetting()` and `RequestFailedException` handling

## Context

App Configuration centralizes application settings. This tests whether the .NET
docs cover the key-value model including labels and the feature flag pattern,
which are the two primary usage scenarios.
