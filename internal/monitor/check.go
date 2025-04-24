package monitor

import (
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/kapi1023/word-monitor/internal/cache"
	"github.com/kapi1023/word-monitor/internal/config"
	"github.com/kapi1023/word-monitor/internal/infocar"
	"github.com/kapi1023/word-monitor/internal/state"
	"github.com/kapi1023/word-monitor/internal/webhook"
)

func Check(cfg *config.Config, i *infocar.InfocarClient, storage *state.Storage, c *cache.Cache[infocar.Word]) (bool, string, error) {
	now := time.Now()
	end := now.Add(time.Duration(cfg.Word.MaxDays) * 24 * time.Hour)

	schedule, err := i.GetExamSchedule(cfg.Word.Category, cfg.Word.WordId, now, end)
	if err != nil {
		return false, "", err
	}

	key := state.Key(cfg.Word.WordId, cfg.Word.Category)
	var messages []string
	var word infocar.Word
	wordId, _ := strconv.Atoi(cfg.Word.WordId)
	word, err = i.GetWordById(c, wordId)
	if err != nil {
		slog.Warn("Nie udało się pobrać danych WORD", "id", cfg.Word.WordId, "error", err)
	}

	for _, day := range schedule.Schedule.ScheduledDays {
		examDate, err := time.Parse("2006-01-02", day.Day)
		if err != nil || examDate.After(end) {
			continue
		}
		for _, hour := range day.ScheduledHours {
			if len(hour.PracticeExams) == 0 && len(hour.TheoryExams) == 0 {
				continue
			}
			hasPractice := cfg.Monitor.PracticeExams && len(hour.PracticeExams) > 0
			hasTheory := cfg.Monitor.TheoryExams && len(hour.TheoryExams) > 0

			if !hasPractice && !hasTheory {
				continue
			}

			if storage.Exists(key, day.Day, hour.Time) {
				continue
			}

			if hasPractice {
				msg := fmt.Sprintf(
					"**Wolny termin egzaminu praktycznego!**\n📅 Data: `%s`\n⏰ Godzina: `%s`\n📍 WORD: `%s (%s)`\n📁 Kategoria: `%s`\n🆔 ID: `%s`\n📂 Dostępne: `%d`",
					day.Day,
					hour.Time,
					word.Name,
					word.Address,
					cfg.Word.Category,
					cfg.Word.WordId,
					len(hour.PracticeExams),
				)
				messages = append(messages, msg)
			}

			if hasTheory {
				msg := fmt.Sprintf(
					"**Wolny termin egzaminu teoretycznego!**\n📅 Data: `%s`\n⏰ Godzina: `%s`\n📍 WORD: `%s (%s)`\n📁 Kategoria: `%s`\n🆔 ID: `%s`\n🤦‍♂️ Dostępne: `%d`",
					day.Day,
					hour.Time,
					word.Name,
					word.Address,
					cfg.Word.Category,
					cfg.Word.WordId,
					len(hour.TheoryExams),
				)
				messages = append(messages, msg)
			}

			var practiceIDs, theoryIDs []string
			if hasPractice {
				for _, p := range hour.PracticeExams {
					practiceIDs = append(practiceIDs, p.ID)
				}
			}
			if hasTheory {
				for _, t := range hour.TheoryExams {
					theoryIDs = append(theoryIDs, t.ID)
				}
			}

			storage.Add(key, state.ExamSlot{
				Day:         day.Day,
				Time:        hour.Time,
				PracticeIDs: practiceIDs,
				TheoryIDs:   theoryIDs,
			})
			slog.Warn("Znaleziono NOWY termin", "data", day.Day, "godzina", hour.Time)
		}
	}

	if len(messages) > 0 {
		for _, msg := range messages {
			webhook.Send(cfg.Webhook.DiscordURL, msg)
			time.Sleep(250 * time.Millisecond)
		}
		return true, "", nil
	}

	return false, "", nil
}
