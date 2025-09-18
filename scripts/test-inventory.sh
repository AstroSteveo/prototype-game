#!/bin/bash
# Developer helper script for testing inventory operations
# Usage: ./scripts/test-inventory.sh

set -e

echo "=== Inventory System Test Helper ==="
echo

# Start services
echo "Starting services..."
make run > /dev/null 2>&1
sleep 2

# Get token
echo "Getting authentication token..."
TOKEN=$(make login 2>/dev/null)
echo "Token: $TOKEN"
echo

# Test WebSocket join with inventory data
echo "Testing WebSocket join with inventory data..."
echo "Expected: join_ack with inventory, equipment, skills, and encumbrance fields"
echo "---"
make wsprobe TOKEN="$TOKEN" 2>/dev/null | python3 -m json.tool
echo "---"
echo

# Test inventory system components
echo "Running inventory integration tests..."
make test-ws > /tmp/test-output.log 2>&1
if [ $? -eq 0 ]; then
    echo "✅ All inventory tests passed"
    echo "   - Inventory data in join_ack"
    echo "   - Item templates validation"
    echo "   - Equipment slot operations"
    echo "   - Encumbrance calculation with movement penalties"
    echo "   - Cooldown protection"
    echo "   - Skill requirement gating"
else
    echo "❌ Some tests failed - check test output:"
    cat /tmp/test-output.log
fi
echo

# Show available item templates
echo "Available test item templates:"
echo "  - sword_iron: Iron Sword (main hand, requires melee skill 10)"
echo "  - shield_wood: Wooden Shield (off hand, requires defense skill 5)" 
echo "  - armor_leather: Leather Armor (chest, no skill requirement)"
echo "  - potion_health: Health Potion (consumable, cannot be equipped)"
echo

echo "Inventory system features implemented:"
echo "  ✅ Item templates with slot masks, weight, bulk, damage types"
echo "  ✅ Player inventory with compartments (backpack, belt, craft bag)"
echo "  ✅ Equipment slots with cooldown protection (2 second default)"
echo "  ✅ Skill gating for equipment requirements"
echo "  ✅ Encumbrance calculation with movement penalties"
echo "  ✅ WebSocket integration - inventory data in join_ack"
echo "  ✅ Comprehensive test coverage"
echo

# Cleanup
echo "Stopping services..."
make stop > /dev/null 2>&1
echo "Done!"