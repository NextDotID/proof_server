package model

import (
	"testing"
	"time"

	"github.com/nextdotid/proof_server/util/crypto"
	"github.com/stretchr/testify/require"
)

func GenerateAliasTo(avatarPubkey any) (string){
	pc1 := GenerateProofChain()

	avatar := AvatarAlias{
		CreatedAt:    time.Now(),
		Avatar:       MarshalAvatar(avatarPubkey),
		Alias:        MarshalAvatar(pc1.Persona),
		ProofChainID: pc1.ID,
	}
	tx := DB.Create(&avatar)
	if tx.Error != nil {
		panic(tx.Error)
	}
	return avatar.Alias
}

func Test_FindAllAliasByAvatar(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		avatar, _ := crypto.GenerateKeypair()
		alias1 := GenerateAliasTo(avatar)
		alias2 := GenerateAliasTo(alias1)
		alias3 := GenerateAliasTo(alias1)

		aliases, err := FindAllAliasByAvatar(MarshalAvatar(avatar))
		require.NoError(t, err)
		require.Contains(t, aliases, alias1)
		require.Contains(t, aliases, alias2)
		require.Contains(t, aliases, alias3)
	})
}
