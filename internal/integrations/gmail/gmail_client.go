package gmail

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	neturl "net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

func NewService(ctx context.Context, credentialsPath, tokenPath, accountEmail string) (*gmail.Service, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	b, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("read credentials: %w", err)
	}
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}
	cl := getClient(ctx, config, tokenPath, accountEmail)
	return gmail.NewService(ctx, option.WithHTTPClient(cl))
}

func getClient(ctx context.Context, config *oauth2.Config, tokenPath, accountEmail string) *http.Client {
	tok, err := tokenFromFile(tokenPath)
	if err != nil {
		tok = getTokenFromWeb(config, accountEmail)
		saveToken(tokenPath, tok)
	}
	return config.Client(ctx, tok)
}

func getTokenFromWeb(config *oauth2.Config, accountEmail string) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	fmt.Println("Open this link in your browser and paste the full redirect link here:")
	fmt.Println(authURL)
	fmt.Println()
	fmt.Println(`Example: http://localhost/?state=state-token&code=YOURTOKENHERE&scope=https://www.googleapis.com/auth/gmail.readonly`)

	reader := bufio.NewReader(os.Stdin)

	for {
		if accountEmail != "" {
			fmt.Println()
			fmt.Printf("Enter redirect link for account (%s) and press Enter:\n> ", accountEmail)
		} else {
			fmt.Println()
			fmt.Print("Enter redirect link:\n> ")
		}

		rawInput, _ := reader.ReadString('\n')
		rawInput = strings.TrimSpace(rawInput)

		code := rawInput
		if strings.HasPrefix(rawInput, "http://") || strings.HasPrefix(rawInput, "https://") {
			if u, err := neturl.Parse(rawInput); err == nil {
				if c := u.Query().Get("code"); c != "" {
					code = c
				} else {
					fmt.Println("Redirect link does not contain a 'code' parameter. Paste the full redirect link or enter the code directly.")
					continue
				}
			} else {
				fmt.Println("Invalid URL. Paste the full redirect link from your browser or enter the code directly.")
				continue
			}
		}

		if code == "" {
			fmt.Println("Auth code is empty. Please try again.")
			continue
		}

		tok, err := config.Exchange(context.Background(), code)
		if err != nil {
			fmt.Printf("Token exchange failed: %v\nPlease paste the redirect link/code again.\n", err)
			continue
		}
		return tok
	}
}

func tokenFromFile(path string) (*oauth2.Token, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var tok oauth2.Token
	err = json.NewDecoder(f).Decode(&tok)
	return &tok, err
}

func saveToken(path string, token *oauth2.Token) {
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_ = json.NewEncoder(f).Encode(token)
}

func FetchDynamicAuthCode(ctx context.Context, credentialsPath, tokenPath string, waitUntil time.Duration) (string, error) {
	srv, err := NewService(ctx, credentialsPath, tokenPath /*accountEmail=*/, "")
	if err != nil {
		return "", err
	}

	q := `from:authentication@notification.dynamicauth.com newer_than:1d subject:"Poseidon's login code"`
	rx := regexp.MustCompile(`\b(\d{6})\b`)

	deadline := time.Now().Add(waitUntil)
	for {
		list, err := srv.Users.Messages.List("me").
			Q(q).
			MaxResults(10).
			Do()
		if err != nil {
			return "", err
		}
		if len(list.Messages) > 0 {
			var newest *gmail.Message
			for _, m := range list.Messages {
				msg, err := srv.Users.Messages.Get("me", m.Id).Format("metadata").MetadataHeaders("Subject").Do()
				if err != nil {
					continue
				}
				if newest == nil || msg.InternalDate > newest.InternalDate {
					newest = msg
				}
			}
			if newest != nil {
				subject := ""
				for _, h := range newest.Payload.Headers {
					if h.Name == "Subject" {
						subject = h.Value
						break
					}
				}
				if subject != "" {
					if m := rx.FindStringSubmatch(subject); len(m) == 2 {
						return m[1], nil
					}
				}
			}
		}

		if time.Now().After(deadline) {
			return "", fmt.Errorf("login code email not found within %s", waitUntil)
		}
		time.Sleep(5 * time.Second)
	}
}
