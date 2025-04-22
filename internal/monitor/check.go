package monitor

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/kapi1023/word-monitor/internal/cache"
	"github.com/kapi1023/word-monitor/internal/config"
	"github.com/kapi1023/word-monitor/internal/infocar"
	"github.com/kapi1023/word-monitor/internal/state"
	"github.com/kapi1023/word-monitor/internal/webhook"
)
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
	var word *infocar.Word
	var err error
	word, err = i.GetWordById(c, cfg.Word.WordId)
	if err != nil {
		word, err =i.GetWordById(c, cfg.Word.WordId)
		if err == nil{
			continue
		}
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

			if storage.Exists(key, day.Day, hour.Time) {
				continue
			}
			msg := fmt.Sprintf(
				"**Wolny termin egzaminu!**\n📅 Data: `%s`\n⏰ Godzina: `%s`\n📍 WORD: `%s (%s)`\n📁 Kategoria: `%s`\n🆔 ID: `%s`\n📂 Praktyczne: `%d`\n🤦‍♂️ Teoretyczne: `%d`",
				day.Day,
				hour.Time,
				word.Name,
				word.City,
				cfg.Word.Category,
				strconv.Itoa(cfg.Word.WordId),
				len(hour.PracticeExams),
				len(hour.TheoryExams),
			)
			messages = append(messages, msg)

			var practiceIDs, theoryIDs []string
			for _, p := range hour.PracticeExams {
				practiceIDs = append(practiceIDs, p.ID)
			}
			for _, t := range hour.TheoryExams {
				theoryIDs = append(theoryIDs, t.ID)
			}

			slot := state.ExamSlot{
				Day:         day.Day,
				Time:        hour.Time,
				PracticeIDs: practiceIDs,
				TheoryIDs:   theoryIDs,
			}
			storage.Add(key, slot)
			slog.Warn("Znaleziono NOWY termin", "data", day.Day, "godzina", hour.Time)
		}
	}

	if len(messages) > 0 {
		for _, msg := range messages {
			webhook.Send(cfg, msg)
			time.Sleep(250 * time.Millisecond)
		}
		return true, "", nil
	}

	return false, "", nil
}
