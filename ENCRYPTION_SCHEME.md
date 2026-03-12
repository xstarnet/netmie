# V2Ray Config Encryption Scheme

## Overview

This document describes the encryption scheme used for V2Ray configuration distribution.

**Encryption Algorithm:** AES-256-GCM (Galois/Counter Mode)

## Why AES-256-GCM?

- **AES-256**: Military-grade symmetric encryption (256-bit key)
- **GCM Mode**: Provides both confidentiality and authenticity
  - Detects tampering (authentication tag)
  - Prevents replay attacks (random nonce)
  - Authenticated encryption with associated data (AEAD)

## Technical Details

### Key Specifications
- **Key Size**: 256 bits (32 bytes)
- **Key Format**: Base64-encoded string
- **Nonce Size**: 12 bytes (96 bits) - randomly generated per encryption
- **Authentication Tag**: 16 bytes (128 bits)

### Encryption Process (Server Side)

```
Input: plaintext config (JSON)
  ↓
1. Generate random 12-byte nonce
2. Create AES-256 cipher with key
3. Create GCM mode from cipher
4. Encrypt plaintext with nonce
5. GCM automatically appends authentication tag
6. Prepend nonce to ciphertext: [nonce (12 bytes)] + [ciphertext + tag]
7. Base64 encode the result
  ↓
Output: base64-encoded encrypted data
```

### Decryption Process (Client Side)

```
Input: base64-encoded encrypted data
  ↓
1. Base64 decode
2. Extract nonce (first 12 bytes)
3. Extract ciphertext + tag (remaining bytes)
4. Create AES-256 cipher with key
5. Create GCM mode from cipher
6. Decrypt and verify authentication tag
7. If tag verification fails, decryption fails
  ↓
Output: plaintext config (JSON)
```

## Implementation

### Go Implementation (Client & Server)

Both client and server use the same `util/crypt` package:

```go
import "github.com/netbirdio/netbird/util/crypt"

// Server: Encrypt config
cipher, err := crypt.NewFieldEncrypt(base64Key)
encrypted, err := cipher.Encrypt(jsonConfigString)
// encrypted is base64-encoded and ready for HTTP transmission

// Client: Decrypt config
cipher, err := crypt.NewFieldEncrypt(base64Key)
decrypted, err := cipher.Decrypt(base64EncodedString)
// decrypted is the original JSON config
```

### Key Generation

```go
import "github.com/netbirdio/netbird/util/crypt"

key, err := crypt.GenerateKey()
// key is a base64-encoded 32-byte random key
```

## Server Implementation Example

### Python Example

```python
import base64
import os
from cryptography.hazmat.primitives.ciphers.aead import AESGCM
from cryptography.hazmat.primitives import hashes
from cryptography.hazmat.backends import default_backend

def encrypt_config(config_json: str, base64_key: str) -> str:
    """Encrypt config using AES-256-GCM"""
    # Decode base64 key
    key = base64.b64decode(base64_key)
    assert len(key) == 32, "Key must be 32 bytes"

    # Generate random nonce
    nonce = os.urandom(12)

    # Create cipher and encrypt
    cipher = AESGCM(key)
    ciphertext = cipher.encrypt(nonce, config_json.encode(), None)

    # Combine nonce + ciphertext and base64 encode
    encrypted_data = nonce + ciphertext
    return base64.b64encode(encrypted_data).decode()

def decrypt_config(encrypted_b64: str, base64_key: str) -> str:
    """Decrypt config using AES-256-GCM"""
    # Decode base64 key
    key = base64.b64decode(base64_key)
    assert len(key) == 32, "Key must be 32 bytes"

    # Decode encrypted data
    encrypted_data = base64.b64decode(encrypted_b64)

    # Extract nonce and ciphertext
    nonce = encrypted_data[:12]
    ciphertext = encrypted_data[12:]

    # Create cipher and decrypt
    cipher = AESGCM(key)
    plaintext = cipher.decrypt(nonce, ciphertext, None)

    return plaintext.decode()

# Usage
key = "dGVzdGtleWZvcmRlbW9vbmx5cGxlYXNlY2hhbmdlaW5wcm9kdWN0aW9u"
config = '{"inbounds": [...], "outbounds": [...]}'
encrypted = encrypt_config(config, key)
print(encrypted)

# Verify
decrypted = decrypt_config(encrypted, key)
assert decrypted == config
```

