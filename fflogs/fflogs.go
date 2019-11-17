package fflogs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/viper"
)

func checkNewContent() {
	name := "Hexa Shell"
	fmt.Println(url.PathEscape(name))
}

func getZones() (Zones, error) {
	apiKey := viper.GetString("token")
	url := "https://www.fflogs.com:443/v1/zones?api_key=" + apiKey

	resp, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data Zones
	if err := json.Unmarshal(html, &data); err != nil {
		return nil, err
	}

	return data, nil
}

type Zones []struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Frozen     bool   `json:"frozen"`
	Encounters []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"encounters"`
	Brackets struct {
		Min    int     `json:"min"`
		Max    float64 `json:"max"`
		Bucket float64 `json:"bucket"`
		Type   string  `json:"type"`
	} `json:"brackets"`
	Partitions []struct {
		Name    string `json:"name"`
		Compact string `json:"compact"`
		Area    int    `json:"area,omitempty"`
		Default bool   `json:"default,omitempty"`
	} `json:"partitions,omitempty"`
}

func getZoneFromID(id int, zones Zones) (string, error) {
	for i, zone := range zones {
		if zone.ID == id {
			return zone.Name, nil
		}
	}

	return "", errors.New("Cannot retrieve name of this zone")
}

func getLastsParsesForCharacter(name string, server string, region string) (Parses, error) {
	apiKey := viper.GetString("token")
	url := "https://www.fflogs.com/v1/parses/character/" + url.PathEscape(name) + "/" + url.PathEscape(server) + "/" + strings.ToUpper(region) + "?api_key=" + apiKey

	resp, err := http.Get(url)

	if err != nil {
		return Parses{}, err
	}

	defer resp.Body.Close()

	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Parses{}, err
	}

	var data Parses
	if err := json.Unmarshal(html, &data); err != nil {
		return Parses{}, err
	}

	if data[0].Status != 0 {
		return Parses{}, errors.New(data[0].Error)
	}

	return data, nil
}

type Parses []struct {
	Status         int     `json:"status"`
	Error          string  `json:"error"`
	EncounterID    int     `json:"encounterID"`
	EncounterName  string  `json:"encounterName"`
	Class          string  `json:"class"`
	Spec           string  `json:"spec"`
	Rank           int     `json:"rank"`
	OutOf          int     `json:"outOf"`
	Duration       int     `json:"duration"`
	StartTime      int64   `json:"startTime"`
	ReportID       string  `json:"reportID"`
	FightID        int     `json:"fightID"`
	Difficulty     int     `json:"difficulty"`
	CharacterID    int     `json:"characterID"`
	CharacterName  string  `json:"characterName"`
	Server         string  `json:"server"`
	Percentile     int     `json:"percentile"`
	IlvlKeyOrPatch float64 `json:"ilvlKeyOrPatch"`
	Total          float64 `json:"total"`
	Estimated      bool    `json:"estimated"`
}
