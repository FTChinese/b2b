package model

import (
	"github.com/guregu/null"
	"strings"
)

// Assignee represents a reader who can accept
// an invitation email, and who can be granted
// a licence.
type Assignee struct {
	FtcID    null.String `json:"ftcId" db:"ftc_id"`
	Email    null.String `json:"email" db:"email"`
	UserName null.String `json:"userName" db:"user_name"`
	IsVIP    bool        `json:"isVip" db:"is_vip"`
}

// NormalizeName tries to find a proper way to greet user
// in email.
func (a Assignee) NormalizeName() string {
	if a.UserName.Valid {
		return a.UserName.String
	}

	return strings.Split(a.Email.String, "@")[0]
}

func (a Assignee) TeamMember(teamID string) Staffer {
	return Staffer{
		Email:  a.Email.String,
		TeamID: teamID,
		FtcID:  a.FtcID,
	}
}

// Staffer is a member belong to a team under admin's
// management.
type Staffer struct {
	ID     int64       `json:"id" db:"id"`
	Email  string      `json:"email" db:"email"`
	TeamID string      `json:"teamId" db:"team_id"`
	FtcID  null.String `json:"ftcId" db:"ftc_id"`
}

// StaffList contains a list of assignee rows and the total number of rows for current team.
type StaffList struct {
	Total int64     `json:"total"`
	Data  []Staffer `json:"data"`
	Err   error     `json:"-"` // Contains possible error when used to pass data from a goroutine.
}
