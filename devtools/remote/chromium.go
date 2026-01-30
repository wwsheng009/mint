// Package remote provides remote debugging support for DevTools.
package remote

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/wwsheng009/mint/devtools"
	"github.com/wwsheng009/mint/devtools/snapshot"
)

// =============================================================================
// Chromium DevTools Bridge
// =============================================================================

// ChromiumBridge bridges TUI DevTools with Chromium DevTools Protocol.
type ChromiumBridge struct {
	mu         sync.RWMutex
	server     *Server
	devtools   *devtools.DevTools
	snapshots  *snapshot.Manager
	enabled    bool
	inspectURL string
}

// NewChromiumBridge creates a new Chromium DevTools bridge.
func NewChromiumBridge(dt *devtools.DevTools, sm *snapshot.Manager) *ChromiumBridge {
	return &ChromiumBridge{
		devtools:  dt,
		snapshots: sm,
		server:    NewServer(9222, "/debug"),
		enabled:   false,
	}
}

// Enable enables the bridge.
func (b *ChromiumBridge) Enable() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.enabled = true
	b.server.Enable()

	// Set up message handler
	b.server.msgHandler = b.handleMessage

	return nil
}

// Disable disables the bridge.
func (b *ChromiumBridge) Disable() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.enabled = false
	b.server.Disable()
}

// IsEnabled returns true if the bridge is enabled.
func (b *ChromiumBridge) IsEnabled() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.enabled
}

// GetInspectURL returns the URL for Chromium DevTools inspection.
func (b *ChromiumBridge) GetInspectURL() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.inspectURL
}

// handleMessage handles incoming messages from the client.
func (b *ChromiumBridge) handleMessage(session *Session, msg *Message) *Message {
	session.UpdateActivity()

	switch msg.Type {
	case TypeHandshake:
		return b.handleHandshake(session, msg)
	case TypeGetSnapshot:
		return b.handleGetSnapshot(session, msg)
	case TypeGetRange:
		return b.handleGetRange(session, msg)
	case TypeGetDiff:
		return b.handleGetDiff(session, msg)
	case TypeSubscribe:
		return b.handleSubscribe(session, msg)
	case TypeUnsubscribe:
		return b.handleUnsubscribe(session, msg)
	case TypeSetBreakpoint:
		return b.handleSetBreakpoint(session, msg)
	case TypeClearBreakpoint:
		return b.handleClearBreakpoint(session, msg)
	default:
		return NewError(msg.ID, fmt.Sprintf("unknown message type: %s", msg.Type))
	}
}

// handleHandshake handles the handshake message.
func (b *ChromiumBridge) handleHandshake(session *Session, msg *Message) *Message {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return NewError(msg.ID, "invalid handshake payload")
	}

	clientID, _ := payload["client_id"].(string)
	if clientID != "" {
		session.clientID = clientID
	}

	ack := HandshakeAckPayload{
		ServerID:     "mint-devtools",
		Version:      ProtocolVersion,
		Capabilities: []string{"snapshots", "events", "diffs", "breakpoints"},
		SessionID:    session.ID(),
	}

	return NewMessageWithID(msg.ID, TypeHandshakeAck, ack)
}

// handleGetSnapshot handles snapshot requests.
func (b *ChromiumBridge) handleGetSnapshot(session *Session, msg *Message) *Message {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return NewError(msg.ID, "invalid payload")
	}

	frameIDFloat, _ := payload["frame_id"].(float64)
	frameID := devtools.FrameID(frameIDFloat)

	snap, exists := b.snapshots.Get(frameID)
	if !exists {
		return NewError(msg.ID, fmt.Sprintf("snapshot not found for frame %d", frameID))
	}

	// Convert to remote format
	components := make([]ComponentData, 0, len(snap.States))
	for _, state := range snap.States {
		comp := ComponentData{
			NodeID:  state.NodeID,
			Type:    state.Type,
			Props:   state.Props,
			State:   state.State,
			Bounds:  RectData{
				X:      state.Bounds.X,
				Y:      state.Bounds.Y,
				Width:  state.Bounds.Width,
				Height: state.Bounds.Height,
			},
			Children: state.Children,
			Visible:  state.Visible,
			Focused:  state.Focused,
		}
		components = append(components, comp)
	}

	result := SnapshotPayload{
		FrameID:   snap.FrameID,
		Timestamp: snap.Timestamp,
		WindowState: WindowState{
			Width:  snap.Global.WindowSize.Width,
			Height: snap.Global.WindowSize.Height,
		},
		Components: components,
	}

	return NewMessageWithID(msg.ID, TypeSnapshot, result)
}

