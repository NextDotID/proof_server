package instagram

import (
	"bufio"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	i "github.com/dghubble/go-instagram/instagram"
	"github.com/dghubble/oauth1"
	"github.com/nextdotid/proof_server/config"
	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/util"
	mycrypto "github.com/nextdotid/proof_server/util/crypto"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"

	"github.com/nextdotid/proof_server/validator"
)

type Instagram struct {
	*validator.Base
}

const (
	MATCH_TEMPLATE = "^Sig: (.*)$"
)

var (
	client      *i.Client
	l           = logrus.WithFields(logrus.Fields{"module": "validator", "validator": "instagram"})
	re          = regexp.MustCompile(MATCH_TEMPLATE)
	POST_STRUCT = map[string]string{
		"default": "🎭 Verifying my Instagram ID @%s for @NextDotID.\nSig: %%SIG_BASE64%%\n\nPowered by Next.ID - Connect All Digital Identities.\n",
		"en_US":   "🎭 Verifying my Instagram ID @%s for @NextDotID.\nSig: %%SIG_BASE64%%\n\nPowered by Next.ID - Connect All Digital Identities.\n",
		"zh_CN":   "🎭 正在通过 @NextDotID 验证我的 Twitter 帐号 @%s 。\nSig: %%SIG_BASE64%%\n\n由 Next.ID 支持 - 连接全域数字身份。\n",
	}
)
