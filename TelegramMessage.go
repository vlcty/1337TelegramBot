package main

import (
    "log"
    "encoding/json"
    "net/http"
    "fmt"
    "bytes"
)

type TelegramMessage struct {
    Message struct {
        Chat struct {
            ID int
            Title string
        }
        From struct {
            Username string
        }
        Text string
    }
}

type TelegramResponse struct {
    Chat_ID int `json:"chat_id"`
    Text string `json:"text"`
    // Disable_notification bool `json:"disable_notification"`
}

func (message *TelegramMessage) GetChatID() int {
    return message.Message.Chat.ID
}

func (message *TelegramMessage) GetChatTitle() string {
    return message.Message.Chat.Title
}

func (message *TelegramMessage) GetUsername() string {
    return message.Message.From.Username
}

func (message *TelegramMessage) GetMessage() string {
    return message.Message.Text
}

func (message *TelegramMessage) Answer(whatToSay string) {
    log.Printf("Sending message to chat %d: %s", message.GetChatID(), whatToSay)

    telegramResponse := TelegramResponse{
        Chat_ID: message.GetChatID(),
        Text: whatToSay,
        //Disable_notification: true
    }

    b := new(bytes.Buffer)
    json.NewEncoder(b).Encode(telegramResponse)

    _, err := http.Post(fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken),
        "application/json;charset=utf-8", b )

    if err != nil {
        log.Println(err)
    }
}
