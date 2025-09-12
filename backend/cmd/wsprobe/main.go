package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	nws "nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func main() {
	var (
		url   = flag.String("url", "ws://localhost:8081/ws", "websocket url")
		token = flag.String("token", "", "dev token from gateway /login")
		moveX = flag.Float64("move_x", 0, "movement intent x (-1..1)")
		moveZ = flag.Float64("move_z", 0, "movement intent z (-1..1)")
	)
	flag.Parse()
	if *token == "" {
		log.Fatal("-token is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, _, err := nws.Dial(ctx, *url, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close(nws.StatusNormalClosure, "bye")

	if err := wsjson.Write(ctx, c, map[string]string{"token": *token}); err != nil {
		log.Fatal(err)
	}
	// Read join_ack
	var raw json.RawMessage
	if err := wsjson.Read(ctx, c, &raw); err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(raw))

	// Optionally send one input and read one state
	if *moveX != 0 || *moveZ != 0 {
		in := map[string]any{
			"type": "input",
			"seq":  1,
			"dt":   0.05,
			"intent": map[string]float64{
				"x": *moveX,
				"z": *moveZ,
			},
		}
		if err := wsjson.Write(ctx, c, in); err != nil {
			log.Fatal(err)
		}
		// Read state
		stateCtx, cancelState := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancelState()
		if err := wsjson.Read(stateCtx, c, &raw); err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(raw))
	}
}
