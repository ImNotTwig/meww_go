package queue

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/ImNotTwig/dca"
	"github.com/bwmarrin/discordgo"
)

type Song struct {
	Title    string
	Url      *string
	Pos      int
	Duration *string
}

type Queue struct {
	Songs                 []Song
	Current_Pos           int
	Voice                 *discordgo.VoiceConnection
	Session               *discordgo.Session
	Message               *discordgo.MessageCreate
	Stream                *dca.StreamingSession
	Encoding_Session      *dca.EncodeSession
	Loop                  bool
	Shuffle               bool
	Already_Played_Tracks []int
	End_Of_Queue          bool
	Paused                bool
}

func NewQueue(s *discordgo.Session, m *discordgo.MessageCreate) *Queue {
	return &Queue{
		Songs:                 make([]Song, 0),
		Current_Pos:           0,
		Session:               s,
		Message:               m,
		Loop:                  false,
		Shuffle:               false,
		Already_Played_Tracks: make([]int, 0),
		End_Of_Queue:          false,
		Paused:                false,
	}
}

func (q *Queue) SetVoice(v *discordgo.VoiceConnection) {
	q.Voice = v
}

func (q *Queue) SetMessageChannel(m *discordgo.MessageCreate) {
	q.Message = m
}

func (q *Queue) CheckPlaying() {
	if !q.Paused {
		if q.DoesStreamExist() {
			return
		} else if !q.DoesStreamExist() {
			q.Shift()
			if q.DoesStreamExist() {
				q.UnpauseQuiet()
				return
			}

		} else if q.End_Of_Queue {
			q.MoveToQuiet(q.Len())
			if q.DoesStreamExist() {
				q.UnpauseQuiet()
			}
			return
		}
	}
}

func (q *Queue) Enqueue(song Song) {
	q.Songs = append(q.Songs, song)
}

func (q *Queue) Remove(index int) {
	if index == q.Current_Pos {
		q.Session.ChannelMessageSend(q.Message.ChannelID, "You cannot remove the currently playing song.")
		return
	}
	q.Songs = append(q.Songs[:index-1], q.Songs[index:]...)
	q.ValidateSongPos()
}

func (q Queue) DoesStreamExist() bool {
	if q.Stream != nil {
		return true
	} else {
		return false
	}
}

func (q *Queue) Insert(song Song, index int) {
	if len(q.Songs) == index {
		q.Songs = append(q.Songs, song)
	}
	q.Songs = append(q.Songs[:index+1], q.Songs[index:]...)
	q.Songs[index] = song
}

func (q *Queue) ValidateSongPos() {
	for i := 0; i < q.Len(); i++ {
		q.Songs[i].Pos = i + 1
	}
}

func (q Queue) Len() int {
	return len(q.Songs)
}

func (q Queue) CurrentTime() string {
	if q.Stream != nil {
		string_time := q.Stream.PlaybackPosition().String()
		minutes_and_seconds := strings.Split(string_time, ".")[0]
		return minutes_and_seconds + "s"
	}
	return "0s"
}

func (q Queue) TotalTime() string {
	if q.Stream != nil {
		if q.Current().Duration != nil {
			return *q.Current().Duration
		} else {
			return "0s"
		}
	}
	return "0s"
}

func (q Queue) Current() Song {
	return q.Songs[q.Current_Pos-1]
}

func (q *Queue) FuckOff() {
	q.Songs = make([]Song, 0)
	if q.Stream != nil && q.Encoding_Session != nil {
		q.Stream.SetPaused(true)
		q.Encoding_Session.Stop()
	}
	q.Current_Pos = 0
	if q.Voice != nil && q.Session != nil {
		q.Session.ChannelMessageSend(q.Message.ChannelID, "The bot has left the Voice channel, and the queue has been cleared.")
	}
	q.Voice.Disconnect()
}

func (q *Queue) MoveTo(index int) {
	q.Current_Pos = index - 1
	if q.Stream != nil && q.Encoding_Session != nil {
		q.Stream.SetPaused(true)
		q.Encoding_Session.Stop()
	}
	q.Session.ChannelMessageSend(q.Message.ChannelID, "Moved to "+strconv.Itoa(index)+" in the queue.")
	q.End_Of_Queue = false
	q.Paused = false
	defer q.Shift()
}

func (q *Queue) MoveToQuiet(index int) {
	q.Current_Pos = index
	if q.Stream != nil && q.Encoding_Session != nil {
		q.Stream.SetPaused(true)
		q.Encoding_Session.Stop()
	}
	q.End_Of_Queue = false
	q.Paused = false
	defer q.Play()
}

func (q *Queue) Clear() {
	q.Songs = make([]Song, 0)
	q.Stream.SetPaused(true)
	q.Encoding_Session.Stop()
	q.Session.ChannelMessageSend(q.Message.ChannelID, "The queue has been cleared, and the player has stopped.")
	q.Current_Pos = 0
}

