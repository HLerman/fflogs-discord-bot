package users

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hlerman/fflogs-discord-bot/fflogs"
	"github.com/hlerman/fflogs-discord-bot/lodestone"
	log "github.com/sirupsen/logrus"
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
	db := Connect()
	defer db.Close()

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
	db := Connect()
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
	db := Connect()
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

	db := Connect()
	defer db.Close()

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

// BETA
func Check(m *discordgo.MessageCreate, s *discordgo.Session) {
	// listUsers()
	users, err := listUsers()
	if err != nil {
		log.Fatal(err)
	}

	// Loop
	for i := range users {
		// isThisUserStillInThisDiscord(id)
		u, err := isThisUserStillInThisDiscord(users[i].DiscordID, m, s)
		if err != nil {
			log.Warn(err)
		}

		characters, err := listCharacters(users[i].DiscordID)
		if err != nil {
			log.Warn(err)
		}

		// if false -> listCharacters, removeCharacter (loop), removeUser
		if u == false {
			for j := range characters {
				err := removeCharacter(characters[j].LodestoneID)
				if err != nil {
					log.Warn(err)
				}
			}

			err = removeUser(users[i].DiscordID)
			if err != nil {
				log.Warn(err)
			}

			continue
		}

		for j := range characters {
			// Check the last report of this character
			dpsMeter, err := fflogs.GetLastDpsMeter(characters[j].LodestoneID)
			if err != nil {
				log.Fatal(err)
			}

			// Check if we already send it
			// if yes, next character
			if fflogs.ReportIsAlreadyInDatabase(dpsMeter.ReportID) == true {
				continue
			}

			// Send to Discord
			var message string

			message = "**" + dpsMeter.Name + " (" + dpsMeter.ZoneName + ")**\n"
			for _, dps := range dpsMeter.Dps {
				minutes := fmt.Sprintf("%02d", dps.FightLength.Minute)
				seconds := fmt.Sprintf("%02d", dps.FightLength.Second)

				message = message + "**" + dps.Name + "** * " + dps.Type + " * " + strconv.Itoa(dps.ADPS) + " dps * [" + minutes + ":" + seconds + "]\n"
			}

			s.ChannelMessageSend(m.ChannelID, message)

			// Register id report in db
			err := fflogs.SaveReportInDb(dpsMeter.ReportID)
			if err != nil {
				log.Warn(err)
			}
		}
	}

	// checkLastReport of this character
	// if already in database -> next character
	// if not -> loop boss fights
	// register id of report in db

	// Fin loop users
}

func isThisUserStillInThisDiscord(id string, m *discordgo.MessageCreate, s *discordgo.Session) (bool, error) {
	members, err := s.GuildMembers(m.GuildID, "0", 1000)
	if err != nil {
		return false, err
	}

	for i := range members {
		if id == members[i].User.ID {
			return true, nil
		}

	}

	return false, nil
}

type User struct {
	DiscordID  string
	Characters Characters
}

type Users []User

type Character struct {
	LodestoneID int
}

type Characters []Character

func listUsers() (Users, error) {
	var users Users

	db := Connect()
	defer db.Close()

	rows, err := db.Query("SELECT discord_id FROM users")

	if err != nil {
		return Users{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var discordID string
		err := rows.Scan(&discordID)
		if err != nil {
			return Users{}, err
		}
		users = append(users, User{DiscordID: discordID})
	}

	if err = rows.Err(); err != nil {
		return Users{}, err
	}

	return users, nil
}

func listCharacters(discordID string) (Characters, error) {
	var characters Characters

	db := Connect()
	defer db.Close()

	rows, err := db.Query("SELECT lodestone_id FROM characters WHERE discord_id = ?", discordID)

	if err != nil {
		return Characters{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var lodestoneID int
		err := rows.Scan(&lodestoneID)
		if err != nil {
			return Characters{}, err
		}
		characters = append(characters, Character{LodestoneID: lodestoneID})
	}

	if err = rows.Err(); err != nil {
		return Characters{}, err
	}

	return characters, nil
}

func removeCharacter(lodestoneID int) error {
	db := Connect()
	defer db.Close()

	stmt, err := db.Prepare("DELETE FROM characters WHERE lodestone_id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(lodestoneID)
	if err != nil {
		return err
	}

	return nil
}

func removeUser(discordID string) error {
	db := Connect()
	defer db.Close()

	stmt, err := db.Prepare("DELETE FROM users WHERE discord_id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(discordID)
	if err != nil {
		return err
	}

	return nil
}

// BETA
func Watch(m *discordgo.MessageCreate, s *discordgo.Session) {
	if isUserAlreadyExist(m.Author.ID) == true {
		dpsMeter, err := fflogs.GetLastDpsMeter(387383)
		if err != nil {
			log.Fatal(err)
		}

		var message string

		message = "**" + dpsMeter.Name + " (" + dpsMeter.ZoneName + ")**\n"
		for _, dps := range dpsMeter.Dps {
			minutes := fmt.Sprintf("%02d", dps.FightLength.Minute)
			seconds := fmt.Sprintf("%02d", dps.FightLength.Second)

			message = message + "**" + dps.Name + "** * " + dps.Type + " * " + strconv.Itoa(dps.ADPS) + " dps * [" + minutes + ":" + seconds + "]\n"
		}

		s.ChannelMessageSend(m.ChannelID, message)
		/***Eden Prime (Eden's Gate)**
		- **Hexa Shell** * DRK * 2589.20 dps * **85%** * [10:07]*/
		s.ChannelMessageSendEmbed(m.ChannelID)
		fmt.Println(message)
	}
}
