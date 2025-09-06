package model

type CampaignScript struct {
	Script struct {
		IsActive     bool        `json:"is_active"`
		ID           string      `json:"id"`
		LanguageID   string      `json:"language_id"`
		BaseScriptID interface{} `json:"base_script_id"`
		Language     struct {
			ID   string `json:"id"`
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"language"`
		FileKey          string      `json:"file_key"`
		RomanizedFileKey interface{} `json:"romanized_file_key"`
		CreatedAt        string      `json:"created_at"`
		UpdatedAt        string      `json:"updated_at"`
		Content          string      `json:"content"`
		RomanizedContent interface{} `json:"romanized_content"`
		FileURL          string      `json:"file_url"`
		RomanizedFileURL interface{} `json:"romanized_file_url"`
	} `json:"script"`
	BaseEnglishScript interface{} `json:"base_english_script"`
	HasRomanization   bool        `json:"has_romanization"`
	AssignmentID      string      `json:"assignment_id"`
	AssignedAt        string      `json:"assigned_at"`
}
