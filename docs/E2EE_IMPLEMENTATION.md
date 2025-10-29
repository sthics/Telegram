# End-to-End Encryption (Optional)

## X3DH Flow
1. Client generates identity key (IK), signed pre-key (SPK), one-time pre-key (OPK)
2. Uploads bundle to `PUT /v1/keys/bundle`
3. Sender fetches bundle, runs X3DH, derives SK
4. SK feeds Double-Ratchet

## Double-Ratchet
- Header = ephemeral public key + previous chain length
- Message encrypted with AES-256-GCM, HKDF chain key rotation
- Server sees only ciphertext blob

## Protobuf Envelope
```proto
message CipherMessage {
  bytes header  = 1; // ratchet header
  bytes cipher  = 2; // ciphertext
  bytes mac     = 3; // optional HMAC
}
```

## Key Storage
- Private keys in device keychain (iOS) / Keystore (Android)
- Server stores public bundles, no private material

## Backward Compatibility
- Per-chat toggle "Secret mode"; fallback to plaintext if other side unsupported
