package exception

import (
	"encoding/json"
	"fmt"

	"github.com/widiskel/poseidon-voice-bot/internal/client/apiclient"
	"github.com/widiskel/poseidon-voice-bot/internal/utils/logger"
)

func HandleError(log *logger.ClassLogger, err error) (shouldStop bool) {
	if apiErr, ok := err.(*apiclient.Error); ok {
		return handleAPIError(log, apiErr)
	}
	return handleTransportError(log, err)
}

func handleAPIError(log *logger.ClassLogger, apiErr *apiclient.Error) (shouldStop bool) {
	msg := extractErrMessage(apiErr.Body)

	switch {
	case apiErr.IsStatus(401):
		log.JustLog(fmt.Sprintf("401 Unauthorized: %s", msg))
		log.Log("JWT invalid/expired.", 3000)
		return false

	case apiErr.IsStatus(502):
		log.JustLog(fmt.Sprintf("502 Cloudflare Block: %s", msg))
		log.Log("Blocked by cloudflare, please open page on browser for unblock.", 3000)
		return true

	case apiErr.IsStatus(403):
		log.JustLog(fmt.Sprintf("403 Forbidden: %s", msg))
		log.Log("Forbidden. Retrying after 30 seconds…", 30000)
		return false

	case apiErr.IsStatus(404):
		log.JustLog(fmt.Sprintf("404 Not Found: %s", msg))
		log.Log("Resource not found. Retrying after 30 seconds…", 30000)
		return false

	case apiErr.IsStatus(429):
		log.JustLog(fmt.Sprintf("429 Too Many Requests: %s", msg))
		log.Log("Rate limited. Backing off 60 seconds…", 60000)
		return false

	case apiErr.IsServerError():
		log.JustLog(fmt.Sprintf("%d Server Error: %s", apiErr.StatusCode, msg))
		log.Log("Server error. Retrying after 15 seconds…", 15000)
		return false

	default:
		log.JustLog(fmt.Sprintf("%d Client Error: %s", apiErr.StatusCode, msg))
		log.Log("Client error. Retrying after 30 seconds…", 30000)
		return false
	}
}

func handleTransportError(log *logger.ClassLogger, err error) (shouldStop bool) {
	log.JustLog(fmt.Sprintf("HTTP transport error: %v", err))
	log.Log("Network error. Retrying after 10 seconds…", 10000)
	return false
}

func extractErrMessage(body string) string {
	var m map[string]any
	if err := json.Unmarshal([]byte(body), &m); err == nil {
		if v, ok := m["message"].(string); ok && v != "" {
			return v
		}
		if v, ok := m["detail"].(string); ok && v != "" {
			return v
		}
	}
	return body
}
