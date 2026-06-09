#!/usr/bin/env bash
set -euo pipefail

# Integration test: boots the server, runs CLI commands end-to-end

PORT=19876
API_URL="http://localhost:${PORT}/v1"
DB_PATH="/tmp/cb-test-$$.db"
SERVER_PID=""
CB_CLI="./dist/cb-linux-amd64"

cleanup() {
    if [ -n "${SERVER_PID}" ]; then
        kill "${SERVER_PID}" 2>/dev/null || true
        wait "${SERVER_PID}" 2>/dev/null || true
    fi
    rm -f "${DB_PATH}"
}
trap cleanup EXIT

# Build if binary doesn't exist
if [ ! -f "${CB_CLI}" ]; then
    echo "Building CLI..."
    go build -o dist/cb-test .
    CB_CLI="./dist/cb-test"
fi

echo "Building server..."
go build -o dist/cb-server-test ./server/main.go

echo "Starting server on port ${PORT}..."
CB_DB_PATH="${DB_PATH}" CB_ADDR=":${PORT}" CB_JWT_SECRET="test-secret" \
    ./dist/cb-server-test &
SERVER_PID=$!

# Wait for server to be ready
echo "Waiting for server..."
for i in $(seq 1 30); do
    if curl -s "http://localhost:${PORT}/health" | grep -q '"ok"'; then
        echo "Server is ready."
        break
    fi
    if [ "$i" -eq 30 ]; then
        echo "Server failed to start"
        exit 1
    fi
    sleep 0.5
done

echo ""
echo "=== Integration Test Suite ==="
echo ""

# Configure CLI
mkdir -p /tmp/cb-test-config-$$
export HOME="/tmp/cb-test-config-$$"

# Test 1: Register
echo "Test 1: Register"
${CB_CLI} login --api-url "${API_URL}" --user "testuser" --password "password123" 2>&1
echo "  PASS: Registration successful"
echo ""

# Test 2: Send a snippet
echo "Test 2: Send snippet"
OUTPUT=$(${CB_CLI} --api-url "${API_URL}" send "hello world from integration test")
echo "  Output: ${OUTPUT}"
echo "  PASS: Send successful"
echo ""

# Test 3: Get latest snippet
echo "Test 3: Get latest snippet"
OUTPUT=$(${CB_CLI} --api-url "${API_URL}" get)
EXPECTED="hello world from integration test"
if [ "${OUTPUT}" = "${EXPECTED}" ]; then
    echo "  PASS: Got correct content"
else
    echo "  FAIL: Expected '${EXPECTED}', got '${OUTPUT}'"
    exit 1
fi
echo ""

# Test 4: Save with alias
echo "Test 4: Save snippet with alias"
OUTPUT=$(${CB_CLI} --api-url "${API_URL}" save mycmd "kubectl get pods -A")
echo "  Output: ${OUTPUT}"
echo "  PASS: Save with alias successful"
echo ""

# Test 5: Get by alias
echo "Test 5: Get snippet by alias"
OUTPUT=$(${CB_CLI} --api-url "${API_URL}" get mycmd)
EXPECTED="kubectl get pods -A"
if [ "${OUTPUT}" = "${EXPECTED}" ]; then
    echo "  PASS: Got correct aliased content"
else
    echo "  FAIL: Expected '${EXPECTED}', got '${OUTPUT}'"
    exit 1
fi
echo ""

# Test 6: List snippets
echo "Test 6: List snippets"
OUTPUT=$(${CB_CLI} --api-url "${API_URL}" list)
echo "  Output:"
echo "${OUTPUT}" | sed 's/^/    /'
if echo "${OUTPUT}" | grep -q "mycmd"; then
    echo "  PASS: Alias appears in list"
else
    echo "  FAIL: Alias not found in list"
    exit 1
fi
echo ""

# Test 7: Send with TTL
echo "Test 7: Send with TTL"
OUTPUT=$(${CB_CLI} --api-url "${API_URL}" send --ttl 1h "temp message")
echo "  Output: ${OUTPUT}"
echo "  PASS: TTL send successful"
echo ""

# Test 8: Delete snippet
echo "Test 8: Delete snippet"
OUTPUT=$(${CB_CLI} --api-url "${API_URL}" rm mycmd)
echo "  Output: ${OUTPUT}"
echo "  PASS: Delete successful"
echo ""

# Test 9: Verify deleted snippet is gone
echo "Test 9: Verify deletion"
OUTPUT=$(${CB_CLI} --api-url "${API_URL}" get mycmd 2>&1) && RC=0 || RC=$?
if [ "${RC}" -ne 0 ]; then
    echo "  PASS: Deleted snippet no longer accessible (exit code: ${RC})"
else
    echo "  FAIL: Deleted snippet still accessible"
    echo "  Output: ${OUTPUT}"
    exit 1
fi
echo ""

# Test 10: Encrypt and decrypt
echo "Test 10: Encrypt/Decrypt"
export CB_MASTER_PASS="test-master-password-123"
OUTPUT=$(${CB_CLI} --api-url "${API_URL}" send --encrypt "this is a secret")
echo "  Send output: ${OUTPUT}"

# Get the encrypted snippet and decrypt
OUTPUT=$(${CB_CLI} --api-url "${API_URL}" get)
EXPECTED="this is a secret"
if [ "${OUTPUT}" = "${EXPECTED}" ]; then
    echo "  PASS: Encrypt/decrypt cycle works"
else
    echo "  FAIL: Expected '${EXPECTED}', got '${OUTPUT}'"
    exit 1
fi
unset CB_MASTER_PASS
echo ""

echo "=== All integration tests passed! ==="
