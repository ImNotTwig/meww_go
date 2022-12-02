package fun

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

func Cats(s *discordgo.Session, m *discordgo.MessageCreate, args *string) {
	res, err := http.Get("https://api.thecatapi.com/v1/images/search")
	if err != nil {
		log.Println(err)
		return
	}
	var json_map []map[string]interface{}

	err = json.NewDecoder(res.Body).Decode(&json_map)
	if err != nil {
		log.Println(err)
		return
	}
	var cat_url string
	if _, ok := json_map[0]["url"]; ok {
		cat_url = json_map[0]["url"].(string)
	} else {
		return
	}

	embed := discordgo.MessageEmbed{
		Image: &discordgo.MessageEmbedImage{URL: cat_url},
	}

	s.ChannelMessageSendEmbed(m.ChannelID, &embed)
}

func CatBomb(s *discordgo.Session, m *discordgo.MessageCreate, args *string) {
	res, err := http.Get("https://api.thecatapi.com/v1/images/search?limit=10")
	if err != nil {
		log.Println(err)
		return
	}
	var json_map []map[string]interface{}

	err = json.NewDecoder(res.Body).Decode(&json_map)
	if err != nil {
		log.Println(err)
		return
	}
	if _, ok := json_map[0]["url"]; ok {
	} else {
		return
	}

	var embed_list []*discordgo.MessageEmbed

	for i := 0; i < len(json_map); i++ {
		if _, ok := json_map[i]["url"]; ok {
			embed_list = append(embed_list, &discordgo.MessageEmbed{
				Image: &discordgo.MessageEmbedImage{URL: json_map[i]["url"].(string)},
			})
		}
	}

	message := &discordgo.MessageSend{
		Embeds: embed_list,
	}

	s.ChannelMessageSendComplex(m.ChannelID, message)
}
