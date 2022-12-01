package command_parsing

import (
	"meww_go/commands/moderation"
	"meww_go/commands/music"
	"meww_go/commands/utility"

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
			HelpMessage:  `Plays a song if nothing is playing, adds a song to the queue if something is, if no song is given will unpause the queue.`,
			Syntax:       `~play <optional_song>`,
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
			Names:        []string{"q", "queue"},
			Command:      music.ShowQueue,
			HelpMessage:  `Sends an embed that shows the queue, every song has the number in queue, and the current song has a -> next to it.`,
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
			HelpMessage:  `Shows the position, title, and current position/duration of the currently playing song.`,
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
			Names:        []string{"unpause", "resume"},
			Command:      music.Unpause,
			HelpMessage:  `Resumes the queue. (This is the same as using ~play without a song argument)`,
			Syntax:       `~resume`,
			CommandGroup: "MusicCommands",
		},
		{
			Names:        []string{"next", "skip", "s"},
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
			Names:        []string{"setprefix", "prefix"},
			Command:      utility.SetPrefix,
			HelpMessage:  `Sets a prefix for this server, if you use this Command again you have to use the new prefix`,
			Syntax:       `~prefix <new prefix>`,
			CommandGroup: "UtilityCommands",
		},
		{
			Names:        []string{"help"},
			Command:      Help,
			HelpMessage:  `Shows the help menu, if you don't input a module or command to get help for, it will show all the modules that you can get help for.`,
			Syntax:       `~help <optional command>`,
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
	}
)

func Help(s *discordgo.Session, m *discordgo.MessageCreate, args *string) {

	switch *args {
	case "":
		// if the user wants all the modules
	case "music":

	case "moderation":

	case "utility":

	}
}
