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
	"github.com/zackradisic/soundcloud-api"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

var SS = &clientcredentials.Config{
	ClientID:     config.ReadConfig().Tokens.Spotify_id,
	ClientSecret: config.ReadConfig().Tokens.Spotify_secret,
	TokenURL:     spotifyauth.TokenURL,
}
var SoundCloud, _ = soundcloudapi.New(soundcloudapi.APIOptions{})

var QueueDict = make(map[string]*queue.Queue)

func PlayNext(s *discordgo.Session, message *discordgo.MessageCreate, song *string) {
	if _, ok := QueueDict[message.GuildID]; ok {
		guild, err := s.State.Guild(message.GuildID)

		if err != nil {
			log.Println(err)
		}

		var voice_channel string

		for i := 0; i < len(guild.VoiceStates); i++ {
			if guild.VoiceStates[i].UserID == message.Author.ID {
				voice_channel = guild.VoiceStates[i].ChannelID
				break
			}
		}

		voice, err := s.ChannelVoiceJoin(message.GuildID, voice_channel, false, true)

		if QueueDict[message.GuildID].GetVoice() == nil {
			QueueDict[message.GuildID].SetVoice(voice)
		}

		if err != nil {
			s.ChannelMessageSend(message.ChannelID, "There was an error while connecting to the voice channel.")
			return
		}

		song_list, err := queue.YtSearch(*song, *QueueDict[message.GuildID])
		if err != nil {
			log.Println(err)
			s.ChannelMessageSend(message.ChannelID, "Could not get song(s) from given query.")
			return
		}
		QueueDict[message.GuildID].Insert(song_list[0], QueueDict[message.GuildID].CurrentPos()+1)
		QueueDict[message.GuildID].ValidateSongPos()
	}
}

func Play(s *discordgo.Session, message *discordgo.MessageCreate, song *string) {

	if _, ok := QueueDict[message.GuildID]; !ok {
		QueueDict[message.GuildID] = queue.NewQueue(s, message)
	}

	if strings.TrimSpace(*song) == "" {
		if QueueDict[message.GuildID].Paused() == true {
			if QueueDict[message.GuildID].DoesStreamExist() {
				QueueDict[message.GuildID].Unpause()
			}
		}
		return
	}

	guild, err := s.State.Guild(message.GuildID)

	if err != nil {
		log.Println(err)
	}

	var voice_channel string

	for i := 0; i < len(guild.VoiceStates); i++ {
		if guild.VoiceStates[i].UserID == message.Author.ID {
			voice_channel = guild.VoiceStates[i].ChannelID
			break
		}
	}

	voice, err := s.ChannelVoiceJoin(message.GuildID, voice_channel, false, true)
	if err != nil {
		if _, ok := s.VoiceConnections[message.GuildID]; ok {
			voice = s.VoiceConnections[message.GuildID]
		}
	}

	if QueueDict[message.GuildID].GetVoice() == nil {
		QueueDict[message.GuildID].SetVoice(voice)
	}

	if err != nil {
		s.ChannelMessageSend(message.ChannelID, "There was an error while connecting to the voice channel.")
		return
	}
	token, _ := SS.Token(context.Background())
	http_client := spotifyauth.New().Client(context.Background(), token)
	client := spotify.New(http_client)
	var track_list []queue.Song

	track_words := strings.Split(*song, "/")
	var new_song string
	if strings.HasPrefix(*song, "https://open.spotify.com") {
		new_song = track_words[len(track_words)-1]
	}

	if strings.HasPrefix(*song, "https://open.spotify.com/album") {
		tracks, err := client.GetAlbum(context.Background(), spotify.ID(new_song))
		if err == nil {
			current_pos := QueueDict[message.GuildID].CurrentPos()
			for i := 0; i < len(tracks.Tracks.Tracks); i++ {
				var track queue.Song
				track.Title = fmt.Sprintf("%v - %v", tracks.Tracks.Tracks[i].Name, tracks.Artists[0].Name)
				duration := strconv.Itoa(tracks.Tracks.Tracks[i].Duration / 1000)
				track.Duration = &duration
				track.Pos = current_pos + 1
				current_pos++
				track_list = append(track_list, track)
			}
		} else if err != nil {
			log.Println(err)
			s.ChannelMessageSend(message.ChannelID, "Could not get song(s) from given query.")
			return
		}
	} else if strings.HasPrefix(*song, "https://open.spotify.com/playlist") {
		tracks, err := client.GetPlaylistItems(context.Background(), spotify.ID(new_song))
		if err == nil {
			current_pos := QueueDict[message.GuildID].CurrentPos()
			for i := 0; i < len(tracks.Items); i++ {
				var track queue.Song
				track.Title = fmt.Sprintf("%v - %v", tracks.Items[i].Track.Track.Name, tracks.Items[i].Track.Track.Artists[0].Name)
				duration := strconv.Itoa(tracks.Items[i].Track.Track.Duration / 1000)
				track.Duration = &duration
				track.Pos = current_pos + 1
				current_pos++
				track_list = append(track_list, track)
			}
		} else if err != nil {
			log.Println(err)
			s.ChannelMessageSend(message.ChannelID, "Could not get song(s) from given query.")
			return
		}
	} else if strings.HasPrefix(*song, "https://open.spotify.com/track") {
		tracks, err := client.GetTrack(context.Background(), spotify.ID(new_song))
		current_pos := QueueDict[message.GuildID].CurrentPos()
		if err == nil {
			var track queue.Song
			track.Title = fmt.Sprintf("%v - %v", tracks.Name, tracks.Artists[0].Name)
			duration := strconv.Itoa(tracks.Duration / 1000)
			track.Duration = &duration
			track.Pos = current_pos + 1
			current_pos++
			track_list = append(track_list, track)
		} else if err != nil {
			log.Println(err)
			s.ChannelMessageSend(message.ChannelID, "Could not get song(s) from given query.")
			return
		}
	} else if strings.HasPrefix(*song, "https://soundcloud") {
		split_url := strings.Split(*song, "/")
		if split_url[4] == "sets" {
			playlist_info, err := SoundCloud.GetPlaylistInfo(*song)
			if err != nil {
				log.Println(err)
				s.ChannelMessageSend(message.ChannelID, "Could not get song(s) from given query.")
				return
			} else {
				tracks := playlist_info.Tracks
				current_pos := QueueDict[message.GuildID].CurrentPos()
				for i := 0; i < len(tracks); i++ {
					var track queue.Song
					track.Title = tracks[i].Title
					track.Url = &tracks[i].PermalinkURL
					track.Pos = current_pos + 1
					current_pos++
					track_list = append(track_list, track)
				}
			}
		}
	}

	var song_list []queue.Song
	if len(track_list) == 0 {
		song_list, err = queue.YtSearch(*song, *QueueDict[message.GuildID])
		if err != nil {
			log.Println(err)
			s.ChannelMessageSend(message.ChannelID, "Could not get song(s) from given query.")
			return
		}
	}

	for i := 0; i < len(song_list); i++ {
		QueueDict[message.GuildID].Enqueue(song_list[i])
	}

	if len(track_list) != 0 {
		for i := 0; i < len(track_list); i++ {
			QueueDict[message.GuildID].Enqueue(track_list[i])
		}
	}

	go QueueDict[message.GuildID].CheckPlaying()
}

