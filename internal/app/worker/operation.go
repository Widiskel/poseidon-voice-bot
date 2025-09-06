package worker

import (
	"fmt"
	"os"
	"time"

	"github.com/widiskel/poseidon-voice-bot/internal/client/apiclient"
	"github.com/widiskel/poseidon-voice-bot/internal/client/tts"
	"github.com/widiskel/poseidon-voice-bot/internal/integrations/dynamic"
	"github.com/widiskel/poseidon-voice-bot/internal/model"
	"github.com/widiskel/poseidon-voice-bot/internal/utils"
	"github.com/widiskel/poseidon-voice-bot/internal/utils/fingerprint"
	"github.com/widiskel/poseidon-voice-bot/internal/utils/logger"
)

type Operation struct {
	session      *model.Session
	api          *apiclient.ApiClient
	log          *logger.ClassLogger
	signedIn     bool
	UserInfo     model.UserInfo
	CampaignList model.Paginate[model.Campaign]
}

func NewOperation(session *model.Session) *Operation {
	return &Operation{
		session: session,
		api:     apiclient.New(session),
		log:     logger.NewNamed(fmt.Sprintf("Operation - Account %d", session.AccIdx+1), session),
	}
}

func (op *Operation) ResetJWT() {
	if op.session != nil && op.session.Email != "" {
		_ = utils.DeleteToken(op.session.Email)
	}
	op.session.JWT = ""
	op.signedIn = false
}

func (op *Operation) buildCommonHeaders() map[string]string {
	return map[string]string{
		"Authorization":            "Bearer " + op.session.JWT,
		"Origin":                   "https://app.psdn.ai",
		"Referer":                  "https://app.psdn.ai/",
		"X-Fingerprint-Request-Id": fingerprint.MakeRequestID(),
	}
}

func (op *Operation) LoginIfNeeded() error {
	if op.session.JWT != "" {
		return nil
	}

	if op.session.Email != "" {
		if st, err := utils.LoadToken(op.session.Email); err == nil && st.JWT != "" {
			if utils.IsExpired(st, 60) {
				op.log.Log("Stored JWT expired. Removing and re-signing…", 1200)
				_ = utils.DeleteToken(op.session.Email)
			} else {
				op.session.JWT = st.JWT
				op.log.Log("Loaded JWT from disk.")
				return nil
			}
		}
	}

	op.log.Log("Signing in via Dynamic Auth…", 1500)
	if err := dynamic.SignIn(op.session, op.api); err != nil {
		return err
	}
	op.signedIn = true
	op.log.Log("Sign-in success.", 1200)
	return nil
}

func (op *Operation) GetUserInformation() error {
	headers := op.buildCommonHeaders()
	op.log.Log("Getting user Information...", 1500)

	resp, err := op.api.Call("https://poseidon-depin-server.storyapis.com/users/me", "GET", nil, headers)
	if err != nil {
		return err
	}

	var userInfo model.UserInfo
	if err := resp.Decode(&userInfo); err != nil {
		op.log.JustLog("Decode /users/me failed: " + err.Error())
		return err
	}

	op.session.Point = userInfo.Points
	op.session.ID = userInfo.ID
	if op.session.Email == "" {
		op.session.Email = userInfo.Email
	}
	op.UserInfo = userInfo
	op.log.Log("Successfully Get User Information")
	return nil
}

func (op *Operation) GetCampaign() error {
	headers := op.buildCommonHeaders()
	op.log.Log("Getting Available Campaign...", 1500)

	resp, err := op.api.Call("https://poseidon-depin-server.storyapis.com/campaigns?page=1&size=100", "GET", nil, headers)
	if err != nil {
		return err
	}

	var list model.Paginate[model.Campaign]
	if err := resp.Decode(&list); err != nil {
		op.log.JustLog("Decode /campaigns failed: " + err.Error())
		return err
	}

	op.CampaignList = list
	op.log.Log("Successfully Get Available Campaign")
	return nil
}

func (op *Operation) CheckCampaignAccess(c model.Campaign) (bool, error) {
	headers := op.buildCommonHeaders()
	op.log.Log(fmt.Sprintf("Checking Access For Campaign %s...", c.CampaignName), 1500)

	resp, err := op.api.Call(
		fmt.Sprintf("https://poseidon-depin-server.storyapis.com/campaigns/%s/access", c.VirtualID),
		"GET", nil, headers,
	)
	if err != nil {
		return false, err
	}

	var access model.Access
	if err := resp.Decode(&access); err != nil {
		op.log.JustLog("Decode /access failed: " + err.Error())
		return false, err
	}
	return access.Allowed, nil
}

func (op *Operation) ProcessCampaign(c model.Campaign) error {
	headers := op.buildCommonHeaders()
	op.log.Log(fmt.Sprintf("Prepairing to Process Campaign %s...", c.CampaignName), 1500)

	resp, err := op.api.Call(
		fmt.Sprintf("https://poseidon-depin-server.storyapis.com/scripts/next?language_code=%s&campaign_id=%s",
			c.SupportedLanguages[0], c.VirtualID),
		"GET", nil, headers,
	)
	if err != nil {
		return err
	}

	var script model.CampaignScript
	if err := resp.Decode(&script); err != nil {
		op.log.JustLog("Decode /scripts failed: " + err.Error())
		return err
	}

	webmPath, err := tts.SynthesizeToWebM(op.session, script.Script.Content, tts.Options{
		Language: script.Script.Language.Code,
		Bitrate:  "48k",
	})
	if err != nil {
		return fmt.Errorf("tts synth: %w", err)
	}
	defer os.Remove(webmPath)

	fileName := fmt.Sprintf("audio_recording_%d.webm", time.Now().UnixMilli())

	initBody := model.FileUploadRequest{
		ContentType:        "audio/webm",
		FileName:           fileName,
		ScriptAssignmentID: script.AssignmentID,
	}
	respInit, err := op.api.Call(
		fmt.Sprintf("https://poseidon-depin-server.storyapis.com/files/uploads/%s", c.VirtualID),
		"POST", initBody, headers,
	)
	if err != nil {
		return err
	}

	var up model.FileUploadResponse
	if err := respInit.Decode(&up); err != nil {
		op.log.JustLog("Decode init upload failed: " + err.Error())
		return err
	}

	if err := tts.PutPresignedWebM(up.PresignedURL, webmPath); err != nil {
		return err
	}

	dg, err := tts.ComputeSHA256AndSize(webmPath)
	if err != nil {
		return err
	}

	validateBody := model.FileUploadValidationRequest{
		ContentType: "audio/webm",
		ObjectKey:   up.ObjectKey,
		Sha256Hash:  dg.HashHex,
		Filesize:    dg.FileSize,
		FileName:    fileName,
		VirtualID:   up.FileID,
		CampaignID:  c.VirtualID,
	}

	respVal, err := op.api.Call(
		"https://poseidon-depin-server.storyapis.com/files",
		"POST", validateBody, headers,
	)
	if err != nil {
		return err
	}

	var val model.FileUploadValidationResponse
	if err := respVal.Decode(&val); err != nil {
		op.log.JustLog("Decode validation failed: " + err.Error())
		return err
	}

	if val.FileStatus != "UPLOADED" {
		return fmt.Errorf("file status unexpected: %s", val.FileStatus)
	}

	op.log.Log(fmt.Sprintf("Upload validated. Awarded=%d verified=%v", val.PointsAwarded, val.IsVerifiedQuality), 1200)
	return nil
}
