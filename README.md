# OpenEnvX

**.env enhanced. Secured. Local-first. Agent-ready by design.**


Zero-configuration secret management. Your keys, your data-Vault-style envelope encryption that runs on your machine. Built for developers and AI agents.

Unbiased. No vendor lock-in. No accounts. No cloud required.

Made for people, by people. Fully open source.

> ⚠️ **Early Development**: This tool is currently in early development. Features and APIs may change.


---

## Quick Start

```bash
# Install (macOS and Linux)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/xmazu/openenvx/main/scripts/install.sh)"

# Initialize workspace (creates keypair + encrypts .env files)
openenvx init

# Store a secret
openenvx set DATABASE_URL

# Run with decrypted env
openenvx run -- npm start
```

Or install via Go: `go install github.com/openenvx/cli@latest`

---

## Security Architecture

### Envelope Encryption

Every secret gets its own Data Encryption Key (DEK):

```
Secret Value → Random DEK → AES-256-GCM → Ciphertext
                     │                           │
                     └─ Wrapped by Master Key ←─┘
```

- **Master Key**: 256-bit key derived from age private key via HMAC-SHA256
- **DEK**: 256-bit random key, unique per secret
- **Algorithm**: AES-256-GCM with random nonces
- **Associated Data**: Key names used for authentication (AEAD)

### File Format

```ini
DATABASE_URL=envx:<base64-wrapped-dek>:<base64-ciphertext>
API_KEY=envx:<base64-wrapped-dek>:<base64-ciphertext>
```

### Key Storage

- **Public key**: Stored in `.openenvx.yaml` at workspace root (safe to commit)
- **Private key**: Stored in `~/.config/openenvx/keys.yaml` (never commit)
- **Resolution**: Environment variable `OPENENVX_PRIVATE_KEY` takes precedence

### Audit Logging

All envelope and secret operations logged with hash-chaining:

```bash
openenvx audit show --last=20
openenvx audit verify
```

Stored in `.envx/audit.logl`. Each entry includes a SHA-256 hash of the previous entry for tamper detection.

---

## Security & Scanning

Scan your repository for leaked secrets before they reach git:

```bash
# Scan current directory
openenvx scan

# Scan specific path
openenvx scan --path ./src

# CI-friendly (fails on high-severity findings)
openenvx scan
```

Detects API keys, tokens, passwords, and other sensitive patterns. Respects `.gitignore` exclusions.

---

## FAQ

**Q: Is the `.env` file really safe to commit?**  
A: Yes. It only contains the public key (which is meant to be public) and encrypted ciphertext. Without the private key, it's useless.

**Q: What if I lose my private key?**  
A: Your secrets are gone. There's no backdoor. Store your private key safely (1Password, hardware key, etc.).

---

MIT License · [GitHub](https://github.com/xmazu/openenvx)
