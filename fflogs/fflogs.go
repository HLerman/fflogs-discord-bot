package fflogs

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/hlerman/fflogs-discord-bot/lodestone"
	"github.com/spf13/viper"
)

func Connect() *sql.DB {
	db, err := sql.Open("mysql", "root:@/fflogs-discord-bot")
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func ReportIsAlreadyInDatabase(id string) bool {
	db := Connect()
	defer db.Close()

	var reportID string
	err := db.QueryRow("SELECT report_id FROM reports WHERE report_id = ?", id).Scan(&reportID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false
		} else {
			log.Fatal(err)
		}
	}

	return true
}

func SaveReportInDb(id string) error {
	db := Connect()
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO reports (report_id) VALUES( ? )")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}

	return nil
}

type Date struct {
	Second int
	Minute int
	Hour   int
}

func convertSecondToDate(s int) Date {
	hours := int(math.Floor(float64(s) / 60 / 60))
	seconds := s % (60 * 60)
	minutes := int(math.Floor(float64(seconds) / 60))
	seconds = s % 60

	return Date{Second: seconds, Minute: minutes, Hour: hours}
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

func GetLastDpsMeter(characterID int) (DPSMeters, error) {
	// Get Name and Server From Character
	name, server, err := lodestone.IsCharacterIDExistInLodestone(characterID)

	if err != nil {
		return DPSMeters{}, err
	}

	// Get last fight for character (reportID)
	parse, err := getLastParseForCharacter(name, server, "EU")
	if err != nil {
		return DPSMeters{}, err
	}

	date := parse.StartTime / 1000

	// Get list of fight from reportID
	fights, reportID, err := getReportFights(parse.ReportID)
	if err != nil {
		return DPSMeters{}, err
	}

	// Get list of boss fights from fights
	bossFights, reportID, err := getBossFights(fights, reportID)
	if err != nil {
		return DPSMeters{}, err
	}

	// Loop all the boss fights -- for debug, we take the first fight :)
	var dpsMeters DPSMeters
	dpsMeters.Date = date

	var boss Boss
	for i := range bossFights {
		boss = bossFights[i]

		// Oups, get the ZoneName and Name
		bossName := boss.Name
		zoneName := boss.ZoneName

		// Get information from the boss fight
		tables, err := getReportTables(reportID, boss.StartTime, boss.EndTime)
		if err != nil {
			return DPSMeters{}, err
		}

		// Finaly, get dps meter
		dpsMeter, err := getFightInformationFromTables(tables, bossName, zoneName)
		if err != nil {
			return DPSMeters{}, err
		}

		dpsMeter.Name = bossName
		dpsMeter.ZoneName = zoneName
		dpsMeter.ReportID = reportID

		dpsMeters.Meters = append(dpsMeters.Meters, dpsMeter)
	}

	return dpsMeters, nil
}

type DPSMeters struct {
	Date   int64
	Meters []DPSMeter
}

type DPSMeter struct {
	Dps      []DPS
	Name     string
	ZoneName string
	ReportID string
}

type DPS struct {
	Name        string
	Type        string
	RDPS        int
	ADPS        int
	FightLength Date
}

func getFightInformationFromTables(tables Tables, bossName string, zoneName string) (DPSMeter, error) {
	dpsMeter := DPSMeter{Name: bossName, ZoneName: zoneName, Dps: []DPS{}}

	for i := range tables.Entries {
		var dps DPS

		dps.Name = tables.Entries[i].Name
		dps.Type = tables.Entries[i].Type
		dps.ADPS = int(math.RoundToEven(tables.Entries[i].TotalADPS)) / (int(tables.TotalTime) / 1000)
		dps.RDPS = int(math.RoundToEven(tables.Entries[i].TotalRDPS)) / (int(tables.TotalTime) / 1000)
		dps.FightLength = convertSecondToDate(tables.TotalTime / 1000)

		dpsMeter.Dps = append(dpsMeter.Dps, dps)
	}

	sort.Slice(dpsMeter.Dps, func(i, j int) bool { return dpsMeter.Dps[i].ADPS > dpsMeter.Dps[j].ADPS })

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
