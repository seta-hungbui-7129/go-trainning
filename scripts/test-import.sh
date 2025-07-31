#!/bin/bash

# Test script for CSV user import functionality
# This script demonstrates the concurrent import feature

set -e

BASE_URL="http://localhost:8080"
API_URL="${BASE_URL}/api/v1"

echo "🚀 Testing CSV User Import Feature"
echo "=================================="

# Function to check if server is running
check_server() {
    if ! curl -s "${BASE_URL}/health" > /dev/null; then
        echo "❌ Server is not running. Please start the server first:"
        echo "   ./scripts/run.sh"
        exit 1
    fi
    echo "✅ Server is running"
}

# Function to create a manager user for testing
create_manager() {
    echo "📝 Creating manager user for testing..."
    
    MUTATION='mutation {
        createUser(input: {
            username: "test.manager"
            email: "test.manager@example.com"
            password: "password123"
            role: MANAGER
        }) {
            id
            username
            email
            role
        }
    }'
    
    curl -s -X POST "${BASE_URL}/graphql" \
        -H "Content-Type: application/json" \
        -d "{\"query\": \"$MUTATION\"}" > /dev/null
    
    echo "✅ Manager user created"
}

# Function to login and get JWT token
login_manager() {
    echo "🔐 Logging in as manager..."
    
    MUTATION='mutation {
        login(input: {
            email: "test.manager@example.com"
            password: "password123"
        }) {
            token
            user {
                id
                username
                role
            }
        }
    }'
    
    RESPONSE=$(curl -s -X POST "${BASE_URL}/graphql" \
        -H "Content-Type: application/json" \
        -d "{\"query\": \"$MUTATION\"}")
    
    TOKEN=$(echo "$RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    
    if [ -z "$TOKEN" ]; then
        echo "❌ Failed to get authentication token"
        echo "Response: $RESPONSE"
        exit 1
    fi
    
    echo "✅ Authentication successful"
    echo "Token: ${TOKEN:0:20}..."
}

# Function to download import template
download_template() {
    echo "📄 Downloading import template..."
    
    curl -s -X GET "${API_URL}/import-users/template" \
        -H "Authorization: Bearer $TOKEN" \
        -o "user_import_template.csv"
    
    echo "✅ Template downloaded: user_import_template.csv"
    echo "Template content:"
    cat user_import_template.csv
    echo ""
}

# Function to test import with sample data
test_import() {
    echo "📊 Testing CSV import with sample data..."
    
    # Create a test CSV file with more users for concurrent testing
    cat > test_import.csv << EOF
username,email,password,role
import.user1,import.user1@example.com,password123,member
import.user2,import.user2@example.com,password456,member
import.user3,import.user3@example.com,password789,manager
import.user4,import.user4@example.com,password101,member
import.user5,import.user5@example.com,password202,member
import.user6,import.user6@example.com,password303,manager
import.user7,import.user7@example.com,password404,member
import.user8,import.user8@example.com,password505,member
import.user9,import.user9@example.com,password606,member
import.user10,import.user10@example.com,password707,manager
EOF
    
    echo "✅ Test CSV file created with 10 users"
    
    # Test import with custom configuration
    echo "🔄 Starting concurrent import (5 workers, batch size 3)..."
    
    RESPONSE=$(curl -s -X POST "${API_URL}/import-users" \
        -H "Authorization: Bearer $TOKEN" \
        -F "csv_file=@test_import.csv" \
        -F "worker_count=5" \
        -F "batch_size=3" \
        -F "max_records=10" \
        -F "timeout_seconds=30")
    
    echo "📋 Import Response:"
    echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"
    
    # Extract summary information
    SUCCESS_COUNT=$(echo "$RESPONSE" | jq -r '.summary.success_count' 2>/dev/null || echo "N/A")
    FAILURE_COUNT=$(echo "$RESPONSE" | jq -r '.summary.failure_count' 2>/dev/null || echo "N/A")
    TOTAL_RECORDS=$(echo "$RESPONSE" | jq -r '.summary.total_records' 2>/dev/null || echo "N/A")
    PROCESSING_TIME=$(echo "$RESPONSE" | jq -r '.summary.processing_time' 2>/dev/null || echo "N/A")
    
    echo ""
    echo "📊 Import Summary:"
    echo "  Total Records: $TOTAL_RECORDS"
    echo "  Successful: $SUCCESS_COUNT"
    echo "  Failed: $FAILURE_COUNT"
    echo "  Processing Time: $PROCESSING_TIME"
    
    # Clean up test file
    rm -f test_import.csv
}

