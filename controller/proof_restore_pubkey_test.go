package controller

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/util"
	"github.com/nextdotid/proof_server/validator"
	"github.com/nextdotid/proof_server/validator/twitter"
	"github.com/stretchr/testify/require"
)

var (
	sampleRequest = ProofRestorePubkeyRequest{
		Action:    types.Actions.Create,
		Platform:  types.Platforms.Twitter,
		Identity:  "yeiwb",
		Uuid:      "f26593f3-3b05-4979-934a-f823a7380d05",
		CreatedAt: "1694961172",
		Previous:  "0aF+vdyS8bU0eA/beKVjrPgAIqeWwD6a6wvb3xLYz/lO4IfYATztpJTggoqUco0C9pI4lNJ5Vd9DNbNmuD9DUgE=",
		Signature: "cakVrOig6RCA5U8iffo7D6BXkeIpELvzY/H7m6V+Vw0qSlSGOnbhyvE+J54Cmwv/9S/6QwU41MwSY8nRwtqM6Rs=",
	}
	sampleRequestWithProofPost = ProofRestorePubkeyRequest{
		Action:    types.Actions.Create,
		Platform:  types.Platforms.Twitter,
		Identity:  "yeiwb",
		ProofPost: "￼ Verify @yeiwb with @NextDotID .\nSig: cakVrOig6RCA5U8iffo7D6BXkeIpELvzY/H7m6V+Vw0qSlSGOnbhyvE+J54Cmwv/9S/6QwU41MwSY8nRwtqM6Rs=\nMisc: f26593f3-3b05-4979-934a-f823a7380d05|1694961172|0aF+vdyS8bU0eA/beKVjrPgAIqeWwD6a6wvb3xLYz/lO4IfYATztpJTggoqUco0C9pI4lNJ5Vd9DNbNmuD9DUgE=",
	}
	sampleResponse = ProofRestorePubkeyResponse{
		PublicKey: "0x027e55e1b78e873c6f7d585064b41cd2735000bacc0092fe947c11ab7742ed351f",
	}
)

func TestProofRestorePubkey(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		req := sampleRequest
		resp := ProofRestorePubkeyResponse{}
		APITestCall(Engine, "POST", "/v1/proof/restore_pubkey", &req, &resp)
		require.Equal(t, resp.PublicKey, sampleResponse.PublicKey)
	})

	t.Run("param missing", func(t *testing.T) {
		req := sampleRequest
		req.Signature = ""
		resp := ErrorResponse{}
		APITestCall(Engine, "POST", "/v1/proof/restore_pubkey", &req, &resp)
		require.Contains(t, resp.Message, "param missing")
	})

	t.Run("param wrong: created_at format", func(t *testing.T) {
		req := sampleRequest
		req.CreatedAt = "abc123"
		resp := ErrorResponse{}
		APITestCall(Engine, "POST", "/v1/proof/restore_pubkey", &req, &resp)
		require.Contains(t, resp.Message, "created_at")
	})

	t.Run("param wrong: signature length", func(t *testing.T) {
		req := sampleRequest
		wrongSig := make([]byte, 64)
		rand.Read(wrongSig)
		req.Signature = base64.StdEncoding.EncodeToString(wrongSig)
		resp := ErrorResponse{}
		APITestCall(Engine, "POST", "/v1/proof/restore_pubkey", &req, &resp)
		require.Contains(t, resp.Message, "signature length error")
		require.Contains(t, resp.Message, "64")
	})

	t.Run("param wrong: previous signature length", func(t *testing.T) {
		req := sampleRequest
		wrongSig := make([]byte, 64)
		rand.Read(wrongSig)
		req.Previous = base64.StdEncoding.EncodeToString(wrongSig)
		resp := ErrorResponse{}
		APITestCall(Engine, "POST", "/v1/proof/restore_pubkey", &req, &resp)
		require.Contains(t, resp.Message, "previous signature length error")
		require.Contains(t, resp.Message, "64")
	})
}

func Test_ScanFunctions(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		base := validator.Base{
			Action:   types.Actions.Create,
			Platform: types.Platforms.Twitter,
			Identity: "yeiwb",
		}
		scanSig := func(text string) {
			re := regexp.MustCompile(twitter.MATCH_TEMPLATE)
			matched := re.FindStringSubmatch(text)
			if len(matched) < 2 {
				return
			}
			sigBytes, err := util.DecodeString(matched[1])
			if err != nil {
				return
			}
			base.Signature = sigBytes
		}

		scanMisc := func(text string) {
			re := regexp.MustCompile(twitter.MATCH_POST_CONTENT)
			matched := re.FindStringSubmatch(text)
			if len(matched) < 4 {
				return
			}
			parsedUUID, err := uuid.Parse(matched[1])
			if err != nil {
				return
			}
			parsedCreatedAt, err := util.TimestampStringToTime(matched[2])
			if err != nil {
				return
			}
			if len(matched[3]) > 0 {
				base.Previous = matched[3]
			}
			base.CreatedAt = parsedCreatedAt
			base.Uuid = parsedUUID
		}
		scanner := bufio.NewScanner(strings.NewReader(sampleRequestWithProofPost.ProofPost))
		for scanner.Scan() {
			text := scanner.Text()
			scanSig(text)
			scanMisc(text)
		}
		require.NotEmpty(t, base.Signature)
		require.Equal(t, uuid.MustParse("f26593f3-3b05-4979-934a-f823a7380d05"), base.Uuid)
		createdAt, _ := util.TimestampStringToTime("1694961172")
		require.Equal(t, createdAt, base.CreatedAt)
		require.Equal(t, "0aF+vdyS8bU0eA/beKVjrPgAIqeWwD6a6wvb3xLYz/lO4IfYATztpJTggoqUco0C9pI4lNJ5Vd9DNbNmuD9DUgE=", base.Previous)
	})
}

func Test_RestorePubkeyWithProofPost(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		req := sampleRequestWithProofPost
		resp := ProofRestorePubkeyResponse{}

		APITestCall(Engine, "POST", "/v1/proof/restore_pubkey", &req, &resp)
		fmt.Printf("Resp: %+v", resp)
		require.Equal(t, sampleResponse.PublicKey, resp.PublicKey)
	})
}
