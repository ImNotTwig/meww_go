package command_parsing

import (
	"encoding/json"
	"io/ioutil"
	"meww_go/config"
	"strings"
	"unicode/utf8"

	"github.com/bwmarrin/discordgo"
)

type ParsedCommand struct {
	Command   string
	Args      string
	IsCommand bool
}

var empty_command = ParsedCommand{
	Command:   "",
	Args:      "",
	IsCommand: false,
}

func ParseCommand(m *discordgo.MessageCreate) *ParsedCommand {

	prefixes_file, _ := ioutil.ReadFile("./command_parsing/server_prefixes.json")
	prefix := config.ReadConfig().Prefix
	var prefixes_map map[string]string
	err := json.Unmarshal(prefixes_file, &prefixes_map)
	if err == nil {
		if value, ok := prefixes_map[m.GuildID]; ok == true {
			prefix = value
		}
	}

	if utf8.RuneCountInString(m.Content) == 1 {
		return &ParsedCommand{Command: "", Args: "", IsCommand: false}
	}

	command_with_args := strings.Split(m.Content, " ")

	command := strings.TrimPrefix(command_with_args[0], prefix)
	args := strings.TrimPrefix(m.Content, command_with_args[0])

	args = strings.TrimSpace(args)
	if strings.HasPrefix(m.Content, prefix) {
		return &ParsedCommand{Command: command, Args: args, IsCommand: true}
	} else {
		return &ParsedCommand{Command: command, Args: args, IsCommand: false}
	}
}

func contains(s []string, i string) bool {
	for _, v := range s {
		if v == i {
			return true
		}
	}

	return false
}

func HandleCommand(s *discordgo.Session, msg *discordgo.MessageCreate, command *ParsedCommand) {

	has_command := false

	for i := 0; i < len(command_list); i++ {
		if contains(command_list[i].names, strings.ToLower(command.Command)) {
			command_list[i].command(s, msg, &command.Args)
			has_command = true
			break
		}
	}
	if has_command == false && command.IsCommand == true {
		s.ChannelMessageSendComplex(
			msg.ChannelID, &discordgo.MessageSend{
				Content:         "The command " + command.Command + " could not be found.",
				Reference:       msg.Reference(),
				AllowedMentions: &discordgo.MessageAllowedMentions{RepliedUser: false},
			},
		)
	}
}