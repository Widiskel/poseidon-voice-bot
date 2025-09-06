package model

type Access struct {
	Allowed            bool        `json:"allowed"`
	Reason             string      `json:"reason"`
	Cap                int         `json:"cap"`
	UsedToday          int         `json:"used_today"`
	Remaining          int         `json:"remaining"`
	TimeoutUntil       interface{} `json:"timeout_until"`
	IncreasedCapActive bool        `json:"increased_cap_active"`
}
