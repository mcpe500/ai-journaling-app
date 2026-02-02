#!/bin/sh
set -e

# =============================================================================
# PocketBase Entrypoint Script with Auto Superuser Creation
# =============================================================================
# This script automatically creates a PocketBase superuser from environment
# variables on first container startup, eliminating manual setup.
#
# Required Environment Variables:
#   PB_ADMIN_EMAIL     - Superuser email address
#   PB_ADMIN_PASSWORD  - Superuser password (min 8 characters)
#
# Optional Environment Variables:
#   PB_HTTP           - HTTP bind address (default: 0.0.0.0:8090)
# =============================================================================

# Configuration
PB_BIND_ADDRESS="${PB_HTTP:-0.0.0.0:8090}"
HEALTH_CHECK_URL="http://localhost:8090/api/health"
MAX_WAIT_SECONDS=30

# Color codes for output (if terminal supports it)
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# =============================================================================
# VALIDATION FUNCTIONS
# =============================================================================

# Validate email format (basic RFC 5322 subset)
validate_email() {
    local email="$1"
    # Basic regex: checks for @ symbol and domain format
    case "$email" in
        *@*.*) return 0 ;;
        *) return 1 ;;
    esac
}

# Validate password meets minimum requirements
validate_password() {
    local password="$1"
    # Minimum 8 characters
    if [ "${#password}" -lt 8 ]; then
        return 1
    fi
    return 0
}

# Check if PocketBase binary exists
check_binary() {
    if [ ! -f "/pb" ]; then
        log_error "PocketBase binary not found at /pb"
        log_error "Please ensure the binary is installed at the correct path"
        exit 1
    fi
    
    if [ ! -x "/pb" ]; then
        log_error "PocketBase binary at /pb is not executable"
        exit 1
    fi
}

# =============================================================================
# HEALTH CHECK FUNCTIONS
# =============================================================================

# Check if PocketBase is healthy and responding
check_pb_health() {
    # Use wget or curl to check health endpoint
    if command -v wget >/dev/null 2>&1; then
        wget -q --spider "$HEALTH_CHECK_URL" 2>/dev/null
        return $?
    elif command -v curl >/dev/null 2>&1; then
        curl -s -o /dev/null -w "%{http_code}" "$HEALTH_CHECK_URL" | grep -q "200"
        return $?
    else
        # Fallback: wait for port to be open
        (echo > /dev/tcp/localhost/8090) 2>/dev/null
        return $?
    fi
}

# Wait for PocketBase to be ready
wait_for_pb() {
    local attempt=1
    
    log_info "Waiting for PocketBase to initialize..."
    
    while [ $attempt -le $MAX_WAIT_SECONDS ]; do
        if check_pb_health; then
            log_info "PocketBase is ready after ${attempt} seconds"
            return 0
        fi
        
        if [ $((attempt % 5)) -eq 0 ]; then
            log_info "Still waiting... (${attempt}/${MAX_WAIT_SECONDS}s)"
        fi
        
        sleep 1
        attempt=$((attempt + 1))
    done
    
    log_error "PocketBase failed to start within ${MAX_WAIT_SECONDS} seconds"
    return 1
}

# =============================================================================
# SUPERUSER MANAGEMENT
# =============================================================================

# Check if superuser already exists
check_superuser_exists() {
    local email="$1"
    
    # Try to list superusers and check for email
    # PocketBase CLI doesn't have a direct "list" command, so we'll try to create
    # and check the exit code/response
    local output
    output=$(/pb superuser upsert "$email" "dummy_password_for_check" 2>&1) || true
    
    # If output contains "already exists" or similar, user exists
    if echo "$output" | grep -qi "already exists\|duplicate\|conflict"; then
        return 0
    fi
    
    # Alternative: Try to authenticate (this is safer)
    # For now, we'll rely on the upsert command's behavior
    # If the user doesn't exist, upsert will create it
    # If it exists, it will update (which is fine for our use case)
    
    return 1
}

# Create superuser with error handling
create_superuser() {
    local email="$1"
    local password="$2"
    
    log_info "Attempting to create superuser: $email"
    
    # Create or update superuser
    if /pb superuser upsert "$email" "$password" 2>/dev/null; then
        log_info "Superuser created/updated successfully!"
        return 0
    else
        # Check if it's because user already exists
        log_warn "Superuser creation returned non-zero exit code"
        log_warn "This may be because the user already exists"
        return 1
    fi
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    log_info "Starting PocketBase initialization..."
    
    # Validate binary exists
    check_binary
    
    # Check if admin credentials are provided
    if [ -z "$PB_ADMIN_EMAIL" ] || [ -z "$PB_ADMIN_PASSWORD" ]; then
        log_warn "Admin credentials not provided (PB_ADMIN_EMAIL and/or PB_ADMIN_PASSWORD)"
        log_warn "Superuser will not be created automatically"
        log_warn "You can manually create admin via: http://localhost:8090/_/"
        
        # Start PocketBase directly without superuser creation
        log_info "Starting PocketBase on $PB_BIND_ADDRESS"
        exec /pb serve --http="$PB_BIND_ADDRESS"
    fi
    
    # Validate email format
    if ! validate_email "$PB_ADMIN_EMAIL"; then
        log_error "Invalid email format: $PB_ADMIN_EMAIL"
        log_error "Please provide a valid email address (e.g., admin@example.com)"
        exit 1
    fi
    
    # Validate password strength
    if ! validate_password "$PB_ADMIN_PASSWORD"; then
        log_error "Password does not meet minimum requirements"
        log_error "Password must be at least 8 characters long"
        log_error "Current length: ${#PB_ADMIN_PASSWORD} characters"
        exit 1
    fi
    
    log_info "Admin credentials validated successfully"
    log_info "Email: $PB_ADMIN_EMAIL"
    log_info "Password: [REDACTED - ${#PB_ADMIN_PASSWORD} characters]"
    
    # Start PocketBase in background for superuser creation
    log_info "Starting PocketBase temporarily for superuser setup..."
    /pb serve --http="$PB_BIND_ADDRESS" &
    PB_PID=$!
    
    # Wait for PocketBase to be ready
    if ! wait_for_pb; then
        log_error "Failed to start PocketBase for superuser creation"
        kill $PB_PID 2>/dev/null || true
        exit 1
    fi
    
    # Create the superuser
    if create_superuser "$PB_ADMIN_EMAIL" "$PB_ADMIN_PASSWORD"; then
        log_info "You can now login at: http://localhost:8090/_/"
        log_info "Email: $PB_ADMIN_EMAIL"
    else
        log_warn "Superuser may already exist or another issue occurred"
        log_info "You can try logging in with the provided credentials"
    fi
    
    # Stop the background PocketBase process
    log_info "Stopping temporary PocketBase instance..."
    kill $PB_PID 2>/dev/null || true
    wait $PB_PID 2>/dev/null || true
    
    # Give it a moment to fully shutdown
    sleep 2
    
    # Start PocketBase in foreground (main process)
    log_info "Starting PocketBase in production mode on $PB_BIND_ADDRESS"
    exec /pb serve --http="$PB_BIND_ADDRESS"
}

# Handle signals for graceful shutdown
cleanup() {
    log_warn "Received shutdown signal, cleaning up..."
    if [ -n "$PB_PID" ]; then
        kill $PB_PID 2>/dev/null || true
    fi
    exit 0
}

trap cleanup TERM INT

# Run main function
main "$@"