// handleGetRange handles range requests.
func (b *ChromiumBridge) handleGetRange(session *Session, msg *Message) *Message {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return NewError(msg.ID, "invalid payload")
	}

	fromFloat, _ := payload["from"].(float64)
	toFloat, _ := payload["to"].(float64)

	from := devtools.FrameID(fromFloat)
	to := devtools.FrameID(toFloat)

	snapshots := b.snapshots.GetRange(from, to)

	frames := make([]FrameSummary, 0, len(snapshots))
	for _, snap := range snapshots {
		frames = append(frames, FrameSummary{
			FrameID:   snap.FrameID,
			Timestamp: snap.Timestamp,
			Events:    snap.Metadata.FramesSinceLast,
			Mutations: snap.Metadata.MutationsCount,
			Layouts:   snap.Metadata.LayoutsCount,
		})
	}

	result := RangePayload{Frames: frames}
	return NewMessageWithID(msg.ID, TypeGetRange, result)
}

// handleGetDiff handles diff requests.
func (b *ChromiumBridge) handleGetDiff(session *Session, msg *Message) *Message {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return NewError(msg.ID, "invalid payload")
	}

	fromFloat, _ := payload["from"].(float64)
	toFloat, _ := payload["to"].(float64)

	from := devtools.FrameID(fromFloat)
	to := devtools.FrameID(toFloat)

	fromSnap, fromOk := b.snapshots.Get(from)
	toSnap, toOk := b.snapshots.Get(to)

	if !fromOk || !toOk {
		return NewError(msg.ID, "snapshot not found for diff")
	}

	differ := snapshot.NewDiffer()
	diff := differ.Compare(fromSnap, toSnap)

	changes := make([]ChangeData, 0, len(diff.Changes))
	for _, change := range diff.Changes {
		changes = append(changes, ChangeData{
			NodeID:   change.NodeID,
			Type:     change.ChangeType.String(),
			Path:     change.Path,
			OldValue: change.OldValue,
			NewValue: change.NewValue,
		})
	}

	result := DiffPayload{
		From:    from,
		To:      to,
		Changes: changes,
	}

	return NewMessageWithID(msg.ID, TypeDiff, result)
}

// handleSubscribe handles subscription requests.
func (b *ChromiumBridge) handleSubscribe(session *Session, msg *Message) *Message {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return NewError(msg.ID, "invalid payload")
	}

	eventType, _ := payload["event_type"].(string)
	session.Subscribe(eventType)

	return NewMessageWithID(msg.ID, "subscribed", map[string]string{
		"event_type": eventType,
	})
}

// handleUnsubscribe handles unsubscribe requests.
func (b *ChromiumBridge) handleUnsubscribe(session *Session, msg *Message) *Message {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return NewError(msg.ID, "invalid payload")
	}

	eventType, _ := payload["event_type"].(string)
	session.Unsubscribe(eventType)

	return NewMessageWithID(msg.ID, "unsubscribed", map[string]string{
		"event_type": eventType,
	})
}

// handleSetBreakpoint handles breakpoint setting.
func (b *ChromiumBridge) handleSetBreakpoint(session *Session, msg *Message) *Message {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return NewError(msg.ID, "invalid payload")
	}

	bpData, _ := payload["breakpoint"].(map[string]interface{})
	bpID, _ := bpData["id"].(string)
	nodeIDStr, _ := bpData["node_id"].(string)

	bp := &BreakpointData{
		ID:     bpID,
		NodeID: devtools.NodeID(nodeIDStr),
		Enabled: true,
	}

	session.AddBreakpoint(bpID, bp)

	return NewMessageWithID(msg.ID, "breakpoint_set", map[string]string{
		"breakpoint_id": bpID,
	})
}

