package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/routing"
	"io"
	"log"
	"p2p-chat/internal/db"
	"p2p-chat/internal/p2p"
	"time"
)

const PrivateChatProtocol = protocol.ID("/p2p-chat/private/1.0.0")

// PrivateMessage represents a single private chat message.
type PrivateMessage struct {
	ID        string `json:"id"`
	SenderID  string `json:"sender_id"`
	RecipientID string `json:"recipient_id"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
	IsSent    bool   `json:"is_sent"`
}

// PrivateChatManager handles private chat operations.
type PrivateChatManager struct {
	host     host.Host
	db       *db.LevelDBStore
	notifier Notifier
	dht      routing.Routing
}

// NewPrivateChatManager creates a new PrivateChatManager.
func NewPrivateChatManager(h host.Host, store *db.LevelDBStore, notifier Notifier, dht routing.Routing) *PrivateChatManager {
	return &PrivateChatManager{
		host:     h,
		db:       store,
		notifier: notifier,
		dht:      dht,
	}
}

// HandlePrivateChatStream sets up a stream handler for private chat messages.
func (pcm *PrivateChatManager) HandlePrivateChatStream(s network.Stream) {
	log.Printf("New private chat stream from %s\n", s.Conn().RemotePeer().String())
	defer s.Close()

	buf := make([]byte, 1024)
	n, err := s.Read(buf)
	if err != nil {
		if err != io.EOF {
			log.Printf("Error reading from private chat stream: %v\n", err)
		}
		return
	}
	messageContent := string(buf[:n])

	// Store message in LevelDB
	msg := &PrivateMessage{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		SenderID:  s.Conn().RemotePeer().String(),
		RecipientID: pcm.host.ID().String(),
		Content:   messageContent,
		Timestamp: time.Now().Unix(),
		IsSent:    false, // This is a received message
	}
	err = pcm.storeMessage(msg)
	if err != nil {
		log.Printf("Failed to store received message: %v", err)
	}

	// Notify frontend via WebSocket
	if pcm.notifier != nil {
		pcm.notifier.NotifyNewMessage(msg.SenderID, msg.Content, "private")
	}

	fmt.Printf("Private message from %s: %s\n", s.Conn().RemotePeer().String(), messageContent)
}

// SendPrivateMessage sends an encrypted private message to a peer.
func (pcm *PrivateChatManager) SendPrivateMessage(ctx context.Context, peerIDStr string, message string) error {
	peerID, err := peer.Decode(peerIDStr)
	if err != nil {
		return fmt.Errorf("invalid peer ID: %w", err)
	}

	s, err := pcm.host.NewStream(ctx, peerID, PrivateChatProtocol)
	if err != nil {
		return fmt.Errorf("failed to open private chat stream: %w", err)
	}
	defer s.Close()

	_, err = s.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to write to private chat stream: %w", err)
	}

	// Store sent message in LevelDB
	msg := &PrivateMessage{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		SenderID:  pcm.host.ID().String(),
		RecipientID: peerIDStr,
		Content:   message,
		Timestamp: time.Now().Unix(),
		IsSent:    true,
	}
	err = pcm.storeMessage(msg)
	if err != nil {
		log.Printf("Failed to store sent message: %v", err)
	}

	log.Printf("Sent private message to %s\n", peerID.String())
	return nil
}

// SendInitialMessage sends an initial message to a peer to start a chat.
func (pcm *PrivateChatManager) SendInitialMessage(ctx context.Context, peerIDStr string) error {
	peerID, err := peer.Decode(peerIDStr)
	if err != nil {
		return fmt.Errorf("invalid peer ID: %w", err)
	}

	s, err := pcm.host.NewStream(ctx, peerID, PrivateChatProtocol)
	if err != nil {
		return fmt.Errorf("failed to open private chat stream: %w", err)
	}
	defer s.Close()

	_, err = s.Write([]byte("Hello"))
	if err != nil {
		return fmt.Errorf("failed to write to private chat stream: %w", err)
	}

	log.Printf("Sent initial message to %s\n", peerID.String())
	return nil
}


func (pcm *PrivateChatManager) storeMessage(msg *PrivateMessage) error {
	key := fmt.Sprintf("chat/private/%s/%s", msg.RecipientID, msg.ID)
	if !msg.IsSent {
		key = fmt.Sprintf("chat/private/%s/%s", msg.SenderID, msg.ID)
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return pcm.db.Put([]byte(key), data)
}

// GetChatHistory retrieves chat history for a specific peer.
func (pcm *PrivateChatManager) GetChatHistory(peerIDStr string) ([]*PrivateMessage, error) {
	prefix := fmt.Sprintf("chat/private/%s/", peerIDStr)
	iter := pcm.db.NewIteratorWithPrefix([]byte(prefix))
	defer iter.Release()

	var messages []*PrivateMessage
	for iter.Next() {
		var msg PrivateMessage
		err := json.Unmarshal(iter.Value(), &msg)
		if err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}
		messages = append(messages, &msg)
	}

	return messages, iter.Error()
}

// SearchPeer searches for a peer by username.
func (pcm *PrivateChatManager) SearchPeer(ctx context.Context, username string) (peer.AddrInfo, error) {
	return p2p.FindPeerByUsername(ctx, pcm.dht, username)
}


