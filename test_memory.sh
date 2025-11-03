#!/bin/bash
# Memory test script for single-instance architecture

set -e

echo "=================================================="
echo "Shine Single-Instance Memory Test"
echo "=================================================="
echo

# Clean up any existing processes
echo "[1/5] Cleaning up existing processes..."
pkill -f "shine" || true
sleep 1
rm -f /tmp/shine*.sock* 2>/dev/null || true
echo "  ✓ Clean slate"
echo

# Launch shine with both components
echo "[2/5] Launching shine with chat and bar..."
./bin/shine &
SHINE_PID=$!
sleep 3
echo "  ✓ Shine launched (PID: $SHINE_PID)"
echo

# Count Kitty processes
echo "[3/5] Counting Kitty processes..."
KITTY_COUNT=$(pgrep -f "kitty.*panel" | wc -l)
echo "  Kitty panel processes: $KITTY_COUNT"

if [ "$KITTY_COUNT" -eq 1 ]; then
    echo "  ✓ SUCCESS: Only 1 Kitty process (shared instance working!)"
elif [ "$KITTY_COUNT" -eq 2 ]; then
    echo "  ✗ FAILURE: 2 Kitty processes (old architecture)"
else
    echo "  ? UNEXPECTED: $KITTY_COUNT Kitty processes"
fi
echo

# Check socket
echo "[4/5] Checking socket files..."
ls -lh /tmp/shine*.sock* 2>/dev/null || echo "  No sockets found"
SOCKET_COUNT=$(ls /tmp/shine*.sock* 2>/dev/null | wc -l)
echo "  Socket count: $SOCKET_COUNT"
echo

# Measure memory
echo "[5/5] Measuring memory usage..."
echo

# Get all Kitty panel processes
KITTY_PIDS=$(pgrep -f "kitty.*panel" || echo "")

if [ -z "$KITTY_PIDS" ]; then
    echo "  ✗ No Kitty processes found!"
else
    TOTAL_RSS=0
    for PID in $KITTY_PIDS; do
        if [ -f "/proc/$PID/status" ]; then
            RSS=$(grep "^VmRSS:" /proc/$PID/status | awk '{print $2}')
            RSS_MB=$((RSS / 1024))
            echo "  Kitty PID $PID: ${RSS_MB} MB RSS"
            TOTAL_RSS=$((TOTAL_RSS + RSS))
        fi
    done

    TOTAL_MB=$((TOTAL_RSS / 1024))
    echo
    echo "  TOTAL RSS: ${TOTAL_MB} MB"
    echo

    if [ "$KITTY_COUNT" -eq 1 ]; then
        echo "  ✓ Expected: ~230-270 MB for single instance"
    elif [ "$KITTY_COUNT" -eq 2 ]; then
        echo "  Expected: ~460-540 MB for dual instance (OLD)"
    fi
fi

echo
echo "=================================================="
echo "Test Results Summary"
echo "=================================================="
echo "Kitty processes: $KITTY_COUNT (expected: 1)"
echo "Total memory: ${TOTAL_MB:-0} MB"
echo

if [ "$KITTY_COUNT" -eq 1 ]; then
    echo "✓ PASS: Single-instance architecture working!"
    echo "  Memory savings: ~50% compared to old architecture"
else
    echo "✗ FAIL: Multiple Kitty instances detected"
    echo "  Single-instance mode not working correctly"
fi

echo
echo "Cleaning up..."
kill $SHINE_PID 2>/dev/null || true
sleep 1
pkill -f "kitty.*panel" || true

echo "=================================================="