func (q *Queue) Shift() {
	q.End_Of_Queue = false
	if !q.Shuffle {
		q.CancelSong()
		q.Current_Pos += 1
	} else {
		q.CancelSong()
		old_pos := q.Current_Pos
		q.Current_Pos = rand.Intn(q.Len()) + 1

		for contains(q.Already_Played_Tracks, q.Current_Pos) && q.Current_Pos != old_pos {
			q.Current_Pos = rand.Intn(q.Len()) + 1
		}

		if len(q.Already_Played_Tracks) == q.Len() {
			q.Session.ChannelMessageSend(q.Message.ChannelID, "Every song in the queue has been played, Shuffle has been turned off.")
			q.Encoding_Session.Stop()
		}
		q.Already_Played_Tracks = append(q.Already_Played_Tracks, old_pos)
	}

	if q.Len() < q.Current_Pos {
		q.End_Of_Queue = true
	}
	q.Paused = false
	defer q.Play()
}

func (q *Queue) CancelSong() {
	if q.Stream != nil {
		q.Stream.SetPaused(true)
	}
	if q.Encoding_Session != nil {
		q.Encoding_Session.Stop()
	}
	q.Stream = nil
	q.Encoding_Session = nil
	q.Paused = true
}

func contains(s []int, i int) bool {
	for _, v := range s {
		if v == i {
			return true
		}
	}

	return false
}

func (q *Queue) Pause() {
	if q.Stream != nil && q.Encoding_Session != nil {
		q.Stream.SetPaused(true)
		q.Session.ChannelMessageSend(q.Message.ChannelID, "Paused the queue.")
	}
	q.Paused = true
}

func (q *Queue) UnpauseQuiet() {
	if q.Stream != nil && q.Encoding_Session != nil {
		q.Stream.SetPaused(false)

		q.End_Of_Queue = false
	}
	q.Paused = false
}

func (q *Queue) PauseQuiet() {
	if q.Stream != nil && q.Encoding_Session != nil {
		q.Stream.SetPaused(true)
		q.End_Of_Queue = true
	}
	q.Paused = true
}

func (q *Queue) Unpause() {
	if q.Stream != nil && q.Encoding_Session != nil {
		q.Stream.SetPaused(false)
		q.Session.ChannelMessageSend(q.Message.ChannelID, "Un-Paused the queue.")
	}
	q.Paused = false
}

func (q *Queue) Skip() {
	if q.Encoding_Session != nil {
		q.Session.ChannelMessageSend(q.Message.ChannelID, "Skipped the current song.")
		q.Shift()
		if q.Stream != nil && q.Encoding_Session != nil {
			q.Stream.SetPaused(true)
			q.Encoding_Session.Stop()
		}
		if q.Current_Pos < q.Len() {
			q.End_Of_Queue = false
		}
		q.Play()
	}
}

func (q *Queue) SkipQuiet() {
	if q.Encoding_Session != nil {
		q.Shift()
		if q.Stream != nil && q.Encoding_Session != nil {
			q.Stream.SetPaused(true)
			q.Encoding_Session.Stop()
		}
		if q.Current_Pos < q.Len() {
			q.End_Of_Queue = false
		}
		q.Play()
	}
}

func (q *Queue) SetLoop() {
	if !q.Loop {
		q.Loop = true
		q.Session.ChannelMessageSend(q.Message.ChannelID, "Turned the Loop on.")
	} else {
		q.Loop = false
		q.Session.ChannelMessageSend(q.Message.ChannelID, "Turned the Loop off.")
	}
}

func (q *Queue) SetShuffle() {
	if !q.Shuffle {
		q.Shuffle = true
		q.Session.ChannelMessageSend(q.Message.ChannelID, "Turned Shuffle on.")
	} else {
		q.Shuffle = false
		q.Session.ChannelMessageSend(q.Message.ChannelID, "Turned Shuffle off.")
		q.Already_Played_Tracks = make([]int, 0)
	}
}

