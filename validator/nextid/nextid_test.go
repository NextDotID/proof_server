package nextid

import (
	"crypto/ecdsa"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/nextdotid/proof_server/types"
	mycrypto "github.com/nextdotid/proof_server/util/crypto"
	"github.com/nextdotid/proof_server/validator"
	"github.com/stretchr/testify/require"
)

func GenerateNextIDTestData() (nextid *NextID, avatarSK, targetAvatarSK *ecdsa.PrivateKey) {
	avatar, avatarSK := mycrypto.GenerateSecp256k1Keypair()
	targetAvatar, targetAvatarSK := mycrypto.GenerateSecp256k1Keypair()
	nextid = &NextID{
		&validator.Base{
			Platform:         types.Platforms.NextID,
			Previous:         "",
			Action:           types.Actions.Create,
			Pubkey:           avatar,
			Identity:         mycrypto.CompressedPubkeyHex(targetAvatar),
			AltID:            "",
			ProofLocation:    "",
			Signature:        []byte{},
			SignaturePayload: "",
			Text:             "",
			Extra:            map[string]string{"target_signature": ""},
			CreatedAt:        time.Now(),
			Uuid:             uuid.New(),
		},
	}

	nextid.SignaturePayload = nextid.GenerateSignPayload()
	nextid.Signature, _ = mycrypto.SignPersonal([]byte(nextid.SignaturePayload), avatarSK)
	targetSig, _ := mycrypto.SignPersonal([]byte(nextid.SignaturePayload), targetAvatarSK)
	nextid.Extra["target_signature"] = common.Bytes2Hex(targetSig)

	return nextid, avatarSK, targetAvatarSK
}

func Test_Validate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		nextid, _, _ := GenerateNextIDTestData()
		require.NoError(t, nextid.Validate())
	})
}
