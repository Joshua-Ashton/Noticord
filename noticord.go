package main

import (
    "log"
    "net/http"
    "net/url"
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "github.com/gotify/go-api-client/auth"
    "github.com/gotify/go-api-client/client/message"
    "github.com/gotify/go-api-client/gotify"
    "github.com/gotify/go-api-client/models"

    "github.com/bwmarrin/discordgo"
)

const (
    gotifyURL        = ""
    gotifyToken      = ""
    
    discordEmail     = ""
    discordPassword  = ""

    myID             = ""
    boyfriendID      = ""
)

func sendNotification(username string, content string, priority int) {
    myURL, _ := url.Parse(gotifyURL)
    client := gotify.NewClient(myURL, &http.Client{})
    versionResponse, err := client.Version.GetVersion(nil)

    if err != nil {
        log.Fatal("Gotify: Could not request version ", err)
        return
    }
    version := versionResponse.Payload
    log.Println("Gotify: Found version", *version)

    params := message.NewCreateMessageParams()
    params.Body = &models.MessageExternal{
        Title:    "Noticord - " + username,
        Message:  content,
        Priority: priority,
    }
    _, err = client.Message.CreateMessage(params, auth.TokenAuth(gotifyToken))

    if err != nil {
        log.Fatalf("Gotify: Could not send message %v", err)
        return
    }
    log.Println("Gotify: Message Sent!")
}

func discordMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
    username     := m.Author.Username
    content      := m.Content
    priority     := 5

    if (m.Author.ID == boyfriendID) {
        priority = 10
    }

    // No messages from ourselves.
    if (m.Author.ID == myID) {
        return
    }

    // Only allow DMs.
    if m.GuildID != "" {
        return
    }

    if m.Type == discordgo.MessageTypeCall {
        content  = "Is calling you! 📞"
        priority = 10
    } else if m.Type != discordgo.MessageTypeDefault {
        return
    }

    if content == "" {
        return
    }

    log.Printf("Discord: Got message from %v: %v (%v)", username, content, priority)
    sendNotification(username, content, priority)
}

func main() {
    dg, err := discordgo.New(discordEmail, discordPassword)
    if err != nil {
        fmt.Println("Discord: Error creating Discord session ", err)
        return
    }

    dg.AddHandler(discordMessageCreate)

    err = dg.Open()
    if err != nil {
        fmt.Println("Discord: Error opening connection ", err)
        return
    }
    
    fmt.Println("Bot is now running. Press CTRL-C to exit.")
    sc := make(chan os.Signal, 1)
    signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
    <-sc

    dg.Close()
}
