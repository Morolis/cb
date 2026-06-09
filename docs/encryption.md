# Encryption Design

## Overview

ClipBoard supports client-side encryption using AES-256-GCM. When enabled, plaintext never reaches the server -- only ciphertext is stored.

## Key Derivation

A master password is expanded into a 256-bit key via PBKDF2:

| Parameter  | Value            |
|------------|------------------|
| Algorithm  | PBKDF2-HMAC-SHA256 |
| Iterations | 100,000          |
| Salt       | User ID          |
| Output     | 256-bit key      |

Using the user ID as salt ties the derived key to the account. Two users with the same master password will produce different keys.

## Encryption

| Parameter | Value              |
|-----------|--------------------|
| Algorithm | AES-256-GCM        |
| Nonce     | 12 random bytes, generated per encryption |
| Tag       | Appended to ciphertext (standard GCM tag) |

The nonce is prepended to the ciphertext before storage. Both are required for decryption.

## Master Password Sources

The master password is resolved in this order:

1. **`CB_MASTER_PASS` environment variable** -- use for CI pipelines and automation where interactive prompts are not possible.
2. **Interactive prompt** (default) -- the CLI prompts for the password at runtime.

## Usage

Add the `--encrypt` flag to any of these commands:

```
cb send --encrypt "some secret content"
cb save --encrypt snippet.json
cb stash --encrypt
```

When fetching an encrypted snippet, the CLI prompts for the master password (or reads `CB_MASTER_PASS`) and decrypts locally.

## Threat Model

- **Server compromise**: The server only stores ciphertext. Without the master password, snippet content is unreadable.
- **Transport**: Encryption is independent of TLS. Even over an unencrypted connection, the payload is protected. TLS is still recommended.
- **Key loss**: If the master password is lost, encrypted snippets cannot be recovered. There is no key escrow.
