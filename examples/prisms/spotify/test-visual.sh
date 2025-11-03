#!/bin/bash
# Test visual appearance of Spotify prism
set -e

echo "Testing Spotify Prism Visual Appearance"
echo "========================================"
echo ""

# Test 1: Not running state
echo "Test 1: Spotify not running (should show friendly message)"
timeout 2 ./shine-spotify 2>&1 | head -20 || true
echo ""

# Test 2: Mock mode
echo "Test 2: Mock mode with sample track"
timeout 2 ./shine-spotify --mock 2>&1 | head -20 || true
echo ""

echo "Visual tests complete!"
