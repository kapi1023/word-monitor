package monitor

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/kapi1023/word-monitor/internal/client"
	"github.com/kapi1023/word-monitor/internal/config"
	"github.com/kapi1023/word-monitor/internal/webhook"
)

func Check(cfg *config.Config, c *client.Client) (bool, string, error) {
	now := time.Now()
	end := now.Add(time.Duration(cfg.Word.MaxDays) * 24 * time.Hour)

	schedule, err := c.GetExamSchedule(cfg.Word.Category, cfg.Word.WordId, now, end)
	if err != nil {
		return false, "", err
	}

	for _, day := range schedule.Schedule.ScheduledDays {
		examDate, err := time.Parse("2006-01-02", day.Day)
		if err != nil {
			continue
		}
		if examDate.After(end) {
			continue
		}

		for _, hour := range day.ScheduledHours {
			if len(hour.PracticeExams) == 0 {
				continue
			}

			msg := fmt.Sprintf(
				"**Wolny termin egzaminu!**\n📅 Data: `%s`\n⏰ Godzina: `%s`\n📍 Kategoria: `%s`\n🆔 WORD ID: `%s`",
				day.Day,
				hour.Time,
				cfg.Word.Category,
				cfg.Word.WordId,
			)

			slog.Warn("Znaleziono termin", "data", day.Day, "godzina", hour.Time)

			if err := webhook.Send(cfg, msg); err != nil {
				slog.Error("Błąd wysyłki webhooka", "err", err)
			}

			return true, msg, nil
		}
	}

	slog.Debug("Brak wolnych terminów")
	return false, "", nil
}
