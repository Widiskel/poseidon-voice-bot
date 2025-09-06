package model

type Campaign struct {
	CampaignName       string   `json:"campaign_name"`
	IPID               string   `json:"ip_id"`
	EndDate            string   `json:"end_date"`
	ThumbnailImageURL  string   `json:"thumbnail_image_url"`
	IsFeatured         bool     `json:"is_featured"`
	Description        string   `json:"description"`
	CampaignType       string   `json:"campaign_type"`
	Tags               []string `json:"tags"`
	IsScripted         bool     `json:"is_scripted"`
	SupportedLanguages []string `json:"supported_languages"`

	VirtualID          string  `json:"virtual_id"`
	UserID             string  `json:"user_id"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
	ParticipantCount   int     `json:"participant_count"`
	RegistrationStatus string  `json:"registration_status"`
	RegistrationError  *string `json:"registration_error"`
	CollectionAddress  string  `json:"collection_address"`
}
