# Prototype Game Backend API Reference

## API Overview

The Prototype Game Backend provides real-time multiplayer game functionality through WebSocket connections and HTTP endpoints. The API is designed for low-latency communication and efficient handling of game state updates.

### Transport Protocols

- **Primary**: WebSocket for real-time bidirectional communication
- **Secondary**: HTTP/HTTPS for health checks and configuration
- **Internal**: gRPC for service-to-service communication (not client-accessible)

### General Conventions

- **Message Format**: JSON for WebSocket messages, standard HTTP for REST endpoints
- **Authentication**: JWT-based authentication for secure connections
- **Error Handling**: Standardized error codes and descriptive messages
- **Versioning**: API version specified in connection headers

## Authentication Flow

### JWT Token Authentication

All client connections require valid JWT authentication tokens.

#### 1. Token Acquisition

**Endpoint**: `POST /auth/login`
```http
POST /auth/login HTTP/1.1
Host: gateway-server:9090
Content-Type: application/json

{
  "username": "player123",
  "password": "secure_password"
}
```

**Response**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2024-01-15T15:30:00Z",
  "player_id": "player_123",
  "session_id": "sess_abc123"
}
```

#### 2. Token Validation

**Endpoint**: `GET /auth/validate`
```http
GET /auth/validate HTTP/1.1
Host: gateway-server:9090
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response**:
```json
{
  "valid": true,
  "player_id": "player_123",
  "expires_at": "2024-01-15T15:30:00Z"
}
```

#### 3. Token Refresh

**Endpoint**: `POST /auth/refresh`
```http
POST /auth/refresh HTTP/1.1
Host: gateway-server:9090
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2024-01-15T16:30:00Z"
}
```

## WebSocket Protocol

### Connection Establishment

#### 1. WebSocket Handshake

**URL**: `ws://gateway-server:9090/ws`

