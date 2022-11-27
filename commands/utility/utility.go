package utility

import (
	"encoding/json"
	"io/fs"
	"io/ioutil"

	"github.com/bwmarrin/discordgo"
)

func SetPrefix(s *discordgo.Session, m *discordgo.MessageCreate, arg *string) {
	file_bytes, err := ioutil.ReadFile("./command_parsing/server_prefixes.json")

	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not get the server_prefixes.json file.")
		return
	}
	var prefix_map map[string]string
	err = json.Unmarshal(file_bytes, &prefix_map)

	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not get the server_prefixes.json file.")
		return
	}

	prefix_map[m.GuildID] = *arg

	json_data, err := json.Marshal(prefix_map)

	ioutil.WriteFile("./command_parsing/server_prefixes.json", json_data, fs.ModeAppend)
}
