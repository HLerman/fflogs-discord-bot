package users

import (
	"github.com/bwmarrin/discordgo"
)

func addUser(id string) error {
	return nil
}

func AddCharacter(m *discordgo.MessageCreate, s *discordgo.Session, characterId int) string {
	if isUserAlreadyExist(m.Author.ID) == false {
		err := addUser(m.Author.ID)
		if err != nil {
			return "Oups, je n'arrive pas à vous ajouter"
		}

		s.ChannelMessageSend(m.ChannelID, "Nous ne nous sommes jamais vu non ? Je vous enregistre")
	}

	// add Character
	if isCharacterAlreadyExist() == true {
		s.ChannelMessageSend(m.ChannelID, "Votre personnage est déjà suivi")
	}

	return ""
}

func isUserAlreadyExist(id string) bool {
	return false
}

func isCharacterAlreadyExist() bool {
	return false
}
