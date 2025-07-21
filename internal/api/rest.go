package api

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"p2p-chat/internal/chat"
	"p2p-chat/internal/db"
	"p2p-chat/internal/p2p"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)


// API represents the REST API server.
type API struct {
	host                host.Host
	db                  *db.LevelDBStore
	privateChatManager  *chat.PrivateChatManager
	groupChatManager    *chat.GroupChatManager
	fileTransferManager *chat.FileTransferManager
	restPort            int
	wsPort              int
	staticFiles         embed.FS
}

// NewAPI creates a new API instance.
func NewAPI(h host.Host, store *db.LevelDBStore, pcm *chat.PrivateChatManager, gcm *chat.GroupChatManager, ftm *chat.FileTransferManager, restPort, wsPort int, staticFiles embed.FS) *API {
	return &API{
		host:                h,
		db:                  store,
		privateChatManager:  pcm,
		groupChatManager:    gcm,
		fileTransferManager: ftm,
		restPort:            restPort,
		wsPort:              wsPort,
		staticFiles:         staticFiles,
	}
}

// StartRestServer starts the REST API server.
func (api *API) StartRestServer(port int) {
	// Serve static files from the embedded filesystem
	staticFS, err := fs.Sub(api.staticFiles, "static")
	if err != nil {
		log.Fatalf("Failed to create static file system: %v", err)
	}
	http.Handle("/", http.FileServer(http.FS(staticFS)))


	// API endpoints
	http.HandleFunc("/api/ports", api.handleGetPorts)
	http.HandleFunc("/peer/connect", api.handleConnectPeer)
	http.HandleFunc("/peer/search", api.handleSearchPeer)
	http.HandleFunc("/chat/private/send", api.handleSendPrivateMessage)
	http.HandleFunc("/group/create", api.handleCreateGroup)
	http.HandleFunc("/group/add_member", api.handleAddMemberToGroup)
	http.HandleFunc("/group/send_message", api.handleSendGroupMessage)
	http.HandleFunc("/file/send", api.handleSendFile)

	log.Printf("REST API server listening on :%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func (api *API) handleGetPorts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{
		"rest_port": api.restPort,
		"ws_port":   api.wsPort,
	})
}

func (api *API) handleConnectPeer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PeerMultiaddr string `json:"peer_multiaddr"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = p2p.ConnectToPeer(api.host, req.PeerMultiaddr)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect to peer: %v", err), http.StatusInternalServerError)
		return
	}

	// After connecting, send an initial message to start the chat
	maddr, err := multiaddr.NewMultiaddr(req.PeerMultiaddr)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid multiaddress: %v", err), http.StatusBadRequest)
		return
	}

	peerInfo, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse peer info from multiaddress: %v", err), http.StatusBadRequest)
		return
	}

	err = api.privateChatManager.SendInitialMessage(r.Context(), peerInfo.ID.String())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to send initial message: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "connected"})
}

func (api *API) handleSearchPeer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "Query parameter 'query' is required", http.StatusBadRequest)
		return
	}

	peerInfo, err := api.privateChatManager.SearchPeer(r.Context(), query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to search peer: %v", err), http.StatusInternalServerError)
		return
	}

	addrs := make([]string, len(peerInfo.Addrs))
	for i, addr := range peerInfo.Addrs {
		addrs[i] = addr.String()
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"peer_id": peerInfo.ID.String(),
		"addrs":   addrs,
	})
}

func (api *API) handleSendPrivateMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PeerID  string `json:"peer_id"`
		Message string `json:"message"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = api.privateChatManager.SendPrivateMessage(r.Context(), req.PeerID, req.Message)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to send private message: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "message sent"})
}

func (api *API) handleCreateGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		GroupID   string `json:"group_id"`
		GroupName string `json:"group_name"`
		AdminID   string `json:"admin_id"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	adminPeerID, err := peer.Decode(req.AdminID)
	if err != nil {
		http.Error(w, "Invalid Admin ID", http.StatusBadRequest)
		return
	}

	err = api.groupChatManager.CreateGroup(req.GroupID, req.GroupName, adminPeerID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create group: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "group created"})
}

func (api *API) handleAddMemberToGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		GroupID  string `json:"group_id"`
		MemberID string `json:"member_id"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	memberPeerID, err := peer.Decode(req.MemberID)
	if err != nil {
		http.Error(w, "Invalid Member ID", http.StatusBadRequest)
		return
	}

	err = api.groupChatManager.AddMemberToGroup(req.GroupID, memberPeerID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to add member to group: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "member added"})
}

func (api *API) handleSendGroupMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		GroupID string `json:"group_id"`
		Message string `json:"message"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = api.groupChatManager.SendGroupMessage(r.Context(), req.GroupID, req.Message)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to send group message: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "group message sent"})
}

func (api *API) handleSendFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PeerID   string `json:"peer_id"`
		FilePath string `json:"file_path"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = api.fileTransferManager.SendFile(r.Context(), req.PeerID, req.FilePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to send file: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "file sent"})
}
