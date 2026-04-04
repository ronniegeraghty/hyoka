---
id: storage-dp-java-encrypted-uploader
properties:
  service: storage
  plane: data-plane
  language: java
  category: crud
  difficulty: advanced
  description: 'Can an agent implement client-side envelope encryption for Azure Blob Storage using Key Vault Keys for key
    wrapping, with AES-GCM local encryption, wrapped DEK stored as blob metadata, and proper key material lifecycle?

    '
  sdk_package: com.azure:azure-storage-blob
  doc_url: https://learn.microsoft.com/en-us/java/api/overview/azure/storage-blob-readme
  created: '2026-03-25'
  author: JonathanGiles, samvaity
tags:
- blob-storage
- key-vault
- encryption
- envelope-encryption
- aes-gcm
- key-wrap
- cryptography-client
- multi-service
- async
- reactor
---

# Encrypted Uploader: Azure Blob Storage + Key Vault Keys (Java)

## Prompt

Create a small Java 17 Maven project that uploads files to Azure Blob Storage with client-side encryption, where the encryption key material is managed in Azure Key Vault.

The project needs:

- A **key management class** (both sync and async versions) that interacts with Azure Key Vault's Keys service (not Secrets) to perform cryptographic operations. It should implement envelope encryption: generate a data encryption key locally, use Key Vault to protect (wrap) it, and store the protected key alongside the encrypted blob. For decryption, have Key Vault recover (unwrap) the data key, then decrypt locally. The raw data key should never be persisted anywhere, and the vault's key material should never leave Key Vault.

- A **blob uploader/downloader class** (both sync and async versions) that handles the actual encryption and storage. For upload: generate a data key, encrypt the data locally, protect the data key via Key Vault, then upload the ciphertext to Blob Storage with the protected key and any necessary cryptographic parameters stored as blob metadata. For download: read the blob and its metadata, recover the data key via Key Vault, and decrypt. Should handle errors from both services (e.g., the vault key may have been disabled, or the blob may not exist).

- A **configuration class** that builds the necessary Azure connections for both Blob Storage and Key Vault. It should read endpoints from environment variables and authenticate with managed identity. All connections should share a single credential instance.

- A **Main class** that demos both implementations: runs the full encrypt-upload-download-decrypt round-trip using the sync implementation first, then repeats with the async implementation. Print the vault key ID used, the wrapped DEK (base64), and the decrypted output.

Include a complete `pom.xml` with the necessary Azure SDK dependencies.

## Evaluation Criteria

### Dependencies
- Uses `com.azure:azure-storage-blob`
- Uses `com.azure:azure-security-keyvault-keys` (Keys, NOT Secrets)
- Uses `com.azure:azure-identity`
- No `com.microsoft.azure` groupId anywhere
- Specifies Java 17
- Uses `javax.crypto` or `java.security` for local AES-GCM encryption

### Authentication
- Uses `DefaultAzureCredential` shared between Blob Storage and Key Vault clients
- No hardcoded keys, connection strings, or SAS tokens
- Reads endpoints from environment variables

### Client Construction
- Uses `BlobServiceClientBuilder` for Blob Storage
- Uses `KeyClient` / `CryptographyClient` builder for Key Vault Keys (NOT `SecretClient`)
- Both use `.endpoint()` / `.vaultUrl()` and `.credential()`

### SDK Patterns — Key Vault Keys (critical)
- Uses Key Vault **Keys** service, NOT Secrets
- Uses `CryptographyClient` for `wrapKey()` and `unwrapKey()` operations
- Specifies RSA key wrap algorithm (`KeyWrapAlgorithm.RSA_OAEP` or `RSA_OAEP_256`)
- Key material never leaves Key Vault (wrap/unwrap is server-side)

### SDK Patterns — Envelope Encryption (critical)
- Generates a random AES-256 DEK locally (32 bytes)
- Encrypts data with AES-GCM locally using the DEK
- Wraps the DEK via Key Vault `wrapKey()`
- Stores wrapped DEK as blob metadata
- Stores IV (initialization vector) in blob metadata
- Stores vault key identifier in blob metadata
- For decryption: retrieves wrapped DEK from metadata, unwraps via Key Vault, decrypts locally

### AES-GCM
- Uses AES-GCM (not AES-CBC, AES-ECB, or other modes)
- Generates random IV for each encryption (typically 12 bytes for GCM)

### Error Handling
- Handles `BlobStorageException` for blob errors
- Handles Key Vault errors (key disabled, key not found)
- Catches specific exceptions rather than generic `Exception`

### Async Quality
- Uses `BlobAsyncClient` and `CryptographyAsyncClient`
- Uses Project Reactor types (`Mono`, `Flux`)
- Does not call `.block()` inside the async implementation

### Anti-Patterns (should NOT appear)
- Using `SecretClient` instead of `KeyClient`/`CryptographyClient`
- Encrypting data directly with the vault key (should be envelope encryption)
- Storing raw DEK in plaintext
- AES-CBC or AES-ECB mode
- `com.microsoft.azure.*` imports

## Context

This is the most advanced scenario, combining two Azure services (Blob Storage + Key Vault)
with client-side cryptography. It tests a critical security pattern: envelope encryption where
a locally-generated data encryption key (DEK) is used to encrypt data with AES-GCM, then the
DEK itself is protected by wrapping it with an RSA key managed in Key Vault. The vault's key
material never leaves the HSM boundary. LLMs frequently make two mistakes here: using Key Vault
Secrets instead of Keys (wrong service), and encrypting data directly with the vault key instead
of implementing the DEK/KEK envelope pattern.
