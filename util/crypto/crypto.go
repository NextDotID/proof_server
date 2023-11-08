package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/xerrors"
)

// ValidatePersonalSignature checks whether (eth.personal.sign) signature,
// payload and pubkey are matched.
// Pubkey and signature should be without "0x".
func ValidatePersonalSignature(payload string, signature []byte, pubkey *ecdsa.PublicKey) (err error) {
	pubkeyRecovered, err := RecoverPubkeyFromPersonalSignature(payload, signature)
	if err != nil {
		return xerrors.Errorf("%w", err)
	}

	if crypto.PubkeyToAddress(*pubkey) != crypto.PubkeyToAddress(*pubkeyRecovered) {
		return xerrors.Errorf("bad signature")
	}
	return nil
}

// RecoverPubkeyFromPersonalSignature extract a public key from signature
func RecoverPubkeyFromPersonalSignature(payload string, signature []byte) (pubkey *ecdsa.PublicKey, err error) {
	// Recover pubkey from signature
	if len(signature) != 65 {
		return nil, xerrors.Errorf("Error: Signature length invalid: %d instead of 65", len(signature))
	}
	if signature[64] == 27 || signature[64] == 28 {
		signature[64] -= 27
	}

	if signature[64] != 0 && signature[64] != 1 {
		return nil, xerrors.Errorf("Error: Signature Recovery ID not supported: %d", signature[64])
	}

	pubkeyRecovered, err := crypto.SigToPub(signPersonalHash([]byte(payload)), signature)
	if err != nil {
		return nil, xerrors.Errorf("Error when recovering pubkey from signature: %s", err.Error())
	}

	return pubkeyRecovered, nil
}

// GenerateSecp256k1Keypair generates a keypair.
// For test purpose only.
func GenerateSecp256k1Keypair() (publicKey *ecdsa.PublicKey, privateKey *ecdsa.PrivateKey) {
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

// StringToSecp256k1Pubkey is compatible with comressed / uncompressed pubkey
// hex, and with / without '0x' head.
func StringToSecp256k1Pubkey(pkHex string) (*ecdsa.PublicKey, error) {
	pkBytes := common.Hex2Bytes(strings.ToLower(strings.TrimPrefix(pkHex, "0x")))
	return BytesToSecp256k1PubKey(pkBytes)
}

// BytesToSecp256k1PubKey is compatible with comressed / uncompressed pubkey
// bytes.
func BytesToSecp256k1PubKey(pkBytes []byte) (result *ecdsa.PublicKey, err error) {
	if len(pkBytes) == 33 { // compressed
		result, err = crypto.DecompressPubkey(pkBytes)
	} else {
		result, err = crypto.UnmarshalPubkey(pkBytes)
	}
	return
}

// CompressedPubkeyHex has no "0x".
func CompressedPubkeyHex(pk *ecdsa.PublicKey) string {
	return common.Bytes2Hex(crypto.CompressPubkey(pk))
}

// StringToSecp256r1Pubkey is compatible with
// `X_CONCAT_Y_64_BYTES_HEXSTRING` public key representation.
func StringToSecp256r1Pubkey(pkHex string) (*ecdsa.PublicKey, error) {
	pkBinary := common.FromHex(strings.ToLower(strings.TrimPrefix(pkHex, "0x")))
	if len(pkBinary) != 64 { // X:32 Y:32
		return nil, xerrors.Errorf("wrong public key length: expect 64, got %d", len(pkBinary))
	}
	return &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(pkBinary[:32]),
		Y:     new(big.Int).SetBytes(pkBinary[32:]),
	}, nil
}
