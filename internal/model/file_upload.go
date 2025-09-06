package model

type FileUploadRequest struct {
	ContentType        string `json:"content_type"`
	FileName           string `json:"file_name"`
	ScriptAssignmentID string `json:"script_assignment_id"`
}
type FileUploadResponse struct {
	PresignedURL string `json:"presigned_url"`
	ObjectKey    string `json:"object_key"`
	FileID       string `json:"file_id"`
}

type FileUploadValidationRequest struct {
	ContentType string `json:"content_type"`
	ObjectKey   string `json:"object_key"`
	Sha256Hash  string `json:"sha256_hash"`
	Filesize    int    `json:"filesize"`
	FileName    string `json:"file_name"`
	VirtualID   string `json:"virtual_id"`
	CampaignID  string `json:"campaign_id"`
}

type FileUploadValidationResponse struct {
	FileName     string `json:"file_name"`
	FilePath     string `json:"file_path"`
	FileType     string `json:"file_type"`
	FileSize     int    `json:"file_size"`
	FileHash     string `json:"file_hash"`
	FileURL      string `json:"file_url"`
	FileStatus   string `json:"file_status"`
	FileMetadata struct {
	} `json:"file_metadata"`
	VirtualID          string      `json:"virtual_id"`
	CampaignID         string      `json:"campaign_id"`
	ID                 string      `json:"id"`
	UserID             string      `json:"user_id"`
	IPID               interface{} `json:"ip_id"`
	CreatedAt          string      `json:"created_at"`
	UpdatedAt          string      `json:"updated_at"`
	RegistrationStatus interface{} `json:"registration_status"`
	RegistrationError  interface{} `json:"registration_error"`
	Score              interface{} `json:"score"`
	IsVerifiedQuality  bool        `json:"is_verified_quality"`
	PointsAwarded      int         `json:"points_awarded"`
	IsRewarded         bool        `json:"is_rewarded"`
	IsFlaggedDuplicate bool        `json:"is_flagged_duplicate"`
	IsFlaggedBot       bool        `json:"is_flagged_bot"`
	IsFlaggedSpam      bool        `json:"is_flagged_spam"`
}
