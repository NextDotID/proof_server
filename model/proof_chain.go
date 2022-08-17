package model

import (
	"crypto/ecdsa"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/util/crypto"
	"github.com/nextdotid/proof-server/validator"
	"github.com/samber/lo"
	"golang.org/x/xerrors"
	"gorm.io/datatypes"
)

// ProofChain is a chain of a persona's proof modification log.
type ProofChain struct {
	ID               int64          `gorm:"primarykey"`
	CreatedAt        time.Time      `gorm:"column:created_at"`
	Action           types.Action   `gorm:"index;not null"`
	Persona          string         `gorm:"index;not null"`
	Identity         string         `gorm:"index;not null"`
	Platform         types.Platform `gorm:"index;not null"`
	Location         string         `gorm:"not null"`
	Signature        string         `gorm:"not null"`
	SignaturePayload string         `gorm:"column:signature_payload"`
	Extra            datatypes.JSON `gorm:"default:'{}'"`
	Uuid             string         `gorm:"index;column:uuid"`
	ArweaveID        string         `gorm:"column:arweave_id;not null;default:''"`
	PreviousID       sql.NullInt64  `gorm:"index"`
	Previous         *ProofChain
}

// Output version of the proof chain
type ProofChainItem struct {
	Action           types.Action   `json:"action"`
	Platform         types.Platform `json:"platform"`
	Identity         string         `json:"identity"`
	ProofLocation    string         `json:"proof_location"`
	CreatedAt        string         `json:"created_at"`
	Signature        string         `json:"signature"`
	SignaturePayload string         `json:"signature_payload"`
	Uuid             string         `json:"uuid"`
	Extra            datatypes.JSON `json:"extra"`
	ArweaveID        string         `json:"arweave_id"`
}

// Arweave data ID
type ProofChainArweaveDocument struct {
	Avatar            string         `json:"avatar"`
	Action            types.Action   `json:"action"`
	Platform          types.Platform `json:"platform"`
	Identity          string         `json:"identity"`
	ProofLocation     string         `json:"proof_location"`
	CreatedAt         string         `json:"created_at"`
	Signature         string         `json:"signature"`
	SignaturePayload  string         `json:"signature_payload"`
	Uuid              string         `json:"uuid"`
	Extra             datatypes.JSON `json:"extra"`
	PreviousUuid      string         `json:"previous_uuid"`
	PreviousArweaveID string         `json:"previous_arweave_id"`
}

func (ProofChain) TableName() string {
	return "proof_chains"
}

func (pc *ProofChain) Pubkey() *ecdsa.PublicKey {
	pubkey, err := crypto.StringToPubkey(pc.Persona)
	if err != nil {
		return nil
	}
	return pubkey
}

// Apply applies current ProofChain modification to Proof model.
func (pc *ProofChain) Apply() (err error) {
	switch pc.Action {
	case types.Actions.Create:
		return pc.createProof()
	case types.Actions.Delete:
		return pc.deleteProof()
	default:
		return xerrors.Errorf("unknown action: %s", string(pc.Action))
	}
}

func (pc *ProofChain) ToProofChainItem() ProofChainItem {
	return ProofChainItem{
		Action:           pc.Action,
		Platform:         pc.Platform,
		Identity:         pc.Identity,
		ProofLocation:    pc.Location,
		CreatedAt:        strconv.FormatInt(pc.CreatedAt.Unix(), 10),
		Signature:        pc.Signature,
		SignaturePayload: pc.SignaturePayload,
		Uuid:             pc.Uuid,
		Extra:            pc.Extra,
		ArweaveID:        pc.ArweaveID,
	}
}

func (pc *ProofChain) createProof() (err error) {
	proof_condition := Proof{
		Persona:  pc.Persona,
		Platform: pc.Platform,
		Identity: pc.Identity,
		Location: pc.Location,
	}
	proof_create := &Proof{
		ProofChainID:  pc.ID,
		Persona:       pc.Persona,
		Platform:      pc.Platform,
		Identity:      pc.Identity,
		Location:      pc.Location,
		LastCheckedAt: time.Now(),
		IsValid:       true,
		InvalidReason: "",
	}
	tx := DB.FirstOrCreate(proof_create, proof_condition)
	if tx.Error != nil {
		return xerrors.Errorf("%w", tx.Error)
	}

	return nil
}

func (pc *ProofChain) deleteProof() (err error) {
	tx := DB.Delete(&Proof{}, Proof{
		Persona:  pc.Persona,
		Platform: pc.Platform,
		Identity: pc.Identity,
		Location: pc.Location,
	})
	if tx.Error != nil {
		return xerrors.Errorf("%w", tx.Error)
	}
	return nil
}

