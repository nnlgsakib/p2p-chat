# P2P Chat Application

This project is a fully decentralized private P2P chatting application built using Go, libp2p, LevelDB, and Svelte. It provides secure peer-to-peer communication without relying on centralized servers.

## Features

- **Private P2P chatting**: Direct encrypted communication between peers
- **Group chat creation**: Create and manage group conversations
- **Curve-based encryption**: ECDSA encryption for secure communications
- **REST and WebSocket APIs**: Full API support for all operations
- **CLI interface**: Command-line tools for initialization and node management
- **Peer discovery**: Automatic peer discovery using mDNS and DHT
- **Connection management**: Request/accept mechanism for peer connections
- **Search functionality**: Find peers by username or multinode address
- **File transfer**: Send and receive files between peers
- **Bootstrap nodes**: Standalone nodes to help with peer discovery
- **Embedded Svelte frontend**: Modern web interface for easy interaction

## Technologies Used

- **Backend**: Go 1.22+
- **P2P Network**: libp2p
- **Data Storage**: LevelDB
- **CLI**: Cobra
- **Encryption**: ECDSA (Elliptic Curve Digital Signature Algorithm)
- **Frontend**: Svelte with SvelteKit
- **APIs**: REST and WebSocket

## Project Structure

```
p2p-chat/
├── cmd/
│   ├── bootnode/
│   │   └── main.go         # Entry point for bootstrap node
│   └── chat/
│       └── main.go         # Entry point for chat client
├── internal/
│   ├── api/
│   │   ├── rest.go         # REST API handlers
│   │   └── websocket.go    # WebSocket handlers
│   ├── crypto/
│   │   └── crypto.go       # Curve-based encryption utilities
│   ├── db/
│   │   └── leveldb.go      # LevelDB storage for messages and peers
│   ├── p2p/
│   │   ├── node.go         # Libp2p node setup and peer discovery
│   │   ├── protocol.go     # Custom libp2p protocols for chat and file transfer
│   │   └── dht.go          # DHT for peer discovery and username mapping
│   ├── chat/
│   │   ├── private.go      # Private P2P chat logic
│   │   ├── group.go        # Group chat logic
│   │   └── file.go         # File transfer logic
│   └── cli/
│       ├── root.go         # Root CLI command
│       ├── init.go         # CLI command for initialization
│       ├── serve.go        # CLI command to start the node
│       └── bootnode.go     # CLI command for bootstrap node
├── frontend/               # Svelte frontend application
├── static/                 # Built frontend files (served by backend)
├── go.mod                  # Go module file
├── go.sum                  # Go module checksums
└── README.md               # Project documentation
```

## Getting Started

### Prerequisites

- Go 1.22 or later
- Node.js 18+ (for frontend development)

### Installation

1. Clone or download the project files
2. Navigate to the project directory:
   ```bash
   cd p2p-chat
   ```

3. Install Go dependencies:
   ```bash
   go mod tidy
   ```

4. Build the application:
   ```bash
   go build -o p2p-chat ./cmd/chat
   ```

### Usage

#### 1. Initialize a New Node

Before starting a chat node, you need to initialize the encryption keys and database:

```bash
./p2p-chat init --db ./my-node-db
```

This command will:
- Generate ECDSA encryption keys
- Generate libp2p identity keys
- Store them securely in the specified LevelDB database

#### 2. Start a Chat Node

Start your P2P chat node with:

```bash
./p2p-chat serve --datadir ./my-node-db --rest-port 8080 --ws-port 8081 --libp2p-port 4001 --username myusername
```

Parameters:
- `--datadir`: Path to the LevelDB database (same as used in init)
- `--rest-port`: Port for REST API (default: 8080)
- `--ws-port`: Port for WebSocket API (default: 8081)
- `--libp2p-port`: Port for libp2p networking (0 for random port)
- `--username`: Your username on the network (optional, generates random if not provided)

#### 3. Access the Web Interface

Once your node is running, open your web browser and navigate to:
```
http://localhost:8080
```

The web interface provides:
- **Chat Tab**: Send private messages to connected peers
- **Peers Tab**: Connect to new peers and search for existing ones
- **Groups Tab**: Create and manage group chats
- **Files Tab**: Send and receive files
- **Node Info Tab**: View your node's network information

#### 4. Connect to Other Peers

To connect to another peer, you need their multiaddress. In the web interface:

1. Go to the "Peers" tab
2. Click "Connect to Peer"
3. Enter the peer's multiaddress (format: `/ip4/IP/tcp/PORT/p2p/PEER_ID`)
4. Click "Connect"

You can find your own multiaddress in the "Node Info" tab.

#### 5. Start a Bootstrap Node (Optional)

To help with peer discovery, you can run a bootstrap node:

```bash
./p2p-chat bootnode --port 4001
```

This creates a standalone node that helps other peers discover each other.

### API Endpoints

The application provides both REST and WebSocket APIs:

#### REST API Endpoints

- `POST /peer/connect` - Connect to a peer
- `GET /peer/search` - Search for peers
- `POST /chat/private/send` - Send private message
- `POST /group/create` - Create a group
- `POST /group/add_member` - Add member to group
- `POST /group/send_message` - Send group message
- `POST /file/send` - Send a file

#### WebSocket API

Connect to `ws://localhost:8081/ws` for real-time updates:
- Peer connection status
- New message notifications
- Group updates
- File transfer status

### Example Usage Scenario

1. **Node A** initializes and starts:
   ```bash
   ./p2p-chat init --db ./nodeA-db
   ./p2p-chat serve --datadir ./nodeA-db --libp2p-port 4001 --username alice
   ```

2. **Node B** initializes and starts:
   ```bash
   ./p2p-chat init --db ./nodeB-db
   ./p2p-chat serve --datadir ./nodeB-db --libp2p-port 4002 --username bob
   ```

3. **Node B** connects to **Node A** using Node A's multiaddress
4. Both nodes can now chat privately or create group chats
5. Files can be transferred between the connected nodes

### Security Features

- **End-to-end encryption**: All communications use ECDSA encryption
- **Decentralized architecture**: No central servers or single points of failure
- **Peer authentication**: Cryptographic verification of peer identities
- **Local data storage**: All data stored locally using LevelDB

### Troubleshooting

1. **Connection issues**: Ensure firewall allows the specified ports
2. **Peer discovery problems**: Try connecting to peers manually using their multiaddresses
3. **Database errors**: Make sure the database directory is writable
4. **Port conflicts**: Use different ports if the default ones are occupied

### Development

To modify the frontend:

1. Navigate to the frontend directory:
   ```bash
   cd frontend
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Start development server:
   ```bash
   npm run dev
   ```

4. Build for production:
   ```bash
   npm run build
   ```

5. Copy built files to static directory:
   ```bash
   cp -r build/* ../static/
   ```

### Contributing

This is a fully functional P2P chat application with all requested features implemented. The codebase is modular and extensible for future enhancements.

### License

This project is provided as-is for educational and development purposes.


