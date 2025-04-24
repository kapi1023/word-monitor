package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type Credential struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Pesel    string `yaml:"pesel"`
	PKK      string `yaml:"pkk"`
	Email    string `yaml:"email"`
	Phone    string `yaml:"phone"`
}

type Webhook struct {
	DiscordURL            string `yaml:"discord_url"`
	DiscordHealthCheckUrl string `yaml:"discord_health_check_url"`
}

type WORD struct {
	WordId   string `yaml:"word_id"`
	Category string `yaml:"category"`
	MaxDays  int    `yaml:"max_days"`
}

type Monitor struct {
	UrlLogin            string `yaml:"url_login"`
	UrlCheck            string `yaml:"url_check"`
	Interval            int    `yaml:"interval"`
	HealthCheckInterval int    `yaml:"health_check_interval"`
	Proxy               bool   `yaml:"proxy"`
	ProxyAddress        string `yaml:"proxy_address"`
	Debug               bool   `yaml:"debug"`
	PracticeExams       bool   `yaml:"practice_exams"`
	TheoryExams         bool   `yaml:"theory_exams"`
}

type State struct {
	SecretKey string `yaml:"secret_key"`
}
type Config struct {
	Credential Credential `yaml:"credential"`
	Webhook    Webhook    `yaml:"webhook"`
	Monitor    Monitor    `yaml:"monitor"`
	Word       WORD       `yaml:"word"`
	State      State      `yaml:"state"`
}

const (
	UrlLogin     = "https://info-car.pl/oauth2/login"
	UrlRefresh   = "https://info-car.pl/oauth2/authorize?response_type=id_token%20token&client_id=client&redirect_uri=https://info-car.pl/new/assets/refresh.html&scope=openid%20profile%20email%20resource.read&prompt=none"
	UrlUserInfo  = "https://info-car.pl/oauth2/userinfo"
	UrlScheadule = "https://info-car.pl/api/word/word-centers/exam-schedule"
	UrlWords     = "https://info-car.pl/api/word/word-centers"
)

func NewConfig() *Config {
	config := &Config{}
	config.Edit()
	return config
}

func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Create(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (c *Config) Show() {
	fmt.Println("\n--- KONFIGURACJA ---")
	fmt.Printf("Login: %s\n", c.Credential.Username)
	fmt.Printf("PKK: %s\n", c.Credential.PKK)
	fmt.Printf("PESEL: %s\n", c.Credential.Pesel)
	fmt.Printf("Phone: %s\n", c.Credential.Phone)
	fmt.Printf("Email: %s\n", c.Credential.Email)

	fmt.Printf("WORD ID: %s\n", c.Word.WordId)
	fmt.Printf("Kategoria: %s\n", c.Word.Category)
	fmt.Printf("Max dni do egzaminu: %d\n", c.Word.MaxDays)

	fmt.Printf("Interval (sekundy): %d\n", c.Monitor.Interval)
	fmt.Printf("Proxy: %t (%s)\n", c.Monitor.Proxy, c.Monitor.ProxyAddress)

	fmt.Printf("Discord: %s\n", c.Webhook.DiscordURL)
}

func (c *Config) Edit() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("\n--- EDYCJA KONFIGURACJI ---")

	input := func(label, current string) string {
		fmt.Printf("%s [%s]: ", label, current)
		scanner.Scan()
		t := scanner.Text()
		if t == "" {
			return current
		}
		return t
	}

	inputBool := func(label string, current bool) bool {
		fmt.Printf("%s [%t]: ", label, current)
		scanner.Scan()
		t := strings.ToLower(scanner.Text())
		if t == "" {
			return current
		}
		return t == "true" || t == "1" || t == "yes"
	}

	inputInt := func(label string, current int) int {
		fmt.Printf("%s [%d]: ", label, current)
		scanner.Scan()
		t := scanner.Text()
		if t == "" {
			return current
		}
		v, err := strconv.Atoi(t)
		if err != nil {
			return current
		}
		return v
	}

	// Dane logowania
	c.Credential.Username = input("Login", c.Credential.Username)
	c.Credential.Password = input("Hasło", c.Credential.Password)
	c.Credential.Pesel = input("PESEL", c.Credential.Pesel)
	c.Credential.PKK = input("PKK", c.Credential.PKK)
	c.Credential.Email = input("Email", c.Credential.Email)
	c.Credential.Phone = input("Telefon", c.Credential.Phone)

	// WORD
	c.Word.WordId = input("WORD ID", c.Word.WordId)
	c.Word.Category = input("Kategoria", c.Word.Category)
	c.Word.MaxDays = inputInt("Max dni do egzaminu", c.Word.MaxDays)

	// Monitor
	c.Monitor.UrlLogin = input("URL logowania", c.Monitor.UrlLogin)
	c.Monitor.UrlCheck = input("URL sprawdzania", c.Monitor.UrlCheck)
	c.Monitor.Interval = inputInt("Interwał (sekundy)", c.Monitor.Interval)
	c.Monitor.HealthCheckInterval = inputInt("Interwał sprawdzania zdrowia ilosc interwalow", c.Monitor.HealthCheckInterval)
	c.Monitor.Proxy = inputBool("Używać proxy?", c.Monitor.Proxy)
	c.Monitor.ProxyAddress = input("Adres proxy", c.Monitor.ProxyAddress)
	c.Monitor.Debug = inputBool("Debug", c.Monitor.Debug)
	c.Monitor.PracticeExams = inputBool("Sprawdzać praktyczne egzaminy? puste = false", c.Monitor.PracticeExams)
	c.Monitor.TheoryExams = inputBool("Sprawdzać teoretyczne egzaminy? puste = false", c.Monitor.TheoryExams)

	// Webhook
	c.Webhook.DiscordURL = input("Discord webhook URL", c.Webhook.DiscordURL)
	c.Webhook.DiscordHealthCheckUrl = input("Discord webhook URL health check", c.Webhook.DiscordHealthCheckUrl)
}
