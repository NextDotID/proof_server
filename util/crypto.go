package util

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

// ValidatePersonalSignature checks whether (eth.personal.sign) signature,
// payload and pubkey are matched.  Pubkey should be "0x...." string.
func ValidatePersonalSignature(payload, signature, pubkey string) bool {
	pubkeyGiven, err := crypto.UnmarshalPubkey(common.Hex2Bytes(pubkey))
	if err != nil {
		logrus.Warnf("Error when converting pubkey: %s", err.Error())
		return false
	}

	// Recover pubkey from signature
	signBytes := common.Hex2Bytes(signature)
	if signBytes[64] != 27 && signBytes[64] != 28 {
		logrus.Warn("Error: Signature Recovery ID not supported")
		return false
	}
	signBytes[64] -= 27

	pubkeyRecovered, err := crypto.SigToPub(signPersonalHash([]byte(payload)), signBytes)
	if err != nil {
		logrus.Warnf("Error when recovering pubkey from signature: %s", err.Error())
		return false
	}

	return pubkeyGiven.Equal(&pubkeyRecovered)
}

// GenerateKeypair generates a keypair.
// For test purpose only.
func GenerateKeypair() (publicKey *ecdsa.PublicKey, privateKey *ecdsa.PrivateKey) {
	privateKey, _ = crypto.GenerateKey()
	publicKey = &privateKey.PublicKey
	return publicKey, privateKey
}

// SignPersonal signs a payload using given secret key.
// For test purpose only.
func SignPersonal(payload []byte, sk *ecdsa.PrivateKey) (signature []byte, err error) {
	hash := signPersonalHash(payload)
	signature, err = crypto.Sign(hash, sk)
	if err != nil {
		return nil, xerrors.Errorf("%w", err)
	}

	return signature, nil
}

func signPersonalHash(data []byte) []byte {
	messsage := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)
	return crypto.Keccak256([]byte(messsage))
}
