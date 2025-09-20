package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	nws "nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type GameClient struct {
	conn        *nws.Conn
	ctx         context.Context
	seq         int
	playerID    string
	inventory   map[string]interface{}
	equipment   map[string]interface{}
	skills      map[string]interface{}
	encumbrance map[string]interface{}
}

func main() {
	var (
		url         = flag.String("url", "ws://localhost:8081/ws", "websocket url")
		token       = flag.String("token", "", "dev token from gateway /login")
		moveX       = flag.Float64("move_x", 0, "movement intent x (-1..1)")
		moveZ       = flag.Float64("move_z", 0, "movement intent z (-1..1)")
		demo        = flag.Bool("demo", false, "run equipment demo")
		interactive = flag.Bool("interactive", false, "interactive equipment management mode")
	)
	flag.Parse()
	if *token == "" {
		log.Fatal("-token is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := &GameClient{ctx: ctx, seq: 1}

	if err := client.connect(*url, *token); err != nil {
		log.Fatal(err)
	}
	defer client.disconnect()

	if *demo {
		client.runEquipmentDemo()
	} else if *interactive {
		client.runInteractiveMode()
	} else {
		client.runBasicProbe(*moveX, *moveZ)
	}
}

func (c *GameClient) connect(url, token string) error {
	conn, _, err := nws.Dial(c.ctx, url, nil)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}
	c.conn = conn

	// Send hello
	hello := map[string]string{"token": token}
	if err := wsjson.Write(c.ctx, conn, hello); err != nil {
		return fmt.Errorf("send hello: %w", err)
	}

	// Read join_ack
	var response map[string]interface{}
	if err := wsjson.Read(c.ctx, conn, &response); err != nil {
		return fmt.Errorf("read join_ack: %w", err)
	}

	if response["type"] != "join_ack" {
		return fmt.Errorf("expected join_ack, got %v", response["type"])
	}

	data := response["data"].(map[string]interface{})
	c.playerID = data["player_id"].(string)
	c.inventory = data["inventory"].(map[string]interface{})
	c.equipment = data["equipment"].(map[string]interface{})
	c.skills = data["skills"].(map[string]interface{})
	c.encumbrance = data["encumbrance"].(map[string]interface{})

	fmt.Printf("Connected as player: %s\n", c.playerID)
	c.printPlayerStatus()

	return nil
}

func (c *GameClient) disconnect() {
	if c.conn != nil {
		c.conn.Close(nws.StatusNormalClosure, "demo complete")
	}
}

func (c *GameClient) printPlayerStatus() {
	fmt.Println("\n=== PLAYER STATUS ===")
	fmt.Printf("Player ID: %s\n", c.playerID)

	// Print encumbrance
	fmt.Println("\n--- Encumbrance ---")
	if c.encumbrance != nil {
		weight := c.encumbrance["current_weight"].(float64)
		maxWeight := c.encumbrance["max_weight"].(float64)
		penalty := c.encumbrance["movement_penalty"].(float64)
		fmt.Printf("Weight: %.1f/%.1f (%.1f%%) - Movement: %.1f%%\n",
			weight, maxWeight, (weight/maxWeight)*100, penalty*100)
	}

	// Print equipment
	fmt.Println("\n--- Equipment ---")
	if c.equipment != nil {
		slots := c.equipment["slots"].(map[string]interface{})
		if len(slots) == 0 {
			fmt.Println("No equipment equipped")
		} else {
			for slot, item := range slots {
				itemData := item.(map[string]interface{})
				instance := itemData["instance"].(map[string]interface{})
				fmt.Printf("%s: %s (ID: %s)\n", slot,
					instance["template_id"], instance["instance_id"])
			}
		}
	}

	// Print inventory
	fmt.Println("\n--- Inventory ---")
	if c.inventory != nil {
		items := c.inventory["items"].([]interface{})
		if len(items) == 0 {
			fmt.Println("Inventory is empty")
		} else {
			for _, item := range items {
				itemData := item.(map[string]interface{})
				instance := itemData["instance"].(map[string]interface{})
				compartment := itemData["compartment"].(string)
				fmt.Printf("[%s] %s (ID: %s) x%v\n",
					compartment, instance["template_id"],
					instance["instance_id"], instance["quantity"])
			}
		}
	}

	// Print skills
	fmt.Println("\n--- Skills ---")
	if c.skills != nil && len(c.skills) > 0 {
		for skill, level := range c.skills {
			fmt.Printf("%s: %v\n", skill, level)
		}
	} else {
		fmt.Println("No skills developed")
	}
	fmt.Println("=====================\n")
}

func (c *GameClient) runBasicProbe(moveX, moveZ float64) {
	fmt.Println("Running basic probe...")

	// Optionally send movement and read one state
	if moveX != 0 || moveZ != 0 {
		c.sendMovement(moveX, moveZ)
		c.readNextMessage()
	}
}

func (c *GameClient) runEquipmentDemo() {
	fmt.Println("🎮 Starting Equipment System Demo...")

	// Demo flow: This simulates a typical equipment interaction
	// Note: This demo assumes items are available via dev commands or pre-populated

	fmt.Println("\n1. Current player state:")
	c.printPlayerStatus()

	fmt.Println("2. Attempting to equip a test sword...")
	// Try to equip an item (this will fail if no items in inventory)
	c.sendEquipCommand("sword_001", "main_hand")
	time.Sleep(100 * time.Millisecond)
	c.readAllMessages()

	fmt.Println("\n3. Attempting to unequip from main hand...")
	c.sendUnequipCommand("main_hand", "backpack")
	time.Sleep(100 * time.Millisecond)
	c.readAllMessages()

	fmt.Println("\n4. Final state:")
	c.printPlayerStatus()

	fmt.Println("✅ Equipment demo complete!")
}

func (c *GameClient) runInteractiveMode() {
	fmt.Println("🎮 Interactive Equipment Management Mode")
	fmt.Println("Commands: equip <item_id> <slot> | unequip <slot> [compartment] | status | move <x> <z> | quit")
	fmt.Println("Example: equip sword_001 main_hand")
	fmt.Println("Example: unequip main_hand backpack")

	scanner := bufio.NewScanner(os.Stdin)

	// Start background message reader
	go c.readMessagesInBackground()

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		cmd := strings.TrimSpace(scanner.Text())
		if cmd == "" {
			continue
		}

		parts := strings.Fields(cmd)
		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "quit", "exit", "q":
			fmt.Println("Goodbye!")
			return
		case "status", "s":
			c.printPlayerStatus()
		case "equip":
			if len(parts) < 3 {
				fmt.Println("Usage: equip <item_id> <slot>")
				continue
			}
			c.sendEquipCommand(parts[1], parts[2])
		case "unequip":
			if len(parts) < 2 {
				fmt.Println("Usage: unequip <slot> [compartment]")
				continue
			}
			compartment := "backpack"
			if len(parts) >= 3 {
				compartment = parts[2]
			}
			c.sendUnequipCommand(parts[1], compartment)
		case "move":
			if len(parts) < 3 {
				fmt.Println("Usage: move <x> <z>")
				continue
			}
			x, err1 := strconv.ParseFloat(parts[1], 64)
			z, err2 := strconv.ParseFloat(parts[2], 64)
			if err1 != nil || err2 != nil {
				fmt.Println("Invalid coordinates")
				continue
			}
			c.sendMovement(x, z)
		case "help", "h":
			fmt.Println("Commands:")
			fmt.Println("  equip <item_id> <slot>    - Equip item to slot")
			fmt.Println("  unequip <slot> [compartment] - Unequip item to compartment")
			fmt.Println("  move <x> <z>              - Send movement intent")
			fmt.Println("  status                    - Show player status")
			fmt.Println("  quit                      - Exit")
		default:
			fmt.Printf("Unknown command: %s (try 'help')\n", parts[0])
		}
	}
}

