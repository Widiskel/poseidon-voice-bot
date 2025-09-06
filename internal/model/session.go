package model

type Session struct {
	JWT    string
	ID     string
	AccIdx int
	Email  string
	Point  int

	VerificationUUID string
	LoginCode        string
}
