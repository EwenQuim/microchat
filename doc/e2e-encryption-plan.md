# End-to-End Encryption Implementation Plan

## Overview

Add optional end-to-end encryption for password-protected rooms. Messages will be encrypted client-side before being sent to the server, ensuring that only users with the room password can decrypt and read messages.

## Architecture

### Key Derivation
- Use the room password to derive an encryption key using PBKDF2 or Argon2
- Separate from the access control password (which is still hashed with bcrypt server-side)
- Derive a 256-bit AES key for symmetric encryption
- Use a room-specific salt stored in room metadata

### Encryption Flow
1. User enters room with password
2. Client derives encryption key from password
3. When sending a message:
   - Generate random nonce/IV
   - Encrypt message content with AES-GCM
   - Send encrypted content + nonce to server
4. When receiving messages:
   - Fetch encrypted content + nonce
   - Decrypt using derived key
   - Display plaintext to user

### Dual Password System
- **Access Password**: Hashed with bcrypt, validated server-side for authorization
- **Encryption Key**: Derived client-side from the same password, never sent to server
- Same password input, different cryptographic purposes

## Implementation Steps

### Phase 1: Data Model Changes

#### Backend (Go)

**1. Update Room Model** (`internal/models/room.go`)
```go
type Room struct {
    Name                 string  `json:"name"`
    MessageCount         int     `json:"message_count"`
    Hidden               bool    `json:"hidden"`
    HasPassword          bool    `json:"has_password"`
    IsEncrypted          bool    `json:"is_encrypted"`      // NEW
    EncryptionSalt       *string `json:"encryption_salt,omitempty"` // NEW - hex encoded
    LastMessageContent   *string `json:"last_message_content,omitempty"`
    LastMessageUser      *string `json:"last_message_user,omitempty"`
    LastMessageTimestamp *string `json:"last_message_timestamp,omitempty"`
}
```

**2. Update Message Model** (`internal/models/message.go`)
```go
type Message struct {
    ID              string    `json:"id"`
    Room            string    `json:"room"`
    User            string    `json:"user"`
    Content         string    `json:"content"`              // Ciphertext if encrypted
    Timestamp       time.Time `json:"timestamp"`
    Signature       string    `json:"signature,omitempty"`
    Pubkey          string    `json:"pubkey,omitempty"`
    SignedTimestamp int64     `json:"signed_timestamp,omitempty"`
    IsEncrypted     bool      `json:"is_encrypted"`         // NEW
    Nonce           string    `json:"nonce,omitempty"`      // NEW - hex encoded
}
```

**3. Database Migration** (`internal/repository/sqlite/migrations/004_add_e2e_encryption.sql`)
```sql
-- Add encryption fields to rooms table
ALTER TABLE rooms ADD COLUMN is_encrypted BOOLEAN DEFAULT FALSE;
ALTER TABLE rooms ADD COLUMN encryption_salt TEXT;

-- Add encryption fields to messages table
ALTER TABLE messages ADD COLUMN is_encrypted BOOLEAN DEFAULT FALSE;
ALTER TABLE messages ADD COLUMN nonce TEXT;
```

**4. Update CreateRoomRequest** (`internal/models/room.go`)
```go
type CreateRoomRequest struct {
    Name        string  `json:"name" validate:"required,min=1,max=50"`
    Password    *string `json:"password,omitempty" validate:"omitempty,min=4,max=72"`
    IsEncrypted bool    `json:"is_encrypted"`  // NEW - enable E2E encryption
}
```

#### Frontend (TypeScript)

**5. Update OpenAPI Schema**
- Regenerate from updated backend API
- Ensure `is_encrypted`, `encryption_salt`, and `nonce` fields are included

### Phase 2: Cryptography Utilities

