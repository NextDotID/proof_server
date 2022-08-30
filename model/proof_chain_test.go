package model

import (
	"database/sql"
	"encoding/base64"
	"reflect"
	"testing"

	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/util"
	"github.com/nextdotid/proof_server/util/crypto"
	"github.com/nextdotid/proof_server/validator"
	"github.com/stretchr/testify/assert"
)

func Test_ProofChainFindLatest(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		pk, _ := crypto.StringToPubkey("0x028c3cda474361179d653c41a62f6bbb07265d535121e19aedf660da2924d0b1e3")

		proof := ProofChain{
			Persona:   "0x" + crypto.CompressedPubkeyHex(pk),
			Platform:  "twitter",
			Identity:  "yeiwb",
			Location:  "1469221200140574721",
			Signature: "gMUJ75eewkdaNrFp7bafzckv9+rlW7rVaxkB7/sYzYgFdFltYG+gn0lYzVNgrAdHWZPmu2giwJniGG7HG9iNigE=",
		}
		tx := DB.Create(&proof)
		assert.Nil(t, tx.Error)

		pc, err := ProofChainFindLatest("0x" + crypto.CompressedPubkeyHex(pk))
		assert.Nil(t, err)
		assert.Equal(t, pc.Identity, proof.Identity)
	})
}

func Test_ProofChainFindBySignature(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		pk, _ := crypto.StringToPubkey("0x028c3cda474361179d653c41a62f6bbb07265d535121e19aedf660da2924d0b1e3")

		proof := ProofChain{
			Persona:   "0x" + crypto.CompressedPubkeyHex(pk),
			Platform:  "twitter",
			Identity:  "yeiwb",
			Location:  "1469221200140574721",
			Signature: "gMUJ75eewkdaNrFp7bafzckv9+rlW7rVaxkB7/sYzYgFdFltYG+gn0lYzVNgrAdHWZPmu2giwJniGG7HG9iNigE=",
		}
		tx := DB.Create(&proof)
		assert.Nil(t, tx.Error)
		assert.Nil(t, proof.Previous)

		proof_found, err := ProofChainFindBySignature("gMUJ75eewkdaNrFp7bafzckv9+rlW7rVaxkB7/sYzYgFdFltYG+gn0lYzVNgrAdHWZPmu2giwJniGG7HG9iNigE=")
		assert.Nil(t, err)
		assert.Equal(t, proof.Persona, proof_found.Persona)
	})

	t.Run("should return empty result", func(t *testing.T) {
		before_each(t)

		proof, err := ProofChainFindBySignature("0xfoobar")
		assert.Nil(t, proof)
		assert.Contains(t, err.Error(), "record not found")
	})
}

