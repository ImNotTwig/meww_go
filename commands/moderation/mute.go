package moderation

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

func GetRoleFromFile(gid string) (*string, error) {
	file, err := os.ReadFile("./commands/moderation/mute_roles.json")
	if err != nil {
		return nil, err
	}

	var muteroledict map[string]string

	json.Unmarshal(file, &muteroledict)

	if val, ok := muteroledict[gid]; ok {
		return &val, nil
	}

	return nil, errors.New("Could not get role from file.")
}

func WriteRolesFile(gid, roleid string) error {
	file, err := os.ReadFile("./commands/moderation/mute_roles.json")
	if err != nil {
		return err
	}

	var muteroledict map[string]string

	json.Unmarshal(file, &muteroledict)

	muteroledict[gid] = roleid

	json_data, err := json.Marshal(muteroledict)
	if err != nil {
		return err
	}

	os.WriteFile("./commands/moderation/mute_roles.json", json_data, fs.ModeAppend)

	return nil
}

func GetRole(s *discordgo.Session, message *discordgo.MessageCreate, args *string) (*discordgo.Role, error) {
	var role *discordgo.Role

	mentions := message.MentionRoles

	if len(mentions) > 0 {
		args = &mentions[0]
	}

	roles, _ := s.GuildRoles(message.GuildID)
	for i := 0; i < len(roles); i++ {
		if *args == roles[i].Name {
			role = roles[i]
		}
	}

	if role == nil {
		return nil, errors.New("could not get Role")
	}
	return role, nil
}

func Mute(s *discordgo.Session, message *discordgo.MessageCreate, time_and_reason *string) {
	hasperms, err := MemberHasPermission(s, message.GuildID, message.Author.ID, message, 8192)
	if hasperms == false {
		if err != nil {
			log.Println(err)
		}
		s.ChannelMessageSend(message.ChannelID, "You do not have the permissions to use this command.")
		return
	}

	user, err := GetUser(s, message, time_and_reason)
	if err != nil {
		log.Println(err)
		return
	}
	role, err := GetRoleFromFile(message.GuildID)
	if err != nil {
		log.Println(err)
		return
	}

	s.GuildMemberRoleAdd(message.GuildID, user.User.ID, *role)

	// TODO: Write the time to unmute to file and have a loop in the on ready function to infinitely check if its time to unmute someone

	// TOOD: Check if a time or reason was given so we can send the right response, and unmute the person at the right time.

}

func MuteRole(s *discordgo.Session, message *discordgo.MessageCreate, role_string *string) {
	role, err := GetRole(s, message, role_string)
	if err != nil {
		s.ChannelMessageSend(message.ChannelID, "Could not get the role from the input name.")
		log.Println(err)
		return
	}
	WriteRolesFile(message.GuildID, role.ID)
	s.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Set the mute role to %v", role.Name))
}
