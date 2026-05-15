
internal/
└── Gatekeeper/
    ├── totp/                      # TOTP generation and validation (RFC 6238)
    │   ├── service.go             # Main TOTP service
    │   ├── generator.go           # Secret generation
    │   ├── validator.go           # OTP verification
    │   ├── recovery.go            # Backup codes generation/validation
    │   ├── qrcode.go              # QR code creation
    │   ├── issuer.go              # otpauth URI builder
    │   ├── clock.go               # Time abstraction for testing
    │   └── errors.go
    │
    ├── webauthn/                  # Future support for security keys (optional)
    │   ├── service.go
    │   ├── registration.go
    │   ├── authentication.go
    │   └── errors.go
    │
    ├── sms/                       # Optional OTP via SMS
    │   ├── provider.go
    │   ├── service.go
    │   └── errors.go
    │
    ├── email/                     # Optional OTP via Email
    │   ├── provider.go
    │   ├── service.go
    │   └── errors.go
    │
    ├── policy/                    # MFA policies and enforcement rules
    │   ├── engine.go
    │   ├── rules.go
    │   └── evaluator.go
    │
    ├── enrollment/                # Setup and activation workflow
    │   ├── service.go
    │   ├── setup.go
    │   ├── activate.go
    │   ├── disable.go
    │   └── status.go
    │
    ├── challenge/                 # Runtime authentication challenge
    │   ├── service.go
    │   ├── begin.go
    │   ├── verify.go
    │   ├── session.go
    │   └── state.go
    │
    ├── backupcodes/               # Standalone backup code management
    │   ├── service.go
    │   ├── generator.go
    │   ├── validator.go
    │   └── hasher.go
    │
    ├── trusteddevices/            # Remember this device
    │   ├── service.go
    │   ├── token.go
    │   ├── cookie.go
    │   └── fingerprint.go
    │
    ├── risk/                      # Adaptive MFA (IP, device, geo, behavior)
    │   ├── engine.go
    │   ├── scorer.go
    │   └── signals.go
    │
    ├── middleware/                # HTTP/gRPC middleware for MFA enforcement
    │   ├── http.go
    │   ├── grpc.go
    │   └── context.go
    ├── raft/
    |   ├── raft.go
    │
    ├── handlers/                  # REST/GraphQL/gRPC handlers
    │   ├── http.go
    │   ├── grpc.go
    │   ├── dto.go
    │   └── mapper.go
    │
    ├── repositories/              # Interfaces
    │   ├── factor_repository.go
    │   ├── challenge_repository.go
    │   ├── backup_code_repository.go
    │   └── trusted_device_repository.go
    │
    ├── pgstore/                   # PostgreSQL implementations
    │   ├── factor_repository.go
    │   ├── challenge_repository.go
    │   ├── backup_code_repository.go
    │   ├── trusted_device_repository.go
    │   └── migrations/
    │       ├── 001_create_twofactor_factors.sql
    │       ├── 002_create_twofactor_challenges.sql
    │       ├── 003_create_twofactor_backup_codes.sql
    │       └── 004_create_twofactor_trusted_devices.sql
    │
    ├── cache/                     # Redis cache/session storage
    │   ├── challenge_cache.go
    │   └── rate_limit.go
    │
    ├── models/                    # Domain entities
    │   ├── factor.go
    │   ├── challenge.go
    │   ├── backup_code.go
    │   ├── trusted_device.go
    │   ├── policy.go
    │   └── enums.go
    │
    ├── contracts/                 # Public interfaces
    │   ├── service.go
    │   ├── provider.go
    │   └── types.go
    │
    ├── events/                    # Domain events
    │   ├── enrolled.go
    │   ├── verified.go
    │   ├── failed.go
    │   ├── disabled.go
    │   └── backup_code_used.go
    │
    ├── audit/                     # Security audit logging
    │   ├── logger.go
    │   └── event_types.go
    │
    ├── metrics/                   # Prometheus metrics
    │   ├── counters.go
    │   ├── histograms.go
    │   └── labels.go
    │
    ├── config/                    # Module configuration
    │   ├── config.go
    │   ├── defaults.go
    │   └── validation.go
    │
    ├── bootstrap/                 # Dependency wiring
    │   ├── module.go
    │   ├── providers.go
    │   └── routes.go
    │
    ├── testutil/                  # Test helpers
    │   ├── fixtures.go
    │   ├── mocks.go
    │   └── fake_clock.go
    │
    ├── docs/                      # Internal documentation
    │   ├── architecture.md
    │   ├── api.md
    │   └── sequence-diagrams.md
    │
    └── README.md