func Test_ProofChainCreateFromValidator(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		pk, _ := crypto.StringToPubkey("0x028c3cda474361179d653c41a62f6bbb07265d535121e19aedf660da2924d0b1e3")

		v := validator.Base{
			Platform:      types.Platforms.Twitter,
			Action:        types.Actions.Create,
			Pubkey:        pk,
			Identity:      "yeiwb",
			ProofLocation: "1469221200140574721",
			Signature:     []byte{1, 2, 3, 4},
			Text:          "",
		}
		pc, err := ProofChainCreateFromValidator(&v)
		assert.Nil(t, err)
		assert.Equal(t, "yeiwb", pc.Identity)
		assert.Equal(t, base64.StdEncoding.EncodeToString(v.Signature), pc.Signature)
		assert.Nil(t, pc.Previous)
		assert.Equal(t, MarshalPersona(pk), pc.Persona)
		assert.Equal(t, "{}", pc.Extra.String())
	})

	t.Run("save extra", func(t *testing.T) {
		before_each(t)
		pk, _ := crypto.StringToPubkey("0x028c3cda474361179d653c41a62f6bbb07265d535121e19aedf660da2924d0b1e3")

		v := validator.Base{
			Platform:      types.Platforms.Ethereum,
			Action:        types.Actions.Create,
			Pubkey:        pk,
			Identity:      "0xWALLET_ADDRESS",
			ProofLocation: "",
			Signature:     []byte{1, 2, 3, 4},
			Text:          "",
			Extra: map[string]string{
				"wallet_signature": "0xTEST",
			},
		}
		pc, err := ProofChainCreateFromValidator(&v)
		assert.Nil(t, err)
		assert.Equal(t, types.Platforms.Ethereum, pc.Platform)
		assert.Equal(t, `{"wallet_signature": "0xTEST"}`, pc.Extra.String())
	})

	t.Run("with previous connected", func(t *testing.T) {
		before_each(t)
		pk, _ := crypto.StringToPubkey("0x028c3cda474361179d653c41a62f6bbb07265d535121e19aedf660da2924d0b1e3")
		v := validator.Base{
			Platform:      types.Platforms.Twitter,
			Action:        types.Actions.Create,
			Pubkey:        pk,
			Identity:      "yeiwb",
			ProofLocation: "1469221200140574721",
			Signature:     []byte{1, 2, 3, 4},
			Text:          "",
		}
		prev, err := ProofChainCreateFromValidator(&v)
		assert.Nil(t, err)

		v2 := validator.Base{
			Previous:      MarshalSignature(v.Signature),
			Platform:      types.Platforms.Twitter,
			Action:        types.Actions.Delete,
			Pubkey:        pk,
			Identity:      "yeiwb",
			ProofLocation: "1469221200140574721",
			Signature:     []byte{5, 6, 7, 8},
			Text:          "",
		}
		current, err := ProofChainCreateFromValidator(&v2)
		assert.Nil(t, err)
		assert.Equal(t, prev.ID, current.Previous.ID)
		assert.Equal(t, prev.ID, current.PreviousID.Int64)
	})

	t.Run("cannot connect to previous", func(t *testing.T) {
		before_each(t)

		pk, _ := crypto.StringToPubkey("0x028c3cda474361179d653c41a62f6bbb07265d535121e19aedf660da2924d0b1e3")
		v := validator.Base{
			Previous:      MarshalSignature([]byte{1, 2, 3, 4}),
			Platform:      types.Platforms.Twitter,
			Action:        types.Actions.Delete,
			Pubkey:        pk,
			Identity:      "yeiwb",
			ProofLocation: "1469221200140574721",
			Signature:     []byte{5, 6, 7, 8},
			Text:          "",
		}
		pc, err := ProofChainCreateFromValidator(&v)
		assert.Nil(t, pc)
		assert.Contains(t, err.Error(), "record not found")
	})
}

func Test_Apply(t *testing.T) {
	t.Run("create and delete", func(t *testing.T) {
		before_each(t)
		pk, _ := crypto.StringToPubkey("0x028c3cda474361179d653c41a62f6bbb07265d535121e19aedf660da2924d0b1e3")

		pc := ProofChain{
			Action:    types.Actions.Create,
			Persona:   MarshalPersona(pk),
			Identity:  "yeiwb",
			Location:  "1469221200140574721",
			Platform:  types.Platforms.Twitter,
			Signature: MarshalSignature([]byte{1, 2, 3, 4}),
		}
		tx := DB.Create(&pc)
		assert.Nil(t, tx.Error)

		err := pc.Apply()
		assert.Nil(t, err)

		proof_found := Proof{ProofChainID: pc.ID}
		DB.First(&proof_found)

		assert.NotZero(t, proof_found.ID)
		assert.Equal(t, pc.Location, proof_found.Location)
		assert.Equal(t, pc.Identity, proof_found.Identity)

		// Duplicated apply
		err = pc.Apply()
		assert.Nil(t, err)
		var count int64
		DB.Model(&Proof{}).Where("proof_chain_id = ?", pc.ID).Count(&count)
		assert.Equal(t, int64(1), count)

		// Delete
		pc_delete := ProofChain{
			Action:    types.Actions.Delete,
			Persona:   MarshalPersona(pk),
			Identity:  "yeiwb",
			Location:  "1469221200140574721",
			Platform:  types.Platforms.Twitter,
			Signature: MarshalSignature([]byte{1, 2, 3, 4}),
		}
		pc_delete.Apply()
		DB.Model(&Proof{}).Where("proof_chain_id = ?", pc.ID).Count(&count)
		assert.Equal(t, int64(0), count)
	})

	t.Run("avoid duplicate", func(t *testing.T) {
		before_each(t)
		pk, _ := crypto.GenerateKeypair()
		pc := ProofChain{
			Action:    types.Actions.Create,
			Persona:   MarshalPersona(pk),
			Identity:  "yeiwb",
			Location:  "1469221200140574721",
			Platform:  types.Platforms.Twitter,
			Signature: MarshalSignature([]byte{1}),
		}
		assert.Nil(t, DB.Create(&pc).Error)
		assert.Nil(t, pc.Apply())

		pc2 := ProofChain{
			Action:     types.Actions.Create,
			Persona:    MarshalPersona(pk),
			Identity:   "yeiwb",
			Platform:   types.Platforms.Twitter,
			Location:   "1469221200140574721",
			Signature:  MarshalSignature([]byte{1}),
			PreviousID: sql.NullInt64{Int64: pc.ID, Valid: true},
			Previous:   &pc,
		}
		assert.Nil(t, DB.Create(&pc2).Error)
		assert.Nil(t, pc.Apply())

		var count int64
		DB.Model(&Proof{}).Count(&count)
		assert.Equal(t, int64(1), count)
	})
}

