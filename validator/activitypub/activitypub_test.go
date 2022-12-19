package activitypub

import (
	"testing"

	"github.com/google/uuid"
	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/util"
	"github.com/nextdotid/proof_server/util/crypto"
	"github.com/nextdotid/proof_server/validator"
	"github.com/stretchr/testify/require"
)

func GenerateMisskeyRecord() (ap *ActivityPub) {
	pk, _ := crypto.StringToPubkey("03c683a83bdf0abae3c344855b55b5978fd22fbedae575bd1f540f919afbc19015")
	ca, _ := util.TimestampStringToTime("1671356397")
	uuid := uuid.MustParse("4d89b36a-4e55-4c1f-93c8-c81b08f71b09")

	return &ActivityPub{
		Base: &validator.Base{
			Platform:      types.Platforms.ActivityPub,
			Action:        types.Actions.Create,
			Pubkey:        pk,
			Identity:      "nykma@t.nyk.app",
			ProofLocation: "98wr1tkc82",
			CreatedAt:     ca,
			Uuid:          uuid,
		},
	}
}

func Test_ExtractSignature(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ap := ActivityPub{
			Base: &validator.Base{
				Text: "Validate my ActivityPub identity @%s for Avatar 0x%s:\n\nSignature: dGVzdDEyMw==\nUUID:%s\nPrevious:%s\nCreatedAt:%d\n\nPowered by Next.ID - Connect All Digital Identities.\n",
			},
		}
		require.NoError(t, ap.ExtractSignature())
		require.Equal(t, "test123", string(ap.Base.Signature))
	})
}

func Test_Validate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ap := GenerateMisskeyRecord()
		require.NoError(t, ap.Validate())
		require.Equal(t, ap.AltID, "8zwtspqtym")
	})
}
