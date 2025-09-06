package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/widiskel/poseidon-voice-bot/internal/model"
	"github.com/widiskel/poseidon-voice-bot/internal/utils/logger"
)

/* ====================== Error Type ====================== */

type Error struct {
	Method     string
	URL        string
	StatusCode int
	Body       string
	Header     http.Header
	Duration   time.Duration
}

func (e *Error) Error() string {
	return fmt.Sprintf("http %s %s -> %d (%s): %s",
		e.Method, e.URL, e.StatusCode, e.Duration, truncate(e.Body, 256))
}

func (e *Error) IsStatus(code int) bool { return e.StatusCode == code }
func (e *Error) IsClientError() bool    { return e.StatusCode >= 400 && e.StatusCode < 500 }
func (e *Error) IsServerError() bool    { return e.StatusCode >= 500 }

/* ====================== Client ====================== */

type ApiClient struct {
	http           *http.Client
	DefaultHeaders map[string]string
	log            *logger.ClassLogger
}

func New(sess *model.Session) *ApiClient {
	timeout := 60 * time.Second
	return &ApiClient{
		http: &http.Client{Timeout: timeout},
		DefaultHeaders: map[string]string{
			"Accept":          "application/json",
			"Accept-Language": "en-US,en;q=0.9,id;q=0.8",
			"Content-Type":    "application/json",
			"User-Agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36",
		},
		log: logger.NewNamed("ApiClient", sess),
	}
}

func (c *ApiClient) BuildHeaders(additional map[string]string) http.Header {
	h := http.Header{}
	for k, v := range c.DefaultHeaders {
		h.Set(k, v)
	}
	for k, v := range additional {
		h.Set(k, v)
	}
	return h
}

/* ====================== Call ====================== */

func (c *ApiClient) Call(
	URL string,
	method string,
	payload any,
	additionalHeaders map[string]string,
) (*model.ApiResponse, error) {

	m := strings.ToUpper(strings.TrimSpace(method))
	if m == "" {
		m = http.MethodGet
	}

	u, err := url.Parse(URL)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}

	var body io.Reader
	var bodyPreview string

	switch m {
	case http.MethodGet, http.MethodDelete, http.MethodHead:
		if payload != nil {
			q := u.Query()
			for k, v := range encodeToQuery(payload) {
				if len(v) > 0 {
					q.Set(k, v[0])
				}
			}
			u.RawQuery = q.Encode()
		}
	default:
		if payload != nil {
			buf, err := json.Marshal(payload)
			if err != nil {
				return nil, fmt.Errorf("json encode: %w", err)
			}
			body = bytes.NewReader(buf)

			bodyPreview = prettyJSON(payload)

			if _, ok := additionalHeaders["Content-Type"]; !ok {
				if _, ok := c.DefaultHeaders["Content-Type"]; !ok {
					if c.DefaultHeaders == nil {
						c.DefaultHeaders = map[string]string{}
					}
					c.DefaultHeaders["Content-Type"] = "application/json"
				}
			}
		}
	}

	req, err := http.NewRequestWithContext(context.Background(), m, u.String(), body)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header = c.BuildHeaders(additionalHeaders)

	c.log.JustLog(fmt.Sprintf(
		"HTTP REQUEST\nMethod : %s\nURL    : %s\nHeaders: %s\nQuery  : %s\nBody   : %s\n",
		m,
		u.String(),
		prettyHeader(maskSensitiveHeaders(req.Header)),
		u.Query().Encode(),
		truncate(bodyPreview, 2000),
	))

	start := time.Now()
	resp, err := c.http.Do(req)
	dur := time.Since(start)
	if err != nil {
		c.log.JustLog(fmt.Sprintf("HTTP ERROR transport %s %s -> %v", m, u.String(), err))
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, rbErr := io.ReadAll(resp.Body)
	if rbErr != nil {
		c.log.JustLog(fmt.Sprintf("HTTP ERROR read body %s %s: %v", m, u.String(), rbErr))
		return &model.ApiResponse{StatusCode: resp.StatusCode, Data: nil}, fmt.Errorf("read body: %w", rbErr)
	}

	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	parsed := make(map[string]any)

	switch {
	case strings.Contains(contentType, "application/json"):
		if err := json.Unmarshal(respBody, &parsed); err != nil {
			c.log.JustLog(fmt.Sprintf("HTTP ERROR json decode %s %s: %v", m, u.String(), err))
			parsed = map[string]any{"message": string(respBody)}
		}
	default:
		parsed = map[string]any{"message": string(respBody)}
	}

	respPreview := tryPrettyJSON(respBody)

	c.log.JustLog(fmt.Sprintf(
		"HTTP RESPONSE\nMethod : %s\nURL    : %s\nStatus : %d\nElapsed: %s\nHeaders: %s\nBody   : %s\n",
		m,
		u.String(),
		resp.StatusCode,
		dur,
		prettyHeader(maskSensitiveHeaders(resp.Header)),
		truncate(respPreview, 4000),
	))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		he := &Error{
			Method:     m,
			URL:        u.String(),
			StatusCode: resp.StatusCode,
			Body:       string(respBody),
			Header:     resp.Header,
			Duration:   dur,
		}
		return &model.ApiResponse{StatusCode: resp.StatusCode, Data: parsed}, he
	}

	return &model.ApiResponse{StatusCode: resp.StatusCode, Data: parsed}, nil
}

/* ====================== Helpers ====================== */

func prettyJSON(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}

func tryPrettyJSON(b []byte) string {
	var tmp any
	if json.Unmarshal(b, &tmp) == nil {
		return prettyJSON(tmp)
	}

	return string(b)
}

func prettyHeader(h http.Header) string {
	if h == nil {
		return "{}"
	}
	flat := make(map[string]string, len(h))
	for k, v := range h {
		flat[k] = strings.Join(v, ", ")
	}
	return prettyJSON(flat)
}

func maskSensitiveHeaders(h http.Header) http.Header {
	if h == nil {
		return h
	}
	cp := h.Clone()
	if auth := cp.Get("Authorization"); auth != "" {

		if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			cp.Set("Authorization", "Bearer ***")
		} else {
			cp.Set("Authorization", "***")
		}
	}
	return cp
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func encodeToQuery(payload any) url.Values {
	values := url.Values{}
	if payload == nil {
		return values
	}

	if m, ok := payload.(map[string]any); ok {
		for k, v := range m {
			values.Add(k, fmt.Sprint(v))
		}
		return values
	}

	rv := reflect.ValueOf(payload)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if rv.IsValid() && rv.Kind() == reflect.Struct {
		var tmp map[string]any
		b, err := json.Marshal(payload)
		if err == nil && json.Unmarshal(b, &tmp) == nil {
			for k, v := range tmp {
				values.Add(k, fmt.Sprint(v))
			}
		}
	}

	return values
}
