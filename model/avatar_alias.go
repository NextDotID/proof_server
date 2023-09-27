package model

import (
	"database/sql"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/samber/lo"
	"golang.org/x/xerrors"
)

type AvatarAlias struct {
	ID        int64     `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"column:created_at"`

	// Avatar public key ("0xPUBKEY_COMPRESSED_HEXSTRING")
	Avatar string `gorm:"column:avatar;not null;index"`
	// Alias avatar of this avatar ("0xPUBKEY_COMPRESSED_HEXSTRING")
	Alias string `gorm:"column:alias;not null;index"`

	// ProofChain record which creates this binding
	ProofChainID int64 `gorm:"index"`
	// ProofChain record which creates this binding
	ProofChain ProofChain
}

func (AvatarAlias) TableName() string {
	return "alias"
}

// FindAllAliasByAvatar finds all alias of an avatar
func FindAllAliasByAvatar(origAvatar string) ([]string, error) {
	originalAvatar := MarshalAvatar(origAvatar)
	if originalAvatar == "" {
		return nil, xerrors.Errorf("invalid avatar")
	}

	aliases := mapset.NewSet(originalAvatar)
	avatarsToQuery := mapset.NewSet(originalAvatar)

	for {
		aliasInstances := make([]AvatarAlias, 0)
		tx := ReadOnlyDB.Model(&AvatarAlias{}).
			Where(
				"avatar IN @avatarsToQuery OR alias IN @avatarsToQuery",
				sql.Named("avatarsToQuery", avatarsToQuery.ToSlice()),
			).
			Find(&aliasInstances)
		if tx.Error != nil {
			return nil, tx.Error
		}
		avatarsToQuery.Clear()
		lo.ForEach(aliasInstances, func(avatarAlias AvatarAlias, index int) {
			if aliases.Add(avatarAlias.Alias) {
				avatarsToQuery.Add(avatarAlias.Alias)
			}
			if aliases.Add(avatarAlias.Avatar) {
				avatarsToQuery.Add(avatarAlias.Avatar)
			}
		})
		if avatarsToQuery.Cardinality() == 0 {
			break
		}
	}
	aliases.Remove(originalAvatar)
	return aliases.ToSlice(), nil
}
