package tts

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	htgotts "github.com/hegedustibor/htgo-tts"

	"github.com/widiskel/poseidon-voice-bot/internal/model"
	"github.com/widiskel/poseidon-voice-bot/internal/utils/logger"
)

type Options struct {
	Language string
	Bitrate  string
}

func SynthesizeToWebM(session *model.Session, text string, opts Options) (string, error) {
	log := logger.NewNamed(fmt.Sprintf("TTS - Account %d", session.AccIdx+1), session)
	startAll := time.Now()

	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("empty text")
	}
	if opts.Language == "" {
		opts.Language = "en"
	}
	if opts.Bitrate == "" {
		opts.Bitrate = "48k"
	}
	if !validBitrate(opts.Bitrate) {
		return "", fmt.Errorf("invalid bitrate: %s (use like 48k, 64k, 96k)", opts.Bitrate)
	}

	tmpDir, err := os.MkdirTemp("", "tts-*")
	if err != nil {
		return "", fmt.Errorf("mktemp: %w", err)
	}

	base := fmt.Sprintf("%s_%d", mapLang(opts.Language), time.Now().UnixNano())
	mp3Path := filepath.Join(tmpDir, base+".mp3")
	webmPath := filepath.Join(tmpDir, base+".webm")

	sp := htgotts.Speech{
		Folder:   tmpDir,
		Language: mapLang(opts.Language),
	}
	log.JustLog(fmt.Sprintf("[TTS] Generating MP3 (lang=%s) -> %s", sp.Language, mp3Path))

	if _, err := sp.CreateSpeechBuff(text, base); err != nil {
		return "", fmt.Errorf("CreateSpeechBuff: %w", err)
	}
	if _, err := os.Stat(mp3Path); err != nil {
		return "", fmt.Errorf("mp3 not found: %w", err)
	}

	cmd := exec.Command("ffmpeg", "-y", "-i", mp3Path, "-c:a", "libopus", "-b:a", opts.Bitrate, webmPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("ffmpeg: %v, out=%s", err, string(out))
	}

	log.JustLog(fmt.Sprintf("[TTS] DONE mp3+webm in %s (tmp=%s)", time.Since(startAll), tmpDir))
	return webmPath, nil
}

func mapLang(code string) string {
	switch strings.ToLower(code) {
	case "en", "en-us", "en_gb":
		return "en"
	case "id", "id-id", "id_id":
		return "id"
	default:
		return code
	}
}

var reBitrate = regexp.MustCompile(`^[1-9]\d{1,3}k$`)

func validBitrate(b string) bool {
	return reBitrate.MatchString(strings.ToLower(strings.TrimSpace(b)))
}

func PutPresignedWebM(url string, webmPath string) error {
	f, err := os.Open(webmPath)
	if err != nil {
		return fmt.Errorf("open webm: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	body := &bytes.Buffer{}
	if _, err := io.Copy(io.MultiWriter(h, body), f); err != nil {
		return fmt.Errorf("read webm: %w", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewReader(body.Bytes()))
	if err != nil {
		return fmt.Errorf("new req: %w", err)
	}
	req.Header.Set("Content-Type", "audio/webm")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("put presigned: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("put presigned failed: %s, %s", resp.Status, string(b))
	}

	return nil
}

type FileDigest struct {
	HashHex  string
	FileSize int
}

func ComputeSHA256AndSize(path string) (FileDigest, error) {
	f, err := os.Open(path)
	if err != nil {
		return FileDigest{}, fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	n, err := io.Copy(h, f)
	if err != nil {
		return FileDigest{}, fmt.Errorf("read: %w", err)
	}
	return FileDigest{
		HashHex:  hex.EncodeToString(h.Sum(nil)),
		FileSize: int(n),
	}, nil
}
