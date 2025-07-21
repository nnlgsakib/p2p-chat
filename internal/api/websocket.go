package api

import (
	"fmt"
	"log"
	"net/http"
	"p2p-chat/internal/chat"
	"p2p-chat/internal/db"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin
	},
}

// WebSocketAPI represents the WebSocket API server.
type WebSocketAPI struct {
	host                host.Host
	db                  *db.LevelDBStore
	privateChatManager  *chat.PrivateChatManager
	groupChatManager    *chat.GroupChatManager
	fileTransferManager *chat.FileTransferManager
	clients             map[*websocket.Conn]bool
	clientsMutex        sync.RWMutex
	bootstrapPeer       string
}

// NewWebSocketAPI creates a new WebSocketAPI instance.
func NewWebSocketAPI(h host.Host, store *db.LevelDBStore, pcm *chat.PrivateChatManager, gcm *chat.GroupChatManager, ftm *chat.FileTransferManager, bootstrapPeer string) *WebSocketAPI {
	return &WebSocketAPI{
		host:                h,
		db:                  store,
		privateChatManager:  pcm,
		groupChatManager:    gcm,
		fileTransferManager: ftm,
		clients:             make(map[*websocket.Conn]bool),
		bootstrapPeer:       bootstrapPeer,
	}
}

// SetManagers sets the chat and file transfer managers for the WebSocket API.
func (wsapi *WebSocketAPI) SetManagers(pcm *chat.PrivateChatManager, gcm *chat.GroupChatManager, ftm *chat.FileTransferManager) {
	wsapi.privateChatManager = pcm
	wsapi.groupChatManager = gcm
	wsapi.fileTransferManager = ftm
}

// StartWebSocketServer starts the WebSocket server.
func (wsapi *WebSocketAPI) StartWebSocketServer(port int) {
	http.HandleFunc("/ws", wsapi.handleWebSocket)

	log.Printf("WebSocket server listening on :%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func (wsapi *WebSocketAPI) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection to WebSocket: %v\n", err)
		return
	}
	defer conn.Close()

	// Add client to the list
	wsapi.clientsMutex.Lock()
	wsapi.clients[conn] = true
	wsapi.clientsMutex.Unlock()

	// Remove client when connection closes
	defer func() {
		wsapi.clientsMutex.Lock()
		delete(wsapi.clients, conn)
		wsapi.clientsMutex.Unlock()
	}()

	log.Printf("New WebSocket connection established\n")

	// Listen for messages from the client
	for {
		var msg map[string]interface{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v\n", err)
			}
			break
		}

		// Handle different message types
		wsapi.handleWebSocketMessage(conn, msg)
	}
}

func (wsapi *WebSocketAPI) handleWebSocketMessage(conn *websocket.Conn, msg map[string]interface{}) {
	msgType, ok := msg["type"].(string)
	if !ok {
		wsapi.sendError(conn, "Invalid message format: missing 'type' field")
		return
	}

	switch msgType {
	case "get_peer_info":
		wsapi.handleGetPeerInfo(conn)
	case "get_connected_peers":
		wsapi.handleGetConnectedPeers(conn)
	case "get_groups":
		wsapi.handleGetGroups(conn)
	case "get_received_files":
		wsapi.handleGetReceivedFiles(conn)
	case "get_chat_history":
		wsapi.handleGetChatHistory(conn, msg)
	default:
		wsapi.sendError(conn, fmt.Sprintf("Unknown message type: %s", msgType))
	}
}

func (wsapi *WebSocketAPI) handleGetPeerInfo(conn *websocket.Conn) {
	peerInfo := map[string]interface{}{
		"type":    "peer_info",
		"peer_id": wsapi.host.ID().String(),
		"addrs":   wsapi.host.Addrs(),
	}

	if err := conn.WriteJSON(peerInfo); err != nil {
		log.Printf("Failed to send peer info: %v\n", err)
	}
}