**Headers**:
```
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==
Sec-WebSocket-Version: 13
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### 2. Connection Confirmation

**Server Response**:
```json
{
  "type": "connection_established",
  "data": {
    "session_id": "sess_abc123",
    "player_id": "player_123",
    "server_time": "2024-01-15T10:30:00.000Z",
    "protocol_version": "1.0"
  }
}
```

#### 3. Authentication Verification

**Client Message**:
```json
{
  "type": "authenticate",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**Server Response**:
```json
{
  "type": "authentication_result",
  "data": {
    "success": true,
    "player_id": "player_123",
    "spawn_position": {"x": 0, "y": 0},
    "game_state": "active"
  }
}
```

### Message Types & Formats

All WebSocket messages follow this structure:
```json
{
  "type": "message_type",
  "timestamp": "2024-01-15T10:30:00.000Z",
  "sequence": 12345,
  "data": {
    // Message-specific payload
  }
}
```

#### Message Properties

- **type**: String identifier for the message type
- **timestamp**: ISO 8601 timestamp when message was sent
- **sequence**: Monotonically increasing sequence number for ordering
- **data**: Message-specific payload object

## Client → Server Messages

### Player Movement

#### move_player
Update player position and movement state.

```json
{
  "type": "move_player",
  "data": {
    "position": {"x": 150.5, "y": 200.0},
    "velocity": {"x": 5.0, "y": 0.0},
    "direction": 45.0,
    "movement_state": "walking"
  }
}
```

**Parameters**:
- `position`: Player's new position coordinates
- `velocity`: Current velocity vector (units per second)
- `direction`: Facing direction in degrees (0-360)
- `movement_state`: One of `"idle"`, `"walking"`, `"running"`

### Player Actions

#### player_action
Execute game-specific player actions.

```json
{
  "type": "player_action",
  "data": {
    "action": "use_item",
    "target": {
      "type": "entity",
      "entity_id": "entity_456"
    },
    "parameters": {
      "item_id": "sword_001",
      "intensity": 0.8
    }
  }
}
```

**Common Actions**:
- `"use_item"`: Use an inventory item
- `"interact"`: Interact with world object
- `"attack"`: Attack target entity
- `"chat"`: Send chat message

### Chat Messages

#### chat_message
Send chat messages to other players.

```json
{
  "type": "chat_message",
  "data": {
    "channel": "global",
    "message": "Hello, world!",
    "recipients": ["player_456", "player_789"]
  }
}
```

**Chat Channels**:
- `"global"`: Visible to all players in the area
- `"private"`: Direct message to specific players
- `"system"`: Server announcements (read-only)

### Connection Management

#### heartbeat
Maintain connection health.

```json
{
  "type": "heartbeat",
  "data": {
    "client_time": "2024-01-15T10:30:00.000Z"
  }
}
```

#### disconnect
Graceful disconnection notification.

```json
{
  "type": "disconnect",
  "data": {
    "reason": "user_logout"
  }
}
```

## Server → Client Messages

### World Updates

#### world_update
Periodic updates of game world state.

```json
{
  "type": "world_update",
  "data": {
    "tick": 12345,
    "entities": [
      {
        "entity_id": "player_456",
        "type": "player",
        "position": {"x": 100.0, "y": 150.0},
        "velocity": {"x": 2.5, "y": 0.0},
        "state": {
          "health": 100,
          "level": 5,
          "equipment": ["sword_001", "armor_002"]
        }
      }
    ],
    "events": [
      {
        "event_type": "entity_spawn",
        "entity_id": "npc_789",
        "position": {"x": 200.0, "y": 200.0}
      }
    ]
  }
}
```

### Player Updates

#### player_update
Updates specific to the connected player.

```json
{
  "type": "player_update",
  "data": {
    "player_id": "player_123",
    "stats": {
      "health": 95,
      "mana": 50,
      "experience": 1250
    },
    "inventory": [
      {"item_id": "sword_001", "quantity": 1},
      {"item_id": "potion_health", "quantity": 3}
    ],
    "position": {"x": 150.5, "y": 200.0}
  }
}
```

### Game Events

#### game_event
Significant game events requiring client attention.

```json
{
  "type": "game_event",
  "data": {
    "event_type": "player_died",
    "source_entity": "monster_001",
    "target_entity": "player_456",
    "details": {
      "damage_dealt": 50,
      "death_location": {"x": 180.0, "y": 220.0}
    }
  }
}
```

**Common Event Types**:
- `"player_died"`: Player has been defeated
- `"item_pickup"`: Item collected by player
- `"level_up"`: Player gained a level
- `"quest_completed"`: Quest objective completed

### Chat and Communication

#### chat_broadcast
Chat messages from other players.

```json
{
  "type": "chat_broadcast",
  "data": {
    "sender_id": "player_456",
    "sender_name": "GameMaster",
    "channel": "global",
    "message": "Welcome to the game!",
    "timestamp": "2024-01-15T10:30:00.000Z"
  }
}
```

### System Messages

#### system_message
Important system notifications.

```json
{
  "type": "system_message",
  "data": {
    "severity": "info",
    "message": "Server maintenance in 10 minutes",
    "action_required": false,
    "auto_dismiss": true,
    "dismiss_after": 30000
  }
}
```

**Severity Levels**:
- `"info"`: General information
- `"warning"`: Important notices
- `"error"`: Error conditions
- `"critical"`: Critical system issues

### Connection Status

#### heartbeat_response
Server response to client heartbeat.

```json
{
  "type": "heartbeat_response",
  "data": {
    "server_time": "2024-01-15T10:30:00.000Z",
    "latency_ms": 45
  }
}
```

## Development Endpoints

These endpoints are available for debugging and development purposes.

### Health Check

**Endpoint**: `GET /health`
```http
GET /health HTTP/1.1
Host: gateway-server:9090
```

**Response**:
```json
{
  "status": "healthy",
  "uptime": "2h30m15s",
  "version": "1.0.0",
  "connections": 42,
  "simulation": {
    "status": "connected",
    "latency_ms": 2.5
  }
}
```

### Metrics

**Endpoint**: `GET /metrics`
```http
GET /metrics HTTP/1.1
Host: gateway-server:9090
```

**Response** (Prometheus format):
```
# HELP gateway_connections_total Current number of WebSocket connections
# TYPE gateway_connections_total gauge
gateway_connections_total 42

# HELP gateway_messages_total Total number of messages processed
# TYPE gateway_messages_total counter
gateway_messages_total{type="player_move"} 12547
gateway_messages_total{type="chat_message"} 892

# HELP gateway_latency_ms Average message processing latency
# TYPE gateway_latency_ms histogram
gateway_latency_ms_bucket{le="1"} 8934
gateway_latency_ms_bucket{le="5"} 12450
gateway_latency_ms_bucket{le="10"} 12467
```

### Debug Information

**Endpoint**: `GET /debug/players`
```http
GET /debug/players HTTP/1.1
Host: gateway-server:9090
Authorization: Bearer admin_token_here
```

**Response**:
```json
{
  "total_players": 42,
  "players": [
    {
      "player_id": "player_123",
      "session_id": "sess_abc123",
      "connected_at": "2024-01-15T10:15:00Z",
      "last_activity": "2024-01-15T10:29:45Z",
      "position": {"x": 150.5, "y": 200.0},
      "shard": "shard_001"
    }
  ]
}
```

## Configuration Reference

### Gateway Service Configuration

#### Command Line Flags

```bash
./bin/gateway [flags]
```

**Available Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-bind-addr` | string | `:9090` | WebSocket server bind address |
| `-sim-addr` | string | `localhost:8080` | Simulation engine address |
| `-max-connections` | int | `1000` | Maximum concurrent connections |
| `-heartbeat-interval` | duration | `30s` | Connection heartbeat interval |
| `-read-timeout` | duration | `60s` | WebSocket read timeout |
| `-write-timeout` | duration | `10s` | WebSocket write timeout |
| `-log-level` | string | `info` | Logging level (debug, info, warn, error) |
| `-jwt-secret` | string | - | JWT signing secret (required) |
| `-cors-origins` | string | `*` | Allowed CORS origins |

#### Environment Variables

```bash
# Gateway configuration via environment
export GATEWAY_BIND_ADDR=":9090"
export GATEWAY_SIM_ADDR="localhost:8080"
export GATEWAY_MAX_CONNECTIONS="1000"
export GATEWAY_JWT_SECRET="your-secret-key"
export GATEWAY_LOG_LEVEL="info"
```

### Simulation Engine Configuration

#### Command Line Flags

```bash
./bin/simulation [flags]
```

**Available Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-bind-addr` | string | `:8080` | Server bind address |
| `-cell-size` | int | `500` | Spatial cell size in units |
| `-tick-rate` | int | `20` | Simulation updates per second |
| `-max-entities` | int | `10000` | Maximum entities per shard |
| `-world-size` | string | `10000x10000` | World dimensions |
| `-log-level` | string | `info` | Logging level |
| `-db-url` | string | - | Database connection URL (optional) |

## Error Codes & Handling

### WebSocket Error Codes

| Code | Name | Description | Client Action |
|------|------|-------------|---------------|
| 4000 | Invalid Message | Malformed JSON or unknown message type | Retry with valid message |
| 4001 | Authentication Failed | Invalid or expired JWT token | Re-authenticate |
| 4002 | Rate Limited | Too many messages sent | Slow down message rate |
| 4003 | Player Not Found | Player ID not found in session | Reconnect and re-authenticate |
| 4004 | Invalid Action | Action not allowed in current state | Check game state |
| 4005 | World Full | Maximum player capacity reached | Try again later |

### HTTP Error Codes

| Code | Description | Common Causes |
|------|-------------|---------------|
| 400 | Bad Request | Invalid JSON, missing parameters |
| 401 | Unauthorized | Missing or invalid authentication |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Endpoint or resource not found |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Server-side error |
| 503 | Service Unavailable | Server overloaded or maintenance |

### Error Response Format

```json
{
  "error": {
    "code": 4001,
    "message": "Authentication failed",
    "details": "JWT token has expired",
    "timestamp": "2024-01-15T10:30:00.000Z",
    "retry_after": 5000
  }
}
```

### Client Error Handling

#### Connection Errors

```javascript
websocket.onerror = function(error) {
  console.error('WebSocket error:', error);
  // Implement exponential backoff reconnection
  setTimeout(() => reconnect(), getBackoffDelay());
};

websocket.onclose = function(event) {
  if (event.code >= 4000 && event.code < 5000) {
    // Server-initiated close, handle specific error
    handleServerError(event.code, event.reason);
  } else {
    // Network issue, attempt reconnection
    scheduleReconnect();
  }
};
```

#### Message Errors

```javascript
function handleServerMessage(message) {
  const data = JSON.parse(message.data);
  
  if (data.type === 'error') {
    switch (data.error.code) {
      case 4001: // Authentication failed
        refreshAuthToken().then(reconnect);
        break;
      case 4002: // Rate limited
        implementRateLimit(data.error.retry_after);
        break;
      default:
        console.error('Server error:', data.error);
    }
  }
}
```

## Protocol Examples

### Complete Connection Flow

#### 1. Initial Connection

```javascript
// Client-side connection establishment
const ws = new WebSocket('ws://localhost:9090/ws', [], {
  headers: {
    'Authorization': 'Bearer ' + jwtToken
  }
});

ws.onopen = function() {
  console.log('WebSocket connected');
};

ws.onmessage = function(event) {
  const message = JSON.parse(event.data);
  handleMessage(message);
};
```

#### 2. Authentication and Setup

```javascript
// Send authentication message
ws.send(JSON.stringify({
  type: 'authenticate',
  data: {
    token: jwtToken
  }
}));

// Handle authentication response
function handleMessage(message) {
  switch (message.type) {
    case 'authentication_result':
      if (message.data.success) {
        console.log('Authenticated as:', message.data.player_id);
        startHeartbeat();
        setupGameLoop();
      } else {
        console.error('Authentication failed');
        ws.close();
      }
      break;
  }
}
```

#### 3. Game Loop Implementation

```javascript
// Regular heartbeat
function startHeartbeat() {
  setInterval(() => {
    ws.send(JSON.stringify({
      type: 'heartbeat',
      data: {
        client_time: new Date().toISOString()
      }
    }));
  }, 30000); // Every 30 seconds
}

// Player movement update
function sendPlayerMovement(position, velocity) {
  ws.send(JSON.stringify({
    type: 'move_player',
    data: {
      position: position,
      velocity: velocity,
      direction: calculateDirection(velocity),
      movement_state: velocity.x === 0 && velocity.y === 0 ? 'idle' : 'walking'
    }
  }));
}

// Handle world updates
function handleWorldUpdate(data) {
  // Update local game state
  data.entities.forEach(entity => {
    updateEntityPosition(entity.entity_id, entity.position);
    updateEntityState(entity.entity_id, entity.state);
  });
  
  // Process events
  data.events.forEach(event => {
    processGameEvent(event);
  });
}
```

### Chat System Example

```javascript
// Send chat message
function sendChatMessage(message, channel = 'global') {
  ws.send(JSON.stringify({
    type: 'chat_message',
    data: {
      channel: channel,
      message: message
    }
  }));
}

// Handle incoming chat
function handleChatMessage(data) {
  const chatDiv = document.getElementById('chat');
  const messageElement = document.createElement('div');
  messageElement.innerHTML = `
    <span class="timestamp">${new Date(data.timestamp).toLocaleTimeString()}</span>
    <span class="sender">${data.sender_name}:</span>
    <span class="message">${data.message}</span>
  `;
  chatDiv.appendChild(messageElement);
  chatDiv.scrollTop = chatDiv.scrollHeight;
}
```

### Error Recovery Example

```javascript
class GameConnection {
  constructor(url, token) {
    this.url = url;
    this.token = token;
    this.reconnectDelay = 1000;
    this.maxReconnectDelay = 30000;
    this.connect();
  }
  
  connect() {
    this.ws = new WebSocket(this.url, [], {
      headers: { 'Authorization': 'Bearer ' + this.token }
    });
    
    this.ws.onopen = () => {
      console.log('Connected');
      this.reconnectDelay = 1000; // Reset delay on successful connection
      this.authenticate();
    };
    
    this.ws.onclose = (event) => {
      if (event.code === 4001) {
        // Authentication failed - refresh token
        this.refreshToken().then(() => this.reconnect());
      } else {
        // Network issue - reconnect with backoff
        this.reconnect();
      }
    };
    
    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };
  }
  
  reconnect() {
    console.log(`Reconnecting in ${this.reconnectDelay}ms...`);
    setTimeout(() => {
      this.connect();
      this.reconnectDelay = Math.min(this.reconnectDelay * 2, this.maxReconnectDelay);
    }, this.reconnectDelay);
  }
  
  async refreshToken() {
    const response = await fetch('/auth/refresh', {
      method: 'POST',
      headers: {
        'Authorization': 'Bearer ' + this.token
      }
    });
    
    if (response.ok) {
      const data = await response.json();
      this.token = data.token;
      return data.token;
    } else {
      throw new Error('Token refresh failed');
    }
  }
}
```

This API reference provides complete documentation for integrating with the Prototype Game Backend. For additional examples and advanced usage patterns, refer to the example implementations in the project repository.