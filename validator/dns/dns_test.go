package dns

import (
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/util"
	"github.com/nextdotid/proof_server/util/crypto"
	"github.com/nextdotid/proof_server/validator"
	"github.com/stretchr/testify/require"
)

func build() DNS {
	pk, _ := crypto.StringToPubkey("0x028568e07ebf497b07a30f8a9d1731980736a4fac9d7c9c9b5682cb82dd3e774d7")
	createdAt, _ := util.TimestampStringToTime("1664267795")
	return DNS{
		Base: &validator.Base{
			Platform:         types.Platforms.DNS,
			Previous:         "",
			Action:           types.Actions.Create,
			Pubkey:           pk,
			Identity:         "testcase.nextnext.id",
			CreatedAt:        createdAt,
			Uuid:             uuid.MustParse("80c98711-f4f6-43c7-b05c-8d86372f6131"),
		},
	}
}

func build_invalid() DNS {
	dns := build()
	dns.Identity = "testcase_invalid.nextnext.id"
	return dns
}

func Test_query(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		_, err := query("nonexist.example.com")
		t.Log(err.Error())
		require.Error(t, err)
	})

	t.Run("found", func(t *testing.T) {
		body, err := query("example.com")
		require.NoError(t, err)
		require.NotNil(t, body)
		require.NotNil(t, body.Answer)
	})
}

func Test_parseTxt(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		txt := "ps:true;v:1;sig:Oyist/0E0MJ5sN3TI33P4EMBGTaCk2S3IQKzYfI5zxpwE2VdHClgLXfmj0L2dPydF8KOXyjbWWuM2AHKdW2DnwE=;ca:1664263102;uuid:26de5ec2-889e-4f59-9fac-dfa8d99e7ce7;prev:null"
		result, err := parseTxt(txt)
		require.NoError(t, err)
		require.Equal(t, "Oyist/0E0MJ5sN3TI33P4EMBGTaCk2S3IQKzYfI5zxpwE2VdHClgLXfmj0L2dPydF8KOXyjbWWuM2AHKdW2DnwE=", result.Signature)
		require.Nil(t, result.Previous)
	})

	t.Run("struct field missing", func(t *testing.T) {
		txt := "ps:true;v:1;sig:Oyist/0E0MJ5sN3TI33P4EMBGTaCk2S3IQKzYfI5zxpwE2VdHClgLXfmj0L2dPydF8KOXyjbWWuM2AHKdW2DnwE=;ca:1664263102;uuid:26de5ec2-889e-4f59-9fac-dfa8d99e7ce7"
		_, err := parseTxt(txt)
		require.Error(t, err)
	})
}

func Test_GeneratePostPayload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dns := build()
		payload_map := dns.GeneratePostPayload()
		payload := payload_map["default"]
		require.Contains(t, payload, dns.Uuid.String())
		require.Contains(t, payload, strconv.FormatInt(dns.CreatedAt.Unix(), 10))
		require.Contains(t, payload, "%SIG_BASE64%")
		require.Contains(t, payload, "prev:null")
	})
}

func Test_GenerateSignPayload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dns := build()
		sp := dns.GenerateSignPayload()
		require.Contains(t, sp, types.Platforms.DNS)
		require.Contains(t, sp, types.Actions.Create)
		require.Contains(t, sp, dns.Uuid.String())
	})
}

func Test_Validate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dns := build()
		require.NoError(t, dns.Validate())
	})

	t.Run("invalid", func(t *testing.T) {
		dns := build_invalid()
		err := dns.Validate()
		require.Error(t, err)
		t.Log(err.Error())
	})
}
