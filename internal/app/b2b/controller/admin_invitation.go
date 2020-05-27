package controller

import (
	"github.com/FTChinese/ftacademy/internal/app/b2b/model"
	"github.com/FTChinese/ftacademy/internal/app/b2b/repository/subs"
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/render"
	"github.com/labstack/echo/v4"
	"net/http"
)

type InvitationRouter struct {
	repo subs.Env
	post postoffice.PostOffice
}

func NewInvitationRouter(env subs.Env, office postoffice.PostOffice) InvitationRouter {
	return InvitationRouter{
		repo: env,
		post: office,
	}
}

// List shows all invitations
// Query: ?page=1&per_page=10
func (router InvitationRouter) List(c echo.Context) error {
	claims := getPassportClaims(c)

	var page gorest.Pagination
	if err := c.Bind(&page); err != nil {
		return render.NewBadRequest(err.Error())
	}

	countCh, listCh := router.repo.AsyncCountInvitation(claims.TeamID.String), router.repo.AsyncListInvitations(claims.TeamID.String, page)

	countResult, listResult := <-countCh, <-listCh
	if listResult.Err != nil {
		return render.NewDBError(listResult.Err)
	}

	listResult.Total = countResult.Total
	return c.JSON(http.StatusOK, listResult)
}

// Send creates an invitation for a licence and send it to a user.
// Input: {email: string, description: string, licenceId: string}
func (router InvitationRouter) Send(c echo.Context) error {
	claims := getPassportClaims(c)

	var input model.InvitationInput
	if err := c.Bind(&input); err != nil {
		return render.NewBadRequest(err.Error())
	}

	if ve := input.Validate(); ve != nil {
		return render.NewUnprocessable(ve)
	}

	input.TeamID = claims.TeamID.String

	invitedLicence, err := router.repo.CreateInvitation(input)
	if err != nil {
		switch err {
		case subs.ErrLicenceUnavailable:
			return &render.ValidationError{
				Message: "The licence is already taken",
				Field:   "licence",
				Code:    "already_taken",
			}

		case subs.ErrInviteeMismatch:
			return &render.ValidationError{
				Message: err.Error(),
				Field:   "invitee",
				Code:    render.CodeAlreadyExists,
			}

		case subs.ErrAlreadyMember:
			return &render.ValidationError{
				Message: "The email to accept the invitation is already a valid member",
				Field:   "membership",
				Code:    render.CodeAlreadyExists,
			}

		default:
			return render.NewDBError(err)
		}
	}

	// Send invitation letter
	go func() {

		accountTeam, err := router.repo.AdminTeam(claims.Id)
		if err != nil {
			return
		}

		parcel, err := model.ComposeInvitationLetter(invitedLicence, accountTeam)
		if err != nil {
			return
		}

		err = router.post.Deliver(parcel)
		if err != nil {
			logger.WithField("trace", "DeliverInvitationLetter").Error(err)
		}
	}()

	// Add the invitee to team member
	go func() {
		_ = router.repo.SaveStaffer(invitedLicence.Assignee.TeamMember(claims.TeamID.String))
	}()

	return c.NoContent(http.StatusNoContent)
}

// Revoke cancels an invitation before it is accepted by user.
// If the invitation is already accepted, revoke has no effect.
// Admin should revoke a licence for this purpose.
func (router InvitationRouter) Revoke(c echo.Context) error {
	invID := c.Param("id") // the invitation id
	claims := getPassportClaims(c)

	err := router.repo.RevokeInvitation(invID, claims.TeamID.String)
	if err != nil {
		return render.NewDBError(err)
	}

	return c.NoContent(http.StatusNoContent)
}