// handleClearBreakpoint handles breakpoint clearing.
func (b *ChromiumBridge) handleClearBreakpoint(session *Session, msg *Message) *Message {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return NewError(msg.ID, "invalid payload")
	}

	bpID, _ := payload["breakpoint_id"].(string)
	session.RemoveBreakpoint(bpID)

	return NewMessageWithID(msg.ID, "breakpoint_cleared", map[string]string{
		"breakpoint_id": bpID,
	})
}

// =============================================================================
// Event Broadcasting
// =============================================================================

// BroadcastEvent sends an event to all subscribed sessions.
func (b *ChromiumBridge) BroadcastEvent(event *EventPayload) {
	if !b.IsEnabled() {
		return
	}

	sessions := b.server.GetSessions()
	for _, session := range sessions {
		if session.IsSubscribed(TypeEvent) {
			msg := NewMessage(TypeEvent, event)
			data, _ := msg.Serialize()
			// In real implementation, send via WebSocket
			_ = data
		}
	}
}

// BroadcastSnapshot sends a snapshot to all subscribed sessions.
func (b *ChromiumBridge) BroadcastSnapshot(snap *snapshot.Snapshot) {
	if !b.IsEnabled() {
		return
	}

	sessions := b.server.GetSessions()
	for _, session := range sessions {
		if session.IsSubscribed(TypeSnapshot) {
			// Convert and send
			_ = session
		}
	}
}

// =============================================================================
// HTML Inspector Page
// =============================================================================

