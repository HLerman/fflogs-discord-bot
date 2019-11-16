package users

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

func connect() *sql.DB {
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

func AddCharacter(m *discordgo.MessageCreate, s *discordgo.Session, characterId int) {
	fmt.Println(isUserAlreadyExist(m.Author.ID))
	if isUserAlreadyExist(m.Author.ID) == false {
		err := addUser(m.Author.ID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Oups, je n'arrive pas à vous ajouter")
			return
		}

		s.ChannelMessageSend(m.ChannelID, "Nous ne nous sommes jamais vu non ? Je vous enregistre")
	}

	// add Character
	if isCharacterAlreadyExist() == true {
		s.ChannelMessageSend(m.ChannelID, "Votre personnage est déjà suivi")
	}
}

func addUser(id string) error {
	db, err := sql.Open("mysql", "root:@/fflogs-discord-bot")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := db.Prepare("INSERT INTO users VALUES( ? )")
	if err != nil {
		log.Warn(err)
		return err
	}
	defer stmt.Close()

	i, err := strconv.Atoi(id)
	if err != nil {
		log.Fatal(err)
	}
	_, err = stmt.Exec(i)
	if err != nil {
		log.Warn(err)
		return err
	}

	return nil
}

func isUserAlreadyExist(id string) bool {
	db := connect()
	defer db.Close()

	i, err := strconv.Atoi(id)
	if err != nil {
		log.Fatal(err)
	}

	var discordID int
	err = db.QueryRow("SELECT discord_id FROM users WHERE discord_id = ?", i).Scan(&discordID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false
		} else {
			log.Fatal(err)
		}
	}

	return true
}

func isCharacterAlreadyExist() bool {
	return false
}
