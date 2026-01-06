# MFA Implementation Plan for zrok

## Overview

Add optional TOTP-based multi-factor authentication to zrok's web console login. Users can enroll an authenticator app in account settings, and will be required to enter a TOTP code when logging in. Includes recovery codes and a global enforcement option.

**Scope**: Web console only (CLI uses pre-stored account tokens and is unaffected)

---

## Database Schema

Create migration `040_v2_0_0_mfa.sql` for PostgreSQL and SQLite3:

### Tables

```sql
-- MFA configuration for accounts
create table account_mfa (
    id                    serial primary key,
    account_id            integer not null unique references accounts(id),
    totp_secret           varchar(64) not null,
    enabled               boolean not null default(false),
    created_at            timestamptz not null default(current_timestamp),
    updated_at            timestamptz not null default(current_timestamp)
);

-- Recovery codes (hashed)
create table mfa_recovery_codes (
    id                    serial primary key,
    account_mfa_id        integer not null references account_mfa(id),
    code_hash             varchar(64) not null,
    used                  boolean not null default(false),
    used_at               timestamptz,
    created_at            timestamptz not null default(current_timestamp)
);

-- Pending MFA sessions for two-step login
create table mfa_pending_auth (
    id                    serial primary key,
    account_id            integer not null references accounts(id),
    pending_token         varchar(64) not null unique,
    expires_at            timestamptz not null,
    created_at            timestamptz not null default(current_timestamp)
);
```

**Files to create:**
- `controller/store/sql/postgresql/040_v2_0_0_mfa.sql`
- `controller/store/sql/sqlite3/040_v2_0_0_mfa.sql`

---

## Backend Implementation

### Dependencies

Add to `go.mod`:
```
github.com/pquerna/otp v1.4.0
```

### TOTP Secret Encryption

TOTP secrets are encrypted at rest using AES-256-GCM. A server-side encryption key is required in the config.

**Config addition** (`controller/config/config.go`):
```go
type Config struct {
    // ...existing fields...
    MfaRequired   bool
    MfaSecretKey  string  // 32-byte base64-encoded key for AES-256-GCM
}
```

**Encryption utilities** (in `controller/totp.go`):
```go
// EncryptTotpSecret encrypts a TOTP secret for database storage
// Uses AES-256-GCM with random nonce prepended to ciphertext
func EncryptTotpSecret(plaintext, key string) (string, error)

// DecryptTotpSecret decrypts a TOTP secret from database
func DecryptTotpSecret(ciphertext, key string) (string, error)
```

**Usage flow**:
1. During `/mfa/setup`: Generate TOTP secret → encrypt → store in `account_mfa.totp_secret`
2. During `/mfa/authenticate`: Read encrypted secret → decrypt → validate TOTP code

**Key generation** (one-time setup):
```bash
openssl rand -base64 32
```

**Operational notes**:
- Key must be present in config before MFA can be used
- Key rotation requires re-encrypting all existing secrets (migration script needed)
- If key is lost, all MFA enrollments become invalid

### New Files

| File | Purpose |
|------|---------|
| `controller/store/accountMfa.go` | Store model and CRUD for MFA tables |
| `controller/totp.go` | TOTP secret generation, validation, QR code, recovery codes |
| `controller/mfaSetup.go` | `POST /mfa/setup` - initiate enrollment, return QR code |
| `controller/mfaVerify.go` | `POST /mfa/verify` - verify code, enable MFA, return recovery codes |
| `controller/mfaDisable.go` | `POST /mfa/disable` - disable MFA (requires password + code) |
| `controller/mfaAuthenticate.go` | `POST /mfa/authenticate` - complete login with TOTP code |
| `controller/mfaStatus.go` | `GET /mfa/status` - check if MFA enabled |
| `controller/mfaRecoveryCodes.go` | `POST /mfa/recoveryCodes` - regenerate recovery codes |
| `controller/maintenanceMfaPendingAuth.go` | Cleanup expired pending auth tokens |

### Files to Modify

| File | Changes |
|------|---------|
| `specs/src/account.yml` | Add MFA endpoints, modify `/login` to return 202 for MFA required |
| `controller/controller.go` | Register new MFA handlers |
| `controller/login.go` | Check MFA status after password verification |
| `controller/config/config.go` | Add `MfaRequired bool` and `MfaSecretKey string` config options |

### API Endpoints

| Endpoint | Auth | Description |
|----------|------|-------------|
| `POST /mfa/setup` | Token | Returns `{secret, qrCode, provisioningUri}` |
| `POST /mfa/verify` | Token | Body: `{code}` → Returns `{recoveryCodes: [...]}` |
| `POST /mfa/disable` | Token | Body: `{password, code}` |
| `GET /mfa/status` | Token | Returns `{enabled, recoveryCodesRemaining}` |
| `POST /mfa/recoveryCodes` | Token | Body: `{code}` → Returns new codes |
| `POST /mfa/authenticate` | None | Body: `{pendingToken, code}` → Returns account token |

### Modified Login Flow