func (c *GameClient) sendEquipCommand(itemID, slot string) {
	cmd := map[string]interface{}{
		"type":        "equip",
		"seq":         c.seq,
		"instance_id": itemID,
		"slot":        slot,
	}
	c.seq++

	fmt.Printf("Sending equip command: %s -> %s\n", itemID, slot)
	if err := wsjson.Write(c.ctx, c.conn, cmd); err != nil {
		fmt.Printf("Failed to send equip command: %v\n", err)
	}
}

func (c *GameClient) sendUnequipCommand(slot, compartment string) {
	cmd := map[string]interface{}{
		"type":        "unequip",
		"seq":         c.seq,
		"slot":        slot,
		"compartment": compartment,
	}
	c.seq++

	fmt.Printf("Sending unequip command: %s -> %s\n", slot, compartment)
	if err := wsjson.Write(c.ctx, c.conn, cmd); err != nil {
		fmt.Printf("Failed to send unequip command: %v\n", err)
	}
}

func (c *GameClient) sendMovement(x, z float64) {
	cmd := map[string]interface{}{
		"type": "input",
		"seq":  c.seq,
		"dt":   0.05,
		"intent": map[string]float64{
			"x": x,
			"z": z,
		},
	}
	c.seq++

	if err := wsjson.Write(c.ctx, c.conn, cmd); err != nil {
		fmt.Printf("Failed to send movement: %v\n", err)
	}
}