func Test_ProofChain_RestoreValidator(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		pk, _ := crypto.StringToPubkey("0x04666b700aeb6a6429f13cbb263e1bc566cd975a118b61bc796204109c1b351d19b7df23cc47f004e10fef41df82bad646b027578f8881f5f1d2f70c80dfcd8031")
		created_at, _ := util.TimestampStringToTime("1647503071")
		pc := ProofChain{
			Action:    types.Actions.Create,
			Persona:   MarshalPersona(pk),
			Identity:  "yeiwb",
			Location:  "1504363098328924163",
			Platform:  types.Platforms.Twitter,
			Signature: "D8i0UOXKrHJ23zCQe6USZDrw7fOjwm4R/eVX0AZXKgomynWWm+Px4Y7I1wtbsHwKj0t9psFqm87EnM93DXOmhwE=",
			Uuid:      "c6fa1483-1bad-4f07-b661-678b191ab4b3",
			CreatedAt: created_at,
		}
		tx := DB.Create(&pc)
		assert.Nil(t, tx.Error)

		v, err := pc.RestoreValidator()
		assert.Nil(t, err)
		assert.Equal(t, v.Platform, types.Platforms.Twitter)
		assert.Equal(t, v.Identity, pc.Identity)
		assert.Equal(t, len(v.Signature), 65)
	})
}

func TestProofChainFindByPersona(t *testing.T) {
	//pk, _ := crypto.StringToPubkey("0x04666b700aeb6a6429f13cbb263e1bc566cd975a118b61bc796204109c1b351d19b7df23cc47f004e10fef41df82bad646b027578f8881f5f1d2f70c80dfcd8031")
	type args struct {
		persona  string
		all_data bool
		from     int
		limit    int
	}
	tests := []struct {
		name      string
		args      args
		wantTotal int64
		wantRs    []ProofChainItem
		wantErr   bool
	}{
		{
			name:      "empty result",
			args:      args{persona: "234234", all_data: false, from: 0, limit: 10},
			wantTotal: 0,
			wantRs:    []ProofChainItem{},
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTotal, gotRs, err := ProofChainFindByPersona(tt.args.persona, tt.args.all_data, tt.args.from, tt.args.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProofChainFindByPersona() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotTotal != tt.wantTotal {
				t.Errorf("ProofChainFindByPersona() gotTotal = %v, want %v", gotTotal, tt.wantTotal)
			}
			if !reflect.DeepEqual(gotRs, tt.wantRs) {
				t.Errorf("ProofChainFindByPersona() gotRs = %v, want %v", gotRs, tt.wantRs)
			}
		})
	}
}
