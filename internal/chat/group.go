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
	"p2p-chat/internal/db"
	"sync"
)

const GroupChatProtocol = protocol.ID("/p2p-chat/group/1.0.0")

// GroupChatManager handles group chat operations.
type GroupChatManager struct {
	host   host.Host
	db     *db.LevelDBStore
	groups map[string]*Group
	mutex  sync.RWMutex
}

// Group represents a chat group.
type Group struct {
	ID      string
	Name    string
	Members []peer.ID
	Admin   peer.ID
}

// NewGroupChatManager creates a new GroupChatManager.
func NewGroupChatManager(h host.Host, store *db.LevelDBStore) *GroupChatManager {
	return &GroupChatManager{
		host:   h,
		db:     store,
		groups: make(map[string]*Group),
	}
}

// CreateGroup creates a new chat group.
func (gcm *GroupChatManager) CreateGroup(groupID, groupName string, adminID peer.ID) error {
	gcm.mutex.Lock()
	defer gcm.mutex.Unlock()

	if _, exists := gcm.groups[groupID]; exists {
		return fmt.Errorf("group %s already exists", groupID)
	}

	group := &Group{
		ID:      groupID,
		Name:    groupName,
		Members: []peer.ID{adminID},
		Admin:   adminID,
	}

	gcm.groups[groupID] = group
	log.Printf("Created group %s with admin %s\n", groupName, adminID.String())

	// TODO: Store group info in LevelDB
	return nil
}

// AddMemberToGroup adds a member to a group.
func (gcm *GroupChatManager) AddMemberToGroup(groupID string, memberID peer.ID) error {
	gcm.mutex.Lock()
	defer gcm.mutex.Unlock()

	group, exists := gcm.groups[groupID]
	if !exists {
		return fmt.Errorf("group %s does not exist", groupID)
	}

	// Check if member is already in the group
	for _, member := range group.Members {
		if member == memberID {
			return fmt.Errorf("member %s is already in group %s", memberID.String(), groupID)
		}
	}

	group.Members = append(group.Members, memberID)
	log.Printf("Added member %s to group %s\n", memberID.String(), groupID)

	// TODO: Store updated group info in LevelDB
	return nil
}

// RemoveMemberFromGroup removes a member from a group.
func (gcm *GroupChatManager) RemoveMemberFromGroup(groupID string, memberID peer.ID) error {
	gcm.mutex.Lock()
	defer gcm.mutex.Unlock()

	group, exists := gcm.groups[groupID]
	if !exists {
		return fmt.Errorf("group %s does not exist", groupID)
	}

	for i, member := range group.Members {
		if member == memberID {
			group.Members = append(group.Members[:i], group.Members[i+1:]...)
			log.Printf("Removed member %s from group %s\n", memberID.String(), groupID)
			// TODO: Store updated group info in LevelDB
			return nil
		}
	}

	return fmt.Errorf("member %s is not in group %s", memberID.String(), groupID)
}

// HandleGroupChatStream sets up a stream handler for group chat messages.
func (gcm *GroupChatManager) HandleGroupChatStream(s network.Stream) {
	log.Printf("New group chat stream from %s\n", s.Conn().RemotePeer().String())

	buf := make([]byte, 1024)
	for {
		n, err := s.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading from group chat stream: %v\n", err)
			}
			break
		}
		// TODO: Parse group ID and message from the received data
		fmt.Printf("Group message from %s: %s\n", s.Conn().RemotePeer().String(), string(buf[:n]))
		// TODO: Store message in LevelDB
	}

	s.Close()
}

// SendGroupMessage sends a message to all members of a group.
func (gcm *GroupChatManager) SendGroupMessage(ctx context.Context, groupID, message string) error {
	gcm.mutex.RLock()
	group, exists := gcm.groups[groupID]
	gcm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("group %s does not exist", groupID)
	}

	// Send message to all group members
	for _, memberID := range group.Members {
		if memberID == gcm.host.ID() {
			continue // Skip sending to self
		}

		go func(peerID peer.ID) {
			s, err := gcm.host.NewStream(ctx, peerID, GroupChatProtocol)
			if err != nil {
				log.Printf("Failed to open group chat stream to %s: %v\n", peerID.String(), err)
				return
			}
			defer s.Close()

			// TODO: Format message with group ID
			formattedMessage := fmt.Sprintf("[%s] %s", groupID, message)
			_, err = s.Write([]byte(formattedMessage))
			if err != nil {
				log.Printf("Failed to write to group chat stream to %s: %v\n", peerID.String(), err)
			}
		}(memberID)
	}

	log.Printf("Sent group message to group %s\n", groupID)
	// TODO: Store sent message in LevelDB
	return nil
}

// ListGroups returns a list of all groups.
func (gcm *GroupChatManager) ListGroups() []*Group {
	gcm.mutex.RLock()
	defer gcm.mutex.RUnlock()

	var groups []*Group
	for _, group := range gcm.groups {
		groups = append(groups, group)
	}

	return groups
}

// GetGroup returns a specific group by ID.
func (gcm *GroupChatManager) GetGroup(groupID string) (*Group, error) {
	gcm.mutex.RLock()
	defer gcm.mutex.RUnlock()

	group, exists := gcm.groups[groupID]
	if !exists {
		return nil, fmt.Errorf("group %s does not exist", groupID)
	}

	return group, nil
}

