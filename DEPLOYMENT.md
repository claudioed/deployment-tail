# Deployment Guide

This guide covers the deployment sequence and procedures for the Deployment Tail application with authentication.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Deployment Sequence](#deployment-sequence)
3. [Database Deployment](#database-deployment)
4. [API Server Deployment](#api-server-deployment)
5. [CLI Distribution](#cli-distribution)
6. [Post-Deployment Validation](#post-deployment-validation)
7. [Disaster Recovery](#disaster-recovery)
8. [Rollback Procedures](#rollback-procedures)
9. [Monitoring Setup](#monitoring-setup)

## Prerequisites

Before deploying, ensure you have:

- [ ] Google OAuth 2.0 credentials created
- [ ] JWT secret generated (minimum 32 characters)
- [ ] MySQL 8.0+ database provisioned
- [ ] All environment variables configured
- [ ] Backup of current database (if upgrading)
- [ ] Access to production servers

## Deployment Sequence

**IMPORTANT**: Follow this sequence exactly to avoid authentication issues.

```
┌─────────────────────────────────────────────────────────────┐
│  Phase 1: Database                                          │
│  - Run migrations                                           │
│  - Create admin user                                        │
│  - Verify schema                                            │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Phase 2: API Server                                        │
│  - Deploy new binary                                        │
│  - Update environment variables                             │
│  - Restart service                                          │
│  - Verify health check                                      │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Phase 3: Validation                                        │
│  - Run smoke tests                                          │
│  - Test authentication flow                                 │
│  - Verify role-based access                                 │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Phase 4: CLI Distribution                                  │
│  - Build CLI binaries                                       │
│  - Distribute to users                                      │
│  - Update documentation                                     │
└─────────────────────────────────────────────────────────────┘
```

## Database Deployment

### Step 1: Backup Current Database

```bash
# Create backup
mysqldump -u root -p deployment_schedules > backup_$(date +%Y%m%d_%H%M%S).sql

# Verify backup
ls -lh backup_*.sql
```

### Step 2: Run Migrations

Migrations run automatically when the server starts, but you can run them manually:

```bash
# Set database connection
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=root
export DB_PASSWORD=your-password
export DB_NAME=deployment_schedules

# Start server (migrations auto-run)
./bin/server
```

**Migrations applied:**
- `000010_add_users_table.up.sql` - Creates users table
- `000011_add_schedule_audit_columns.up.sql` - Adds audit columns to schedules
- `000012_add_revoked_tokens_table.up.sql` - Creates token revocation table

### Step 3: Create Initial Admin User

**Option A: Using seed script (before first login)**

```bash
# If you don't have a Google ID yet
export ADMIN_EMAIL=your-email@example.com
go run scripts/seed_admin_user.go
```

The script will create a temporary Google ID. After your first login:

1. Check server logs for your actual Google ID
2. Update the database:
```sql
UPDATE users SET google_id = 'your-actual-google-id' WHERE email = 'your-email@example.com';
```

**Option B: Direct SQL (if you know your Google ID)**

```sql
INSERT INTO users (id, google_id, email, name, role, created_at, updated_at)
VALUES (
    UUID(),
    'your-google-id',
    'your-email@example.com',
    'Your Name',
    'admin',
    NOW(),
    NOW()
);
```

**Option C: Promote existing user**

```sql
UPDATE users SET role = 'admin' WHERE email = 'your-email@example.com';
```

### Step 4: Verify Database Schema

```bash
# Check users table
mysql -u root -p deployment_schedules -e "DESCRIBE users;"

# Check audit columns in schedules
mysql -u root -p deployment_schedules -e "DESCRIBE schedules;"

# Check revoked tokens table
mysql -u root -p deployment_schedules -e "DESCRIBE revoked_tokens;"

# Verify admin user exists
mysql -u root -p deployment_schedules -e "SELECT id, email, role FROM users WHERE role = 'admin';"
```

## API Server Deployment

### Step 1: Build the Server

```bash
# Build for production
make build

# Or manually
go build -o bin/server cmd/server/main.go
```

### Step 2: Configure Environment Variables

Create or update `/etc/deployment-tail/config.env`:

```bash
# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=deployment_user
DB_PASSWORD=secure-password-here
DB_NAME=deployment_schedules

# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Google OAuth
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_REDIRECT_URL=https://yourdomain.com/auth/google/callback

# JWT Configuration
JWT_SECRET=your-jwt-secret-minimum-32-characters-long
JWT_EXPIRY=24h
JWT_ISSUER=deployment-tail
```

### Step 3: Deploy Binary

```bash
# Copy binary to server
scp bin/server user@production-server:/opt/deployment-tail/bin/

# Or use your deployment tool
# ansible-playbook deploy.yml
# kubectl apply -f k8s/deployment.yaml
```

### Step 4: Update Systemd Service (if using systemd)

Create `/etc/systemd/system/deployment-tail.service`:

```ini
[Unit]
Description=Deployment Tail API Server
After=network.target mysql.service

[Service]
Type=simple
User=deployment
Group=deployment
WorkingDirectory=/opt/deployment-tail
EnvironmentFile=/etc/deployment-tail/config.env
ExecStart=/opt/deployment-tail/bin/server
Restart=always
RestartSec=10

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log/deployment-tail

[Install]
WantedBy=multi-user.target
```

### Step 5: Restart Service

```bash
# Reload systemd
sudo systemctl daemon-reload

# Restart service
sudo systemctl restart deployment-tail

# Check status
sudo systemctl status deployment-tail

# View logs
sudo journalctl -u deployment-tail -f
```

### Step 6: Verify Health Check

```bash
# Test health endpoint (no auth required)
curl http://localhost:8080/health

# Expected response:
# {"status":"ok"}
```

## CLI Distribution

### Step 1: Build CLI Binaries

```bash
# Build for multiple platforms
make build-cli

# Or manually for each platform
GOOS=linux GOARCH=amd64 go build -o bin/deployment-tail-linux-amd64 cmd/cli/main.go
GOOS=darwin GOARCH=amd64 go build -o bin/deployment-tail-darwin-amd64 cmd/cli/main.go
GOOS=darwin GOARCH=arm64 go build -o bin/deployment-tail-darwin-arm64 cmd/cli/main.go
GOOS=windows GOARCH=amd64 go build -o bin/deployment-tail-windows-amd64.exe cmd/cli/main.go
```

### Step 2: Distribute to Users

**Option A: GitHub Releases**
```bash
gh release create v1.0.0 bin/deployment-tail-* --title "Release v1.0.0" --notes "Added authentication"
```

**Option B: Internal artifact repository**
```bash
# Upload to Artifactory/Nexus
curl -u user:pass -T bin/deployment-tail-linux-amd64 https://artifacts.company.com/deployment-tail/v1.0.0/
```

**Option C: Direct distribution**
```bash
# Copy to shared location
cp bin/deployment-tail-* /shared/tools/deployment-tail/
```

### Step 3: User Setup Instructions

Send these instructions to users:

```bash
# 1. Download CLI for your platform
curl -O https://releases.company.com/deployment-tail/latest/deployment-tail

# 2. Make executable
chmod +x deployment-tail

# 3. Move to PATH
sudo mv deployment-tail /usr/local/bin/

# 4. Login
deployment-tail auth login

# 5. Verify
deployment-tail auth status
```

## Post-Deployment Validation

### Automated Smoke Tests

Run the smoke test script:

```bash
./scripts/smoke_test.sh
```

### Manual Validation Checklist

- [ ] Health check returns 200 OK
- [ ] Unauthenticated requests return 401
- [ ] Google OAuth login flow works
- [ ] JWT tokens are issued correctly
- [ ] Token refresh works
- [ ] Token revocation (logout) works
- [ ] Role-based access control enforced
- [ ] Schedule creation includes audit trail
- [ ] Existing schedules still accessible
- [ ] CLI authentication works
- [ ] CLI commands require authentication

### Test Authentication Flow

```bash
# 1. Test unauthenticated request (should fail)
curl http://localhost:8080/api/v1/schedules
# Expected: 401 Unauthorized

# 2. Get OAuth URL
curl http://localhost:8080/auth/google/login
# Should redirect to Google

# 3. After OAuth, test with token
export TOKEN="your-jwt-token"
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/schedules
# Should return schedules

# 4. Test token refresh
curl -X POST -H "Authorization: Bearer $TOKEN" http://localhost:8080/auth/refresh
# Should return new token

# 5. Test logout
curl -X POST -H "Authorization: Bearer $TOKEN" http://localhost:8080/auth/logout
# Token should be revoked
```

## Disaster Recovery

### Scenario 1: Admin Lockout

**Problem**: All admin users are locked out or deleted.

**Solution**:

```bash
# Option 1: Create new admin via database
mysql -u root -p deployment_schedules << EOF
INSERT INTO users (id, google_id, email, name, role, created_at, updated_at)
VALUES (UUID(), 'emergency-admin-$(date +%s)', 'emergency@company.com', 'Emergency Admin', 'admin', NOW(), NOW());
EOF

# Option 2: Promote existing user
mysql -u root -p deployment_schedules << EOF
UPDATE users SET role = 'admin' WHERE email = 'user@company.com';
EOF

# Option 3: Use seed script
export ADMIN_EMAIL=new-admin@company.com
go run scripts/seed_admin_user.go
```

### Scenario 2: OAuth Provider Outage

**Problem**: Google OAuth is down, users cannot authenticate.

**Emergency Solution**: Temporarily disable authentication

```bash
# Run emergency auth disable script
./scripts/emergency_auth_disable.sh

# This creates a migration that sets auth_disabled flag
# REMEMBER TO RE-ENABLE AFTER OAUTH IS RESTORED!

# To re-enable:
mysql -u root -p deployment_schedules << EOF
UPDATE system_config SET config_value = 'false' WHERE config_key = 'auth_disabled';
EOF

# Restart server
systemctl restart deployment-tail
```

### Scenario 3: JWT Secret Compromised

**Problem**: JWT secret has been exposed.

**Solution**:

1. Generate new JWT secret:
```bash
openssl rand -base64 32
```

2. Update environment variable:
```bash
# Update /etc/deployment-tail/config.env
JWT_SECRET=new-secret-here
```

3. Restart server:
```bash
systemctl restart deployment-tail
```

4. All existing tokens are now invalid
5. Users must re-authenticate

### Scenario 4: Database Corruption

**Problem**: Users table is corrupted or lost.

**Solution**:

```bash
# 1. Stop API server
systemctl stop deployment-tail

# 2. Restore from backup
mysql -u root -p deployment_schedules < backup_20260404_120000.sql

# 3. Verify data
mysql -u root -p deployment_schedules -e "SELECT COUNT(*) FROM users;"

# 4. Start API server
systemctl start deployment-tail
```

## Rollback Procedures

### Full Rollback

If the deployment fails and needs to be rolled back:

```bash
# 1. Stop new server
systemctl stop deployment-tail

# 2. Restore old binary
cp /opt/deployment-tail/bin/server.backup /opt/deployment-tail/bin/server

# 3. Rollback database
mysql -u root -p deployment_schedules < backup_before_deploy.sql

# 4. Remove environment variables (if needed)
# Edit /etc/deployment-tail/config.env and remove auth variables

# 5. Start old server
systemctl start deployment-tail

# 6. Verify
curl http://localhost:8080/health
```

### Partial Rollback (Keep Database, Rollback Code)

```bash
# 1. Stop server
systemctl stop deployment-tail

# 2. Restore old binary
cp /opt/deployment-tail/bin/server.backup /opt/deployment-tail/bin/server

# 3. Start server (database migrations are forward-compatible)
systemctl start deployment-tail
```

## Monitoring Setup

### Key Metrics to Monitor

1. **Authentication Failures**
   - Failed login attempts
   - Invalid token attempts
   - Expired token attempts

2. **User Activity**
   - New user registrations
   - Active sessions
   - Token refresh rate

3. **Performance**
   - JWT validation time
   - Database query time
   - OAuth callback latency

4. **Security**
   - Revoked token access attempts
   - Role escalation attempts
   - Brute force detection

### Example Prometheus Metrics

```yaml
# metrics.yml
- name: auth_failures_total
  help: Total number of authentication failures
  type: counter
  labels: [reason]

- name: active_sessions
  help: Number of active user sessions
  type: gauge

- name: jwt_validation_duration_seconds
  help: Time to validate JWT tokens
  type: histogram

- name: user_registrations_total
  help: Total number of new user registrations
  type: counter
```

### Example Alert Rules

```yaml
# alerts.yml
groups:
  - name: authentication
    rules:
      - alert: HighAuthFailureRate
        expr: rate(auth_failures_total[5m]) > 10
        for: 5m
        annotations:
          summary: "High authentication failure rate"

      - alert: NoActiveAdmins
        expr: count(users{role="admin"}) == 0
        annotations:
          summary: "No admin users available"

      - alert: TokenRevocationBacklogHigh
        expr: revoked_tokens_count > 10000
        annotations:
          summary: "Revoked tokens table needs cleanup"
```

## Support

For issues during deployment:

1. Check server logs: `journalctl -u deployment-tail -f`
2. Check application logs: `/var/log/deployment-tail/`
3. Verify environment variables: `systemctl show deployment-tail`
4. Run smoke tests: `./scripts/smoke_test.sh`
5. Contact: devops@company.com

## Checklist

Use this checklist for each deployment:

- [ ] Backup database created
- [ ] Migrations reviewed
- [ ] Environment variables configured
- [ ] Admin user created/verified
- [ ] Binary deployed
- [ ] Service restarted
- [ ] Health check passed
- [ ] Smoke tests passed
- [ ] Authentication flow tested
- [ ] Role-based access verified
- [ ] CLI distributed to users
- [ ] Documentation updated
- [ ] Monitoring configured
- [ ] Alerts configured
- [ ] Rollback plan ready
