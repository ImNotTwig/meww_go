package music

import (
	"context"
	"fmt"
	"log"
	"meww_go/config"
	"meww_go/queue"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/raitonoberu/lyricsapi/lyrics"
	soundcloudapi "github.com/zackradisic/soundcloud-api"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

// spotify
var SpotifyConfig = &clientcredentials.Config{
	ClientID:     config.ReadConfig().Tokens.Spotify_id,
	ClientSecret: config.ReadConfig().Tokens.Spotify_secret,
	TokenURL:     spotifyauth.TokenURL,
}
var token, _ = SpotifyConfig.Token(context.Background())
var http_client = spotifyauth.New().Client(context.Background(), token)
var Spotify = spotify.New(http_client)

// soundcloud
var SoundCloud, _ = soundcloudapi.New(soundcloudapi.APIOptions{})

// lyrics api
var lyricsapi = lyrics.NewLyricsApi("sp_t=f99b345620f5ddbdf4c76c5a667f9e44; sp_m=us; OptanonConsent=isIABGlobal=false&datestamp=Sat+Nov+26+2022+22%3A19%3A25+GMT-0600+(Central+Standard+Time)&version=6.26.0&hosts=&landingPath=NotLandingPage&groups=s00%3A1%2Cf00%3A1%2Cm00%3A1%2Ct00%3A1%2Ci00%3A1%2Cf02%3A1%2Cm02%3A1%2Ct02%3A1&AwaitingReconsent=false; sp_pfhp=2c2ccb58-8a92-4713-a1c0-8b43b3090b49; sp_dc=AQA82HIDVRIJ1L9uCj5_bVlTK2_tsPy9JlB9QlALqRDtjjeiAZBuFRqlqskDvUTKLX5ehsvGHFwSo-oAr7QqCHfzftgc6Ve5k4s9c5X-ODdF3ft_9G6K7zuTpGdqMSCMFcqdghm-EBcMTAel8P6_SnzDudKWwzFp; sp_key=56842b5c-3732-4ec8-bae0-62c3f0d9ab77; sp_landing=https%3A%2F%2Fopen.spotify.com%2F%3Fsp_cid%3Df99b345620f5ddbdf4c76c5a667f9e44%26device%3Ddesktop; OptanonAlertBoxClosed=2022-11-27T04:19:25.186Z")

// Queue Dict
var QueueDict = make(map[string]*queue.Queue)

func GetNonYTSongs(s *discordgo.Session, message *discordgo.MessageCreate, song *string) []queue.Song {
	server_queue := QueueDict[message.GuildID]

	var track_list []queue.Song

	// checking if the song is a spotify link, if so only get the id
	// and assign that to a new variable
	var new_song string
	if strings.HasPrefix(*song, "https://open.spotify.com") {
		track_words := strings.Split(*song, "/")
		new_song = track_words[len(track_words)-1]
	}

	if strings.HasPrefix(*song, "https://open.spotify.com/album") {
		tracks, err := Spotify.GetAlbum(context.Background(), spotify.ID(new_song))
		if err != nil {
			log.Println(err)
			return nil
		}

		current_pos := server_queue.Current_Pos
		for i := 0; i < len(tracks.Tracks.Tracks); i++ {
			duration := strconv.Itoa(tracks.Tracks.Tracks[i].Duration / 1000)
			track := queue.Song{
				Title:    fmt.Sprintf("%v - %v", tracks.Tracks.Tracks[i].Name, tracks.Artists[0].Name),
				Duration: &duration,
				Pos:      current_pos + 1,
			}
			current_pos++
			track_list = append(track_list, track)
		}
		return track_list

	} else if strings.HasPrefix(*song, "https://open.spotify.com/playlist") {
		tracks, err := Spotify.GetPlaylistItems(context.Background(), spotify.ID(new_song))
		if err != nil {
			log.Println(err)
			return nil
		}

		current_pos := server_queue.Current_Pos
		for i := 0; i < len(tracks.Items); i++ {
			duration := strconv.Itoa(tracks.Items[i].Track.Track.Duration / 1000)
			track := queue.Song{
				Title:    fmt.Sprintf("%v - %v", tracks.Items[i].Track.Track.Name, tracks.Items[i].Track.Track.Artists[0].Name),
				Duration: &duration,
				Pos:      current_pos + 1,
			}
			current_pos++
			track_list = append(track_list, track)
		}
		return track_list

	} else if strings.HasPrefix(*song, "https://open.spotify.com/track") {
		tracks, err := Spotify.GetTrack(context.Background(), spotify.ID(new_song))
		current_pos := server_queue.Current_Pos
		if err != nil {
			log.Println(err)
			return nil
		}

		duration := strconv.Itoa(tracks.Duration / 1000)
		track := queue.Song{
			Title:    fmt.Sprintf("%v - %v", tracks.Name, tracks.Artists[0].Name),
			Duration: &duration,
			Pos:      current_pos + 1,
		}
		current_pos++
		track_list = append(track_list, track)
		return track_list

	} else if strings.HasPrefix(*song, "https://soundcloud") {
		split_url := strings.Split(*song, "/")
		if split_url[4] == "sets" {
			playlist_info, err := SoundCloud.GetPlaylistInfo(*song)
			if err != nil {
				log.Println(err)
				return nil
			}
			tracks := playlist_info.Tracks
			current_pos := server_queue.Current_Pos
			for i := 0; i < len(tracks); i++ {
				duration := fmt.Sprintf("%v", tracks[i].DurationMS/1000)
				track := queue.Song{
					Title:    tracks[i].Title,
					Url:      &tracks[i].PermalinkURL,
					Pos:      current_pos + 1,
					Duration: &duration,
				}
				current_pos++
				track_list = append(track_list, track)
			}

		}
		return track_list
	} else {
		return nil
	}
}

func GetVoice(s *discordgo.Session, message *discordgo.MessageCreate) (*discordgo.VoiceConnection, error) {

	guild, err := s.State.Guild(message.GuildID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var voice_channel string
	for i := 0; i < len(guild.VoiceStates); i++ {
		if guild.VoiceStates[i].UserID == message.Author.ID {
			voice_channel = guild.VoiceStates[i].ChannelID
			break
		}
	}

	voice, err := s.ChannelVoiceJoin(message.GuildID, voice_channel, false, true)
	if err == nil {
		return voice, nil
	} else {
		return nil, err
	}

}

// PlayNext function, this function adds one song to the next position in the queue
// This command will not work if shuffle is enabled in the queue
func PlayNext(s *discordgo.Session, message *discordgo.MessageCreate, song *string) {

	// checking if this server doesnt have a queue yet.
	// if they don't then make one
	if _, ok := QueueDict[message.GuildID]; !ok {
		QueueDict[message.GuildID] = queue.NewQueue(s, message)
	}
	server_queue := QueueDict[message.GuildID]

	// if a song was not given, try to resume the queue.
	if *song == "" || song == nil {
		s.ChannelMessageSend(message.ChannelID, "You cannot play nothing next.")
		return
	}

	// checking if the queue for this server has shuffle enabled
	if server_queue.Shuffle {
		s.ChannelMessageSend(message.ChannelID, "You can't use play next when shuffle is enabled.")
		return
	}

	// checking if the queue is paused, if so tell the user that they should just use the Play command.
	if !server_queue.Paused {

		// getting the voice connection
		voice, err := GetVoice(s, message)
		if err == nil {
			server_queue.SetVoice(voice)
		} else if voice != server_queue.Voice {
			s.ChannelMessageSend(message.ChannelID, "The bot is already in another voice channel.")
			return
		} else if err != nil && voice == nil {
			s.ChannelMessageSend(message.ChannelID, "There was an error while connecting to the voice channel.")
			return
		}

		track_list := GetNonYTSongs(s, message, song)
		if track_list != nil {
			server_queue.Insert(track_list[0], server_queue.Current_Pos+1)
		} else {
			song_list, err := queue.YtSearch(*song, *server_queue)
			if err != nil {
				log.Println(err)
				s.ChannelMessageSend(message.ChannelID, "Could not get song(s) from given query.")
				return
			}
			server_queue.Insert(song_list[0], server_queue.Current_Pos+1)
		}
		server_queue.ValidateSongPos()
		server_queue.SetMessageChannel(message)
	} else if server_queue.Paused {
		s.ChannelMessageSend(message.ChannelID, "The queue is paused, use the play command not playnext.")
	}
}

// Play command, this command will play a song, and if a playlist is inputted will add the rest of the songs to the queue
// if a song is already playing, this function will add the inputted songs to the queue instead
func Play(s *discordgo.Session, message *discordgo.MessageCreate, song *string) {

	// checking if this server doesnt have a queue yet.
	// if they don't then make one
	if _, ok := QueueDict[message.GuildID]; !ok {
		QueueDict[message.GuildID] = queue.NewQueue(s, message)
	}
	server_queue := QueueDict[message.GuildID]

	// if a song was not given, try to resume the queue.
	if *song == "" || song == nil {
		if server_queue.Paused {
			if server_queue.DoesStreamExist() {
				server_queue.Unpause()
			}
		}
		return
	}

	// get a voice connection, and if we already have one compare the two
	// if it's different tell the user that they cant do this because the bot is connected elsewhere.
	voice, err := GetVoice(s, message)
	if err == nil {
		server_queue.SetVoice(voice)
	} else if voice != server_queue.Voice {
		s.ChannelMessageSend(message.ChannelID, "The bot is already in another voice channel.")
		return
	} else if err != nil && voice == nil {
		s.ChannelMessageSend(message.ChannelID, "There was an error while connecting to the voice channel.")
		return
	}

	// seeing if we get any songs from querying the non-yt song searcher (spotify and soundcloud)
	// if we dont then just search yt-dlp
	track_list := GetNonYTSongs(s, message, song)
	if track_list == nil {
		var song_list []queue.Song
		if strings.HasPrefix(*song, "https://open.spotify") || strings.HasPrefix(*song, "https://soundcloud") {
			s.ChannelMessageSend(message.ChannelID, "Could not get song(s) from given query.")
			return
		}
		song_list, err = queue.YtSearch(*song, *server_queue)
		if err != nil {
			log.Println(err)
			s.ChannelMessageSend(message.ChannelID, "Could not get song(s) from given query.")
			return
		}

		for i := 0; i < len(song_list); i++ {
			server_queue.Enqueue(song_list[i])
		}
	} else {
		for i := 0; i < len(track_list); i++ {
			server_queue.Enqueue(track_list[i])
		}
	}

	server_queue.SetMessageChannel(message)
	go server_queue.CheckPlaying()
}

// Pause function, this pauses the queue, and the currently playing song.
func Pause(s *discordgo.Session, message *discordgo.MessageCreate, args *string) {
	if _, ok := QueueDict[message.GuildID]; ok {
		server_queue := QueueDict[message.GuildID]
		server_queue.Pause()
		server_queue.SetMessageChannel(message)
	}
}

// Unpause function, this unpauses the queue and the current song.
func Unpause(s *discordgo.Session, message *discordgo.MessageCreate, args *string) {
	if _, ok := QueueDict[message.GuildID]; ok {
		server_queue := QueueDict[message.GuildID]
		server_queue.Unpause()
		server_queue.SetMessageChannel(message)
	}
}

// Skip function, this skips the currently playing song.
// if you skip on the last song, it will pause and go to the last song and when you play a new song will skip to that one.
func Skip(s *discordgo.Session, message *discordgo.MessageCreate, args *string) {
	if _, ok := QueueDict[message.GuildID]; ok {
		server_queue := QueueDict[message.GuildID]
		server_queue.Skip()
		server_queue.SetMessageChannel(message)
	}
}

// ShowQueue function, this function will send an embed to discord that shows the current queue, with track positions.
// The current song will have a -> next to it
func ShowQueue(s *discordgo.Session, message *discordgo.MessageCreate, args *string) {
	if _, ok := QueueDict[message.GuildID]; ok {
		server_queue := QueueDict[message.GuildID]

		title_list := make([]string, 0)

		songs := server_queue.Songs

		for i := 0; i < len(songs); i++ {
			if songs[i].Pos == server_queue.Current_Pos {
				title_list = append(title_list, "-> "+strconv.Itoa(songs[i].Pos)+" - "+songs[i].Title)
			} else {
				title_list = append(title_list, strconv.Itoa(songs[i].Pos)+" - "+songs[i].Title)
			}
		}

		embed_desc := strings.Join(title_list, "\n")

		embed_to_send := discordgo.MessageEmbed{
			Title:       "The Queue",
			Description: embed_desc,
		}
		server_queue.SetMessageChannel(message)
		s.ChannelMessageSendEmbed(message.ChannelID, &embed_to_send)
	}
}

// Loop function, this function will enable the loop if it is disabled, and vice versa.
func Loop(s *discordgo.Session, message *discordgo.MessageCreate, args *string) {
	if _, ok := QueueDict[message.GuildID]; ok {
		server_queue := QueueDict[message.GuildID]
		server_queue.SetLoop()
		server_queue.SetMessageChannel(message)
	}
}

// Shuffle function, this function will enable shuffle if it is disabled, and vice versa.
func Shuffle(s *discordgo.Session, message *discordgo.MessageCreate, args *string) {
	if _, ok := QueueDict[message.GuildID]; ok {
		server_queue := QueueDict[message.GuildID]
		server_queue.SetShuffle()
		server_queue.SetMessageChannel(message)
	}
}

// Clear function, this function will clear the queue, and stop the currently playing song.
func Clear(s *discordgo.Session, message *discordgo.MessageCreate, args *string) {
	if _, ok := QueueDict[message.GuildID]; ok {
		server_queue := QueueDict[message.GuildID]
		server_queue.Clear()
		server_queue.SetMessageChannel(message)
	}
}

// Goto function, this function will move the current position to the one specified and start playing that song.
func GoTo(s *discordgo.Session, message *discordgo.MessageCreate, arg *string) {
	if _, ok := QueueDict[message.GuildID]; ok {
		server_queue := QueueDict[message.GuildID]
		int, err := strconv.Atoi(*arg)
		if err != nil {
			s.ChannelMessageSend(message.ChannelID, "You did not pass a valid integer.")
			return
		}
		if int > server_queue.Len() {
			int = server_queue.Len()
		}
		server_queue.MoveTo(int)
		server_queue.SetMessageChannel(message)
	}
}

// FuckOff function, this function will make the bot leave the voice channel and delete the queue.
func FuckOff(s *discordgo.Session, message *discordgo.MessageCreate, args *string) {
	if _, ok := QueueDict[message.GuildID]; ok {
		server_queue := QueueDict[message.GuildID]
		server_queue.FuckOff()
		delete(QueueDict, message.GuildID)
		server_queue.SetMessageChannel(message)
	}
}

// Remove function, this function will remove the song specified
// this function cannot remove the currently playing song.
func Remove(s *discordgo.Session, message *discordgo.MessageCreate, arg *string) {
	if _, ok := QueueDict[message.GuildID]; ok {
		server_queue := QueueDict[message.GuildID]
		int, err := strconv.Atoi(*arg)
		if err != nil {
			s.ChannelMessageSend(message.ChannelID, "You did not pass a valid integer.")
			return
		}
		if int > server_queue.Len() {
			s.ChannelMessageSend(message.ChannelID, "You cannot remove a song thats not in the queue.")
			return
		}
		server_queue.Remove(int)
		server_queue.SetMessageChannel(message)
	}
}

// NowPlaying function, this function will send a message to discord showing the currently playing song, and the time played / duration
func NowPlaying(s *discordgo.Session, message *discordgo.MessageCreate, args *string) {
	if _, ok := QueueDict[message.GuildID]; ok {
		server_queue := QueueDict[message.GuildID]
		current_time := server_queue.CurrentTime()
		total_time := server_queue.TotalTime()

		embed_to_send := discordgo.MessageEmbed{
			Title:       (strconv.Itoa(server_queue.Current().Pos) + " - " + server_queue.Current().Title),
			Description: current_time + "/" + total_time,
		}

		s.ChannelMessageSendEmbed(message.ChannelID, &embed_to_send)
		server_queue.SetMessageChannel(message)

	}
}

// Lyrics function, this function simply shows the lyrics of whatever song you gave it
func Lyrics(s *discordgo.Session, message *discordgo.MessageCreate, args *string) {
	if args != nil && strings.TrimSpace(*args) != "" {
		song_lyrics, err := lyricsapi.GetByName(*args)
		if err != nil {
			fmt.Println(err)
			s.ChannelMessageSend(message.ChannelID, "Could not get the lyrics for the given song.")
			return
		}
		var lyrics_list []string
		for i := 0; i < len(song_lyrics.Lyrics.Lines); i++ {
			lyrics_list = append(lyrics_list, song_lyrics.Lyrics.Lines[i].Words)
		}
		embed_desc := strings.Join(lyrics_list, "\n")

		embed_to_send := discordgo.MessageEmbed{
			Title:       fmt.Sprintf("Lyrics for '%v'", *args),
			Description: embed_desc,
		}

		s.ChannelMessageSendEmbed(message.ChannelID, &embed_to_send)
	} else {
		s.ChannelMessageSend(message.ChannelID, "You need to input a song to query.")
	}
}
