package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/kapi1023/word-monitor/internal/cache"
	"github.com/kapi1023/word-monitor/internal/config"
	"github.com/kapi1023/word-monitor/internal/infocar"
	"github.com/kapi1023/word-monitor/internal/monitor"
	"github.com/kapi1023/word-monitor/internal/state"
	"github.com/kapi1023/word-monitor/internal/webhook"
)

const (
	path      = "internal/config/config.yaml"
	statePath = "internal/state/state.enc"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = path
	}
	slog.Info("Używana konfiguracja", "path", configPath)
	statePath := os.Getenv("STATE_PATH")
	if statePath == "" {
		statePath = statePath
	}
	slog.Info("Używana konfiguracja", "path", statePath)

	cfg, err := config.Load(configPath)
	if err != nil {
		slog.Error("Błąd ładowania konfiguracji", "err", err)
		cfg = config.NewConfig()
		if err := cfg.Create(configPath); err != nil {
			slog.Error("Błąd zapisu konfiguracji", "err", err)
			return
		}
		cfg.Show()
		slog.Info("Utworzono nową konfigurację")
	} else {
		slog.Info("Wczytano konfigurację")
	}
	level := slog.LevelInfo
	if cfg.Monitor.Debug {
		level = slog.LevelDebug
	}

	storage, err := state.New(statePath, cfg.State.SecretKey)
	if err != nil {
		slog.Error("Błąd inicjalizacji state storage", "err", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)
	reader := bufio.NewScanner(os.Stdin)
	c := cache.New[infocar.Word]()

	for {
		fmt.Println("\n--- WORD MONITOR ---")
		fmt.Println("1. Start monitoringu")
		fmt.Println("2. Pokaż konfigurację")
		fmt.Println("3. Edytuj konfigurację")
		fmt.Println("4. Zapisz konfigurację")
		fmt.Println("5. Pokaz dostepne wordy")
		fmt.Println("6. Pokaz dostepne wordy w województwie")
		fmt.Println("7. Pokaz wojewodztwa")
		fmt.Println("5. Wyjdź")
		fmt.Print("Wybierz opcję: ")

		if !reader.Scan() {
			break
		}
		choice := reader.Text()

		switch choice {
		case "1":
			startMonitoring(cfg, storage, c)
		case "2":
			cfg.Show()
		case "3":
			cfg.Edit()
		case "4":
			if err := cfg.Save(configPath); err != nil {
				slog.Error("Błąd zapisu konfiguracji", "err", err)
			} else {
				slog.Info("Konfiguracja zapisana")
			}
		case "5":
			words, err := infocar.GetWords()
			if err != nil || words == nil {
				slog.Error("Błąd pobierania dostępnych WORDów", "err", err)
			}
			fmt.Println("--- DOSTĘPNE WORDY ---")
			for _, word := range words {
				fmt.Printf("ID: %d, Nazwa: %s\n", word.ID, word.Name)
			}
		case "6":
			fmt.Println("Podaj nazwe województwa:")
			if !reader.Scan() {
				break
			}
			provinceName := reader.Text()
			words, err := infocar.GetWordsByProvince(provinceName)
			if err != nil || words == nil {
				slog.Error("Błąd pobierania dostępnych WORDów w regionie", "err", err)
			}
			fmt.Println("--- DOSTĘPNE WORDY W WOJEWÓDZTWIE ---")
			for _, word := range words {
				fmt.Printf("ID: %v, Nazwa: %s\n", word.ID, word.Name)
			}
		case "7":
			regions, err := infocar.GetProvinces()
			if err != nil || regions == nil {
				slog.Error("Błąd pobierania dostępnych województw", "err", err)
			}
			fmt.Println("--- DOSTĘPNE WOJEWÓDZTWA ---")
			for _, region := range regions {
				fmt.Printf("Nazwa: %s\n", region)
			}
		case "8":
			fmt.Println("--- EXIT ---")
			os.Exit(0)

		default:
			fmt.Println("Nieprawidłowa opcja.")
		}
	}
}

func startMonitoring(cfg *config.Config, storage *state.Storage, c *cache.Cache[infocar.Word]) {
	slog.Info("Rozpoczęcie monitoringu...")
	client := infocar.NewCLient()
	if err := client.Login(cfg.Credential.Username, cfg.Credential.Password); err != nil {
		slog.Error("Błąd logowania", "err", err)
		return
	}

	slog.Info("Zalogowano pomyślnie. Start monitoringu...")
	var i int
	for {
		found, _, err := monitor.Check(cfg, client, storage, c)
		if err != nil {
			if err.Error() == "token is empty or expired" {
				slog.Info("Token wygasł, ponowne logowanie...")
				if err := client.Login(cfg.Credential.Username, cfg.Credential.Password); err != nil {
					slog.Error("Błąd logowania", "err", err)
					return
				}
				slog.Info("Zalogowano pomyślnie. Start monitoringu...")
				continue
			}
			slog.Error("Błąd podczas sprawdzania dostępności", "err", err)
		}
		if !found {
			slog.Debug("Brak dostępnych terminów")
		}
		time.Sleep(time.Duration(cfg.Monitor.Interval) * time.Second)
		i++
		if cfg.Monitor.HealthCheckInterval != 0 && i%cfg.Monitor.HealthCheckInterval == 0 {
			webhook.Send(cfg.Webhook.DiscordHealthCheckUrl, "Health check")
			slog.Debug("Wysłano health check")
		}
	}
}