func (wsapi *WebSocketAPI) handleGetConnectedPeers(conn *websocket.Conn) {
	peers := wsapi.host.Network().Peers()
	var peerList []string
	var bootstrapPeerID peer.ID
	if wsapi.bootstrapPeer != "" {
		p, err := peer.AddrInfoFromString(wsapi.bootstrapPeer)
		if err == nil {
			bootstrapPeerID = p.ID
		}
	}

	for _, peer := range peers {
		if peer != bootstrapPeerID {
			peerList = append(peerList, peer.String())
		}
	}

	response := map[string]interface{}{
		"type":  "connected_peers",
		"peers": peerList,
	}

	if err := conn.WriteJSON(response); err != nil {
		log.Printf("Failed to send connected peers: %v\n", err)
	}
}

func (wsapi *WebSocketAPI) handleGetGroups(conn *websocket.Conn) {
	groups := wsapi.groupChatManager.ListGroups()
	var groupList []map[string]interface{}

	for _, group := range groups {
		var memberList []string
		for _, member := range group.Members {
			memberList = append(memberList, member.String())
		}

		groupInfo := map[string]interface{}{
			"id":      group.ID,
			"name":    group.Name,
			"admin":   group.Admin.String(),
			"members": memberList,
		}
		groupList = append(groupList, groupInfo)
	}

	response := map[string]interface{}{
		"type":   "groups",
		"groups": groupList,
	}

	if err := conn.WriteJSON(response); err != nil {
		log.Printf("Failed to send groups: %v\n", err)
	}
}

func (wsapi *WebSocketAPI) handleGetReceivedFiles(conn *websocket.Conn) {
	files, err := wsapi.fileTransferManager.ListReceivedFiles()
	if err != nil {
		wsapi.sendError(conn, fmt.Sprintf("Failed to list received files: %v", err))
		return
	}

	response := map[string]interface{}{
		"type":  "received_files",
		"files": files,
	}

	if err := conn.WriteJSON(response); err != nil {
		log.Printf("Failed to send received files: %v\n", err)
	}
}

func (wsapi *WebSocketAPI) handleGetChatHistory(conn *websocket.Conn, msg map[string]interface{}) {
	peerID, ok := msg["peer_id"].(string)
	if !ok {
		wsapi.sendError(conn, "Invalid message format: missing 'peer_id' field")
		return
	}

	history, err := wsapi.privateChatManager.GetChatHistory(peerID)
	if err != nil {
		wsapi.sendError(conn, fmt.Sprintf("Failed to get chat history: %v", err))
		return
	}

	response := map[string]interface{}{
		"type":    "chat_history",
		"peer_id": peerID,
		"history": history,
	}

	if err := conn.WriteJSON(response); err != nil {
		log.Printf("Failed to send chat history: %v\n", err)
	}
}

func (wsapi *WebSocketAPI) sendError(conn *websocket.Conn, message string) {
	errorMsg := map[string]interface{}{
		"type":  "error",
		"error": message,
	}

	if err := conn.WriteJSON(errorMsg); err != nil {
		log.Printf("Failed to send error message: %v\n", err)
	}
}

// BroadcastMessage sends a message to all connected WebSocket clients.
func (wsapi *WebSocketAPI) BroadcastMessage(message map[string]interface{}) {
	wsapi.clientsMutex.RLock()
	defer wsapi.clientsMutex.RUnlock()

	for conn := range wsapi.clients {
		if err := conn.WriteJSON(message); err != nil {
			log.Printf("Failed to broadcast message to client: %v\n", err)
			// Remove the client if there's an error
			wsapi.clientsMutex.Lock()
			delete(wsapi.clients, conn)
			wsapi.clientsMutex.Unlock()
			conn.Close()
		}
	}
}

// NotifyNewMessage notifies all clients about a new message.
func (wsapi *WebSocketAPI) NotifyNewMessage(senderID, message, messageType string) {
	notification := map[string]interface{}{
		"type":         "new_message",
		"sender_id":    senderID,
		"message":      message,
		"message_type": messageType,
		"timestamp":    time.Now().Unix(),
	}

	wsapi.BroadcastMessage(notification)
}
