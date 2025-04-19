package client

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/kapi1023/word-monitor/internal/config"
)

type Client struct {
	client       *http.Client
	token        string
	tokenExpires time.Time
}

type UserInfo struct {
	Sub               string `json:"sub"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	Email             string `json:"email"`
}

type ExamScheduleRequest struct {
	Category string `json:"category"`
	WordID   string `json:"wordId"`
	Start    string `json:"startDate"`
	End      string `json:"endDate"`
}

type ExamScheduleResponse struct {
	Schedule struct {
		ScheduledDays []struct {
			Day            string `json:"day"`
			ScheduledHours []struct {
				Time          string        `json:"time"`
				PracticeExams []interface{} `json:"practiceExams"`
			} `json:"scheduledHours"`
		} `json:"scheduledDays"`
	} `json:"schedule"`
}

func NewCLient() *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		client: &http.Client{
			Jar: jar,
		},
	}
}

func (c *Client) DoRequest(req *http.Request, tag string) (*http.Response, error) {
	if err := c.BearerAuth(req); err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	slog.Debug(tag, slog.String("status", resp.Status))
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("request failed: " + resp.Status + " " + tag)
	}
	return resp, nil
}

func (c *Client) GetCSRFToken(targetURL string) (string, error) {
	resp, err := c.client.Get(targetURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	csrf, exists := doc.Find("input[name='_csrf']").Attr("value")
	if !exists {
		return "", errors.New("_csrf token not found")
	}

	return csrf, nil
}

func (c *Client) Login(username, password string) error {
	csrfToken, err := c.GetCSRFToken(config.UrlLogin)
	if err != nil {
		return err
	}

	slog.Debug("CSRF token", slog.String("token", csrfToken))
	form := url.Values{}
	form.Add("username", username)
	form.Add("password", password)
	form.Add("_csrf", csrfToken)

	req, err := http.NewRequest("POST", config.UrlLogin, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("X-CSRF-Token", csrfToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	slog.Debug("Login response", slog.String("status", resp.Status))
	if resp.StatusCode != http.StatusOK {
		return errors.New("login failed: " + resp.Status)
	}

	return c.RefreshToken()
}

func (c *Client) RefreshToken() error {
	resp, err := c.client.Get(config.UrlRefresh)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fragment := resp.Request.URL.Fragment
	if fragment == "" {
		return errors.New("no token found in URL fragment")
	}

	values, err := url.ParseQuery(fragment)
	if err != nil {
		return err
	}

	token := values.Get("access_token")
	expiresIn := values.Get("expires_in")
	if token == "" || expiresIn == "" {
		return errors.New("access token or expires_in not found in URL fragment")
	}

	duration, err := time.ParseDuration(expiresIn + "s")
	if err != nil {
		return err
	}
	c.token = token
	c.tokenExpires = time.Now().Add(duration)

	slog.Debug("Token", slog.String("bearer", c.token), slog.Time("expires", c.tokenExpires))
	return nil
}

func (c *Client) BearerAuth(req *http.Request) error {
	if c.token == "" || time.Now().After(c.tokenExpires) {
		return errors.New("token is empty or expired")
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	return nil
}

func (c *Client) GetUserInfo() (*UserInfo, error) {
	req, err := http.NewRequest("GET", config.UrlUserInfo, nil)
	if err != nil {
		return nil, err
	}
	if err := c.BearerAuth(req); err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("request failed: " + resp.Status)
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}
	return &userInfo, nil
}

var format string = "2025-04-21T19:23:03.483Z"

func (c *Client) GetExamSchedule(category, wordID string, start, end time.Time) (*ExamScheduleResponse, error) {
	reqBody := ExamScheduleRequest{
		Category: category,
		WordID:   wordID,
		Start:    start.Format("2006-01-02T15:04:05Z"),
		End:      end.Format("2006-01-02T15:04:05Z"),
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	slog.Debug("ExamScheduleRequest", slog.String("body", string(body)))
	req, err := http.NewRequest("POST", config.UrlScheadule, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.DoRequest(req, "GetExamSchedule")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var scheduleResponse ExamScheduleResponse
	if err := json.NewDecoder(resp.Body).Decode(&scheduleResponse); err != nil {
		return nil, err
	}
	return &scheduleResponse, nil
}

type AvailableWords struct {
	Provinces []Provinces `json:"provinces"`
	Words     []Word      `json:"words"`
}

type Provinces struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
	Zoom      int    `json:"zoom"`
}
type Word struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Address    string `json:"address"`
	Latitude   string `json:"latitude"`
	Longitude  string `json:"longitude"`
	ProvinceID int    `json:"provinceId"`
	Offline    bool   `json:"offline"`
}

func GetWords() ([]Word, error) {
	req, err := http.NewRequest("GET", config.UrlWords, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("request failed: " + resp.Status)
	}

	var availableWords AvailableWords
	if err := json.NewDecoder(resp.Body).Decode(&availableWords); err != nil {
		return nil, err
	}
	return availableWords.Words, nil
}

func GetProvinces() ([]string, error) {
	req, err := http.NewRequest("GET", config.UrlWords, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("request failed: " + resp.Status)
	}

	var availableWords AvailableWords
	if err := json.NewDecoder(resp.Body).Decode(&availableWords); err != nil {
		return nil, err
	}

	var provinces []string
	for _, province := range availableWords.Provinces {
		provinces = append(provinces, province.Name)
	}
	return provinces, nil
}

func GetWordsByProvince(provinceName string) ([]Word, error) {
	req, err := http.NewRequest("GET", config.UrlWords, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("request failed: " + resp.Status)
	}

	var availableWords AvailableWords
	if err := json.NewDecoder(resp.Body).Decode(&availableWords); err != nil {
		return nil, err
	}
	var searchedProvinceId int
	for _, province := range availableWords.Provinces {
		if strings.Contains(strings.ToLower(province.Name), strings.ToLower(provinceName)) {
			searchedProvinceId = province.ID
			break
		}
	}
	if searchedProvinceId == 0 {
		return nil, errors.New("province not found")
	}

	var searchedWords []Word
	for _, word := range availableWords.Words {
		if word.ProvinceID == searchedProvinceId {
			searchedWords = append(searchedWords, word)
		}
	}

	return searchedWords, nil
}
