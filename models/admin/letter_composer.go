package admin

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/postoffice"
	"strings"
)

const baseUrl = "https://www.ftacademy.cn/b2b"

type InvitationLetter struct {
	AdminEmail string
	TeamName   string
	Tier       enum.Tier
	URL        string
}

func ComposeInvitationLetter(il InvitedLicence, at AccountTeam) (postoffice.Parcel, error) {
	data := struct {
		AssigneeName string
		TeamName     string
		Tier         enum.Tier
		URL          string
		AdminEmail   string
	}{
		AssigneeName: il.Assignee.NormalizeName(),
		TeamName:     at.TeamName.String,
		Tier:         il.Plan.Tier,
		URL:          baseUrl + "/accept-invitation/" + il.Invitation.Token,
		AdminEmail:   at.Email,
	}

	var body strings.Builder
	err := tmpl.ExecuteTemplate(&body, "invitation", data)

	if err != nil {
		return postoffice.Parcel{}, err
	}

	return postoffice.Parcel{
		FromAddress: "no-reply@ftchinese.com",
		FromName:    "FT中文网",
		ToAddress:   il.Assignee.Email.String,
		ToName:      data.AssigneeName,
		Subject:     "[FT中文网B2B]会员邀请",
		Body:        body.String(),
	}, nil
}

func ComposeVerificationLetter(a Account, verifier AccountInput) (postoffice.Parcel, error) {

	data := struct {
		Name     string
		URL      string
		IsSignUp bool
	}{
		Name:     a.NormalizeName(),
		URL:      baseUrl + "/verify/" + verifier.Token,
		IsSignUp: verifier.IsSignUp,
	}
	var body strings.Builder
	err := tmpl.ExecuteTemplate(&body, "verification", data)

	if err != nil {
		return postoffice.Parcel{}, err
	}

	return postoffice.Parcel{
		FromAddress: "no-reply@ftchinese.com",
		FromName:    "FT中文网",
		ToAddress:   a.Email,
		ToName:      a.NormalizeName(),
		Subject:     "[FT中文网B2B]验证账号",
		Body:        body.String(),
	}, nil
}

func ComposePwResetLetter(a Account, bearer AccountInput) (postoffice.Parcel, error) {

	data := struct {
		Name string
		URL  string
	}{
		Name: a.NormalizeName(),
		URL:  baseUrl + "/password-reset/token/" + bearer.Token,
	}
	var body strings.Builder
	err := tmpl.ExecuteTemplate(&body, "passwordReset", data)

	if err != nil {
		return postoffice.Parcel{}, err
	}

	return postoffice.Parcel{
		FromAddress: "no-reply@ftchinese.com",
		FromName:    "FT中文网",
		ToAddress:   a.Email,
		ToName:      data.Name,
		Subject:     "[FT中文网B2B]重置密码",
		Body:        body.String(),
	}, nil
}