**6. Create Crypto Utility** (`app/src/lib/crypto/e2e.ts`)
```typescript
// Key derivation
async function deriveEncryptionKey(
  password: string,
  salt: Uint8Array
): Promise<CryptoKey>

// Encryption
async function encryptMessage(
  plaintext: string,
  key: CryptoKey
): Promise<{ ciphertext: string; nonce: string }>

// Decryption
async function decryptMessage(
  ciphertext: string,
  nonce: string,
  key: CryptoKey
): Promise<string>

// Salt generation (for room creation)
function generateSalt(): string
```

**Implementation Details:**
- Use Web Crypto API (SubtleCrypto)
- Algorithm: AES-GCM (authenticated encryption)
- PBKDF2 parameters:
  - Iterations: 100,000 (or Argon2 if available)
  - Hash: SHA-256
  - Key length: 256 bits
- Nonce: 96-bit random value (12 bytes)
- Encoding: Hex strings for storage/transmission

### Phase 3: Frontend Integration

**7. Update Room Creation** (`app/src/components/chat/CreateRoomDialog.tsx`)
- Add checkbox: "Enable end-to-end encryption"
- Only show if password is provided
- Generate salt when E2E encryption is enabled
- Include `is_encrypted` flag in CreateRoomRequest

**8. Update Room Password Hook** (`app/src/hooks/useRoomPassword.ts`)
- Store encryption keys in memory (not sessionStorage)
- Derive key when password is entered
- Clear keys when user leaves room

```typescript
interface RoomKeys {
  [roomName: string]: CryptoKey;
}

// In-memory only (not persisted)
const encryptionKeys = new Map<string, CryptoKey>();

async function deriveAndStoreKey(
  roomName: string,
  password: string,
  salt: string
): Promise<void>

function getEncryptionKey(roomName: string): CryptoKey | undefined
```

**9. Update Send Message Hook** (`app/src/hooks/useSendMessage.ts`)
- Check if room is encrypted
- If encrypted:
  - Get encryption key for room
  - Encrypt message content
  - Send ciphertext + nonce
- If not encrypted:
  - Send plaintext as before

**10. Update Messages Hook** (`app/src/hooks/useMessages.ts`)
- Check if each message is encrypted
- If encrypted:
  - Get encryption key for room
  - Decrypt message content
  - Handle decryption errors gracefully
- If not encrypted:
  - Display content as-is

**11. Update Room Password Dialog** (`app/src/components/chat/RoomPasswordDialog.tsx`)
- When user enters password:
  - Validate with server (access control)
  - If room is encrypted, derive encryption key
  - Store key in memory
  - Fetch and decrypt messages

### Phase 4: Backend Updates

**12. Update Room Repository** (`internal/repository/sqlite/store.go`)
- Update `CreateRoom` to accept and store `is_encrypted` and `encryption_salt`
- Update `GetRooms` to return encryption metadata
- No changes needed for message storage (already stores content as-is)

**13. Update Room Handlers** (`internal/handlers/rooms.go`)
- Pass through `is_encrypted` flag from CreateRoomRequest
- Generate and return salt in room response

**14. Update Message Handlers** (`internal/handlers/messages.go`)
- No decryption logic needed (server stores ciphertext)
- Pass through `is_encrypted` and `nonce` fields
- Note: Server cannot validate message content for encrypted rooms

### Phase 5: UX Enhancements

**15. Visual Indicators**
- Add lock icon to encrypted rooms in sidebar
- Show encryption status in chat header
- Display warning if encryption key is missing

**16. Error Handling**
- Graceful handling of decryption failures
- Clear error messages for users
- Fallback display for undecryptable messages

**17. Password Management**
- Prompt for password when joining encrypted room
- Re-prompt if decryption fails (wrong password)
- Clear encryption keys on logout/session end

### Phase 6: Security Considerations

**18. Key Management**
- Store encryption keys only in memory (never persisted)
- Clear keys when user navigates away
- Re-derive on page reload (must re-enter password)

**19. Salt Storage**
- Store salt in database (public, not secret)
- Use consistent salt per room
- Generate cryptographically random salt on room creation

