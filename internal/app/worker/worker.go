package worker

import (
	"github.com/widiskel/poseidon-voice-bot/internal/client/apiclient"
	"github.com/widiskel/poseidon-voice-bot/internal/model"
	"github.com/widiskel/poseidon-voice-bot/internal/utils/exception"
)

func Run(session *model.Session) {
	op := NewOperation(session)

	for {
		if err := op.LoginIfNeeded(); err != nil {
			op.log.JustLog("Failed to login: " + err.Error())
			if stop := exception.HandleError(op.log, err); stop {
				return
			}
			continue
		}

		if err := op.GetUserInformation(); err != nil {
			op.log.Log("Failed to get user information: " + err.Error())
			if apiErr, ok := err.(*apiclient.Error); ok && apiErr.IsStatus(401) {
				op.log.Log("Session Expired Resetting JWT: " + err.Error())
				op.ResetJWT()
				continue
			}
			if stop := exception.HandleError(op.log, err); stop {
				return
			}
			continue
		}

		if err := op.GetCampaign(); err != nil {
			op.log.JustLog("Failed to get campaigns: " + err.Error())
			if stop := exception.HandleError(op.log, err); stop {
				return
			}
			continue
		}

		for _, c := range op.CampaignList.Items {
			allowed, err := op.CheckCampaignAccess(c)
			if err != nil {
				op.log.JustLog("Failed to check campaign access: " + err.Error())
				if stop := exception.HandleError(op.log, err); stop {
					return
				}
				continue
			}
			if !allowed {
				op.log.JustLog("No access to campaign: " + c.CampaignName)
				continue
			}

			op.log.Log("Processing campaign: "+c.CampaignName, 800)

			if err := op.ProcessCampaign(c); err != nil {
				op.log.JustLog("Failed to get campaigns: " + err.Error())
				if stop := exception.HandleError(op.log, err); stop {
					return
				}
				continue
			}
		}

		op.log.Log("Account processing complete. Sleeping...", 1_000_000)
	}
}
