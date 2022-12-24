package moderation

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// this function just gets a user from a server,
// if you supplied a message without a mention it will find a member with that username
func GetUser(s *discordgo.Session, message *discordgo.MessageCreate, args *string) (*discordgo.Member, error) {
	guild, err := s.State.Guild(message.GuildID)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var user *discordgo.Member
	seperated_message := strings.Split(*args, " ")

	users, _ := s.GuildMembersSearch(message.GuildID, seperated_message[0], 1)
	if users != nil {
		user = users[0]
	}

	if user == nil {
		if len(message.Mentions) != 0 {
			member_id := message.Mentions[0].ID
			for i := 0; i < len(guild.Members); i++ {
				if member_id == guild.Members[i].User.ID {
					user = guild.Members[i]
				}
			}
		}
	}

	if user == nil {
		return nil, errors.New("could not get User")
	}
	return user, nil
}

// this function checks if a member has certain permissions
func MemberHasPermission(s *discordgo.Session, guildID string, userID string, m *discordgo.MessageCreate, permission int) (bool, error) {
	perms, err := s.UserChannelPermissions(userID, m.ChannelID)
	if err != nil {
		return false, err
	}
	if perms&int64(permission) != 0 {
		return true, nil
	}

	return false, nil
}

func Kick(s *discordgo.Session, message *discordgo.MessageCreate, reason *string) {
	hasperms, err := MemberHasPermission(s, message.GuildID, message.Author.ID, message, 2)
	if hasperms == false {
		if err != nil {
			log.Println(err)
		}
		s.ChannelMessageSend(message.ChannelID, "You do not have the permissions to use this command.")
		return
	}

	kicked_user, err := GetUser(s, message, reason)
	if err != nil {
		log.Println(err)
		s.ChannelMessageSend(message.ChannelID, "Could not get user, please @ them if you haven't done so.")
		return
	}

	new_reason := strings.Replace(*reason, kicked_user.User.Username, "", -1)
	new_reason = strings.Replace(new_reason, kicked_user.Nick, "", -1)
	new_reason = strings.TrimSpace(new_reason)

	if new_reason == "" {
		err = s.GuildMemberDelete(message.GuildID, kicked_user.User.ID)
	} else {
		err = s.GuildMemberDeleteWithReason(message.GuildID, kicked_user.User.ID, new_reason)
	}
	if err != nil {
		log.Println(err)
		s.ChannelMessageSend(message.ChannelID, "Could not kick user, or there was some other type of error.")
	} else {
		if new_reason == "" {
			s.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Kicked %v.", kicked_user.User.Username))
		} else {
			s.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Kicked %v for %v", kicked_user.User.Username, new_reason))
		}
	}
}

func Ban(s *discordgo.Session, message *discordgo.MessageCreate, reason *string) {
	hasperms, err := MemberHasPermission(s, message.GuildID, message.Author.ID, message, 2)
	if hasperms == false {
		if err != nil {
			log.Println(err)
		}
		s.ChannelMessageSend(message.ChannelID, "You do not have the permissions to use this command.")
		return
	}

	banned_user, err := GetUser(s, message, reason)
	if err != nil {
		log.Println(err)
		s.ChannelMessageSend(message.ChannelID, "Could not get user, please @ them if you haven't done so.")
		return
	}

	new_reason := strings.Replace(*reason, banned_user.User.Username, "", -1)
	new_reason = strings.Replace(new_reason, banned_user.Nick, "", -1)
	new_reason = strings.TrimSpace(new_reason)

	if new_reason == "" {
		err = s.GuildBanCreate(message.GuildID, banned_user.User.ID, 0)
	} else {
		err = s.GuildBanCreateWithReason(message.GuildID, banned_user.User.ID, new_reason, 0)
	}
	if err != nil {
		log.Println(err)
		s.ChannelMessageSend(message.ChannelID, "Could not ban user, or there was some other type of error.")
	} else {
		if new_reason == "" {
			s.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Banned %v.", banned_user.User.Username))
		} else {
			s.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Banned %v for %v", banned_user.User.Username, new_reason))
		}
	}
}

func UnBan(s *discordgo.Session, message *discordgo.MessageCreate, reason *string) {
	hasperms, err := MemberHasPermission(s, message.GuildID, message.Author.ID, message, 2)
	if hasperms == false {
		if err != nil {
			log.Println(err)
		}
		s.ChannelMessageSend(message.ChannelID, "You do not have the permissions to use this command.")
		return
	}

	guild_bans, err := s.GuildBans(message.GuildID, 1000, "", "")

	if err != nil {
		log.Println(err)
		return
	}
	var unbanned_user *discordgo.User
	seperated_message := strings.Split(*reason, " ")
	for i := 0; i < len(guild_bans); i++ {
		if seperated_message[0] == guild_bans[i].User.Username || seperated_message[0] == guild_bans[i].User.ID {
			unbanned_user = guild_bans[i].User
		}
	}

	err = s.GuildBanDelete(message.GuildID, unbanned_user.ID)
	if err != nil {
		log.Println(err)
		s.ChannelMessageSend(message.ChannelID, "Could not unban user, or there was some other type of error.")
		return
	}
	s.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Unbanned %v.", unbanned_user.Username))
}
