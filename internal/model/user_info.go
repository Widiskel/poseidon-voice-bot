package model

import "time"

type UserInfo struct {
	ID                       string        `json:"id"`
	Name                     interface{}   `json:"name"`
	AuthProvider             string        `json:"auth_provider"`
	DynamicWallet            string        `json:"dynamic_wallet"`
	WorldWallet              interface{}   `json:"world_wallet"`
	CreatedAt                string        `json:"created_at"`
	UpdatedAt                string        `json:"updated_at"`
	WorldIDVerified          bool          `json:"world_id_verified"`
	ReferralCode             string        `json:"referral_code"`
	Points                   int           `json:"points"`
	IsNewUser                bool          `json:"is_new_user"`
	CurrentRank              int           `json:"current_rank"`
	ReferrerID               string        `json:"referrer_id"`
	ReferralPointsAwarded    bool          `json:"referral_points_awarded"`
	ReferralQualifyingFileID interface{}   `json:"referral_qualifying_file_id"`
	ReferralPointsAwardedAt  interface{}   `json:"referral_points_awarded_at"`
	WorldIDBonusAwarded      bool          `json:"world_id_bonus_awarded"`
	WorldIDBonusAwardedAt    interface{}   `json:"world_id_bonus_awarded_at"`
	AvatarURL                string        `json:"avatar_url"`
	AvatarNftURL             string        `json:"avatar_nft_url"`
	Email                    string        `json:"email"`
	VoicePhrase              string        `json:"voice_phrase"`
	VoicePhraseCreatedAt     time.Time     `json:"voice_phrase_created_at"`
	PersonalWallet           interface{}   `json:"personal_wallet"`
	PersonalWalletLinkedAt   interface{}   `json:"personal_wallet_linked_at"`
	Gender                   interface{}   `json:"gender"`
	BirthYear                interface{}   `json:"birth_year"`
	Nationality              interface{}   `json:"nationality"`
	PrimaryLanguage          interface{}   `json:"primary_language"`
	KnownLanguages           []interface{} `json:"known_languages"`
}
