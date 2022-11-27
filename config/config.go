package config

import (
	"github.com/spf13/viper"
)

type Tokens struct {
	Discord_token  string
	Genius_token   string
	Spotify_id     string
	Spotify_secret string
}

type SpamSettings struct {
	Antispam   bool
	Spam_count int
}

type LevelSystem struct {
	Levels_on           bool
	Xp_per_message      []int
	Cooldown_in_seconds int
}

type BotConfig struct {
	Prefix        string
	Tokens        Tokens
	Spam_settings SpamSettings
	Level_system  LevelSystem
}

func ReadConfig() BotConfig {
	viper.SetConfigName("config")
	viper.AddConfigPath("./")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	var config BotConfig

	config.Tokens.Discord_token = viper.GetString("tokens.discord_token")
	config.Tokens.Genius_token = viper.GetString("tokens.genius_token")
	config.Tokens.Spotify_id = viper.GetString("tokens.spotify_id")
	config.Tokens.Spotify_secret = viper.GetString("tokens.spotify_secret")

	config.Spam_settings.Antispam = viper.GetBool("spam_settings.antispam")
	config.Spam_settings.Spam_count = viper.GetInt("spam_settings.spam_count")

	config.Level_system.Levels_on = viper.GetBool("level_system.levels_on")
	config.Level_system.Xp_per_message = viper.GetIntSlice("level_system.xp_per_message")
	config.Level_system.Cooldown_in_seconds = viper.GetInt("level_system.cooldown_in_seconds")

	config.Prefix = viper.GetString("prefix")

	return config
}
