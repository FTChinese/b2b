package model

import (
	"github.com/FTChinese/b2b/internal/pkg/plan"
	"github.com/FTChinese/b2b/internal/pkg/validator"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/rand"
	"github.com/FTChinese/go-rest/render"
	"github.com/guregu/null"
	"strings"
	"time"
)

// Invitation is an email sent to team member to accept a licence.
// An invitation could in 3 phases:
// Initially created: it indicates an email is sent to reader;
// Accepted: reader clicked the link in the invitation email,
// it should not be used any longer;
// Revoked: admin could revoke an invitation before it is accepted.
// An accepted invitation could not be revoked since that is meaningless.
type Invitation struct {
	ID             string           `json:"id" db:"invitation_id"`
	LicenceID      string           `json:"licenceId" db:"licence_id"`
	TeamID         string           `json:"teamId" db:"team_id"`
	Token          string           `json:"-" db:"token"` // This field is used only when inserting data. Retrieval does not include this field. However, it is included when saving to the JSON column in licence.
	Email          string           `json:"email" db:"email"`
	Description    null.String      `json:"description" db:"description"`
	ExpirationDays int64            `json:"expiresInDays" db:"expiration_days"`
	Status         InvitationStatus `json:"status" db:"current_status"`
	CreatedUTC     chrono.Time      `json:"createdUtc" db:"created_utc"`
	UpdatedUTC     chrono.Time      `json:"updatedUtc" db:"updated_utc"`
}

// Expires tests whether the invitation is expired.
func (i Invitation) Expired() bool {
	now := time.Now().Unix()

	created := i.CreatedUTC.Time.Unix()

	// Default 7 days * 24 * 60 * 60
	return (created + i.ExpirationDays*86400) < now
}

// IsValid determines whether an invitation is valid.
// A valid invitation must be not expires, not revoked by admin, not accepted by any one.
// A valid invitation can be accepted or revoked.
func (i Invitation) IsValid() bool {
	return i.Status == InvitationStatusCreated && !i.Expired()
}

func (i Invitation) CanBeRevoked() bool {
	return i.Status == InvitationStatusCreated
}

// Revoke invalidates an invitation by admin.
func (i Invitation) Revoke() Invitation {
	i.Status = InvitationStatusRevoked
	i.UpdatedUTC = chrono.TimeNow()

	return i
}

func (i Invitation) CanBeAccepted() bool {
	return i.Status == InvitationStatusCreated
}

// Accept invalidates an invitation after reader accepted the licence associated with it.
func (i Invitation) Accept() Invitation {
	i.Status = InvitationStatusAccepted
	i.UpdatedUTC = chrono.TimeNow()

	return i
}

// InvitationInput contains the essential data client
// submitted to create a new invitation.
type InvitationInput struct {
	Email       string      `json:"email"` // To whom the invitation should be sent.
	Description null.String `json:"description"`
	LicenceID   string      `json:"licenceId"` // Which licence is being granted.
	TeamID      string      `json:"-"`
}

// NewInvitation creates a new Invitation instance based
// on user input.
func (i InvitationInput) NewInvitation() (Invitation, error) {
	token, err := GenerateToken()
	if err != nil {
		return Invitation{}, err
	}

	return Invitation{
		ID:             "inv_" + rand.String(12),
		LicenceID:      i.LicenceID,
		TeamID:         i.TeamID,
		Token:          token,
		Email:          i.Email,
		Description:    i.Description,
		ExpirationDays: 7,
		Status:         InvitationStatusCreated,
		CreatedUTC:     chrono.TimeNow(),
		UpdatedUTC:     chrono.TimeNow(),
	}, nil
}

func (i *InvitationInput) Validate() *render.ValidationError {
	i.Email = strings.TrimSpace(i.Email)
	desc := strings.TrimSpace(i.Description.String)
	i.Description = null.NewString(desc, desc != "")
	i.LicenceID = strings.TrimSpace(i.LicenceID)

	ve := validator.New("email").Required().Email().Validate(i.Email)
	if ve != nil {
		return ve
	}

	ve = validator.New("description").Max(128).Validate(i.Description.String)
	if ve != nil {
		return ve
	}

	return validator.New("licenceId").Required().Validate(i.LicenceID)
}

// InvitedLicence wraps all related information after
// an invitation is created.
type InvitedLicence struct {
	Invitation Invitation
	Licence    BaseLicence   // The licence to grant
	Plan       plan.BasePlan // The plan of this licence
	Assignee   Assignee      // Who will be granted the licence.
}

// InvitationList is used for restful output.
type InvitationList struct {
	Total int64        `json:"total"`
	Data  []Invitation `json:"data"`
	Err   error        `json:"-"`
}