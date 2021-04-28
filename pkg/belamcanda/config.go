package belamcanda

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/TensRoses/iris/internal/configparser"
)

const (
	logLevel            int           = 2 // refers to internal/irislog/irislog.go for level definition
	defaultPomDuration  time.Duration = 25 * time.Minute
	baseAuthURLTemplate string        = "https://discord.com/api/oauth2/authorize?client_id=%s&scope=bot"
)

// TODO: add options to check for valid configuration name.
var (
	ClientID      = configparser.Register("iris.clientid", "ClientID of the bot", nil)
	ClientSecrets = configparser.Register("iris.clientsecret", "ClientSecret of the bot", nil)
	BotToken      = configparser.Register("iris.authtoken", "authentication token of the bot", nil)
	Loaded        = false

	// VERSION is defined via git.
	VERSION     = "unknown"
	IrisSession *discordgo.Session
	IrisUser    *discordgo.User
)

// LoadConfig loads required configs
func LoadConfig() error {
	if Loaded {
		return nil
	}

	Loaded = true
	configparser.AddSource(&configparser.EnvSource{})
	configparser.Load()

	required := []*configparser.Options{
		ClientID,
		ClientSecrets,
		BotToken,
	}

	for _, v := range required {
		if v.LoadedValue == nil {
			env := strings.ToUpper(strings.ReplaceAll(v.Name, ".", "_"))
			return fmt.Errorf("didn't contain required config options: %q (%s as envars)", v.Name, env)
		}
	}
	return nil
}