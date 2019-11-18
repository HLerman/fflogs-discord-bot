package lodestone

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func IsCharacterIDExistInLodestone(id int) (string, string, error) {
	characterID := strconv.Itoa(id)
	url := "https://xivapi.com/character/" + characterID

	resp, err := http.Get(url)

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var data XIVApi
	if err := json.Unmarshal(html, &data); err != nil {
		log.Fatal(err)
	}

	if data.Error {
		return "", "", errors.New(data.Message)
	}

	return data.Character.Name, data.Character.Server, nil
}

type XIVApi struct {
	Error              bool        `json:"Error"`
	Message            string      `json:"Message"`
	Achievements       interface{} `json:"Achievements"`
	AchievementsPublic interface{} `json:"AchievementsPublic"`
	Character          struct {
		ActiveClassJob struct {
			ClassID       int    `json:"ClassID"`
			ExpLevel      int    `json:"ExpLevel"`
			ExpLevelMax   int    `json:"ExpLevelMax"`
			ExpLevelTogo  int    `json:"ExpLevelTogo"`
			IsSpecialised bool   `json:"IsSpecialised"`
			JobID         int    `json:"JobID"`
			Level         int    `json:"Level"`
			Name          string `json:"Name"`
		} `json:"ActiveClassJob"`
		Avatar    string `json:"Avatar"`
		Bio       string `json:"Bio"`
		ClassJobs []struct {
			ClassID       int    `json:"ClassID"`
			ExpLevel      int    `json:"ExpLevel"`
			ExpLevelMax   int    `json:"ExpLevelMax"`
			ExpLevelTogo  int    `json:"ExpLevelTogo"`
			IsSpecialised bool   `json:"IsSpecialised"`
			JobID         int    `json:"JobID"`
			Level         int    `json:"Level"`
			Name          string `json:"Name"`
		} `json:"ClassJobs"`
		DC            string `json:"DC"`
		FreeCompanyID string `json:"FreeCompanyId"`
		GearSet       struct {
			Attributes struct {
				Num1  int `json:"1"`
				Num2  int `json:"2"`
				Num3  int `json:"3"`
				Num4  int `json:"4"`
				Num5  int `json:"5"`
				Num6  int `json:"6"`
				Num7  int `json:"7"`
				Num8  int `json:"8"`
				Num19 int `json:"19"`
				Num20 int `json:"20"`
				Num21 int `json:"21"`
				Num22 int `json:"22"`
				Num24 int `json:"24"`
				Num27 int `json:"27"`
				Num33 int `json:"33"`
				Num34 int `json:"34"`
				Num44 int `json:"44"`
				Num45 int `json:"45"`
				Num46 int `json:"46"`
			} `json:"Attributes"`
			ClassID int `json:"ClassID"`
			Gear    struct {
				Body struct {
					Creator interface{} `json:"Creator"`
					Dye     interface{} `json:"Dye"`
					ID      int         `json:"ID"`
					Materia []int       `json:"Materia"`
					Mirage  int         `json:"Mirage"`
				} `json:"Body"`
				Bracelets struct {
					Creator interface{} `json:"Creator"`
					Dye     interface{} `json:"Dye"`
					ID      int         `json:"ID"`
					Materia []int       `json:"Materia"`
					Mirage  interface{} `json:"Mirage"`
				} `json:"Bracelets"`
				Earrings struct {
					Creator interface{} `json:"Creator"`
					Dye     interface{} `json:"Dye"`
					ID      int         `json:"ID"`
					Materia []int       `json:"Materia"`
					Mirage  interface{} `json:"Mirage"`
				} `json:"Earrings"`
				Feet struct {
					Creator interface{} `json:"Creator"`
					Dye     interface{} `json:"Dye"`
					ID      int         `json:"ID"`
					Materia []int       `json:"Materia"`
					Mirage  int         `json:"Mirage"`
				} `json:"Feet"`
				Hands struct {
					Creator interface{} `json:"Creator"`
					Dye     interface{} `json:"Dye"`
					ID      int         `json:"ID"`
					Materia []int       `json:"Materia"`
					Mirage  int         `json:"Mirage"`
				} `json:"Hands"`
				Head struct {
					Creator interface{} `json:"Creator"`
					Dye     interface{} `json:"Dye"`
					ID      int         `json:"ID"`
					Materia []int       `json:"Materia"`
					Mirage  int         `json:"Mirage"`
				} `json:"Head"`
				Legs struct {
					Creator interface{} `json:"Creator"`
					Dye     interface{} `json:"Dye"`
					ID      int         `json:"ID"`
					Materia []int       `json:"Materia"`
					Mirage  int         `json:"Mirage"`
				} `json:"Legs"`
				MainHand struct {
					Creator interface{} `json:"Creator"`
					Dye     interface{} `json:"Dye"`
					ID      int         `json:"ID"`
					Materia []int       `json:"Materia"`
					Mirage  int         `json:"Mirage"`
				} `json:"MainHand"`
				Necklace struct {
					Creator interface{} `json:"Creator"`
					Dye     interface{} `json:"Dye"`
					ID      int         `json:"ID"`
					Materia []int       `json:"Materia"`
					Mirage  interface{} `json:"Mirage"`
				} `json:"Necklace"`
				Ring1 struct {
					Creator interface{} `json:"Creator"`
					Dye     interface{} `json:"Dye"`
					ID      int         `json:"ID"`
					Materia []int       `json:"Materia"`
					Mirage  interface{} `json:"Mirage"`
				} `json:"Ring1"`
				Ring2 struct {
					Creator interface{} `json:"Creator"`
					Dye     interface{} `json:"Dye"`
					ID      int         `json:"ID"`
					Materia []int       `json:"Materia"`
					Mirage  interface{} `json:"Mirage"`
				} `json:"Ring2"`
				SoulCrystal struct {
					Creator interface{}   `json:"Creator"`
					Dye     interface{}   `json:"Dye"`
					ID      int           `json:"ID"`
					Materia []interface{} `json:"Materia"`
					Mirage  interface{}   `json:"Mirage"`
				} `json:"SoulCrystal"`
				Waist struct {
					Creator interface{} `json:"Creator"`
					Dye     interface{} `json:"Dye"`
					ID      int         `json:"ID"`
					Materia []int       `json:"Materia"`
					Mirage  interface{} `json:"Mirage"`
				} `json:"Waist"`
			} `json:"Gear"`
			GearKey string `json:"GearKey"`
			JobID   int    `json:"JobID"`
			Level   int    `json:"Level"`
		} `json:"GearSet"`
		Gender       int `json:"Gender"`
		GrandCompany struct {
			NameID int `json:"NameID"`
			RankID int `json:"RankID"`
		} `json:"GrandCompany"`
		GuardianDeity int           `json:"GuardianDeity"`
		ID            int           `json:"ID"`
		Lang          interface{}   `json:"Lang"`
		Minions       []interface{} `json:"Minions"`
		Mounts        []interface{} `json:"Mounts"`
		Name          string        `json:"Name"`
		Nameday       string        `json:"Nameday"`
		ParseDate     int           `json:"ParseDate"`
		Portrait      string        `json:"Portrait"`
		PvPTeamID     interface{}   `json:"PvPTeamId"`
		Race          int           `json:"Race"`
		Server        string        `json:"Server"`
		Title         int           `json:"Title"`
		TitleTop      bool          `json:"TitleTop"`
		Town          int           `json:"Town"`
		Tribe         int           `json:"Tribe"`
	} `json:"Character"`
	FreeCompany        interface{} `json:"FreeCompany"`
	FreeCompanyMembers interface{} `json:"FreeCompanyMembers"`
	Friends            interface{} `json:"Friends"`
	FriendsPublic      interface{} `json:"FriendsPublic"`
	PvPTeam            interface{} `json:"PvPTeam"`
}

func getAllCharacters() {

}
