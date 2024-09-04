package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	tele "gopkg.in/telebot.v3"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Token      string `yaml:"token" validate:"required"`
	WelcomeMsg string `yaml:"welcome_msg" default:"Welcome! Please enter a domain or IP address to scan."`
}

var userStates = make(map[int64]string)
var userTargets = make(map[int64]string)

func loadConfig(filename string) (Config, error) {
	var cfg Config
	file, err := os.Open(filename)
	if err != nil {
		return cfg, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func setupLogging() {
	log.SetOutput(os.Stdout)
	log.Println("Logging started")
	
}

func main() {
	setupLogging()

	cfg, err := loadConfig("/etc/tg-nmap-bot/config.yml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	pref := tele.Settings{
		Token:  cfg.Token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
		OnError: func(e error, c tele.Context) {
			log.Println("Error:", e)
		},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatalf("Creating bot: %v", err)
	}

	b.Handle("/start", func(c tele.Context) error {
		userStates[c.Chat().ID] = "waiting_for_target"
		log.Printf("User %d started interaction with the bot", c.Chat().ID)
		return c.Send(cfg.WelcomeMsg)
	})

	b.Handle("/help", func(c tele.Context) error {
		userStates[c.Chat().ID] = "waiting_for_target"
		log.Printf("User %d enter help command", c.Chat().ID)
		return c.Send("Тут это, короче, могу только морально поддрежать. А вообще тыка на клавиатуре /start , а после по подсказкам.")
	})

	b.Handle(tele.OnText, func(c tele.Context) error {
		state, exists := userStates[c.Chat().ID]
		if !exists {
			return c.Send("Пожалуйста, начните с команды /start.")
		}

		switch state {
		case "waiting_for_target":
			userTargets[c.Chat().ID] = c.Text()
			userStates[c.Chat().ID] = "waiting_for_ports"
			log.Printf("User %d entered target address: %s", c.Chat().ID, c.Text())
			return c.Send("Введите диапазон портов для сканирования (например, 1-100):")

		case "waiting_for_ports":
			portRange := c.Text()
			target := userTargets[c.Chat().ID]
			startMsg := fmt.Sprintf("Начинаю сканирование %s на портах %s..", target, portRange)
			_ = c.Send(startMsg)

			output, err := runNmap(target, portRange)
			if err != nil {
				log.Printf("Error executing nmap: %v", err)
				return c.Send(fmt.Sprintf("Error: %s", err))
			}

			finishMsg := "Сканирование завершено:\n" + output
			userStates[c.Chat().ID] = "waiting_for_target"
			log.Printf("Scan completed for user %d", c.Chat().ID)
			return c.Send(finishMsg)

		default:
			return c.Send("Ошибка. Пожалуйста, начните с командой /start.")
		}
	})

	b.Start()
}

func runNmap(target, portRange string) (string, error) {
	args := []string{"-p", portRange, target}
	cmd := exec.Command("nmap", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(output), nil
}