func (pc *ProofChain) SignatureBytes() (sig []byte) {
	byte, err := base64.StdEncoding.DecodeString(pc.Signature)
	if err != nil {
		return nil
	}
	return byte
}

// RestoreValidator rebuilds `validator.Base` from current `ProofChain` record.
func (pc *ProofChain) RestoreValidator() (v *validator.Base, err error) {
	previousSig := ""
	if pc.Previous != nil {
		previousSig = pc.Previous.Signature
	}

	extra := map[string]string{}
	if pc.Extra.String() != "" {
		err = json.Unmarshal([]byte(pc.Extra.String()), &extra)
		if err != nil {
			return nil, xerrors.Errorf("%w", err)
		}
	}
	parsedUuid, err := uuid.Parse(pc.Uuid)
	if err != nil {
		return nil, xerrors.Errorf("UUID parse error: %w", err)
	}
	v = &validator.Base{
		Platform:      pc.Platform,
		Previous:      previousSig,
		Action:        pc.Action,
		Pubkey:        pc.Pubkey(),
		Identity:      pc.Identity,
		ProofLocation: pc.Location,
		Signature:     pc.SignatureBytes(),
		Extra:         extra,
		Uuid:          parsedUuid,
		CreatedAt:     pc.CreatedAt,
	}

	return v, nil
}

// MarshalPersona accepts *ecdsa.Pubkey | string type of pubkey,
// returns a string to be stored into DB.
func MarshalPersona(persona any) string {
	switch p := persona.(type) {
	case *ecdsa.PublicKey:
		return "0x" + crypto.CompressedPubkeyHex(p)
	case string:
		pubkey, err := crypto.StringToPubkey(p)
		if err != nil {
			return ""
		}
		return MarshalPersona(pubkey)
	default:
		return ""
	}
}

// MarshalSignature converts []byte signature into string.
func MarshalSignature(signature []byte) string {
	return base64.StdEncoding.EncodeToString(signature)
}

func ProofChainFindLatest(persona string) (pc *ProofChain, err error) {
	pc = new(ProofChain)
	tx := DB.Where("persona = ?", MarshalPersona(persona)).Order("id DESC").Take(pc)
	if tx.Error != nil {
		if strings.Contains(tx.Error.Error(), "record not found") {
			return nil, nil
		}
		return nil, xerrors.Errorf("%w", tx.Error)
	}

	return pc, nil
}

func ProofChainFindBySignature(signature string) (pc *ProofChain, err error) {
	previous := &ProofChain{}
	tx := DB.Where("signature = ?", signature).Take(previous)
	if tx.Error != nil || previous.ID == int64(0) {
		return nil, xerrors.Errorf("error finding previous proof chain: %w", tx.Error)
	}

	return previous, nil
}

func ProofChainCreateFromValidator(validator *validator.Base) (pc *ProofChain, err error) {
	pc = &ProofChain{
		Action:           validator.Action,
		Persona:          MarshalPersona(validator.Pubkey),
		Identity:         validator.Identity, // TODO: exception may occur
		Platform:         validator.Platform,
		Location:         validator.ProofLocation,
		Signature:        MarshalSignature(validator.Signature),
		SignaturePayload: validator.SignaturePayload,
		CreatedAt:        validator.CreatedAt,
		Uuid:             validator.Uuid.String(),
		Previous:         nil,
		ArweaveID:        "",
	}

	if validator.Previous != "" {
		previous, err := ProofChainFindBySignature(validator.Previous)
		if err != nil {
			return nil, xerrors.Errorf("%w", err)
		}

		pc.Previous = previous
	}

	if len(validator.Extra) != 0 {
		extra_json, err := json.Marshal(validator.Extra)
		if err != nil {
			return nil, xerrors.Errorf("%w", err)
		}
		pc.Extra = datatypes.JSON(extra_json)
	}

	tx := DB.Create(pc)
	if tx.Error != nil {
		return nil, xerrors.Errorf("%w", err)
	}

	return pc, nil
}

func ProofChainFindByPersona(persona string, all_data bool, from int, limit int) (total int64, rs []ProofChainItem, err error) {
	rs = make([]ProofChainItem, 0, 0)
	proofs := make([]ProofChain, 0, 0)

	tx := DB.Model(&ProofChain{})
	tx = tx.Where("persona = ?", persona)

	countTx := tx // Value-copy another query for total amount calculation
	countTx.Count(&total)

	if !all_data {
		tx = tx.Offset(from).Limit(limit).Find(&proofs)
	} else {
		tx = tx.Find(&proofs)
	}

	if tx.Error != nil || tx.RowsAffected == int64(0) || len(proofs) == 0 {
		return total, rs, tx.Error
	}

	rs = lo.Map(proofs, func(item ProofChain, index int) ProofChainItem {
		return item.ToProofChainItem()
	})

	return total, rs, nil
}
