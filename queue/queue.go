package queue

import (
	"encoding/json"
	"fmt"
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
	Duration string
}

type Queue struct {
	songs                 []Song
	current_pos           int
	voice                 *discordgo.VoiceConnection
	session               *discordgo.Session
	message               *discordgo.MessageCreate
	stream                *dca.StreamingSession
	encoding_session      *dca.EncodeSession
	loop                  bool
	shuffle               bool
	already_played_tracks []int
	end_of_queue          bool
	paused                bool
}

func NewQueue(s *discordgo.Session, m *discordgo.MessageCreate) *Queue {
	return &Queue{
		songs:                 make([]Song, 0),
		current_pos:           0,
		session:               s,
		message:               m,
		loop:                  false,
		shuffle:               false,
		already_played_tracks: make([]int, 0),
		end_of_queue:          false,
		paused:                false,
	}
}

func (q *Queue) SetVoice(v *discordgo.VoiceConnection) {
	q.voice = v
}

func (q Queue) GetVoice() *discordgo.VoiceConnection {
	return q.voice
}

func (q *Queue) CheckPlaying() {
	if q.Paused() {
		if q.DoesStreamExist() {
			return
		} else if q.DoesStreamExist() == false {
			q.Shift()
			if q.DoesStreamExist() {
				q.UnpauseQuiet()
				return
			}

		} else if q.EndOfQueue() {
			q.MoveToQuiet(q.Len())
			if q.DoesStreamExist() {
				q.UnpauseQuiet()
			}
			return
		}
	}
}

func (q *Queue) Enqueue(song Song) {
	q.songs = append(q.songs, song)
}

func (q *Queue) Remove(index int) {
	if index-1 == q.current_pos {
		q.session.ChannelMessageSend(q.message.ChannelID, "You cannot remove the currently playing song.")
		return
	}
	q.songs = append(q.songs[:index-1], q.songs[index:]...)
	q.ValidateSongPos()
}

func (q Queue) DoesStreamExist() bool {
	if q.stream != nil {
		return true
	} else {
		return false
	}
}

func (q *Queue) Insert(song Song, index int) {
	if len(q.songs) == index {
		q.songs = append(q.songs, song)
	}
	q.songs = append(q.songs[:index+1], q.songs[index:]...)
	q.songs[index] = song
}

func (q *Queue) ValidateSongPos() {
	for i := 0; i < q.Len(); i++ {
		q.songs[i].Pos = i + 1
	}
}

func (q Queue) GetSongs() []Song {
	return q.songs
}

func (q Queue) CurrentPos() int {
	return q.current_pos
}

func (q Queue) Len() int {
	return len(q.songs)
}

func (q Queue) CurrentTime() string {
	if q.stream != nil {
		string_time := q.stream.PlaybackPosition().String()
		minutes_and_seconds := strings.Split(string_time, ".")[0]
		return minutes_and_seconds + "s"
	}
	return "0s"
}

func (q Queue) TotalTime() string {
	if q.stream != nil {
		return q.Current().Duration
	}
	return "0s"
}

func (q Queue) Current() Song {
	return q.songs[q.current_pos-1]
}

func (q *Queue) FuckOff() {
	q.songs = make([]Song, 0)
	if q.stream != nil && q.encoding_session != nil {
		q.stream.SetPaused(true)
		q.encoding_session.Stop()
	}
	q.current_pos = 0
	if q.voice != nil {
		q.session.ChannelMessageSend(q.message.ChannelID, "The bot has left the voice channel, and the queue has been cleared.")
	}
	q.voice.Disconnect()
}

func (q *Queue) MoveTo(index int) {
	q.current_pos = index - 1
	if q.stream != nil && q.encoding_session != nil {
		q.stream.SetPaused(true)
		q.encoding_session.Stop()
	}
	q.session.ChannelMessageSend(q.message.ChannelID, "Moved to "+strconv.Itoa(index)+" in the queue.")
	q.end_of_queue = false
	q.paused = false
	defer q.Shift()
}

func (q *Queue) MoveToQuiet(index int) {
	q.current_pos = index
	if q.stream != nil && q.encoding_session != nil {
		q.stream.SetPaused(true)
		q.encoding_session.Stop()
	}
	q.end_of_queue = false
	q.paused = false
	defer q.Play()
}

