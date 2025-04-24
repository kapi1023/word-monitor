package webhook

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

type discordPayload struct {
	Content string `json:"content"`
}

func Send(discordUrl string, message string) error {
	if discordUrl == "" {
		return errors.New("brakuje adresu Discord webhook")
	}

	payload := discordPayload{
		Content: message,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", discordUrl, bytes.NewReader(body))
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
