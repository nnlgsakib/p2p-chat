package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	cryptoLocal "p2p-chat/internal/crypto"
	"p2p-chat/internal/db"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize encryption keys and libp2p key",
	Run: func(cmd *cobra.Command, args []string) {
		dbPath, _ := cmd.Flags().GetString("db")
		if dbPath == "" {
			fmt.Println("Error: --db flag is required for database path.")
			return
		}

		store, err := db.NewLevelDBStore(dbPath)
		if err != nil {
			fmt.Printf("Error opening database: %v\n", err)
			return
		}
		defer store.Close()

		// Generate ECDSA keys for encryption
		privKeyECDSA, err := cryptoLocal.GenerateKeyPair()
		if err != nil {
			fmt.Printf("Error generating ECDSA keys: %v\n", err)
			return
		}
		privKeyHex, err := cryptoLocal.EncodePrivateKey(privKeyECDSA)
		if err != nil {
			fmt.Printf("Error encoding ECDSA private key: %v\n", err)
			return
		}
		pubKeyHex, err := cryptoLocal.EncodePublicKey(&privKeyECDSA.PublicKey)
		if err != nil {
			fmt.Printf("Error encoding ECDSA public key: %v\n", err)
			return
		}

		// Store ECDSA keys
		store.Put([]byte("ecdsa_private_key"), []byte(privKeyHex))
		store.Put([]byte("ecdsa_public_key"), []byte(pubKeyHex))
		fmt.Println("ECDSA encryption keys generated and stored.")

		// Generate libp2p keys
		privKeyLibp2p, pubKeyLibp2p, err := cryptoLocal.GenerateKeyPairLibp2p()
		if err != nil {
			fmt.Printf("Error generating libp2p keys: %v\n", err)
			return
		}
		privKeyLibp2pBytes, err := cryptoLocal.MarshalPrivateKeyLibp2p(privKeyLibp2p)
		if err != nil {
			fmt.Printf("Error marshaling libp2p private key: %v\n", err)
			return
		}
		pubKeyLibp2pBytes, err := cryptoLocal.MarshalPublicKeyLibp2p(pubKeyLibp2p)
		if err != nil {
			fmt.Printf("Error marshaling libp2p public key: %v\n", err)
			return
		}

		// Store libp2p keys
		store.Put([]byte("libp2p_private_key"), privKeyLibp2pBytes)
		store.Put([]byte("libp2p_public_key"), pubKeyLibp2pBytes)
		fmt.Println("Libp2p keys generated and stored.")
	},
}

func init() {
	initCmd.Flags().String("db", "", "Path to the LevelDB database")
	RootCmd.AddCommand(initCmd)
}