### Node.js Example

```javascript
const crypto = require('crypto');

function encryptConfig(configJson, base64Key) {
    // Decode base64 key
    const key = Buffer.from(base64Key, 'base64');
    if (key.length !== 32) throw new Error('Key must be 32 bytes');

    // Generate random nonce
    const nonce = crypto.randomBytes(12);

    // Create cipher and encrypt
    const cipher = crypto.createCipheriv('aes-256-gcm', key, nonce);
    let encrypted = cipher.update(configJson, 'utf8', 'binary');
    encrypted += cipher.final('binary');

    // Get authentication tag
    const authTag = cipher.getAuthTag();

    // Combine nonce + ciphertext + authTag and base64 encode
    const encryptedData = Buffer.concat([nonce, Buffer.from(encrypted, 'binary'), authTag]);
    return encryptedData.toString('base64');
}

function decryptConfig(encryptedB64, base64Key) {
    // Decode base64 key
    const key = Buffer.from(base64Key, 'base64');
    if (key.length !== 32) throw new Error('Key must be 32 bytes');

    // Decode encrypted data
    const encryptedData = Buffer.from(encryptedB64, 'base64');

    // Extract nonce, ciphertext, and authTag
    const nonce = encryptedData.slice(0, 12);
    const authTag = encryptedData.slice(-16);
    const ciphertext = encryptedData.slice(12, -16);

    // Create decipher and decrypt
    const decipher = crypto.createDecipheriv('aes-256-gcm', key, nonce);
    decipher.setAuthTag(authTag);

    let decrypted = decipher.update(ciphertext, 'binary', 'utf8');
    decrypted += decipher.final('utf8');

    return decrypted;
}

// Usage
const key = "dGVzdGtleWZvcmRlbW9vbmx5cGxlYXNlY2hhbmdlaW5wcm9kdWN0aW9u";
const config = '{"inbounds": [...], "outbounds": [...]}';
const encrypted = encryptConfig(config, key);
console.log(encrypted);

// Verify
const decrypted = decryptConfig(encrypted, key);
console.log(decrypted === config);
```

### Java Example

