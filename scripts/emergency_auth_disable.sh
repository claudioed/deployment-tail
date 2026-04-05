#!/bin/bash

# Emergency Authentication Disable Script
#
# This script creates an emergency rollback migration to temporarily disable
# authentication requirements. Use this ONLY in emergency situations where
# authentication is blocking critical operations.
#
# Usage:
#   ./scripts/emergency_auth_disable.sh
#
# WARNING: This will allow unauthenticated access to the API!
# Remember to re-enable authentication after the emergency is resolved.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
MIGRATIONS_DIR="$PROJECT_ROOT/migrations"

# Colors for output
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

echo -e "${RED}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${RED}║          EMERGENCY AUTHENTICATION DISABLE                  ║${NC}"
echo -e "${RED}║                                                            ║${NC}"
echo -e "${RED}║  ⚠️  WARNING: This will disable authentication!  ⚠️         ║${NC}"
echo -e "${RED}║                                                            ║${NC}"
echo -e "${RED}║  This is an EMERGENCY ONLY procedure. Unauthenticated     ║${NC}"
echo -e "${RED}║  access will be allowed to all endpoints.                 ║${NC}"
echo -e "${RED}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${YELLOW}Use cases for this script:${NC}"
echo "  - Admin lockout (all admins locked out)"
echo "  - OAuth provider outage (Google OAuth down)"
echo "  - Critical bug in authentication middleware"
echo "  - Database issues with users table"
echo ""
echo -e "${YELLOW}What this script does:${NC}"
echo "  1. Creates a migration to add 'auth_disabled' flag"
echo "  2. Sets the flag to enable bypass mode"
echo "  3. Provides instructions to re-enable auth"
echo ""

# Confirm action
read -p "Are you sure you want to disable authentication? (type 'YES' to confirm): " CONFIRM
if [ "$CONFIRM" != "YES" ]; then
    echo "Aborted."
    exit 1
fi

# Get next migration number
LAST_MIGRATION=$(ls -1 "$MIGRATIONS_DIR" | grep -E '^[0-9]+_' | tail -1 | cut -d_ -f1)
if [ -z "$LAST_MIGRATION" ]; then
    NEXT_NUMBER="000013"
else
    NEXT_NUMBER=$(printf "%06d" $((10#$LAST_MIGRATION + 1)))
fi

MIGRATION_NAME="${NEXT_NUMBER}_emergency_auth_disable"
UP_FILE="$MIGRATIONS_DIR/${MIGRATION_NAME}.up.sql"
DOWN_FILE="$MIGRATIONS_DIR/${MIGRATION_NAME}.down.sql"

# Create UP migration
cat > "$UP_FILE" << 'EOF'
-- Emergency authentication disable migration
-- This allows the application to bypass authentication checks temporarily

-- Create configuration table if it doesn't exist
CREATE TABLE IF NOT EXISTS system_config (
    config_key VARCHAR(255) PRIMARY KEY,
    config_value TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by VARCHAR(255) DEFAULT 'system',
    description TEXT
);

-- Set authentication disabled flag
INSERT INTO system_config (config_key, config_value, description, created_by)
VALUES (
    'auth_disabled',
    'true',
    'Emergency flag to disable authentication. Set to false to re-enable.',
    'emergency_script'
)
ON DUPLICATE KEY UPDATE
    config_value = 'true',
    updated_at = CURRENT_TIMESTAMP,
    created_by = 'emergency_script';

-- Log the action
INSERT INTO system_config (config_key, config_value, description, created_by)
VALUES (
    'auth_disabled_at',
    NOW(),
    'Timestamp when authentication was disabled',
    'emergency_script'
)
ON DUPLICATE KEY UPDATE
    config_value = NOW(),
    updated_at = CURRENT_TIMESTAMP;
EOF

# Create DOWN migration
cat > "$DOWN_FILE" << 'EOF'
-- Re-enable authentication

UPDATE system_config
SET config_value = 'false', updated_at = CURRENT_TIMESTAMP
WHERE config_key = 'auth_disabled';

-- Log re-enable
INSERT INTO system_config (config_key, config_value, description, created_by)
VALUES (
    'auth_reenabled_at',
    NOW(),
    'Timestamp when authentication was re-enabled',
    'emergency_script'
)
ON DUPLICATE KEY UPDATE
    config_value = NOW(),
    updated_at = CURRENT_TIMESTAMP;
EOF

echo ""
echo -e "${GREEN}✓ Created emergency migrations:${NC}"
echo "  - $UP_FILE"
echo "  - $DOWN_FILE"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo ""
echo "1. Restart the API server to apply the migration:"
echo "   systemctl restart deployment-tail"
echo "   # OR"
echo "   ./bin/server"
echo ""
echo "2. Verify authentication is disabled:"
echo "   curl http://localhost:8080/api/v1/schedules"
echo "   # Should return data without Authorization header"
echo ""
echo "3. ${RED}CRITICAL: Fix the authentication issue${NC}"
echo ""
echo "4. Re-enable authentication by running:"
echo "   mysql deployment_schedules -e \"UPDATE system_config SET config_value = 'false' WHERE config_key = 'auth_disabled';\""
echo ""
echo "5. Restart the API server again"
echo ""
echo "6. Verify authentication is working:"
echo "   curl http://localhost:8080/api/v1/schedules"
echo "   # Should return 401 Unauthorized"
echo ""
echo -e "${RED}⚠️  REMEMBER: Authentication is now disabled!${NC}"
echo -e "${RED}⚠️  Re-enable it as soon as possible!${NC}"
echo ""

# Optionally apply migration immediately
read -p "Do you want to restart the server now to apply this migration? (y/n): " RESTART
if [ "$RESTART" = "y" ] || [ "$RESTART" = "Y" ]; then
    echo ""
    echo "Please restart your server manually:"
    echo "  systemctl restart deployment-tail"
    echo "  # OR"
    echo "  ./bin/server"
fi

echo ""
echo -e "${YELLOW}Note: Update your application code to check the auth_disabled flag:${NC}"
echo ""
echo "  // In middleware/auth.go:"
echo "  func (m *AuthMiddleware) Handler(next http.Handler) http.Handler {"
echo "      return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {"
echo "          // Check if auth is disabled"
echo "          if m.isAuthDisabled() {"
echo "              next.ServeHTTP(w, r)"
echo "              return"
echo "          }"
echo "          // Normal authentication flow..."
echo "      })"
echo "  }"
echo ""