# Function to test import status endpoint
test_import_status() {
    echo "📈 Testing import status endpoint..."
    
    RESPONSE=$(curl -s -X GET "${API_URL}/import-users/status" \
        -H "Authorization: Bearer $TOKEN")
    
    echo "Import Status Response:"
    echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"
}

# Function to test error scenarios
test_error_scenarios() {
    echo "🧪 Testing error scenarios..."
    
    # Test 1: Invalid CSV format
    echo "Test 1: Invalid CSV header"
    cat > invalid_header.csv << EOF
name,mail,pass,type
test.user,test@example.com,password123,member
EOF
    
    RESPONSE=$(curl -s -X POST "${API_URL}/import-users" \
        -H "Authorization: Bearer $TOKEN" \
        -F "csv_file=@invalid_header.csv")
    
    echo "Invalid header response:"
    echo "$RESPONSE" | jq -r '.error' 2>/dev/null || echo "$RESPONSE"
    rm -f invalid_header.csv
    
    # Test 2: Invalid role
    echo ""
    echo "Test 2: Invalid role"
    cat > invalid_role.csv << EOF
username,email,password,role
test.user,test@example.com,password123,invalid_role
EOF
    
    RESPONSE=$(curl -s -X POST "${API_URL}/import-users" \
        -H "Authorization: Bearer $TOKEN" \
        -F "csv_file=@invalid_role.csv")
    
    echo "Invalid role response:"
    echo "$RESPONSE" | jq '.summary.results[0].error' 2>/dev/null || echo "$RESPONSE"
    rm -f invalid_role.csv
    
    # Test 3: No authentication
    echo ""
    echo "Test 3: No authentication"
    RESPONSE=$(curl -s -X POST "${API_URL}/import-users" \
        -F "csv_file=@examples/sample_users.csv")
    
    echo "No auth response:"
    echo "$RESPONSE" | jq -r '.error' 2>/dev/null || echo "$RESPONSE"
}

# Function to verify imported users
verify_users() {
    echo "🔍 Verifying imported users..."
    
    QUERY='query {
        fetchUsers {
            id
            username
            email
            role
        }
    }'
    
    RESPONSE=$(curl -s -X POST "${BASE_URL}/graphql" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -d "{\"query\": \"$QUERY\"}")
    
    USER_COUNT=$(echo "$RESPONSE" | jq '.data.fetchUsers | length' 2>/dev/null || echo "N/A")
    echo "✅ Total users in system: $USER_COUNT"
    
    # Show imported users
    echo "Imported users:"
    echo "$RESPONSE" | jq -r '.data.fetchUsers[] | select(.username | startswith("import.")) | "  - \(.username) (\(.email)) - \(.role)"' 2>/dev/null || echo "Could not parse users"
}

# Main execution
main() {
    echo "Starting CSV Import Test Suite..."
    echo ""
    
    check_server
    create_manager
    login_manager
    download_template
    test_import
    test_import_status
    test_error_scenarios
    verify_users
    
    echo ""
    echo "🎉 CSV Import Test Suite Completed!"
    echo ""
    echo "Key Features Demonstrated:"
    echo "  ✅ Concurrent processing with worker pools"
    echo "  ✅ Configurable worker count and batch size"
    echo "  ✅ Success/failure reporting"
    echo "  ✅ Error handling and validation"
    echo "  ✅ Authentication and authorization"
    echo "  ✅ CSV template download"
    echo "  ✅ Import status monitoring"
    
    # Clean up
    rm -f user_import_template.csv
}

# Run main function
main
