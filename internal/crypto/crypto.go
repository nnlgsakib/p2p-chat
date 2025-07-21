package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"github.com/libp2p/go-libp2p/core/crypto"
)

// GenerateKeyPair generates a new ECDSA private and public key pair.
func GenerateKeyPair() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

// EncodePrivateKey encodes a private key to a hex string.
func EncodePrivateKey(priv *ecdsa.PrivateKey) (string, error) {
	der, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(der), nil
}

// DecodePrivateKey decodes a hex string to a private key.
func DecodePrivateKey(hexKey string) (*ecdsa.PrivateKey, error) {
	der, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, err
	}
	return x509.ParseECPrivateKey(der)
}

// EncodePublicKey encodes a public key to a hex string.
func EncodePublicKey(pub *ecdsa.PublicKey) (string, error) {
	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(der), nil
}

// DecodePublicKey decodes a hex string to a public key.
func DecodePublicKey(hexKey string) (*ecdsa.PublicKey, error) {
	der, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, err
	}
	pub, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return nil, err
	}
	return pub.(*ecdsa.PublicKey), nil
}




// GenerateKeyPairLibp2p generates a new libp2p private and public key pair.
func GenerateKeyPairLibp2p() (crypto.PrivKey, crypto.PubKey, error) {
	return crypto.GenerateKeyPair(crypto.Ed25519, -1)
}

// MarshalPrivateKeyLibp2p marshals a libp2p private key to bytes.
func MarshalPrivateKeyLibp2p(priv crypto.PrivKey) ([]byte, error) {
	return crypto.MarshalPrivateKey(priv)
}

// UnmarshalPrivateKeyLibp2p unmarshals bytes to a libp2p private key.
func UnmarshalPrivateKeyLibp2p(data []byte) (crypto.PrivKey, error) {
	return crypto.UnmarshalPrivateKey(data)
}

// MarshalPublicKeyLibp2p marshals a libp2p public key to bytes.
func MarshalPublicKeyLibp2p(pub crypto.PubKey) ([]byte, error) {
	return crypto.MarshalPublicKey(pub)
}

// UnmarshalPublicKeyLibp2p unmarshals bytes to a libp2p public key.
func UnmarshalPublicKeyLibp2p(data []byte) (crypto.PubKey, error) {
	return crypto.UnmarshalPublicKey(data)
}


