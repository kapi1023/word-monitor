package webhook

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/kapi1023/word-monitor/internal/config"
)

type discordPayload struct {
	Content string `json:"content"`
}

func Send(cfg *config.Config, message string) error {
	if cfg.Webhook.DistordUrl == "" {
		return errors.New("brakuje adresu Discord webhook")
	}

	payload := discordPayload{
		Content: message,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", cfg.Webhook.DistordUrl, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		slog.Error("Błąd webhooka Discord", "status", resp.Status)
		return errors.New("błąd wysyłki webhooka Discord")
	}

	slog.Info("Wysłano powiadomienie do Discorda")
	return nil
}
