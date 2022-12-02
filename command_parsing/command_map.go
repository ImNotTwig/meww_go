package command_parsing

import (
	"log"
	"meww_go/commands/fun"
	"meww_go/commands/moderation"
	"meww_go/commands/music"
	"meww_go/commands/utility"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Command struct {
	Names        []string
	Command      func(*discordgo.Session, *discordgo.MessageCreate, *string)
	HelpMessage  string
	Syntax       string
	CommandGroup string
}

var (
	CommandList = []Command{
		//Music commands
		{
			Names:        []string{"play", "p"},
			Command:      music.Play,
			HelpMessage:  `Plays a song if nothing is playing, adds a song to the queue if something is. If no song is given will unpause the queue.`,
			Syntax:       `~play <optional song>`,
			CommandGroup: "MusicCommands",
		},
		{
			Names:        []string{"playnext", "pn"},
			Command:      music.PlayNext,
			HelpMessage:  `will add a song to the next position in queue. Only gets the first song in a playlist or album.`,
			Syntax:       `~playnext <song>`,
			CommandGroup: "MusicCommands",
		},
		{
			Names:        []string{"queue", "q"},
			Command:      music.ShowQueue,
			HelpMessage:  `Sends an embed that shows the songs in the queue with their positions in queue. The current song has a -> next to it.`,
			Syntax:       `~queue`,
			CommandGroup: "MusicCommands",
		},
		{
			Names:        []string{"clear"},
			Command:      music.Clear,
			HelpMessage:  `Clears the queue.`,
			Syntax:       `~clear`,
			CommandGroup: "MusicCommands",
		},
		{
			Names:        []string{"nowplaying", "np"},
			Command:      music.NowPlaying,
			HelpMessage:  `Shows the position, title, and current position in the duration of the currently playing song.`,
			Syntax:       `~nowplaying`,
			CommandGroup: "MusicCommands",
		},
		{
			Names:        []string{"pause"},
			Command:      music.Pause,
			HelpMessage:  `Pauses the queue. (This includes the currently playing song)`,
			Syntax:       `~pause`,
			CommandGroup: "MusicCommands",
		},
		{
			Names:        []string{"resume", "unpause"},
			Command:      music.Unpause,
			HelpMessage:  `Resumes the queue. (This is the same as using ~play without a song argument)`,
			Syntax:       `~resume`,
			CommandGroup: "MusicCommands",
		},
		{
			Names:        []string{"skip", "next", "s"},
			Command:      music.Skip,
			HelpMessage:  `Skips the currently playing song.`,
			Syntax:       `~skip`,
			CommandGroup: "MusicCommands",
		},
		{
			Names:        []string{"goto", "jumpto"},
			Command:      music.GoTo,
			HelpMessage:  `Moves the current position in the queue to the one specfied`,
			Syntax:       `~goto <number in queue>`,
			CommandGroup: "MusicCommands",
		},
		{
			Names:        []string{"leave", "disconnect", "quit", "fuckoff", "stop"},
			Command:      music.FuckOff,
			HelpMessage:  `Makes the bot leave the voice channel and clear the queue.`,
			Syntax:       `~leave`,
			CommandGroup: "MusicCommands",
		},
		{
			Names:        []string{"remove"},
			Command:      music.Remove,
			HelpMessage:  `Removes a song from the queue at the specified position`,
			Syntax:       `~remove <position in queue>`,
			CommandGroup: "MusicCommands",
		},
		{
			Names:        []string{"loop"},
			Command:      music.Loop,
			HelpMessage:  `Turns the loop on, if it is off, as well as vice versa.`,
			Syntax:       `~loop`,
			CommandGroup: "MusicCommands",
		},
		{
			Names:        []string{"shuffle"},
			Command:      music.Shuffle,
			HelpMessage:  `Turns the shuffle on, if it is on, as well as vice versa.`,
			Syntax:       `~shuffle`,
			CommandGroup: "MusicCommands",
		},
		{
			Names:        []string{"lyrics", "lyric"},
			Command:      music.Lyrics,
			HelpMessage:  `Show lyrics of the given song, if no song is given it checks the queue for the currently playing song.`,
			Syntax:       `~lyrics <song>`,
			CommandGroup: "MusicCommands",
		},

		// Utility Commands
		{
			Names:        []string{"prefix", "setprefix"},
			Command:      utility.SetPrefix,
			HelpMessage:  `Sets a prefix for this server, if you use this command again you have to use the new prefix`,
			Syntax:       `~prefix <new prefix>`,
			CommandGroup: "UtilityCommands",
		},

		// Moderation Commands
		{
			Names:        []string{"kick"},
			Command:      moderation.Kick,
			HelpMessage:  `Kicks a user from the server. You can supply a reason of why the user was kicked, but it is optional.`,
			Syntax:       `~kick <user> <reason>`,
			CommandGroup: "ModerationCommands",
		},
		{
			Names:        []string{"ban"},
			Command:      moderation.Ban,
			HelpMessage:  `Bans a user from the server. You can supply a reason of why the user was banned, but it is optional.`,
			Syntax:       `~ban <user> <reason>`,
			CommandGroup: "ModerationCommands",
		},
		{
			Names:        []string{"unban"},
			Command:      moderation.UnBan,
			HelpMessage:  `Unbans a user from the server.`,
			Syntax:       `~unban <user>`,
			CommandGroup: "ModerationCommands",
		},

		// Fun Commands
		{
			Names:        []string{"cat"},
			Command:      fun.Cats,
			HelpMessage:  `Shows a picture of a cat.`,
			Syntax:       `~cat`,
			CommandGroup: "FunCommands",
		},
		{
			Names:        []string{"catbomb"},
			Command:      fun.CatBomb,
			HelpMessage:  `Shows 10 pictures of cats.`,
			Syntax:       `~catbomb`,
			CommandGroup: "FunCommands",
		},
	}
)

// this command has to be here because it would cause an import loop if I didnt put it in this file
// I might rework the command handler to make it where I can put this command in the UtilityCommands group
func Help(s *discordgo.Session, m *discordgo.MessageCreate, args *string) {

	embed := &discordgo.MessageEmbed{
		Title: "Help",
	}

	var command *Command
	for i := 0; i < len(CommandList); i++ {
		if contains(CommandList[i].Names, *args) {
			command = &CommandList[i]
			break
		}
	}

	if command != nil {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  *args,
			Value: command.HelpMessage,
		})
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "Use",
			Value: command.Syntax,
		})
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "Names",
			Value: strings.Join(command.Names, ", "),
		})
	} else {
		var musiccommands []string
		var utilcommands []string
		var moderationcommands []string
		var funcommands []string

		for i := 0; i < len(CommandList); i++ {
			command := CommandList[i]
			if command.CommandGroup == "ModerationCommands" {
				moderationcommands = append(moderationcommands, command.Names[0])
			} else if command.CommandGroup == "MusicCommands" {
				musiccommands = append(musiccommands, command.Names[0])
			} else if command.CommandGroup == "UtilityCommands" {
				utilcommands = append(utilcommands, command.Names[0])
			} else if command.CommandGroup == "FunCommands" {
				funcommands = append(funcommands, command.Names[0])
			}
		}

		moderationcommandsstring := strings.Join(moderationcommands, " ")

		musiccommandsstring := strings.Join(musiccommands, " ")

		utilcommandsstring := strings.Join(utilcommands, " ")

		funcommandsstring := strings.Join(funcommands, " ")

		embed.Fields = []*discordgo.MessageEmbedField{
			{
				Name:   "Moderation Commands",
				Value:  moderationcommandsstring,
				Inline: false,
			},
			{
				Name:   "Music Commands",
				Value:  musiccommandsstring,
				Inline: false,
			},
			{
				Name:   "Utility Commands",
				Value:  utilcommandsstring,
				Inline: false,
			},
			{
				Name:   "Fun Commands",
				Value:  funcommandsstring,
				Inline: false,
			},
		}

	}

	message := discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{embed},
	}

	_, err := s.ChannelMessageSendComplex(m.ChannelID, &message)
	if err != nil {
		log.Println(err)
	}
}