```java
import javax.crypto.Cipher;
import javax.crypto.spec.GCMParameterSpec;
import javax.crypto.spec.SecretKeySpec;
import java.security.SecureRandom;
import java.util.Base64;

public class ConfigEncryption {
    private static final int GCM_IV_LENGTH = 12;
    private static final int GCM_TAG_LENGTH = 128;

    public static String encryptConfig(String configJson, String base64Key) throws Exception {
        byte[] key = Base64.getDecoder().decode(base64Key);
        if (key.length != 32) throw new IllegalArgumentException("Key must be 32 bytes");

        // Generate random nonce
        byte[] nonce = new byte[GCM_IV_LENGTH];
        new SecureRandom().nextBytes(nonce);

        // Create cipher
        Cipher cipher = Cipher.getInstance("AES/GCM/NoPadding");
        GCMParameterSpec spec = new GCMParameterSpec(GCM_TAG_LENGTH, nonce);
        cipher.init(Cipher.ENCRYPT_MODE, new SecretKeySpec(key, 0, key.length, "AES"), spec);

        // Encrypt
        byte[] ciphertext = cipher.doFinal(configJson.getBytes());

        // Combine nonce + ciphertext and base64 encode
        byte[] encryptedData = new byte[nonce.length + ciphertext.length];
        System.arraycopy(nonce, 0, encryptedData, 0, nonce.length);
        System.arraycopy(ciphertext, 0, encryptedData, nonce.length, ciphertext.length);

        return Base64.getEncoder().encodeToString(encryptedData);
    }

    public static String decryptConfig(String encryptedB64, String base64Key) throws Exception {
        byte[] key = Base64.getDecoder().decode(base64Key);
        if (key.length != 32) throw new IllegalArgumentException("Key must be 32 bytes");

        // Decode encrypted data
        byte[] encryptedData = Base64.getDecoder().decode(encryptedB64);

        // Extract nonce and ciphertext
        byte[] nonce = new byte[GCM_IV_LENGTH];
        System.arraycopy(encryptedData, 0, nonce, 0, GCM_IV_LENGTH);
        byte[] ciphertext = new byte[encryptedData.length - GCM_IV_LENGTH];
        System.arraycopy(encryptedData, GCM_IV_LENGTH, ciphertext, 0, ciphertext.length);

        // Create cipher
        Cipher cipher = Cipher.getInstance("AES/GCM/NoPadding");
        GCMParameterSpec spec = new GCMParameterSpec(GCM_TAG_LENGTH, nonce);
        cipher.init(Cipher.DECRYPT_MODE, new SecretKeySpec(key, 0, key.length, "AES"), spec);

        // Decrypt
        byte[] plaintext = cipher.doFinal(ciphertext);
        return new String(plaintext);
    }
}
```

## Security Considerations

1. **Key Management**
   - Generate keys using cryptographically secure random (not predictable)
   - Store keys securely (environment variables, key management systems)
   - Rotate keys periodically
   - Never commit keys to version control

2. **Nonce Handling**
   - Always use random nonce (never reuse with same key)
   - Nonce is included in ciphertext (safe to transmit)
   - 12-byte nonce is standard for GCM

3. **Authentication**
   - GCM provides authentication tag (detects tampering)
   - Decryption fails if ciphertext is modified
   - Always verify authentication before using plaintext

4. **Transport Security**
   - Use HTTPS for config distribution
   - Encryption provides defense-in-depth (not replacement for HTTPS)

## Testing

```bash
# 1. Generate encryption key (one-time)
go run ./tools/generate-key/main.go
# Output: dGVzdGtleWZvcmRlbW9vbmx5cGxlYXNlY2hhbmdlaW5wcm9kdWN0aW9u

# 2. Set up client key
cp client/cmd/encryption_key.txt.example client/cmd/encryption_key.txt
# Edit encryption_key.txt and paste the key

# 3. Rebuild client
go build -o ./netmie ./client

# 4. On server side, encrypt config using the same key
# See Python/Node.js/Java examples above

# 5. Test decryption (client side)
netmie vconfig https://your-server.com/v2ray-config
```

## Key Management

### Initial Setup
1. Generate key: `go run ./tools/generate-key/main.go`
2. Store key securely on server (environment variable, secrets manager, etc.)
3. Copy key to `client/cmd/encryption_key.txt`
4. Rebuild and deploy client binary

### Key Rotation
1. Generate new key
2. Update server to use new key
3. Update `client/cmd/encryption_key.txt` with new key
4. Rebuild and redeploy client binary
5. Old clients will fail to decrypt (expected behavior)

### Important
- **DO NOT** commit `client/cmd/encryption_key.txt` to git
- Use `client/cmd/encryption_key.txt.example` as template
- Add `client/cmd/encryption_key.txt` to `.gitignore` (already done)

## References

- [NIST SP 800-38D: Recommendation for Block Cipher Modes of Operation: Galois/Counter Mode (GCM)](https://nvlpubs.nist.gov/nistpubs/Legacy/SP/nistspecialpublication800-38d.pdf)
- [Go crypto/cipher package](https://pkg.go.dev/crypto/cipher)
- [Python cryptography library](https://cryptography.io/)
- [Node.js crypto module](https://nodejs.org/api/crypto.html)