func (c *GameClient) readNextMessage() {
	var raw json.RawMessage
	readCtx, cancel := context.WithTimeout(c.ctx, 2*time.Second)
	defer cancel()

	if err := wsjson.Read(readCtx, c.conn, &raw); err != nil {
		fmt.Printf("Read error: %v\n", err)
		return
	}

	c.processMessage(raw)
}

func (c *GameClient) readAllMessages() {
	// Read multiple messages with short timeout
	for i := 0; i < 5; i++ {
		var raw json.RawMessage
		readCtx, cancel := context.WithTimeout(c.ctx, 200*time.Millisecond)
		err := wsjson.Read(readCtx, c.conn, &raw)
		cancel()

		if err != nil {
			break // Timeout or error, stop reading
		}

		c.processMessage(raw)
	}
}

func (c *GameClient) readMessagesInBackground() {
	for {
		var raw json.RawMessage
		if err := wsjson.Read(c.ctx, c.conn, &raw); err != nil {
			if c.ctx.Err() != nil {
				return // Context cancelled
			}
			fmt.Printf("\nConnection error: %v\n> ", err)
			return
		}

		c.processMessage(raw)
		fmt.Print("> ") // Re-prompt after processing message
	}
}

func (c *GameClient) processMessage(raw json.RawMessage) {
	var msg map[string]interface{}
	if err := json.Unmarshal(raw, &msg); err != nil {
		fmt.Printf("Failed to parse message: %v\n", err)
		return
	}

	msgType, ok := msg["type"].(string)
	if !ok {
		fmt.Printf("Message missing type: %v\n", msg)
		return
	}

	switch msgType {
	case "state":
		c.handleStateUpdate(msg)
	case "equipment_result":
		c.handleEquipmentResult(msg)
	case "error":
		c.handleError(msg)
	case "telemetry":
		c.handleTelemetry(msg)
	case "handover":
		c.handleHandover(msg)
	default:
		fmt.Printf("Unknown message type '%s': %v\n", msgType, msg)
	}
}

func (c *GameClient) handleStateUpdate(msg map[string]interface{}) {
	data, ok := msg["data"].(map[string]interface{})
	if !ok {
		return
	}

	// Update local state if deltas are present
	if inventory, exists := data["inventory"]; exists {
		c.inventory = inventory.(map[string]interface{})
		fmt.Println("📦 Inventory updated")
	}

	if equipment, exists := data["equipment"]; exists {
		c.equipment = equipment.(map[string]interface{})
		fmt.Println("⚔️  Equipment updated")
	}

	if skills, exists := data["skills"]; exists {
		c.skills = skills.(map[string]interface{})
		fmt.Println("🎯 Skills updated")
	}

	// Always print encumbrance if present
	if c.inventory != nil {
		if encumbrance, exists := c.inventory["encumbrance"]; exists {
			c.encumbrance = encumbrance.(map[string]interface{})
			weight := c.encumbrance["current_weight"].(float64)
			maxWeight := c.encumbrance["max_weight"].(float64)
			penalty := c.encumbrance["movement_penalty"].(float64)
			fmt.Printf("⚖️  Encumbrance: %.1f/%.1f (%.0f%% speed)\n",
				weight, maxWeight, penalty*100)
		}
	}
}

func (c *GameClient) handleEquipmentResult(msg map[string]interface{}) {
	data, ok := msg["data"].(map[string]interface{})
	if !ok {
		return
	}

	operation := data["operation"].(string)
	slot := data["slot"].(string)
	success := data["success"].(bool)
	code := data["code"].(string)
	message := data["message"].(string)

	status := "❌"
	if success {
		status = "✅"
	}

	fmt.Printf("%s %s %s: %s (%s)\n", status, operation, slot, message, code)
}

func (c *GameClient) handleError(msg map[string]interface{}) {
	data, ok := msg["data"].(map[string]interface{})
	if !ok {
		fmt.Printf("❌ Error: %v\n", msg)
		return
	}

	code := data["code"].(string)
	message := data["message"].(string)
	fmt.Printf("❌ Error [%s]: %s\n", code, message)
}

func (c *GameClient) handleTelemetry(msg map[string]interface{}) {
	data, ok := msg["data"].(map[string]interface{})
	if !ok {
		return
	}

	rtt := data["rtt_ms"].(float64)
	tickRate := data["tick_rate"].(float64)
	fmt.Printf("📊 RTT: %.1fms, Tick Rate: %.0fHz\n", rtt, tickRate)
}

func (c *GameClient) handleHandover(msg map[string]interface{}) {
	data, ok := msg["data"].(map[string]interface{})
	if !ok {
		return
	}

	from := data["from"]
	to := data["to"]
	fmt.Printf("🔄 Handover: %v -> %v\n", from, to)
}