// The main Function for the queue; this function plays the song at the current posistion,
// pauses the queue if at the end, or will Loop it.
//
// This function basically is the control flow for the entire queue
func (q *Queue) Play() {

	if q.Stream != nil {
		if q.Stream.Paused() {
			q.Paused = true
		} else {
			q.Paused = false
		}
	}

	if q.Current_Pos > q.Len() {
		if !q.Loop {

			q.End_Of_Queue = true
			q.Current_Pos = q.Len()

			if q.Stream != nil {
				q.Stream.SetPaused(true)
				q.Paused = true
				q.Encoding_Session.Cleanup()
			}

			guild, _ := q.Session.State.Guild(q.Message.GuildID)
			in_vc := false
			for i := 0; i < len(guild.VoiceStates); i++ {
				if guild.VoiceStates[i].UserID == q.Session.State.User.ID {
					in_vc = true
					break
				}
			}
			if in_vc {
				q.Session.ChannelMessageSend(q.Message.ChannelID, "Reached end of queue, pausing.")
			}

		} else {
			q.Current_Pos = 1
			if q.Stream != nil {
				q.Encoding_Session.Cleanup()
			}
			q.Paused = false
		}
	}

	if q.Paused {
		if q.Stream != nil {
			q.Stream.SetPaused(true)
		}
	} else {
		if q.Stream != nil {
			q.Stream.SetPaused(false)
		}
	}
	if q.Stream != nil {
		if q.Stream.Paused() {
			q.Paused = true
		} else {
			q.Paused = false
		}
	}

	if (!q.End_Of_Queue || q.Loop == true) && !q.Paused {
		guild, _ := q.Session.State.Guild(q.Message.GuildID)
		in_vc := false
		for i := 0; i < len(guild.VoiceStates); i++ {
			if guild.VoiceStates[i].UserID == q.Session.State.User.ID {
				in_vc = true
				break
			}
		}
		if in_vc {
			q.Session.ChannelMessageSend(q.Message.ChannelID, "Now playing: "+q.Current().Title)
		}
		encodingOptions := dca.StdEncodeOptions
		encodingOptions.RawOutput = true
		encodingOptions.Bitrate = 256
		encodingOptions.Application = "lowdelay"
		encodingSession, err := dca.EncodeFile(q.GetStreamUrl(&q.Songs[q.Current_Pos-1]), encodingOptions)

		if err != nil {
			return
		}

		defer encodingSession.Cleanup()
		done := make(chan error)

		q.Stream = dca.NewStream(encodingSession, q.Voice, done)
		q.Encoding_Session = encodingSession
		<-done

		if q.Paused {
			q.Stream.SetPaused(true)
		}

		defer q.Shift()
		q.End_Of_Queue = false
	}
}

func (q Queue) GetStreamUrl(song *Song) string {
	if song.Url == nil {
		song_list, err := YtSearch(song.Title, q)
		if err != nil {
			log.Println(err)
		}
		song.Url = song_list[0].Url
	}
	out, err := exec.Command(
		"yt-dlp",
		"-N",
		"64",
		*song.Url,
		"--no-download",
		"--flat-playlist",
		"--downloader",
		"aria2c",
		"-J",
	).Output()

	if err != nil {
		log.Println(err)
	}

	var json_data map[string]interface{}
	json.Unmarshal(out, &json_data)

	if _, ok := json_data["requested_formats"]; ok {

		requested_formats := json_data["requested_formats"].([]interface{})
		var format_we_want map[string]interface{}
		if len(requested_formats) == 1 {
			format_we_want = requested_formats[0].(map[string]interface{})
		} else {
			format_we_want = requested_formats[1].(map[string]interface{})
		}
		return format_we_want["url"].(string)
	} else {
		return json_data["url"].(string)
	}
}

func YtSearch(song string, queue Queue) ([]Song, error) {
	url, err := url.ParseRequestURI(song)
	var song_list []Song

	is_url := true
	if err != nil {
		is_url = false
	}

	var out []byte

	if is_url {

		out, err = exec.Command(
			"yt-dlp",
			"-N",
			"64",
			url.String(),
			"--no-download",
			"--flat-playlist",
			"--downloader",
			"aria2c",
			"-J",
		).Output()

	} else {

		out, err = exec.Command(
			"yt-dlp",
			"-N",
			"64",
			"ytsearch:"+song,
			"--no-download",
			"--flat-playlist",
			"--downloader",
			"aria2c",
			"-J",
		).Output()

	}

	if err != nil {
		return nil, err
	}

	var json_data map[string]interface{}
	json.Unmarshal(out, &json_data)

	i := 1

	if _, ok := json_data["entries"]; ok {
		entries, _ := json_data["entries"].([]interface{})
		for i = 0; i < len(entries); i++ {
			entry := entries[i].(map[string]interface{})
			var song Song
			song.Title = entry["title"].(string)
			new_url := strings.TrimSuffix(entry["url"].(string), "#__youtubedl_smuggle=%7B%22is_music_url%22%3A+true%7D")
			song.Url = &new_url
			song.Pos = queue.Len() + i + 1
			if _, ok := json_data["duration"]; ok {
				duration := entry["duration"].(float64)
				seconds, _ := time.ParseDuration(fmt.Sprintf("%v", duration) + "s")
				duration_string := seconds.String()
				song.Duration = &duration_string
			}
			song_list = append(song_list, song)
		}
	} else {
		var song Song
		song.Title = json_data["title"].(string)
		new_url := strings.TrimSuffix(json_data["webpage_url"].(string), "#__youtubedl_smuggle=%7B%22is_music_url%22%3A+true%7D")
		song.Url = &new_url
		song.Pos = queue.Len() + i
		if _, ok := json_data["duration"]; ok {
			duration := json_data["duration"].(float64)
			seconds, _ := time.ParseDuration(fmt.Sprintf("%v", duration) + "s")
			duration_string := seconds.String()
			song.Duration = &duration_string
		}
		song_list = append(song_list, song)
	}
	return song_list, nil
}
