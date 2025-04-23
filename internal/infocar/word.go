package infocar

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kapi1023/word-monitor/internal/cache"
	"github.com/kapi1023/word-monitor/internal/config"
)

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
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Address    string    `json:"address"`
	Latitude   string    `json:"latitude"`
	Longitude  string    `json:"longitude"`
	ProvinceID int       `json:"provinceId"`
	Offline    bool      `json:"offline"`
	Time       time.Time `json:"-"`
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
func (i *InfocarClient) GetWordById(cache *cache.Cache[Word], wordId int) (Word, error) {
	idStr := strconv.Itoa(wordId)

	if cachedWord, err := cache.Get(idStr); err == nil {
		return *cachedWord, nil
	}

	req, err := http.NewRequest("GET", config.UrlWords, nil)
	if err != nil {
		return Word{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Word{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Word{}, errors.New("request failed: " + resp.Status)
	}

	var availableWords AvailableWords
	if err := json.NewDecoder(resp.Body).Decode(&availableWords); err != nil {
		return Word{}, err
	}

	for _, word := range availableWords.Words {
		_ = cache.Set(strconv.Itoa(word.ID), word)
		if word.ID == wordId {
			return word, nil
		}
	}
	return Word{}, errors.New("word not found")
}
