package main

import (
	"fmt"
	"meww_go/command_parsing"
	"meww_go/commands/music"
	"meww_go/config"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var println = fmt.Println
var bot_config = config.ReadConfig()

func main() {
	discordgo.Logger = nil
	dg, err := discordgo.New("Bot " + bot_config.Tokens.Discord_token)

	if err != nil {
		panic(err)
	}

	dg.Identify.Intents = discordgo.IntentsAll

	dg.AddHandler(messageCreate)
	dg.AddHandler(MemberJoin)

	dg.AddHandler(func(s *discordgo.Session, evt *discordgo.Ready) {
		println("Logged in as " + s.State.User.Username + "#" + s.State.User.Discriminator)
		dg.UpdateListeningStatus("~help")
	})

	guild_timer_map := make(map[string]int)
	dg.AddHandler(func(s *discordgo.Session, evt *discordgo.VoiceStateUpdate) {
		if _, ok := guild_timer_map[evt.GuildID]; !ok {
			guild_timer_map[evt.GuildID] = 0
		}
		guild, _ := dg.State.Guild(evt.GuildID)
		in_vc := false
		for i := 0; i < len(guild.VoiceStates); i++ {
			if guild.VoiceStates[i].UserID == dg.State.User.ID {
				in_vc = true
				break
			}
		}
		if in_vc == false {
			return
		}
		for guild_timer_map[evt.GuildID] = 0; len(guild.VoiceStates) < 2; guild_timer_map[evt.GuildID] += 1 {
			time.Sleep(time.Duration(1 * time.Second))
			if guild_timer_map[evt.GuildID] == 30 && len(guild.VoiceStates) < 2 {
				if music.QueueDict[evt.GuildID].Voice != nil {
					music.QueueDict[evt.GuildID].FuckOff()
				}
				break
			}
			if len(guild.VoiceStates) >= 2 {
				guild_timer_map[evt.GuildID] = 0
				return
			}
		}
	})

	err = dg.Open()
	if err != nil {
		panic(err)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	if dg.VoiceConnections != nil {
		for _, val := range dg.VoiceConnections {
			if val != nil {
				val.Disconnect()
			}
		}
	}
	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == s.State.User.ID {
		return
	}

	parsed_command := command_parsing.ParseCommand(m)
	command_parsing.HandleCommand(s, m, parsed_command)
}

func MemberJoin(s *discordgo.Session, mj *discordgo.GuildMemberAdd) {

}
