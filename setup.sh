#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_NAME="tacacs4zitadel"
ZITADEL_ADMIN_USER="zitadel-admin@zitadel.localhost"
ZITADEL_ADMIN_PASS="Password1!"
ZITADEL_URL="http://localhost:8080"

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to wait for service
wait_for_service() {
    local url=$1
    local service_name=$2
    local max_attempts=60
    local attempt=1

    print_status "Waiting for $service_name to be ready..."
    
    while [ $attempt -le $max_attempts ]; do
        if curl -f -s "$url" >/dev/null 2>&1; then
            print_success "$service_name is ready!"
            return 0
        fi
        
        echo -n "."
        sleep 5
        attempt=$((attempt + 1))
    done
    
    print_error "$service_name failed to start within expected time"
    return 1
}

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    if ! command_exists docker; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    if ! command_exists docker-compose; then
        print_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    if ! command_exists curl; then
        print_error "curl is not installed. Please install curl first."
        exit 1
    fi
    
    if ! command_exists jq; then
        print_warning "jq is not installed. Installing jq for JSON processing..."
        if command_exists apt-get; then
            sudo apt-get update && sudo apt-get install -y jq
        elif command_exists yum; then
            sudo yum install -y jq
        elif command_exists brew; then
            brew install jq
        else
            print_error "Cannot install jq automatically. Please install it manually."
            exit 1
        fi
    fi
    
    print_success "All prerequisites are satisfied"
}

# Function to start services
start_services() {
    print_status "Starting Zitadel and TACACS+ services..."
    
    # Copy environment file
    if [ ! -f .env ]; then
        cp .env.zitadel .env
        print_status "Environment file created from template"
    fi
    
    # Start services using Zitadel compose file
    docker-compose -f docker-compose.zitadel.yml up -d
    
    wait_for_service "${ZITADEL_URL}/debug/healthz" "Zitadel" || {
        print_error "Zitadel failed to start"
        return 1
    }
    
    wait_for_service "http://localhost:8090/health" "TACACS+ Server" || {
        print_error "TACACS+ Server failed to start"
        return 1
    }
    
    print_success "All services started successfully"
}

# Function to configure Zitadel
configure_zitadel() {
    print_status "Configuring Zitadel..."
    
    # Note: Zitadel configuration is different from Keycloak
    # For now, we'll provide manual configuration instructions
    
    print_status "Zitadel is ready for manual configuration"
    print_warning "Manual configuration required:"
    echo "1. Go to ${ZITADEL_URL}/ui/console"
    echo "2. Login with: ${ZITADEL_ADMIN_USER} / ${ZITADEL_ADMIN_PASS}"
    echo "3. Create a new project for TACACS+"
    echo "4. Create an application with client credentials"
    echo "5. Add users and assign roles (network-admin, network-user, network-readonly)"
    echo "6. Update .env file with:"
    echo "   - ZITADEL_PROJECT_ID"
    echo "   - ZITADEL_CLIENT_ID"
    echo "   - ZITADEL_CLIENT_SECRET"
    echo "7. Restart the TACACS+ server: docker-compose -f docker-compose.zitadel.yml restart tacacs-server"
}

# Function to run tests
run_tests() {
    print_status "Running automated tests..."
    
    docker-compose -f docker-compose.zitadel.yml --profile testing up test-client
    
    if [ $? -eq 0 ]; then
        print_success "All tests passed!"
    else
        print_warning "Some tests failed. Check the logs for details."
    fi
}

# Function to display summary
display_summary() {
    print_success "=== TACACS+ with Zitadel Setup Complete ==="
    echo
    echo "🌐 Service URLs:"
    echo "   • Zitadel Console: ${ZITADEL_URL}/ui/console"
    echo "   • Zitadel Account: ${ZITADEL_URL}/ui/login"
    echo "   • TACACS+ Server: localhost:49"
    echo "   • Health Check: http://localhost:8090/health"
    echo
    echo "👤 Admin Credentials:"
    echo "   • Username: ${ZITADEL_ADMIN_USER}"
    echo "   • Password: ${ZITADEL_ADMIN_PASS}"
    echo
    echo "🔧 TACACS+ Configuration:"
    echo "   • Shared Secret: $(grep TACACS_SECRET .env | cut -d'=' -f2 2>/dev/null || echo 'testing123')"
    echo "   • Listen Address: $(grep TACACS_LISTEN_ADDRESS .env | cut -d'=' -f2 2>/dev/null || echo '0.0.0.0:49')"
    echo
    echo "⚙️  Management Commands:"
    echo "   • View logs: docker-compose -f docker-compose.zitadel.yml logs -f"
    echo "   • Stop services: docker-compose -f docker-compose.zitadel.yml down"
    echo "   • Restart services: docker-compose -f docker-compose.zitadel.yml restart"
    echo "   • Run tests: docker-compose -f docker-compose.zitadel.yml --profile testing up test-client"
    echo
    print_warning "⚠️  Next Steps:"
    echo "   1. Complete Zitadel configuration (see instructions above)"
    echo "   2. Update .env file with Zitadel credentials"
    echo "   3. Restart TACACS+ server"
    echo "   4. Test authentication"
}

# Main function
main() {
    echo "🚀 TACACS+ with Zitadel Setup Script"
    echo "====================================="
    echo
    
    check_prerequisites
    start_services
    configure_zitadel
    
    if [ "${1:-}" != "--no-tests" ]; then
        run_tests
    fi
    
    display_summary
    
    echo
    print_success "Setup completed successfully! 🎉"
    echo "Your TACACS+ system with Zitadel is ready for configuration."
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [options]"
        echo "Options:"
        echo "  --help, -h     Show this help message"
        echo "  --no-tests     Skip running tests"
        exit 0
        ;;
    --no-tests)
        main --no-tests
        ;;
    *)
        main
        ;;
esac