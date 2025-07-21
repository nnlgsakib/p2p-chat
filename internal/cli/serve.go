package cli

import (
	"context"
	"github.com/spf13/cobra"
	"log"
	"p2p-chat/internal/api"
	"p2p-chat/internal/chat"
	"p2p-chat/internal/db"
	"p2p-chat/internal/p2p"
	"p2p-chat/assets"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the P2P chat node",
	Run: func(cmd *cobra.Command, args []string) {
		dbPath, _ := cmd.Flags().GetString("datadir")
		restPort, _ := cmd.Flags().GetInt("rest-port")
		wsPort, _ := cmd.Flags().GetInt("ws-port")
		libp2pPort, _ := cmd.Flags().GetInt("libp2p-port")
		username, _ := cmd.Flags().GetString("username")
		bootstrapPeer, _ := cmd.Flags().GetString("bootstrap-peer")

		if dbPath == "" {
			log.Fatal("Error: --datadir flag is required for database path.")
		}

		store, err := db.NewLevelDBStore(dbPath)
		if err != nil {
			log.Fatalf("Error opening database: %v", err)
		}
		defer store.Close()

		// Load libp2p private key
		_, err = store.Get([]byte("libp2p_private_key"))
		if err != nil {
			log.Fatalf("Error loading libp2p private key: %v", err)
		}

		// Create libp2p host
		host, err := p2p.NewHost(libp2pPort)
		if err != nil {
			log.Fatalf("Error creating libp2p host: %v", err)
		}
		defer host.Close()

		// Setup mDNS discovery
		err = p2p.SetupDiscovery(host)
		if err != nil {
			log.Fatalf("Error setting up mDNS discovery: %v", err)
		}

		// Setup DHT
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		dht, err := p2p.SetupDHT(ctx, host, bootstrapPeer)
		if err != nil {
			log.Fatalf("Error setting up DHT: %v", err)
		}
		defer dht.Close()

		// If username is not provided, generate a random hash
		if username == "" {
			username = host.ID().String()[:20] // Use first 20 chars of Peer ID
			log.Printf("Generated username: %s\n", username)
		}

		// Publish username to DHT
        err = p2p.PublishUsername(ctx, dht, host, username)
        if err != nil {
            log.Printf("Warning: Failed to publish username to DHT: %v\n", err)
        }

		// Start WebSocket API server
		wsAPI := api.NewWebSocketAPI(host, store, nil, nil, nil, bootstrapPeer)
		go wsAPI.StartWebSocketServer(wsPort)

		// Setup chat managers
		privateChatManager := chat.NewPrivateChatManager(host, store, wsAPI, dht)
		groupChatManager := chat.NewGroupChatManager(host, store)
		fileTransferManager := chat.NewFileTransferManager(host, store, "./downloads") // TODO: Make download dir configurable

		// Set up stream handlers
		host.SetStreamHandler(p2p.ChatProtocol, p2p.HandleChatStream)
		host.SetStreamHandler(p2p.FileProtocol, p2p.HandleFileStream)
		host.SetStreamHandler(chat.PrivateChatProtocol, privateChatManager.HandlePrivateChatStream)
		host.SetStreamHandler(chat.GroupChatProtocol, groupChatManager.HandleGroupChatStream)
		host.SetStreamHandler(chat.FileTransferProtocol, fileTransferManager.HandleFileTransferStream)

		// Start REST API server
		restAPI := api.NewAPI(host, store, privateChatManager, groupChatManager, fileTransferManager, restPort, wsPort, assets.StaticFiles)
		go restAPI.StartRestServer(restPort)

		// Assign managers to WebSocket API
		wsAPI.SetManagers(privateChatManager, groupChatManager, fileTransferManager)

		log.Println("P2P Chat Node started successfully!")

		select { // Block forever
		case <-ctx.Done():
			log.Println("Shutting down P2P Chat Node...")
		}
	},
}

func init() {
	serveCmd.Flags().String("datadir", "", "Path to the LevelDB database")
	serveCmd.Flags().Int("rest-port", 8080, "Port for the REST API server")
	serveCmd.Flags().Int("ws-port", 8081, "Port for the WebSocket API server")
	serveCmd.Flags().Int("libp2p-port", 0, "Port for the libp2p host (0 for random)")
	serveCmd.Flags().String("username", "", "Username for this node (generates random if not provided)")
	serveCmd.Flags().String("bootstrap-peer", "", "Bootstrap peer multiaddress")
	RootCmd.AddCommand(serveCmd)
}


