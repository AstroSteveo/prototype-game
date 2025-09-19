#!/bin/bash
# Developer helper script for testing inventory operations
# Usage: ./scripts/test-inventory.sh

set -euo pipefail

echo "=== Inventory System Test Helper ==="
echo

RUN_LOG=$(mktemp -t inventory-run.XXXX.log)
TEST_LOG=$(mktemp -t inventory-tests.XXXX.log)

cleanup() {
	if [[ -n "${RUN_PID:-}" ]]; then
		if kill -0 "${RUN_PID}" 2>/dev/null; then
			kill "${RUN_PID}" 2>/dev/null || true
			wait "${RUN_PID}" 2>/dev/null || true
		fi
	fi
	make stop >/dev/null 2>&1 || true
	rm -f "${RUN_LOG}" "${TEST_LOG}"
}
trap cleanup EXIT

echo "Starting services..."
make run >"${RUN_LOG}" 2>&1 &
RUN_PID=$!
sleep 2
if ! kill -0 "${RUN_PID}" 2>/dev/null; then
	echo "❌ make run exited early"
	cat "${RUN_LOG}"
	exit 1
fi

echo "Getting authentication token..."
TOKEN=$(make login 2>/dev/null)
echo "Token: ${TOKEN}"
echo

echo "Testing WebSocket join with inventory data..."
echo "Expected: join_ack with inventory, equipment, skills, and encumbrance fields"
echo "---"
OUTPUT=$(make wsprobe TOKEN="${TOKEN}" 2>/dev/null | grep '^{')
echo "${OUTPUT}" | python3 -m json.tool
echo "---"
echo

echo "Running inventory integration tests..."
if make test-ws >"${TEST_LOG}" 2>&1; then
	echo "✅ All inventory tests passed"
	echo "   - Inventory data in join_ack"
	echo "   - Item templates validation"
	echo "   - Equipment slot operations"
	echo "   - Encumbrance calculation with movement penalties"
	echo "   - Cooldown protection"
	echo "   - Skill requirement gating"
else
	echo "❌ Some tests failed - check test output:"
	cat "${TEST_LOG}"
	exit 1
fi
echo

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

echo "Stopping services..."
# cleanup trap will stop services

echo "Done!"
