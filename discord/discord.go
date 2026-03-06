package discord

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

type Messenger interface {
	SendMessage(channelID, message string) error
}

type Bot struct {
	Session *discordgo.Session
}

func NewBot(token string) (*Bot, error) {
	session, err := discordgo.New(token)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}

	return &Bot{
		Session: session,
	}, nil
}

func (b *Bot) Start() error {
	err := b.Session.Open()
	if err != nil {
		return fmt.Errorf("error opening Discord session: %w", err)
	}
	log.Println("Discord bot started.")
	return nil
}

func (b *Bot) Stop() {
	b.Session.Close()
	log.Println("Discord bot stopped.")
}

func (b *Bot) SendMessage(channelID, message string) error {
	_, err := b.Session.ChannelMessageSend(channelID, message)
	if err != nil {
		return fmt.Errorf("error sending message to Discord channel %s: %w", channelID, err)
	}
	log.Printf("Message sent to Discord channel %s\n", channelID)
	return nil
}
