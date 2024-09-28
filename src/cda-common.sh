#!/bin/bash
# File: /usr/local/lib/cda-common.sh

set -e  # Exit immediately if a command exits with a non-zero status

# Color definitions
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print a header message in blue
print_header() {
    echo -e "${BLUE}=== $1 ===${NC}"
}

# Print a success message in green
print_success() {
    echo -e "${GREEN}✔ $1${NC}"
}

# Print an error message in red
print_error() {
    echo -e "${RED}✖ $1${NC}"
}

# Print a warning message in yellow
print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# Check the status of the last command and print a message
check_command_status() {
    local command_name="$1"
    if [ $? -eq 0 ]; then
        print_success "$command_name completed successfully"
    else
        print_error "$command_name failed"
        exit 1
    fi
}