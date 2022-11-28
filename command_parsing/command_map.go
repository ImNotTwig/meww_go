package command_parsing

import (
	"meww_go/commands/moderation"
	"meww_go/commands/music"
	"meww_go/commands/utility"

	"github.com/bwmarrin/discordgo"
)

type Command struct {
	names         []string
	command       func(*discordgo.Session, *discordgo.MessageCreate, *string)
	help_message  string
	command_group string
}

var (
	command_list = []Command{
		//Music commands
		{
			names:   []string{"play", "p"},
			command: music.Play,
			help_message: `Plays a song if nothing is playing, adds a song to the queue if something is, if no song is given will unpause the queue.
syntax: ~play <optional_song>`,
			command_group: "MusicCommands",
		},
		{
			names:   []string{"playnext", "pn"},
			command: music.PlayNext,
			help_message: `will add a song to the next position in queue. Only gets the first song in a playlist or album.
syntax: ~playnext <song>`,
			command_group: "MusicCommands",
		},
		{
			names:   []string{"q", "queue"},
			command: music.ShowQueue,
			help_message: `Sends an embed that shows the queue, every song has the number in queue, and the current song has a -> next to it.
syntax: ~queue`,
			command_group: "MusicCommands",
		},
		{
			names:   []string{"clear"},
			command: music.Clear,
			help_message: `Clears the queue.
syntax: ~clear`,
			command_group: "MusicCommands",
		},
		{
			names:   []string{"nowplaying", "np"},
			command: music.NowPlaying,
			help_message: `Shows the position, title, and current position/duration of the currently playing song.
syntax: ~nowplaying`,
			command_group: "MusicCommands",
		},
		{
			names:   []string{"pause"},
			command: music.Pause,
			help_message: `Pauses the queue. (This includes the currently playing song)
syntax: ~pause`,
			command_group: "MusicCommands",
		},
		{
			names:   []string{"unpause", "resume"},
			command: music.Unpause,
			help_message: `Resumes the queue. (This is the same as using ~play without a song argument)
syntax: ~resume`,
			command_group: "MusicCommands",
		},
		{
			names:   []string{"next", "skip", "s"},
			command: music.Skip,
			help_message: `Skips the currently playing song.
syntax: ~skip`,
			command_group: "MusicCommands",
		},
		{
			names:   []string{"goto", "jumpto"},
			command: music.GoTo,
			help_message: `Moves the current position in the queue to the one specfied
syntax: ~goto <number in queue>`,
			command_group: "MusicCommands",
		},
		{
			names:   []string{"leave", "disconnect", "quit", "fuckoff", "stop"},
			command: music.FuckOff,
			help_message: `Makes the bot leave the voice channel and clear the queue.
syntax: ~leave`,
			command_group: "MusicCommands",
		},
		{
			names:   []string{"remove"},
			command: music.Remove,
			help_message: `Removes a song from the queue at the specified position
syntax: ~remove <position in queue>`,
			command_group: "MusicCommands",
		},
		{
			names:   []string{"loop"},
			command: music.Loop,
			help_message: `Turns the loop on, if it is off, as well as vice versa.
syntax: ~loop`,
			command_group: "MusicCommands",
		},
		{
			names:   []string{"shuffle"},
			command: music.Shuffle,
			help_message: `Turns the shuffle on, if it is on, as well as vice versa.
syntax: ~shuffle`,
			command_group: "MusicCommands",
		},
		{
			names:   []string{"lyrics", "lyric"},
			command: music.Lyrics,
			help_message: `Show lyrics of the given song, if no song is given it checks the queue for the currently playing song.
syntax: ~lyrics <song>`,
			command_group: "MusicCommands",
		},

		// Utility Commands
		{
			names:   []string{"setprefix", "prefix"},
			command: utility.SetPrefix,
			help_message: `Sets a prefix for this server, if you use this command again you have to use the new prefix
syntax: ~prefix <new prefix>`,
			command_group: "Utility",
		},

		// Moderation Commands
		{
			names:   []string{"kick"},
			command: moderation.Kick,
			help_message: `Kicks a user from the server. You can supply a reason of why the user was kicked, but it is optional.
syntax: ~kick <user> <reason>`,
			command_group: "Moderation",
		},
		{
			names:   []string{"ban"},
			command: moderation.Ban,
			help_message: `Bans a user from the server. You can supply a reason of why the user was banned, but it is optional.
syntax: ~ban <user> <reason>`,
			command_group: "Moderation",
		},
		{
			names:   []string{"unban"},
			command: moderation.UnBan,
			help_message: `Unbans a user from the server.
syntax: ~unban <user>`,
			command_group: "Moderation",
		},
	}
)