**20. Signature Integration**
- Sign encrypted content (not plaintext)
- Signature proves sender, encryption proves confidentiality
- Both features work independently and complementarily

**21. Metadata Leakage**
- Server still knows: room, sender, timestamp, message length
- Consider padding message length if this is a concern
- Last message preview cannot be shown for encrypted rooms

## Testing Strategy

### Unit Tests
- Encryption/decryption round-trip
- Key derivation consistency
- Nonce uniqueness
- Error handling for wrong password

### Integration Tests
- Create encrypted room
- Send/receive encrypted messages
- Multiple users with same password
- Wrong password handling
- Mixed encrypted/unencrypted rooms

### Security Tests
- Verify server cannot decrypt messages
- Verify salt tampering doesn't break security
- Verify keys are not persisted
- Verify nonce reuse prevention

## Migration Strategy

### Backwards Compatibility
- Default `is_encrypted = false` for existing rooms
- Existing rooms continue to work as before
- No automatic migration of existing messages
- Users opt-in to encryption when creating new rooms

### Gradual Rollout
1. Deploy backend changes (database migration)
2. Deploy frontend changes
3. Monitor for issues
4. Document feature for users

## Performance Considerations

- Encryption/decryption overhead: ~1-2ms per message
- Key derivation (PBKDF2): ~50-100ms (one-time per room)
- Consider Web Workers for heavy crypto operations
- Cache derived keys in memory during session

## Documentation Needed

### User Documentation
- How to create encrypted rooms
- What E2E encryption means
- Password recovery implications
- Security guarantees and limitations

### Developer Documentation
- Crypto implementation details
- Key derivation parameters
- API changes
- Testing procedures

## Future Enhancements

### Optional Improvements
- **Argon2**: More secure key derivation (if WebAssembly version available)
- **Message Padding**: Hide message length from server
- **Key Rotation**: Periodic re-keying for forward secrecy
- **Per-User Keys**: More complex key management for user-based access
- **Encrypted Attachments**: Extend to file uploads
- **Encrypted Room Names**: Full metadata encryption

### Advanced Features
- **Multi-Device Sync**: Encrypted key backup/recovery
- **Password Change**: Re-encrypt all messages with new password
- **Partial Decryption**: Show metadata even without key
- **Audit Log**: Track encryption events

## Security Assumptions

### Trust Model
- Users trust their own devices
- Users securely share room passwords out-of-band
- Server is honest-but-curious (stores data correctly, but may try to read it)
- Network transport is secure (HTTPS)

### Threat Model
**Protected Against:**
- Server admin reading messages
- Database breach exposing message content
- MITM attacks (combined with HTTPS)

**Not Protected Against:**
- Compromised client devices
- Malicious code injection
- Social engineering for passwords
- Timing attacks (not constant-time crypto)
- Physical access to unlocked device

## Open Questions

1. Should encryption be mandatory for password-protected rooms, or optional?
2. Should we support password changes for encrypted rooms (requires re-encryption)?
3. Should we implement key escrow/recovery mechanisms?
4. Should encrypted rooms support search functionality (client-side)?
5. Should we show "message encrypted" placeholders vs. hiding count entirely?

## Estimated Complexity

- **Backend Changes**: Low-Medium (mostly data model updates)
- **Frontend Changes**: Medium (crypto utilities, state management)
- **Testing**: Medium (security testing is critical)
- **Overall**: Medium complexity, high security impact

## Decision: Opt-In Feature

Recommendation: Make E2E encryption an **opt-in feature** when creating rooms.

**Rationale:**
- Maintains backwards compatibility
- Allows users to choose based on their needs
- Simpler rooms don't pay encryption overhead
- Clear security model (encrypted vs. not encrypted)

**UX Flow:**
1. User creates room with password
2. Checkbox: "Enable end-to-end encryption (messages will be encrypted on your device)"
3. If enabled, generate salt and mark room as encrypted
4. Show lock icon and "Encrypted" badge in UI
