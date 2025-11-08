#!/usr/bin/env bash
# Automated test runner for prismctl Phase 1 MVP

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PASS=0
FAIL=0

log_test() {
    echo ""
    echo "========================================="
    echo "TEST: $1"
    echo "========================================="
}

log_pass() {
    echo -e "${GREEN}✓ PASS:${NC} $1"
    PASS=$((PASS + 1))
}

log_fail() {
    echo -e "${RED}✗ FAIL:${NC} $1"
    FAIL=$((FAIL + 1))
}

log_warn() {
    echo -e "${YELLOW}⚠ WARN:${NC} $1"
}

log_info() {
    echo "  $1"
}

cleanup() {
    echo "  Cleaning up..." >&2
    pkill -9 prismctl 2>/dev/null || true
    pkill -9 shine-clock 2>/dev/null || true
    pkill -9 shine-sysinfo 2>/dev/null || true
    sleep 0.5
    rm -f /run/user/$(id -u)/shine/prism-autotest.*.sock 2>/dev/null || true
}

trap cleanup EXIT

# Verify binaries exist
log_test "Prerequisites Check"

if [ ! -f "bin/prismctl" ]; then
    log_fail "bin/prismctl not found - run: go build -o bin/prismctl ./cmd/prismctl"
    exit 1
fi
log_pass "prismctl binary exists"

if [ ! -f "bin/shine-clock" ]; then
    log_fail "bin/shine-clock not found - run: go build -o bin/shine-clock ./cmd/shine-clock"
    exit 1
fi
log_pass "shine-clock binary exists"

if [ ! -f "bin/shine-sysinfo" ]; then
    log_fail "bin/shine-sysinfo not found - run: go build -o bin/shine-sysinfo ./cmd/shine-sysinfo"
    exit 1
fi
log_pass "shine-sysinfo binary exists"

export PATH="$PWD/bin:$PATH"

# Test 1: Basic launch
log_test "Test 1: Basic Launch"

cleanup
timeout 3 prismctl shine-clock autotest >/dev/null 2>&1 &
PRISMCTL_PID=$!
sleep 1

if ! ps -p $PRISMCTL_PID > /dev/null; then
    log_fail "prismctl exited unexpectedly"
else
    log_pass "prismctl started successfully (PID: $PRISMCTL_PID)"
fi

# Check if child process exists
CHILD_PID=$(pgrep -P $PRISMCTL_PID shine-clock || echo "")
if [ -z "$CHILD_PID" ]; then
    log_fail "Child process (shine-clock) not found"
else
    log_pass "Child process running (PID: $CHILD_PID)"
fi

# Check socket exists
SOCKET_PATH=$(ls /run/user/$(id -u)/shine/prism-autotest.*.sock 2>/dev/null | head -1 || echo "")
if [ -z "$SOCKET_PATH" ]; then
    log_fail "IPC socket not created"
else
    log_pass "IPC socket created: $(basename $SOCKET_PATH)"
fi

kill $PRISMCTL_PID 2>/dev/null || true
wait $PRISMCTL_PID 2>/dev/null || true
sleep 0.5

# Verify clean shutdown
if pgrep prismctl > /dev/null || pgrep shine-clock > /dev/null; then
    log_fail "Processes did not exit cleanly"
    cleanup
else
    log_pass "Clean shutdown successful"
fi

# Test 2: Hot-swap via IPC
log_test "Test 2: Hot-Swap via IPC"

cleanup
prismctl shine-clock autotest >/dev/null 2>&1 &
PRISMCTL_PID=$!
sleep 1

SOCKET=$(ls /run/user/$(id -u)/shine/prism-autotest.*.sock 2>/dev/null | head -1 || echo "")
if [ -z "$SOCKET" ]; then
    log_fail "Socket not found"
    kill $PRISMCTL_PID 2>/dev/null || true
else
    log_pass "Found socket: $(basename $SOCKET)"

    # Send swap command
    RESPONSE=$(echo '{"action":"swap","prism":"shine-sysinfo"}' | socat - "UNIX-CONNECT:$SOCKET" 2>/dev/null || echo "")

    if echo "$RESPONSE" | grep -q '"success":true'; then
        log_pass "Hot-swap command accepted"
        sleep 1

        # Verify new child is shine-sysinfo
        if pgrep -f shine-sysinfo > /dev/null; then
            log_pass "shine-sysinfo is now running"
        else
            log_fail "shine-sysinfo not running after swap"
        fi
    else
        log_fail "Hot-swap command failed: $RESPONSE"
    fi

    kill $PRISMCTL_PID 2>/dev/null || true
    wait $PRISMCTL_PID 2>/dev/null || true
fi

cleanup
sleep 0.5

# Test 3: SIGKILL recovery
log_test "Test 4: Crash Recovery (SIGKILL)"

cleanup
prismctl shine-clock autotest >/dev/null 2>&1 &
PRISMCTL_PID=$!
sleep 1

CHILD_PID=$(pgrep -P $PRISMCTL_PID shine-clock || echo "")
if [ -z "$CHILD_PID" ]; then
    log_fail "Child not running before SIGKILL test"
else
    log_pass "Child running (PID: $CHILD_PID)"

    # Kill child with SIGKILL
    kill -9 $CHILD_PID
    log_info "Sent SIGKILL to child"
    sleep 0.5

    # Verify prismctl still running
    if ps -p $PRISMCTL_PID > /dev/null; then
        log_pass "prismctl survived child SIGKILL"

        # Verify socket still works
        SOCKET=$(ls /run/user/$(id -u)/shine/prism-autotest.*.sock 2>/dev/null | head -1 || echo "")
        if [ -n "$SOCKET" ]; then
            RESPONSE=$(echo '{"action":"status"}' | socat - "UNIX-CONNECT:$SOCKET" 2>/dev/null || echo "")
            if echo "$RESPONSE" | grep -q '"success":true'; then
                log_pass "IPC still functional after crash"
            else
                log_fail "IPC not responding after crash"
            fi
        fi
    else
        log_fail "prismctl exited after child SIGKILL"
    fi
fi

kill $PRISMCTL_PID 2>/dev/null || true
wait $PRISMCTL_PID 2>/dev/null || true
cleanup

# Summary
echo ""
echo "========================================="
echo "TEST SUMMARY"
echo "========================================="
echo -e "${GREEN}Passed: $PASS${NC}"
echo -e "${RED}Failed: $FAIL${NC}"

if [ $FAIL -eq 0 ]; then
    echo -e "${GREEN}✓ All automated tests passed!${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Run manual tests from docs/prismtty/TESTING.md"
    echo "  2. Test in Kitty panel for real-world usage"
    echo "  3. Verify terminal state reset visually"
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    exit 1
fi