// GetInspectorHTML returns the HTML for the inspector page.
func (b *ChromiumBridge) GetInspectorHTML() string {
	return `<!DOCTYPE html>
<html>
<head>
	<title>Mint TUI DevTools Inspector</title>
	<style>
		body { font-family: 'Segoe UI', sans-serif; margin: 0; padding: 20px; background: #1e1e1e; color: #d4d4d4; }
		h1 { color: #4ec9b0; }
		.container { max-width: 1200px; margin: 0 auto; }
		.panel { background: #252526; border-radius: 4px; padding: 15px; margin-bottom: 15px; }
		.panel h2 { margin-top: 0; color: #569cd6; border-bottom: 1px solid #3e3e42; padding-bottom: 10px; }
		button { background: #0e639c; border: none; color: white; padding: 8px 16px; border-radius: 4px; cursor: pointer; }
		button:hover { background: #1177bb; }
		#snapshots { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 10px; }
		.snapshot-card { background: #333; padding: 10px; border-radius: 4px; }
		.snapshot-card h3 { margin: 0 0 10px 0; font-size: 14px; }
		.diff { font-family: monospace; white-space: pre-wrap; background: #1e1e1e; padding: 10px; border-radius: 4px; }
		.diff-added { color: #6a9955; }
		.diff-removed { color: #f48771; }
		.diff-changed { color: #dcdcaa; }
	</style>
</head>
<body>
	<div class="container">
		<h1>🔧 Mint TUI DevTools Inspector</h1>

		<div class="panel">
			<h2>Connection</h2>
			<p>Status: <span id="status">Disconnected</span></p>
			<button onclick="connect()">Connect</button>
			<button onclick="disconnect()">Disconnect</button>
		</div>

		<div class="panel">
			<h2>Snapshots</h2>
			<button onclick="loadSnapshots()">Refresh</button>
			<div id="snapshots"></div>
		</div>

		<div class="panel">
			<h2>Diff</h2>
			<input type="number" id="fromFrame" placeholder="From frame (e.g. 0)" value="0">
			<input type="number" id="toFrame" placeholder="To frame (e.g. 9)" value="9">
			<button onclick="loadDiff()">Compare</button>
			<div id="diff" class="diff"></div>
		</div>
	</div>

	<script>
		let ws = null;

		// Load snapshots via HTTP API (fallback)
		async function loadSnapshotsHTTP() {
			try {
				const response = await fetch('/api/snapshots');
				const snapshots = await response.json();
				displaySnapshotsHTTP(snapshots);
			} catch (err) {
				console.error('Failed to load snapshots:', err);
			}
		}

		function displaySnapshotsHTTP(snapshots) {
			const container = document.getElementById('snapshots');
			if (!snapshots || snapshots.length === 0) {
				container.innerHTML = '<p>No snapshots available</p>';
				return;
			}
			let html = '';
			for (const snap of snapshots) {
				html += '<div class="snapshot-card">';
				html += '<h3>Frame ' + snap.frame_id + '</h3>';
				html += '<small>ID: ' + snap.id + '</small><br>';
				html += '<small>Components: ' + snap.components + '</small>';
				html += '</div>';
			}
			container.innerHTML = html;
		}

		function connect() {
			ws = new WebSocket('ws://localhost:9222/ws');
			ws.onopen = () => {
				console.log('WebSocket connected');
				document.getElementById('status').textContent = 'Connected';
				ws.send(JSON.stringify({
					version: '1.0.0',
					type: 'handshake',
					id: 'handshake-' + Date.now(),
					payload: { client_id: 'inspector-' + Date.now() }
				}));
			};
			ws.onmessage = (event) => {
				try {
					const msg = JSON.parse(event.data);
					console.log('WebSocket received:', msg);
					handleMessage(msg);
				} catch (err) {
					console.error('Failed to parse message:', err, event.data);
				}
			};
			ws.onerror = (err) => {
				console.error('WebSocket error:', err);
				document.getElementById('status').textContent = 'Error: ' + err;
			};
			ws.onclose = () => {
				console.log('WebSocket closed');
				document.getElementById('status').textContent = 'Disconnected';
			};
		}

		function disconnect() {
			if (ws) ws.close();
		}

		function handleMessage(msg) {
			switch(msg.type) {
				case 'snapshot':
					displaySnapshot(msg.payload);
					break;
				case 'get_range':
					displaySnapshots(msg.payload);
					break;
				case 'diff':
					displayDiff(msg.payload);
					break;
				case 'handshake_ack':
					console.log('Handshake confirmed:', msg);
					// Load snapshots after handshake
					loadSnapshotsHTTP();
					break;
				case 'error':
					console.error('Server error:', msg.error);
					break;
				default:
					console.log('Unknown message type:', msg.type);
			}
		}

		function loadSnapshots() {
			// Try WebSocket first, fallback to HTTP
			if (ws && ws.readyState === WebSocket.OPEN) {
				ws.send(JSON.stringify({
					version: '1.0.0',
					type: 'get_range',
					id: 'req-' + Date.now(),
					payload: { from: 0, to: 100 }
				}));
			} else {
				loadSnapshotsHTTP();
			}
		}

		async function loadDiff() {
			const from = document.getElementById('fromFrame').value;
			const to = document.getElementById('toFrame').value;
			if (!from || !to) {
				alert('Please enter frame IDs');
				return;
			}

			// Use HTTP API for diff
			try {
				const response = await fetch('/api/diff?from=' + from + '&to=' + to);
				const data = await response.json();
				if (data.error) {
					displayError(data.error + (data.available_frames ? '\nAvailable frames: ' + data.available_frames.join(', ') : ''));
					return;
				}
				displayDiff(data);
			} catch (err) {
				console.error('Failed to load diff:', err);
				displayError('Failed to load diff: ' + err.message);
			}
		}

		function displayError(message) {
			const diffDiv = document.getElementById('diff');
			diffDiv.innerHTML = '<p style="color: #f48771;">' + message + '</p>';
		}

		function displaySnapshot(payload) {
			console.log('Display snapshot:', payload);
		}

		function displaySnapshots(payload) {
			console.log('Display snapshots from WebSocket:', payload);
			const container = document.getElementById('snapshots');
			if (!payload || !payload.frames || payload.frames.length === 0) {
				container.innerHTML = '<p>No snapshots available</p>';
				return;
			}
			let html = '';
			for (const frame of payload.frames) {
				html += '<div class="snapshot-card">';
				html += '<h3>Frame ' + frame.frame_id + '</h3>';
				html += '<small>Events: ' + (frame.events || 0) + ', Mutations: ' + (frame.mutations || 0) + '</small>';
				html += '</div>';
			}
			container.innerHTML = html;
		}

		function displayDiff(payload) {
			console.log('Display diff:', payload);
			const diffDiv = document.getElementById('diff');
			if (!payload || !payload.changes || payload.changes.length === 0) {
				diffDiv.innerHTML = '<p>No differences</p>';
				return;
			}
			let html = '';
			for (const change of payload.changes) {
				if (change.type === 'added') {
					html += '<div class="diff-added">+ ' + change.node_id + ': added</div>';
				} else if (change.type === 'removed') {
					html += '<div class="diff-removed">- ' + change.node_id + ': removed</div>';
				} else if (change.type === 'modified') {
					html += '<div class="diff-changed">~ ' + change.node_id + ': ' + change.path + '</div>';
					html += '<div class="diff-removed">  Old: ' + JSON.stringify(change.old_value) + '</div>';
					html += '<div class="diff-added">  New: ' + JSON.stringify(change.new_value) + '</div>';
				}
			}
			diffDiv.innerHTML = html;
		}

		// Auto-connect on load
		window.onload = () => {
			connect();
			// Also try to load via HTTP immediately
			setTimeout(() => loadSnapshotsHTTP(), 500);
		};
	</script>
</body>
</html>`
}

