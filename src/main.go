package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

var println = fmt.Println

func main() {
	discord, err := discordgo.New("Bot " + "")
	if err != nil {
		println("Error creating Discord session,", err)
		return
	}

	out, err := exec.Command("yt-dlp", "-N", "64", "--no-download", "--flat-playlist", "-J", "--downloader", "aria2c", "https://music.youtube.com/playlist?list=OLAK5uy_ntb7UL7aPoMsaM0Q0vgmRQl2lIsH__kFk").Output()

	if err != nil {
		println("Error:", err)
		return
	}
	output := string(out[:])
	println(output)

	// adding a message create handler
	discord.AddHandler(messageCreate)

	discord.Identify.Intents = discordgo.IntentsGuildMessages

	err = discord.Open()
	if err != nil {
		println("Error opening connection,", err)
		return
	}

	println("Connected to discord as:", discord.State.User.Username)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	println("\nClosing connection to Discord.")
	discord.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}
}
