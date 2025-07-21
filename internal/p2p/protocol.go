package p2p

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"io"
	"log"
)

const (
	ChatProtocol = protocol.ID("/p2p-chat/1.0.0")
	FileProtocol = protocol.ID("/p2p-file/1.0.0")
)

// HandleChatStream sets up a stream handler for chat messages.
func HandleChatStream(s network.Stream) {
	log.Printf("Got a new chat stream from %s\n", s.Conn().RemotePeer().String())

	buf := make([]byte, 1024)
	for {
		n, err := s.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading from chat stream: %v\n", err)
			}
			break
		}
		fmt.Printf("[%s]: %s\n", s.Conn().RemotePeer().String(), string(buf[:n]))
	}

	s.Close()
}

// SendChatMessage sends a chat message to a peer.
func SendChatMessage(ctx context.Context, h host.Host, peerID string, message string) error {
	peerInfo, err := PeerIDFromString(peerID)
	if err != nil {
		return err
	}

	s, err := h.NewStream(ctx, peerInfo.ID, ChatProtocol)
	if err != nil {
		return fmt.Errorf("failed to open chat stream: %w", err)
	}
	defer s.Close()

	_, err = s.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to write to chat stream: %w", err)
	}

	return nil
}

// HandleFileStream sets up a stream handler for file transfers.
func HandleFileStream(s network.Stream) {
	log.Printf("Got a new file stream from %s\n", s.Conn().RemotePeer().String())

	// TODO: Implement file transfer logic
	log.Println("File transfer not yet implemented.")

	s.Close()
}

// SendFile sends a file to a peer.
func SendFile(ctx context.Context, h host.Host, peerID string, filePath string) error {
	peerInfo, err := PeerIDFromString(peerID)
	if err != nil {
		return err
	}

	s, err := h.NewStream(ctx, peerInfo.ID, FileProtocol)
	if err != nil {
		return fmt.Errorf("failed to open file stream: %w", err)
	}
	defer s.Close()

	// TODO: Implement file reading and writing logic
	log.Println("File sending not yet implemented.")

	return nil
}

// PeerIDFromString converts a string representation of a Peer ID to a peer.AddrInfo.
func PeerIDFromString(peerIDStr string) (*peer.AddrInfo, error) {
	peerID, err := peer.Decode(peerIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid peer ID: %w", err)
	}
	return &peer.AddrInfo{ID: peerID}, nil
}