```
POST /login {email, password}
    ↓
Password valid?
    ↓ No → 401 Unauthorized
    ↓ Yes
MFA enabled for account?
    ↓ No → 200 OK {token}
    ↓ Yes
Create pending_auth record (expires in 5 min)
    ↓
202 Accepted {pendingToken}
    ↓
Frontend shows MFA code input
    ↓
POST /mfa/authenticate {pendingToken, code}
    ↓
Code valid (TOTP or recovery)?
    ↓ No → 401 Unauthorized
    ↓ Yes → 200 OK {token}
```

### Global Enforcement

Add to `controller/config/config.go`:
```go
type Config struct {
    // ...existing fields...
    MfaRequired bool
}
```

**Behavior by setting:**

| `MfaRequired` | Behavior |
|---------------|----------|
| `false` (default) | MFA is **opt-in**. Users can enable MFA in account settings if they want. Those who enable it must provide TOTP to log in. Those who don't can log in with just email/password. |
| `true` | MFA is **mandatory**. `/login` returns 403 if user doesn't have MFA enabled. Unenrolled users must complete MFA setup before accessing the app. |

In both modes, the MFA setup/disable endpoints are available and functional.

---

## Frontend Implementation

### New Components

| File | Purpose |
|------|---------|
| `ui/src/MfaSetupModal.tsx` | Multi-step wizard: QR code → verify code → show recovery codes |
| `ui/src/MfaVerifyModal.tsx` | Login flow: enter TOTP code (or recovery code) |
| `ui/src/MfaDisableModal.tsx` | Confirm disable: password + TOTP code |
| `ui/src/MfaRecoveryCodesModal.tsx` | View remaining codes, regenerate |
| `ui/src/RecoveryCodesDownload.tsx` | Download codes as text file |

### Files to Modify

| File | Changes |
|------|---------|
| `ui/src/Login.tsx` | Handle 202 response, show `MfaVerifyModal` |
| `ui/src/AccountPanel.tsx` | Add MFA status indicator and setup/manage button |
| `ui/src/model/user.ts` | Add `mfaEnabled?: boolean` to User interface |

### UI Flow - Enrollment

1. User clicks "Enable MFA" in AccountPanel
2. `MfaSetupModal` opens, calls `POST /mfa/setup`
3. Modal shows QR code and manual entry secret
4. User scans with authenticator, enters verification code
5. `POST /mfa/verify` enables MFA, returns recovery codes
6. Modal shows recovery codes with download button
7. User must acknowledge they saved codes before closing

### UI Flow - Login with MFA

1. User enters email/password, submits
2. API returns 202 with `pendingToken`
3. `MfaVerifyModal` opens
4. User enters TOTP code (or clicks "Use recovery code")
5. `POST /mfa/authenticate` validates and returns token
6. Normal login completion

---

## Security Considerations

- **TOTP secrets**: Encrypted at rest with AES-256-GCM using server-side key
- **Recovery codes**: Hashed with argon2 before storage, 10 codes generated
- **Pending auth expiry**: 5 minutes, cleaned up by maintenance agent
- **Rate limiting**: Consider limiting `/mfa/authenticate` attempts per pending token
- **Time drift**: Accept codes within ±1 time period (30 seconds)
- **Key management**: MfaSecretKey must be securely stored and backed up

---

## Implementation Order

### Phase 1: Backend Foundation
1. Add `github.com/pquerna/otp` dependency to `go.mod`
2. Create database migrations (`controller/store/sql/postgresql/040_v2_0_0_mfa.sql` and sqlite3 equivalent)
3. Create `controller/store/accountMfa.go` - Store model and CRUD operations for MFA tables
4. Create `controller/totp.go` - TOTP utilities (secret generation, validation, QR code, encryption, recovery codes)

### Phase 2: MFA Setup APIs
5. Update OpenAPI spec (`specs/src/account.yml`)
6. Run `make generate`
7. Implement `mfaSetup.go`, `mfaVerify.go`, `mfaStatus.go`
8. Register handlers in `controller.go`

### Phase 3: Login Flow
9. Modify `login.go` for MFA check
10. Implement `mfaAuthenticate.go`
11. Add pending auth cleanup maintenance agent

### Phase 4: Management APIs
12. Implement `mfaDisable.go`, `mfaRecoveryCodes.go`
13. Add global enforcement config option

### Phase 5: Frontend - Enrollment
14. Regenerate TypeScript API client
15. Create `MfaSetupModal.tsx`, `RecoveryCodesDownload.tsx`
16. Update `AccountPanel.tsx`

### Phase 6: Frontend - Login
17. Create `MfaVerifyModal.tsx`
18. Update `Login.tsx`

### Phase 7: Frontend - Management
19. Create `MfaDisableModal.tsx`, `MfaRecoveryCodesModal.tsx`

### Phase 8: Testing
20. Unit tests for TOTP functions
21. Integration tests for MFA flow
22. Manual testing of full enrollment and login flows

---

## Key Reference Files

- `controller/login.go` - Current login handler to modify
- `controller/store/account.go` - Pattern for store models
- `controller/store/passwordResetRequest.go` - Similar token-based flow pattern
- `controller/config/config.go` - Where to add MfaRequired option
- `ui/src/AccountPasswordChangeModal.tsx` - Modal pattern to follow
- `ui/src/Login.tsx` - Login page to modify
- `ui/src/AccountPanel.tsx` - Settings page to add MFA controls
