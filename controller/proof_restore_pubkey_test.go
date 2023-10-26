package controller

import (
	"crypto/rand"
	"encoding/base64"
	"testing"

	"github.com/nextdotid/proof_server/types"
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
	sampleResponse = ProofRestorePubkeyResponse {
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
