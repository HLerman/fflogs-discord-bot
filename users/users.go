package users

import (
	"database/sql"

	"github.com/bwmarrin/discordgo"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hlerman/fflogs-discord-bot/fflogs"
	"github.com/hlerman/fflogs-discord-bot/lodestone"
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

func AddCharacter(m *discordgo.MessageCreate, s *discordgo.Session, characterID int) {
	if isUserAlreadyExist(m.Author.ID) == false {
		err := addUser(m.Author.ID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Oups, je n'arrive pas à vous ajouter")
			return
		}

		s.ChannelMessageSend(m.ChannelID, "Nous ne nous sommes jamais vu non ? Je vous enregistre")
	}

	// add Character
	if isCharacterAlreadyExist(characterID) == true {
		s.ChannelMessageSend(m.ChannelID, "Votre personnage est déjà suivi :thumbsup:")
	} else {
		name, server, err := saveCharacter(m.Author.ID, characterID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Oups, je n'arrive pas à ajouter votre personnage :cry:. Peut être que l'id lodestone est erroné ou que le personnage n'existe pas dans FFLogs ?")
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Votre personnage "+name+" du serveur "+server+" a bien été enregistré :rocket:")
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

	_, err = stmt.Exec(id)
	if err != nil {
		log.Warn(err)
		return err
	}

	return nil
}

func isUserAlreadyExist(id string) bool {
	db := connect()
	defer db.Close()

	var discordID int
	err := db.QueryRow("SELECT discord_id FROM users WHERE discord_id = ?", id).Scan(&discordID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false
		} else {
			log.Fatal(err)
		}
	}

	return true
}

func isCharacterAlreadyExist(id int) bool {
	db := connect()
	defer db.Close()

	var characterID int
	err := db.QueryRow("SELECT lodestone_id FROM characters WHERE lodestone_id = ?", id).Scan(&characterID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false
		} else {
			log.Fatal(err)
		}
	}

	return true
}

func saveCharacter(user string, id int) (string, string, error) {
	name, server, err := lodestone.IsCharacterIDExistInLodestone(id)

	if err != nil {
		log.Warn(err)
		return "", "", err
	}

	_, err = fflogs.GetLastsParsesForCharacter(name, server, "EU")
	if err != nil {
		log.Warn(err)
		return "", "", err
	}

	db, err := sql.Open("mysql", "root:@/fflogs-discord-bot")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := db.Prepare("INSERT INTO characters VALUES( ?, ?, ? )")
	if err != nil {
		log.Warn(err)
		return "", "", err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id, 0, user)
	if err != nil {
		log.Warn(err)
		return "", "", err
	}

	return name, server, nil
}
