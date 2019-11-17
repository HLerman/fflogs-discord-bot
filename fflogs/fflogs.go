package fflogs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

func checkNewContent() {
	name := "Hexa Shell"
	fmt.Println(url.PathEscape(name))
}

func getZones() (Zones, error) {
	apiKey := viper.GetString("fflogsToken")
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
	for _, zone := range zones {
		if zone.ID == id {
			return zone.Name, nil
		}
	}

	return "", errors.New("Cannot retrieve name of this zone")
}

func GetLastsParsesForCharacter(name string, server string, region string) (Parses, error) {
	apiKey := viper.GetString("fflogsToken")
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

func getLastParseForCharacter(name string, server string, region string) (Parse, error) {
	parses, err := GetLastsParsesForCharacter(name, server, region)
	if err != nil {
		return Parse{}, err
	}

	parse := parses[0]

	return parse, nil
}

type DPSMeter []DPS

type DPS struct {
	Name        string
	Type        string
	RDPS        int64
	ADPS        int64
	FightLength int
}

func getFightInformationFromTables(tables Tables) (DPSMeter, error) {
	var dpsMeter DPSMeter

	for i := range tables.Entries {
		var dps DPS

		dps.Name = tables.Entries[i].Name
		dps.Type = tables.Entries[i].Type
		dps.ADPS = int64(math.RoundToEven(tables.Entries[i].TotalADPS)) / (int64(tables.TotalTime) / 1000)
		dps.RDPS = int64(math.RoundToEven(tables.Entries[i].TotalRDPS)) / (int64(tables.TotalTime) / 1000)
		dps.FightLength = tables.TotalTime / 1000

		dpsMeter = append(dpsMeter, dps)
	}

	return dpsMeter, nil
}

func getReportTables(reportID string, startTime int, endTime int) (Tables, error) {
	apiKey := viper.GetString("fflogsToken")
	url := "https://www.fflogs.com/v1/report/tables/damage-done/" + reportID + "?start=" + strconv.Itoa(startTime) + "&end=" + strconv.Itoa(endTime) + "&translate=true&api_key=" + apiKey
	resp, err := http.Get(url)

	if err != nil {
		return Tables{}, err
	}

	defer resp.Body.Close()

	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Tables{}, err
	}

	var data Tables
	if err := json.Unmarshal(html, &data); err != nil {
		return Tables{}, err
	}

	if data.Status != 0 {
		return Tables{}, errors.New(data.Error)
	}

	return data, nil
}

type Tables struct {
	Status  int    `json:"status"`
	Error   string `json:"error"`
	Entries []struct {
		Name              string `json:"name"`
		ID                int    `json:"id"`
		GUID              int    `json:"guid"`
		Type              string `json:"type"`
		Icon              string `json:"icon"`
		Total             int    `json:"total"`
		ActiveTime        int    `json:"activeTime"`
		ActiveTimeReduced int    `json:"activeTimeReduced"`
		Abilities         []struct {
			Name  string `json:"name"`
			Total int    `json:"total"`
			Type  int    `json:"type"`
		} `json:"abilities"`
		DamageAbilities []interface{} `json:"damageAbilities"`
		Targets         []struct {
			Name  string `json:"name"`
			Total int    `json:"total"`
			Type  string `json:"type"`
		} `json:"targets"`
		TotalRDPS      float64 `json:"totalRDPS"`
		TotalRDPSTaken float64 `json:"totalRDPSTaken"`
		TotalRDPSGiven float64 `json:"totalRDPSGiven"`
		TotalADPS      float64 `json:"totalADPS"`
		Given          []struct {
			GUID        int     `json:"guid"`
			Name        string  `json:"name"`
			Total       float64 `json:"total"`
			Type        string  `json:"type"`
			AbilityIcon string  `json:"abilityIcon"`
		} `json:"given"`
		Taken []struct {
			GUID        int     `json:"guid"`
			Name        string  `json:"name"`
			Total       float64 `json:"total"`
			Type        string  `json:"type"`
			AbilityIcon string  `json:"abilityIcon"`
		} `json:"taken"`
		TotalReduced int `json:"totalReduced,omitempty"`
		Pets         []struct {
			Name           string        `json:"name"`
			ID             int           `json:"id"`
			GUID           int           `json:"guid"`
			Type           string        `json:"type"`
			Icon           string        `json:"icon"`
			Total          int           `json:"total"`
			TotalReduced   int           `json:"totalReduced"`
			ActiveTime     int           `json:"activeTime"`
			TotalRDPS      float64       `json:"totalRDPS"`
			TotalRDPSTaken float64       `json:"totalRDPSTaken"`
			TotalRDPSGiven float64       `json:"totalRDPSGiven"`
			TotalADPS      float64       `json:"totalADPS"`
			Given          []interface{} `json:"given"`
			Taken          []struct {
				GUID        int     `json:"guid"`
				Name        string  `json:"name"`
				Total       float64 `json:"total"`
				Type        string  `json:"type"`
				AbilityIcon string  `json:"abilityIcon"`
			} `json:"taken"`
		} `json:"pets,omitempty"`
	} `json:"entries"`
	TotalTime   int `json:"totalTime"`
	LogVersion  int `json:"logVersion"`
	GameVersion int `json:"gameVersion"`
}

type Bosses []Boss

type Boss struct {
	StartTime int
	EndTime   int
	Name      string
	ZoneName  string
}

func getBossFights(fights Fights, reportID string) (Bosses, string, error) {
	var bosses Bosses

	for _, fight := range fights.Fights {
		if fight.Boss > 0 {
			bosses = append(bosses, Boss{StartTime: fight.StartTime, EndTime: fight.EndTime, Name: fight.Name, ZoneName: fight.ZoneName})
		}
	}

	if len(bosses) == 0 {
		return bosses, reportID, errors.New("No bosses during these fights")
	}

	return bosses, reportID, nil
}

func getReportFights(reportID string) (Fights, string, error) {
	apiKey := viper.GetString("fflogsToken")
	url := "https://www.fflogs.com/v1/report/fights/" + reportID + "?api_key=" + apiKey
	resp, err := http.Get(url)

	if err != nil {
		return Fights{}, reportID, err
	}

	defer resp.Body.Close()

	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Fights{}, reportID, err
	}

	var data Fights
	if err := json.Unmarshal(html, &data); err != nil {
		return Fights{}, reportID, err
	}

	if data.Status != 0 {
		return Fights{}, reportID, errors.New(data.Error)
	}

	return data, reportID, nil
}

type Parses []Parse

type Parse struct {
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

type Fights struct {
	Status int    `json:"status"`
	Error  string `json:"error"`
	Fights []struct {
		ID                            int    `json:"id"`
		StartTime                     int    `json:"start_time"`
		EndTime                       int    `json:"end_time"`
		Boss                          int    `json:"boss"`
		Name                          string `json:"name"`
		ZoneID                        int    `json:"zoneID"`
		ZoneName                      string `json:"zoneName"`
		Size                          int    `json:"size,omitempty"`
		Difficulty                    int    `json:"difficulty,omitempty"`
		Kill                          bool   `json:"kill,omitempty"`
		Partial                       int    `json:"partial,omitempty"`
		StandardComposition           bool   `json:"standardComposition,omitempty"`
		BossPercentage                int    `json:"bossPercentage,omitempty"`
		FightPercentage               int    `json:"fightPercentage,omitempty"`
		LastPhaseForPercentageDisplay int    `json:"lastPhaseForPercentageDisplay,omitempty"`
	} `json:"fights"`
}
