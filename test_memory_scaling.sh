#!/bin/bash
# Memory Scaling Test for Shine Widgets
# Tests actual memory usage with 1, 2, 3, and 4 widgets

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="$HOME/.config/shine/shine.toml"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Shine Memory Scaling Test ===${NC}"
echo ""
echo "This test measures actual memory usage with different widget counts."
echo "Each test runs for 5 seconds to allow processes to stabilize."
echo ""

# Function to write config
write_config() {
    local enable_chat=$1
    local enable_bar=$2
    local enable_clock=$3
    local enable_sysinfo=$4

    cat > "$CONFIG_FILE" <<EOF
# Memory scaling test configuration

[chat]
enabled = $enable_chat
edge = "bottom"
lines = 10
margin_left = 10
margin_right = 10
margin_bottom = 10
single_instance = false
hide_on_focus_loss = false
focus_policy = "always"

[bar]
enabled = $enable_bar
edge = "top"
lines = 1
margin_left = 10
margin_right = 10
margin_top = 10
single_instance = false

[clock]
enabled = $enable_clock
edge = "bottom-left"
lines_pixels = 200
columns_pixels = 200
margin_left = 10
margin_bottom = 10
single_instance = false

[sysinfo]
enabled = $enable_sysinfo
edge = "bottom-right"
lines_pixels = 200
columns_pixels = 300
margin_right = 10
margin_bottom = 10
single_instance = false
EOF
}

# Function to measure memory
measure_memory() {
    local test_name=$1
    local widget_count=$2

    echo -e "\n${YELLOW}[$test_name - $widget_count widgets]${NC}"

    # Launch shine
    "$SCRIPT_DIR/bin/shine" &
    SHINE_PID=$!

    # Wait for processes to stabilize
    sleep 5

    # Count Kitty panel processes
    KITTY_PIDS=$(pgrep -f "kitten panel" || echo "")

    if [ -z "$KITTY_PIDS" ]; then
        echo -e "${RED}ERROR: No Kitty panel processes found${NC}"
        kill $SHINE_PID 2>/dev/null || true
        return 1
    fi

    KITTY_COUNT=$(echo "$KITTY_PIDS" | wc -l)

    # Get memory usage
    echo "Process details:"
    echo "$KITTY_PIDS" | while read pid; do
        if [ -n "$pid" ]; then
            ps -o pid,rss,vsz,cmd -p $pid --no-headers 2>/dev/null || true
        fi
    done

    # Sum RSS (resident set size - actual RAM used)
    TOTAL_RSS=0
    echo "$KITTY_PIDS" | while read pid; do
        if [ -n "$pid" ]; then
            RSS=$(ps -o rss -p $pid --no-headers 2>/dev/null | awk '{print $1}' || echo "0")
            TOTAL_RSS=$((TOTAL_RSS + RSS))
        fi
    done

    # Get total RSS properly
    TOTAL_RSS=$(echo "$KITTY_PIDS" | xargs ps -o rss -p 2>/dev/null | awk 'NR>1 {sum+=$1} END {print sum}')

    # Convert to MB
    TOTAL_MB=$((TOTAL_RSS / 1024))
    PER_WIDGET=$((TOTAL_MB / KITTY_COUNT))

    echo ""
    echo -e "${GREEN}Results:${NC}"
    echo "  Kitty processes: $KITTY_COUNT"
    echo "  Total memory: ${TOTAL_MB} MB"
    echo "  Per-process average: ${PER_WIDGET} MB"

    # Save results for analysis
    echo "$widget_count,$KITTY_COUNT,$TOTAL_MB,$PER_WIDGET" >> /tmp/shine_memory_test.csv

    # Cleanup
    kill $SHINE_PID 2>/dev/null || true
    sleep 1
    pkill -f "shine" 2>/dev/null || true
    pkill -f "kitten panel" 2>/dev/null || true
    sleep 1
}

# Initialize results file
echo "widgets,processes,total_mb,per_process_mb" > /tmp/shine_memory_test.csv

# Test 1: Chat only
write_config true false false false
measure_memory "Test 1: Chat Only" 1

# Test 2: Chat + Bar
write_config true true false false
measure_memory "Test 2: Chat + Bar" 2

# Test 3: Chat + Bar + Clock
write_config true true true false
measure_memory "Test 3: Chat + Bar + Clock" 3

# Test 4: All widgets
write_config true true true true
measure_memory "Test 4: All Widgets" 4

# Analysis
echo ""
echo -e "${BLUE}=== Memory Scaling Analysis ===${NC}"
echo ""

# Read results and calculate scaling
cat /tmp/shine_memory_test.csv | tail -n +2 | while IFS=, read widgets processes total_mb per_mb; do
    echo "Widgets: $widgets | Processes: $processes | Total: ${total_mb}MB | Per-process: ${per_mb}MB"
done

echo ""
echo "Detailed results saved to: /tmp/shine_memory_test.csv"
echo ""

# Calculate scaling factor
FIRST_TOTAL=$(awk -F, 'NR==2 {print $3}' /tmp/shine_memory_test.csv)
LAST_TOTAL=$(awk -F, 'NR==5 {print $3}' /tmp/shine_memory_test.csv)
LAST_WIDGETS=$(awk -F, 'NR==5 {print $1}' /tmp/shine_memory_test.csv)

if [ -n "$FIRST_TOTAL" ] && [ -n "$LAST_TOTAL" ] && [ "$FIRST_TOTAL" -gt 0 ]; then
    EXPECTED_LINEAR=$((FIRST_TOTAL * LAST_WIDGETS))
    EFFICIENCY=$(echo "scale=2; $EXPECTED_LINEAR / $LAST_TOTAL" | bc)

    echo -e "${YELLOW}Scaling Analysis:${NC}"
    echo "  1 widget baseline: ${FIRST_TOTAL}MB"
    echo "  4 widgets actual: ${LAST_TOTAL}MB"
    echo "  4 widgets expected (linear): ${EXPECTED_LINEAR}MB"
    echo "  Memory efficiency: ${EFFICIENCY}x"
    echo ""

    if (( $(echo "$EFFICIENCY > 1.2" | bc -l) )); then
        echo -e "${GREEN}✓ Good: Sublinear scaling detected (${EFFICIENCY}x efficiency)${NC}"
        echo "  OS is sharing some memory between processes"
    elif (( $(echo "$EFFICIENCY > 1.0" | bc -l) )); then
        echo -e "${YELLOW}⚠ Fair: Some memory sharing (${EFFICIENCY}x efficiency)${NC}"
        echo "  Minor memory sharing occurring"
    else
        echo -e "${RED}✗ Bad: Pure linear scaling (${EFFICIENCY}x efficiency)${NC}"
        echo "  No memory sharing between processes"
    fi
fi

echo ""
echo -e "${BLUE}Test complete!${NC}"
