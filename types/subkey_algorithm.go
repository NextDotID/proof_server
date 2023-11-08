package types

type SubkeyAlgorithm string

var SubkeyAlgorithms = struct {
	Secp256R1 SubkeyAlgorithm
	Secp256K1 SubkeyAlgorithm
}{
	Secp256R1: "es256",
	Secp256K1: "secp256k1",
}
