package model

import "time"

type Account struct {
	JWT         string `json:"jwt"`
	MinifiedJWT string `json:"minifiedJwt"`
	ExpiresAt   int64  `json:"expiresAt"`
	User        User   `json:"user"`
}

type User struct {
	ID                       string               `json:"id"`
	ProjectEnvironmentID     string               `json:"projectEnvironmentId"`
	VerifiedCredentials      []VerifiedCredential `json:"verifiedCredentials"`
	LastVerifiedCredentialID string               `json:"lastVerifiedCredentialId"`
	SessionID                string               `json:"sessionId"`
	Email                    string               `json:"email"`
	FirstVisit               *time.Time           `json:"firstVisit"`
	LastVisit                *time.Time           `json:"lastVisit"`
	NewUser                  bool                 `json:"newUser"`
	Metadata                 map[string]any       `json:"metadata"`
	MFABackupCodeAck         any                  `json:"mfaBackupCodeAcknowledgement"`
	Lists                    []any                `json:"lists"`
	MissingFields            []any                `json:"missingFields"`
}

type VerifiedCredential struct {
	ID               string `json:"id"`
	Format           string `json:"format"`
	PublicIdentifier string `json:"public_identifier"`
	SignInEnabled    *bool  `json:"signInEnabled,omitempty"`

	Address        string            `json:"address,omitempty"`
	Chain          string            `json:"chain,omitempty"`
	WalletName     string            `json:"wallet_name,omitempty"`
	WalletProvider string            `json:"wallet_provider,omitempty"`
	WalletProps    *WalletProperties `json:"wallet_properties,omitempty"`
	LastSelectedAt *time.Time        `json:"lastSelectedAt,omitempty"`
	NameService    map[string]any    `json:"name_service,omitempty"`

	Email string `json:"email,omitempty"`
}

type WalletProperties struct {
	KeyShares                []KeyShare `json:"keyShares"`
	ThresholdSignatureScheme string     `json:"thresholdSignatureScheme"`
	DerivationPath           string     `json:"derivationPath"`
}

type KeyShare struct {
	ID                 string `json:"id"`
	BackupLocation     string `json:"backupLocation"`
	PasswordEncrypted  bool   `json:"passwordEncrypted"`
	ExternalKeyShareID string `json:"externalKeyShareId"`
}