// =============================================================================
// Statistics
// =============================================================================

// BridgeStats represents bridge statistics.
type BridgeStats struct {
	Enabled      bool          `json:"enabled"`
	Port         int           `json:"port"`
	Path         string        `json:"path"`
	SessionCount int           `json:"session_count"`
	InspectURL   string        `json:"inspect_url"`
}

// GetStats returns bridge statistics.
func (b *ChromiumBridge) GetStats() BridgeStats {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return BridgeStats{
		Enabled:      b.enabled,
		Port:         b.server.port,
		Path:         b.server.path,
		SessionCount: len(b.server.GetSessions()),
		InspectURL:   b.inspectURL,
	}
}

// =============================================================================
// JSON Export for Chromium
// =============================================================================

// ExportForChromium exports data in Chromium-compatible format.
func (b *ChromiumBridge) ExportForChromium() ([]byte, error) {
	snapshots := b.snapshots.GetAll()

	export := struct {
		Version   string                `json:"version"`
		Timestamp time.Time             `json:"timestamp"`
		Frames    []ChromiumFrame       `json:"frames"`
	}{
		Version:   ProtocolVersion,
		Timestamp: time.Now(),
		Frames:    make([]ChromiumFrame, 0, len(snapshots)),
	}

	for _, snap := range snapshots {
		frame := ChromiumFrame{
			FrameID:   int(snap.FrameID),
			Timestamp: snap.Timestamp,
		}

		for _, state := range snap.States {
			node := ChromiumNode{
				ID:       string(state.NodeID),
				Type:     state.Type,
				Props:    state.Props,
				State:    state.State,
				Children: make([]string, len(state.Children)),
			}
			for i, child := range state.Children {
				node.Children[i] = string(child)
			}
			frame.Nodes = append(frame.Nodes, node)
		}

		export.Frames = append(export.Frames, frame)
	}

	return json.MarshalIndent(export, "", "  ")
}

// ChromiumFrame represents a frame in Chromium format.
type ChromiumFrame struct {
	FrameID   int            `json:"frameId"`
	Timestamp time.Time      `json:"timestamp"`
	Nodes     []ChromiumNode `json:"nodes"`
}

// ChromiumNode represents a node in Chromium format.
type ChromiumNode struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Props    map[string]interface{} `json:"props,omitempty"`
	State    map[string]interface{} `json:"state,omitempty"`
	Children []string               `json:"children,omitempty"`
}
