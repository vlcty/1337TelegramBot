package main

import (
    "log"
)

type TelegramMessage struct {
    Message struct {
        Chat struct {
            ID int
        }
        From struct {
            Username string
        }
        Text string
    }
}

func (message *TelegramMessage) GetChatID() int {
    return message.Message.Chat.ID
}

func (message *TelegramMessage) GetUsername() string {
    return message.Message.From.Username
}

func (message *TelegramMessage) GetMessage() string {
    return message.Message.Text
}

func (message *TelegramMessage) Answer(whatToSay string) {
    log.Printf("Sending message to chat %d: %s", message.GetChatID(), whatToSay);
}