func (q *Queue) Clear() {
	q.songs = make([]Song, 0)
	q.stream.SetPaused(true)
	q.encoding_session.Stop()
	q.session.ChannelMessageSend(q.message.ChannelID, "The queue has been cleared, and the player has stopped.")
	q.current_pos = 0
}

func (q *Queue) Shift() {
	q.end_of_queue = false
	if q.shuffle == false {
		q.CancelSong()
		q.current_pos += 1
	} else {
		q.CancelSong()
		old_pos := q.current_pos
		q.current_pos = rand.Intn(q.Len()) + 1

		for contains(q.already_played_tracks, q.current_pos) && q.current_pos != old_pos {
			q.current_pos = rand.Intn(q.Len()) + 1
		}

		if len(q.already_played_tracks) == q.Len() {
			q.session.ChannelMessageSend(q.message.ChannelID, "Every song in the queue has been played, shuffle has been turned off.")
			q.encoding_session.Stop()
		}
		q.already_played_tracks = append(q.already_played_tracks, old_pos)
	}

	if q.Len() < q.current_pos {
		q.end_of_queue = true
	}

	defer q.Play()
}

func (q *Queue) CancelSong() {
	if q.stream != nil {
		q.stream.SetPaused(true)
	}
	if q.encoding_session != nil {
		q.encoding_session.Stop()
	}
	q.stream = nil
	q.encoding_session = nil
}

func contains(s []int, i int) bool {
	for _, v := range s {
		if v == i {
			return true
		}
	}

	return false
}

func (q Queue) IsPlaying() bool {
	if q.paused == false {
		return true
	}
	return false
}

func (q *Queue) Pause() {
	if q.stream != nil && q.encoding_session != nil {
		q.stream.SetPaused(true)
		q.session.ChannelMessageSend(q.message.ChannelID, "Paused the queue.")
	}
	q.paused = true
}

func (q *Queue) UnpauseQuiet() {
	if q.stream != nil && q.encoding_session != nil {
		q.stream.SetPaused(false)

		q.end_of_queue = false
	}
	q.paused = false
}

func (q *Queue) PauseQuiet() {
	if q.stream != nil && q.encoding_session != nil {
		q.stream.SetPaused(true)
		q.end_of_queue = true
	}
	q.paused = true
}

func (q *Queue) Unpause() {
	if q.stream != nil && q.encoding_session != nil {
		q.stream.SetPaused(false)
		q.session.ChannelMessageSend(q.message.ChannelID, "Un-paused the queue.")
	}
	q.paused = false
}

func (q *Queue) Skip() {
	if q.encoding_session != nil {
		q.session.ChannelMessageSend(q.message.ChannelID, "Skipped the current song.")
		q.Shift()
		if q.stream != nil && q.encoding_session != nil {
			q.stream.SetPaused(true)
			q.encoding_session.Stop()
		}
		if q.current_pos < q.Len() {
			q.end_of_queue = false
		}
		q.Play()
	}
}
func (q *Queue) SkipQuiet() {
	if q.encoding_session != nil {
		q.Shift()
		if q.stream != nil && q.encoding_session != nil {
			q.stream.SetPaused(true)
			q.encoding_session.Stop()
		}
		if q.current_pos < q.Len() {
			q.end_of_queue = false
		}
		q.Play()
	}
}

func (q Queue) EndOfQueue() bool {
	return q.end_of_queue
}

func (q Queue) Paused() bool {
	if q.DoesStreamExist() {
		return q.paused
	}
	return true
}

func (q *Queue) Loop() {
	if q.loop == false {
		q.loop = true
		q.session.ChannelMessageSend(q.message.ChannelID, "Turned the loop on.")
	} else {
		q.loop = false
		q.session.ChannelMessageSend(q.message.ChannelID, "Turned the loop off.")
	}
}

func (q *Queue) Shuffle() {
	if q.shuffle == false {
		q.shuffle = true
		q.session.ChannelMessageSend(q.message.ChannelID, "Turned shuffle on.")
	} else {
		q.shuffle = false
		q.session.ChannelMessageSend(q.message.ChannelID, "Turned shuffle off.")
		q.already_played_tracks = make([]int, 0)
	}
}

