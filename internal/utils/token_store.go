package utils

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type SavedToken struct {
	Email     string         `json:"email"`
	JWT       string         `json:"jwt"`
	ExpiresAt int64          `json:"expiresAt,omitempty"`
	Raw       map[string]any `json:"raw,omitempty"`
}

func filePathFor(email string) string {
	safe := strings.ReplaceAll(email, string(os.PathSeparator), "_")
	return filepath.Join("accounts", safe+"-token.json")
}

func SaveToken(email string, raw map[string]any) error {
	jwt, _ := raw["jwt"].(string)
	if jwt == "" {
		return errors.New("tokenstore: empty jwt in response")
	}
	exp, _ := raw["expiresAt"].(float64)
	data := SavedToken{
		Email:     email,
		JWT:       jwt,
		ExpiresAt: int64(exp),
		Raw:       raw,
	}
	if err := os.MkdirAll("accounts", 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePathFor(email), b, 0o600)
}

func LoadToken(email string) (SavedToken, error) {
	var st SavedToken
	b, err := os.ReadFile(filePathFor(email))
	if err != nil {
		return st, err
	}
	if err := json.Unmarshal(b, &st); err != nil {
		return st, err
	}
	return st, nil
}

func DeleteToken(email string) error {
	if email == "" {
		return nil
	}
	return os.Remove(filePathFor(email))
}

func IsExpired(st SavedToken, skewSec int64) bool {
	if st.ExpiresAt == 0 {
		return false
	}
	now := time.Now().Unix()
	return now+skewSec >= st.ExpiresAt
}
