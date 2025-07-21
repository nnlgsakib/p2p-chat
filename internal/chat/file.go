package chat

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"io"
	"log"
	"os"
	"path/filepath"
	"p2p-chat/internal/db"
)

const FileTransferProtocol = protocol.ID("/p2p-chat/file/1.0.0")

// FileTransferManager handles file transfer operations.
type FileTransferManager struct {
	host      host.Host
	db        *db.LevelDBStore
	uploadDir string
}

// NewFileTransferManager creates a new FileTransferManager.
func NewFileTransferManager(h host.Host, store *db.LevelDBStore, uploadDir string) *FileTransferManager {
	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("Failed to create upload directory: %v\n", err)
	}

	return &FileTransferManager{
		host:      h,
		db:        store,
		uploadDir: uploadDir,
	}
}

// HandleFileTransferStream sets up a stream handler for file transfers.
func (ftm *FileTransferManager) HandleFileTransferStream(s network.Stream) {
	log.Printf("New file transfer stream from %s\n", s.Conn().RemotePeer().String())

	// Read file metadata first (filename, size, etc.)
	metadataBuf := make([]byte, 256)
	n, err := s.Read(metadataBuf)
	if err != nil {
		log.Printf("Error reading file metadata: %v\n", err)
		s.Close()
		return
	}

	// TODO: Parse metadata (filename, size)
	filename := string(metadataBuf[:n])
	log.Printf("Receiving file: %s\n", filename)

	// Create file in upload directory
	filePath := filepath.Join(ftm.uploadDir, filename)
	file, err := os.Create(filePath)
	if err != nil {
		log.Printf("Error creating file %s: %v\n", filePath, err)
		s.Close()
		return
	}
	defer file.Close()

	// Copy file data from stream to file
	_, err = io.Copy(file, s)
	if err != nil {
		log.Printf("Error writing file data: %v\n", err)
		s.Close()
		return
	}

	log.Printf("Successfully received file: %s\n", filePath)
	s.Close()

	// TODO: Store file transfer record in LevelDB
}

// SendFile sends a file to a peer.
func (ftm *FileTransferManager) SendFile(ctx context.Context, peerIDStr, filePath string) error {
	peerID, err := peer.Decode(peerIDStr)
	if err != nil {
		return fmt.Errorf("invalid peer ID: %w", err)
	}

	// Check if file exists
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Open stream to peer
	s, err := ftm.host.NewStream(ctx, peerID, FileTransferProtocol)
	if err != nil {
		return fmt.Errorf("failed to open file transfer stream: %w", err)
	}
	defer s.Close()

	// Send file metadata first (filename)
	filename := filepath.Base(filePath)
	_, err = s.Write([]byte(filename))
	if err != nil {
		return fmt.Errorf("failed to send file metadata: %w", err)
	}

	// Send file data
	_, err = io.Copy(s, file)
	if err != nil {
		return fmt.Errorf("failed to send file data: %w", err)
	}

	log.Printf("Successfully sent file %s (%d bytes) to %s\n", filename, fileInfo.Size(), peerID.String())

	// TODO: Store file transfer record in LevelDB
	return nil
}

// ListReceivedFiles lists all files received in the upload directory.
func (ftm *FileTransferManager) ListReceivedFiles() ([]string, error) {
	files, err := os.ReadDir(ftm.uploadDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read upload directory: %w", err)
	}

	var fileNames []string
	for _, file := range files {
		if !file.IsDir() {
			fileNames = append(fileNames, file.Name())
		}
	}

	return fileNames, nil
}

// GetReceivedFilePath returns the full path of a received file.
func (ftm *FileTransferManager) GetReceivedFilePath(filename string) string {
	return filepath.Join(ftm.uploadDir, filename)
}

