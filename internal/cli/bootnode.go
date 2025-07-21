package cli

import (
	"context"
	"github.com/spf13/cobra"
	"log"
	"p2p-chat/internal/p2p"
	"time"
)

var bootnodeCmd = &cobra.Command{
	Use:   "bootnode",
	Short: "Start a standalone bootstrap node",
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")

		// Create libp2p host for bootstrap node
		host, err := p2p.NewHost(port)
		if err != nil {
			log.Fatalf("Error creating libp2p host for bootstrap node: %v", err)
		}
		defer host.Close()

		// Setup DHT for bootstrap node
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		dht, err := p2p.SetupDHT(ctx, host, "")
		if err != nil {
			log.Fatalf("Error setting up DHT for bootstrap node: %v", err)
		}
		defer dht.Close()

		log.Printf("Bootstrap node started on port %d\n", port)
		log.Printf("Bootstrap node multiaddress: /ip4/127.0.0.1/tcp/%d/p2p/%s\n", port, host.ID().String())

		// Keep the bootstrap node running
		for {
			time.Sleep(10 * time.Second)
			connectedPeers := host.Network().Peers()
			log.Printf("Bootstrap node has %d connected peers\n", len(connectedPeers))
		}
	},
}

func init() {
	bootnodeCmd.Flags().Int("port", 4001, "Port for the bootstrap node")
	RootCmd.AddCommand(bootnodeCmd)
}

