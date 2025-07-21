package p2p

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/multiformats/go-multiaddr"
	"log"
	"time"
)

// DiscoveryServiceTag is used to identify our service on the network.
const DiscoveryServiceTag = "p2p-chat-discovery"

// NewHost creates a new libp2p host.
func NewHost(port int) (host.Host, error) {
	listenAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))
	if err != nil {
		return nil, err
	}

	host, err := libp2p.New(libp2p.ListenAddrs(listenAddr))
	if err != nil {
		return nil, err
	}

	fmt.Printf("\n[*] Your Multiaddress: /ip4/127.0.0.1/tcp/%v/p2p/%s\n", port, host.ID().String())

	return host, nil
}

// discoveryNotifee gets notified when new peers are found
type discoveryNotifee struct {
	host host.Host
}

// HandlePeerFound connects to peers discovered via mDNS.
func (n *discoveryNotifee) HandlePeerFound(peerInfo peer.AddrInfo) {
	fmt.Printf("Discovered new peer: %s\n", peerInfo.ID.String())
	err := n.host.Connect(context.Background(), peerInfo)
	if err != nil {
		fmt.Printf("Error connecting to peer %s: %s\n", peerInfo.ID.String(), err)
	}
}

// SetupDiscovery creates an mDNS discovery service and attaches it to the libp2p Host.
func SetupDiscovery(h host.Host) error {
	service := mdns.NewMdnsService(h, DiscoveryServiceTag, &discoveryNotifee{host: h})
	return service.Start()
}

// ConnectToPeer connects to a specific peer using its multiaddress.
func ConnectToPeer(h host.Host, addr string) error {
	maddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return fmt.Errorf("invalid multiaddress: %w", err)
	}

	peerInfo, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return fmt.Errorf("failed to parse peer info from multiaddress: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Printf("Connecting to peer %s...", peerInfo.ID.String())
	err = h.Connect(ctx, *peerInfo)
	if err != nil {
		return fmt.Errorf("failed to connect to peer %s: %w", peerInfo.ID.String(), err)
	}

	log.Printf("Connected to peer %s\n", peerInfo.ID.String())
	return nil
}


