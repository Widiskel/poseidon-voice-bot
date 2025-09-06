package dynamic

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	gethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/widiskel/poseidon-voice-bot/internal/client/apiclient"
	"github.com/widiskel/poseidon-voice-bot/internal/integrations/gmail"
	"github.com/widiskel/poseidon-voice-bot/internal/model"
	"github.com/widiskel/poseidon-voice-bot/internal/utils"
	"github.com/widiskel/poseidon-voice-bot/internal/utils/logger"
)

const (
	sdk           = "b26e23c8-16eb-4076-bd78-110084fe8a6e"
	origin        = "https://app.psdn.ai"
	referer       = "https://app.psdn.ai/"
	dynAPIVersion = "API/0.0.758"
	dynWalletVer  = "WalletKit/4.29.4"
)

func RequestEmailVerification(session *model.Session, api *apiclient.ApiClient, log *logger.ClassLogger) (string, error) {
	body := map[string]any{"email": session.Email}

	if log != nil {
		log.Log("Requesting verification email to Dynamic Auth…", 1500)
	}
	res, err := api.Call(
		fmt.Sprintf("https://app.dynamicauth.com/api/v0/sdk/%s/emailVerifications/create", sdk),
		"POST",
		body,
		nil,
	)
	if err != nil {
		if log != nil {
			log.JustLog("Request verification failed: " + err.Error())
		}
		return "", err
	}
	uuidAny, ok := res.Data["verificationUUID"]
	if !ok {
		return "", errors.New("no verificationUUID in response")
	}
	uuidStr, _ := uuidAny.(string)
	if uuidStr == "" {
		return "", errors.New("empty verificationUUID")
	}
	if log != nil {
		log.Log("Verification email requested. UUID: "+uuidStr, 1500)
	}
	return uuidStr, nil
}

func SignIn(session *model.Session, api *apiclient.ApiClient) error {
	log := logger.NewNamed("DynamicAuth", session)

	verificationUUID, err := RequestEmailVerification(session, api, log)
	if err != nil {
		return err
	}

	log.Log("Delay and Waiting for verification email in Gmail…", 10000)
	code, err := gmail.FetchDynamicAuthCode(
		context.Background(),
		"configs/credentials.json",
		fmt.Sprintf("accounts/%s-data.json", session.Email),
		2*time.Minute,
	)
	if err != nil {
		log.JustLog("Failed to fetch login code: " + err.Error())
		return err
	}
	log.Log("Login code obtained: "+code, 1500)

	pubkey, err := GenerateSessionPublicKey()
	if err != nil {
		log.JustLog("Failed to generate session public key: " + err.Error())
		return err
	}
	log.JustLog("Session public key generated.")

	body := map[string]any{
		"verificationUUID":  verificationUUID,
		"verificationToken": code,
		"sessionPublicKey":  pubkey,
	}
	log.Log("Sending verification to /signin…", 1500)

	additionalHeaders := map[string]string{
		"Accept":                       "*/*",
		"Origin":                       origin,
		"Referer":                      referer,
		"x-dyn-api-version":            dynAPIVersion,
		"x-dyn-version":                dynWalletVer,
		"x-dyn-is-global-wallet-popup": "false",
		"x-dyn-device-fingerprint":     randomHex(16),
		"x-dyn-request-id":             randomHex(24),
		"x-dyn-session-public-key":     pubkey,
	}

	res, err := api.Call(
		fmt.Sprintf("https://app.dynamicauth.com/api/v0/sdk/%s/emailVerifications/signin", sdk),
		"POST",
		body,
		additionalHeaders,
	)
	if err != nil {
		log.JustLog("Sign-in API failed: " + err.Error())
		return err
	}

	if session.Email != "" {
		if err := utils.SaveToken(session.Email, res.Data); err != nil {
			log.JustLog("Warning: failed to persist token: " + err.Error())
		}
	}

	var acc model.Account
	if err := decodeMapInto(res.Data, &acc); err == nil {
		if acc.JWT != "" {
			session.JWT = acc.JWT
		}
		if acc.User.ID != "" {
			session.ID = acc.User.ID
		}
		if acc.User.Email != "" && session.Email == "" {
			session.Email = acc.User.Email
		}
	} else {

		if jwt, ok := res.Data["jwt"].(string); ok && jwt != "" {
			session.JWT = jwt
		}
	}

	return nil
}

func GenerateSessionPublicKey() (string, error) {
	priv, err := gethcrypto.GenerateKey()
	if err != nil {
		return "", err
	}
	comp := gethcrypto.CompressPubkey(&priv.PublicKey)
	if len(comp) != 33 {
		return "", errors.New("unexpected compressed pubkey length")
	}
	return hex.EncodeToString(comp), nil
}

func decodeMapInto[T any](m map[string]any, out *T) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
