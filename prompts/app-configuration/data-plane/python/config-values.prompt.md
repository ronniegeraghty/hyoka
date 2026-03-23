---
id: app-configuration-dp-python-crud
service: app-configuration
plane: data-plane
language: python
category: crud
difficulty: basic
description: >
  Can a developer read and write configuration values and feature flags
  in Azure App Configuration using the Python SDK documentation?
sdk_package: azure-appconfiguration
doc_url: https://learn.microsoft.com/en-us/azure/azure-app-configuration/quickstart-python
tags:
  - app-configuration
  - configuration
  - feature-flags
  - crud
created: 2025-07-28
author: ronniegeraghty
---

# Configuration Values: Azure App Configuration (Python)

## Prompt

Using only the Azure SDK for Python documentation, write a Python script that manages
configuration settings in Azure App Configuration:
1. Create an AzureAppConfigurationClient using from_connection_string()
2. Set a configuration setting with key "app:Settings:FontSize" and value "24"
3. Set a setting with label "Production"
4. Get the setting by key and print its value
5. List all settings matching the key filter "app:Settings:*"
6. Create a FeatureFlagConfigurationSetting for "BetaFeature" that is enabled
7. Delete the setting by key

Show required pip packages and proper error handling with HttpResponseError.

## Evaluation Criteria

The documentation should cover:
- `azure-appconfiguration` pip package
- `AzureAppConfigurationClient.from_connection_string()`
- `set_configuration_setting()` with `ConfigurationSetting` objects
- `get_configuration_setting()` by key and label
- `list_configuration_settings()` with key_filter and label_filter
- `FeatureFlagConfigurationSetting` for feature flags
- `delete_configuration_setting()` and `HttpResponseError` handling

## Context

The Python App Configuration SDK uses a factory method pattern for client creation.
This tests whether the Python docs cover the configuration setting model including
labels and the FeatureFlagConfigurationSetting class.