// The main Function for the queue; this function plays the song at the current posistion,
// pauses the queue if at the end, or will loop it.
//
// This function basically is the control flow for the entire queue
func (q *Queue) Play() {

	if q.stream != nil {
		if q.stream.Paused() {
			q.paused = true
		} else {
			q.paused = false
		}
	}

	if q.current_pos > q.Len() {
		if q.loop == false {

			q.end_of_queue = true
			q.current_pos = q.Len()

			if q.stream != nil {
				q.stream.SetPaused(true)
				q.paused = true
				q.encoding_session.Cleanup()
			}

			guild, _ := q.session.State.Guild(q.message.GuildID)
			in_vc := false
			for i := 0; i < len(guild.VoiceStates); i++ {
				if guild.VoiceStates[i].UserID == q.session.State.User.ID {
					in_vc = true
					break
				}
			}
			if in_vc {
				q.session.ChannelMessageSend(q.message.ChannelID, "Reached end of queue, pausing.")
			}

		} else {
			q.current_pos = 1
		}
	}

	if q.paused {
		if q.stream != nil {
			q.stream.SetPaused(true)
		}
	} else {
		if q.stream != nil {
			q.stream.SetPaused(false)
		}
	}
	if q.stream != nil {
		if q.stream.Paused() {
			q.paused = true
		} else {
			q.paused = false
		}
	}

	if q.end_of_queue == false && q.paused == false {

		if q.paused == false {
			guild, _ := q.session.State.Guild(q.message.GuildID)
			in_vc := false
			for i := 0; i < len(guild.VoiceStates); i++ {
				if guild.VoiceStates[i].UserID == q.session.State.User.ID {
					in_vc = true
					break
				}
			}
			if in_vc {
				q.session.ChannelMessageSend(q.message.ChannelID, "Now playing: "+q.Current().Title)
			}
		}
		encodingOptions := dca.StdEncodeOptions
		encodingOptions.RawOutput = true
		encodingOptions.Bitrate = 256
		encodingOptions.Application = "lowdelay"
		encodingSession, err := dca.EncodeFile(q.GetStreamUrl(q.Current()), encodingOptions)

		if err != nil {
			return
		}

		defer encodingSession.Cleanup()
		done := make(chan error)

		q.stream = dca.NewStream(encodingSession, q.voice, done)
		q.encoding_session = encodingSession
		err = <-done

		if q.paused == true {
			q.stream.SetPaused(true)
		}

		defer q.Shift()
		q.end_of_queue = false
	}
}

func (q Queue) GetStreamUrl(song Song) string {
	if song.Url == nil {
		song_list, err := YtSearch(song.Title, q)
		if err != nil {
			fmt.Println(err)
		}
		song.Url = song_list[0].Url
	}
	out, _ := exec.Command(
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
	var json_data map[string]interface{}
	json.Unmarshal(out, &json_data)

	requested_formats := json_data["requested_formats"].([]interface{})

	format_we_want := requested_formats[1].(map[string]interface{})

	return format_we_want["url"].(string)
}

func YtSearch(song string, queue Queue) ([]Song, error) {
	url, err := url.ParseRequestURI(song)
	var song_list []Song

	is_url := true
	if err != nil {
		is_url = false
	}

	var out []byte

	if is_url == true {

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
			duration := entry["duration"].(float64)
			seconds, _ := time.ParseDuration(fmt.Sprintf("%v", duration) + "s")
			song.Duration = seconds.String()
			song_list = append(song_list, song)
		}
	} else {
		var song Song
		song.Title = json_data["title"].(string)
		new_url := strings.TrimSuffix(json_data["webpage_url"].(string), "#__youtubedl_smuggle=%7B%22is_music_url%22%3A+true%7D")
		song.Url = &new_url
		song.Pos = queue.Len() + i
		duration := json_data["duration"].(float64)
		seconds, _ := time.ParseDuration(fmt.Sprintf("%v", duration) + "s")
		song.Duration = seconds.String()
		song_list = append(song_list, song)
	}

	return song_list, nil
}
