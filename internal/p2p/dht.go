package p2p

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	record "github.com/libp2p/go-libp2p-record"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/multiformats/go-multiaddr"
)

// customValidator for usernames
type customValidator struct{}

func (v customValidator) Validate(key string, value []byte) error {
	// For this example, we allow any username record.
	// In a real-world application, you might want to add validation logic here.
	log.Printf("Validate key: %s", key)
	return nil
}

func (v customValidator) Select(key string, values [][]byte) (int, error) {
	// For this example, we always select the first record.
	log.Printf("Select key: %s", key)
	return 0, nil
}

// SetupDHT creates and bootstraps a DHT for peer discovery.
// SetupDHT creates and bootstraps a DHT for peer discovery.
func SetupDHT(ctx context.Context, h host.Host, bootstrapPeer string) (*dht.IpfsDHT, error) {
	// Create a new DHT
	kademliaDHT, err := dht.New(ctx, h,
		dht.Mode(dht.ModeServer),
		dht.ProtocolPrefix("/p2p-chat"),
		dht.Validator(record.NamespacedValidator{
			"username": customValidator{},
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create DHT: %w", err)
	}

	// Bootstrap the DHT
	if err = kademliaDHT.Bootstrap(ctx); err != nil {
		return nil, fmt.Errorf("failed to bootstrap DHT: %w", err)
	}

	// Custom bootstrap peers
	customBootstrapPeers := []string{
		// "/ip4/127.0.0.1/tcp/4001/p2p/12D3KooWDjdtJTB8hdvJgsh9MRvXdHQ6ThKnjF2pzhw8tXWVXhL4",
		"/ip4/148.251.35.204/tcp/30001/p2p/12D3KooWH4uEYewx2gwwzxQNeGkkTVm1V2dyfvUytcDzr6eh7HSd",
		"/dnsaddr/sg1.bootstrap.libp2p.io/p2p/QmcZf59bWwK5XFi76CZX8cbJ4BhTzzA3gU1ZjYZcYW3dwt",
		"/dnsaddr/sv15.bootstrap.libp2p.io/p2p/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN",
		"/dnsaddr/am6.bootstrap.libp2p.io/p2p/QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb",
		"/dnsaddr/ny5.bootstrap.libp2p.io/p2p/QmQCU2EcMqAqQPR2i9bChDtGNJchTbq5TbXJJ16u19uLTa",
	}

	// Convert string addresses to multiaddr
	var peers []multiaddr.Multiaddr
	for _, addrStr := range customBootstrapPeers {
		addr, err := multiaddr.NewMultiaddr(addrStr)
		if err != nil {
			log.Printf("Failed to parse bootstrap peer address %s: %v", addrStr, err)
			continue
		}
		peers = append(peers, addr)
	}

	// If a specific bootstrap peer is provided, use it instead
	if bootstrapPeer != "" {
		addr, err := multiaddr.NewMultiaddr(bootstrapPeer)
		if err != nil {
			return nil, fmt.Errorf("failed to parse bootstrap peer: %w", err)
		}
		peers = []multiaddr.Multiaddr{addr}
	}

	// Connect to bootstrap peers
	var wg sync.WaitGroup
	for _, peerAddr := range peers {
		peerinfo, err := peer.AddrInfoFromP2pAddr(peerAddr)
		if err != nil {
			log.Printf("Failed to parse peer address %s: %v", peerAddr, err)
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := h.Connect(ctx, *peerinfo); err != nil {
				log.Printf("Failed to connect to bootstrap peer %s: %v", peerinfo.ID, err)
			} else {
				log.Printf("Connected to bootstrap peer %s", peerinfo.ID)
			}
		}()
	}
	wg.Wait()

	// Wait for the routing table to be populated
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for {
		if len(kademliaDHT.RoutingTable().ListPeers()) > 0 {
			break
		}
		select {
		case <-ctx.Done():
			log.Println("Warning: DHT bootstrap timed out")
			return kademliaDHT, nil
		case <-time.After(100 * time.Millisecond):
		}
	}

	return kademliaDHT, nil
}

// PublishUsername publishes a username to the DHT.
func PublishUsername(ctx context.Context, dht routing.Routing, h host.Host, username string) error {
	key := fmt.Sprintf("/username/%s", username)
	value := []byte(h.ID())

	err := dht.PutValue(ctx, key, value)
	if err != nil {
		return fmt.Errorf("failed to publish username: %w", err)
	}

	log.Printf("Published username %s to DHT for peer %s", username, h.ID())
	return nil
}

// FindPeerByUsername searches for a peer by username in the DHT.
func FindPeerByUsername(ctx context.Context, dht routing.Routing, username string) (peer.AddrInfo, error) {
	key := fmt.Sprintf("/username/%s", username)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	value, err := dht.GetValue(ctx, key)
	if err != nil {
		if err == routing.ErrNotFound {
			return peer.AddrInfo{}, fmt.Errorf("user not found")
		}
		return peer.AddrInfo{}, fmt.Errorf("failed to search username: %w", err)
	}

	peerID, err := peer.Decode(string(value))
	if err != nil {
		return peer.AddrInfo{}, fmt.Errorf("failed to decode peer id: %w", err)
	}

	return dht.FindPeer(ctx, peerID)
}