func Pause(s *discordgo.Session, message *discordgo.MessageCreate, args *string) {
	if _, ok := QueueDict[message.GuildID]; ok {
		QueueDict[message.GuildID].Pause()
	}
}

func Unpause(s *discordgo.Session, message *discordgo.MessageCreate, args *string) {
	if _, ok := QueueDict[message.GuildID]; ok {
		QueueDict[message.GuildID].Unpause()
	}
}

func Skip(s *discordgo.Session, message *discordgo.MessageCreate, args *string) {
	if _, ok := QueueDict[message.GuildID]; ok {
		QueueDict[message.GuildID].Skip()
	}
}

func ShowQueue(s *discordgo.Session, message *discordgo.MessageCreate, args *string) {
	if _, ok := QueueDict[message.GuildID]; ok {

		title_list := make([]string, 0)

		songs := QueueDict[message.GuildID].GetSongs()

		for i := 0; i < len(songs); i++ {
			if songs[i].Pos == QueueDict[message.GuildID].CurrentPos() {
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

		s.ChannelMessageSendEmbed(message.ChannelID, &embed_to_send)
	}
}

func Loop(s *discordgo.Session, message *discordgo.MessageCreate, args *string) {
	if _, ok := QueueDict[message.GuildID]; ok {
		QueueDict[message.GuildID].Loop()
	}
}

func Shuffle(s *discordgo.Session, message *discordgo.MessageCreate, args *string) {
	if _, ok := QueueDict[message.GuildID]; ok {
		QueueDict[message.GuildID].Shuffle()
	}
}

func Clear(s *discordgo.Session, message *discordgo.MessageCreate, args *string) {
	if _, ok := QueueDict[message.GuildID]; ok {
		QueueDict[message.GuildID].Clear()
	}
}

func GoTo(s *discordgo.Session, message *discordgo.MessageCreate, arg *string) {
	if _, ok := QueueDict[message.GuildID]; ok {
		int, err := strconv.Atoi(*arg)
		if err != nil {
			s.ChannelMessageSend(message.ChannelID, "You did not pass a valid integer.")
			return
		}
		if int > QueueDict[message.GuildID].Len() {
			int = QueueDict[message.GuildID].Len()
		}
		QueueDict[message.GuildID].MoveTo(int)
	}
}

func FuckOff(s *discordgo.Session, message *discordgo.MessageCreate, args *string) {
	if _, ok := QueueDict[message.GuildID]; ok {
		QueueDict[message.GuildID].FuckOff()
		delete(QueueDict, message.GuildID)
	}
}

func Remove(s *discordgo.Session, message *discordgo.MessageCreate, arg *string) {
	if _, ok := QueueDict[message.GuildID]; ok {
		int, err := strconv.Atoi(*arg)
		if err != nil {
			s.ChannelMessageSend(message.ChannelID, "You did not pass a valid integer.")
			return
		}
		if int > QueueDict[message.GuildID].Len() {
			s.ChannelMessageSend(message.ChannelID, "You cannot remove a song thats not in the queue.")
			return
		}
		QueueDict[message.GuildID].Remove(int)
	}
}

func NowPlaying(s *discordgo.Session, message *discordgo.MessageCreate, args *string) {
	if _, ok := QueueDict[message.GuildID]; ok {
		current_time := QueueDict[message.GuildID].CurrentTime()
		total_time := QueueDict[message.GuildID].TotalTime()

		embed_to_send := discordgo.MessageEmbed{
			Title:       (strconv.Itoa(QueueDict[message.GuildID].Current().Pos) + " - " + QueueDict[message.GuildID].Current().Title),
			Description: current_time + "/" + total_time,
		}

		s.ChannelMessageSendEmbed(message.ChannelID, &embed_to_send)

	}
}
