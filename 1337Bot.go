package main

import (
    "log"
    "strings"
    "database/sql"
    "strconv"
    "time"
    "fmt"
    "net/http"
    "encoding/json"
    "io/ioutil"
    "os"

    _ "github.com/mattn/go-sqlite3"
)

var database *sql.DB
var botToken string

const (
    BotCommand_Start        string = "start"
    BotCommand_Stats        string = "stats"
    BotCommand_1337         string = "1337"
    BotCommand_Bonus        string = "bonus"
    BotCommand_Status       string = "status"

    Bot_Version             string = "1.0"
)

func main() {
    log.Println("Starting 1337Bot")

    GetBotToken()
    OpenDatabase()
    PrepareWebserver()
}

func GetBotToken() {
    botToken = os.Getenv("BOTTOKEN")

    if len(botToken) != 0 {
        log.Printf("Bot token is: %s", botToken)
    } else {
        log.Fatal("Did not get a bot token")
    }
}

func HandleHTTPCall(writer http.ResponseWriter, request *http.Request) {

    body, err := ioutil.ReadAll(request.Body)
    if err != nil {
        log.Println(err)
    }

    telegramMessage := TelegramMessage{}

    log.Println(string(body))

    if err := json.Unmarshal(body, &telegramMessage); err != nil {
        log.Println(err)
    }

    AnalyzePossibleCommand(&telegramMessage);
}

func PrepareWebserver() {
    http.HandleFunc("/", HandleHTTPCall)

    log.Println("Starting HTTP server")

    if _, err := os.Stat("./private.pem"); os.IsNotExist(err) {
        log.Fatal("private.pem not found in binaries location")
    }

    if _, err := os.Stat("./public.pem"); os.IsNotExist(err) {
        log.Fatal("private.pem not found in binaries location")
    }

    if err := http.ListenAndServeTLS("0.0.0.0:1337", "public.pem", "private.pem", nil); err != nil {
        log.Println("Was not able to stat the webserver");
        log.Fatal(err)
    }
}

func OpenDatabase() {
    log.Println("Trying to open database");

    db, err := sql.Open("sqlite3", "./1337.db")

	if err != nil {
		log.Fatal("Error opening database")
	}

    database = db

    CreateTableRegisteredChats();
    CreateTable1337Message();

    log.Println("Wohoo! That worked :-)")
}

func CreateTableRegisteredChats() {

    createStatement := `
    CREATE TABLE IF NOT EXISTS registeredChats (
        chatID INTEGER NOT NULL PRIMARY KEY
    )
    `;

    _, err := database.Exec(createStatement);

    if ( err != nil ) {
        log.Println("Was not able to create database table registeredChats");
        log.Panic(err)
    } else {
        log.Println("I maybe created database table registeredChats");
    }
}

func CreateTable1337Message() {

    createStatement := `
    CREATE TABLE IF NOT EXISTS messages (
        messageID INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        submitDate TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
        chatID INTEGER NOT NULL,
        username VARCHAR(255) NOT NULL
    )
    `;

    _, err := database.Exec(createStatement);

    if ( err != nil ) {
        log.Fatal("Was not able to create database table messages");
    } else {
        log.Println("I maybe created database table messages");
    }
}

func SanitizeChatMessage(message string) string {
    return strings.TrimPrefix(message, "/")
}

func AnalyzePossibleCommand(message *TelegramMessage) {
    log.Printf("Analyzing possible command \"%s\" from %s in chat with ID %d",
        message.GetMessage(),
        message.GetUsername(),
        message.GetChatID())

    switch SanitizeChatMessage(message.GetMessage()) {
        case BotCommand_Start:
            StartBot(message)

        case BotCommand_1337:
            Process1337Message(message)

        case BotCommand_Stats:
            SendStats(message)

        case BotCommand_Status:
            SendStatus(message)

        default:
            log.Println("I don't know the command. Doing nothing")
    }
}

func Broadcast1337() {

}

func SendStatus(message *TelegramMessage) {
    log.Println("Sending status");

    currentTime := GetCurrentTime()

    status := `
Hallo! Ich bin der 1337 Telegram Bot von @veloc1ty

Version: %s
Du bist: @%s
Serverzeit:
  Stunde:   %d
  Minute:   %d
  Sekunde:  %d
  Zeitzone: %s
    `

    message.Answer(fmt.Sprintf(status,
            Bot_Version,
            message.GetUsername(),
            currentTime.Hour(),
            currentTime.Minute(),
            currentTime.Second(),
            currentTime.Location().String()))
}

func SendStats(message *TelegramMessage) {
    log.Println("Processing stats request")

    query := `
    SELECT username, count(username) AS counter
    FROM messages
    WHERE chatID = %d
    GROUP BY username
    ORDER BY counter DESC
    `

    rows, err := database.Query(fmt.Sprintf(query, message.GetChatID()))

	if err != nil {
		log.Println(err)
	}

	defer rows.Close()

    statsString := "Hey %s, hier ist die Tabelle:\n\n"
    rowsFound := false

	for rows.Next() {
        rowsFound = true
		var username string
		var count int

		err = rows.Scan(&username, &count)

        if err != nil {
            log.Println(err)
		}

        statsString += fmt.Sprintf("%s -> %d", username, count)
	}

    if ( rowsFound ) {
        message.Answer(fmt.Sprintf(statsString, message.GetUsername()))
    } else {
        message.Answer("Bis jetzt ist die Liste noch leer")
    }
}

func StartBot(message *TelegramMessage) {
    log.Printf("I should register to this chat");

    log.Printf("Checking if I'm already registered")

    if ! IsBotAlreadyRegisteredForThisChat(message) {
        RegisterBotWithThisChat(message)

        message.Answer("Cool! Ich bin nun hier!")
    } else {
        message.Answer("Hey, ich bin doch schon hier!")
    }
}

func GetCurrentTime() time.Time {
    location, err := time.LoadLocation("Europe/Berlin")

    if err != nil {
        log.Printf("Error loading time location: %s", err)
    }

    return time.Now().In(location)
}

func IsIt1337() bool {

    currentTime := GetCurrentTime()

    log.Printf("Current time is %d:%d", currentTime.Hour(), currentTime.Minute())

    if currentTime.Hour() == 13 && currentTime.Minute() == 37 {
        log.Printf("It's 13:37!")

        return true
    } else {
        log.Printf("It's not 13:37 o'Clock!")

        return false
    }
}

func Process1337Message(message *TelegramMessage) {
    log.Printf("Analyzing a 1337 message")

    if IsIt1337() {
        if HasUserAlreadyVoted(message) {
            message.Answer(
                fmt.Sprintf("Sorry %s, aber du hast heute schon", message.GetUsername()))
        } else {
            AddSuccessfull1337Message(message)

            message.Answer(
                fmt.Sprintf("Hey %s, dein 13:37 wurde gezählt", message.GetUsername()))
        }

    } else {
        message.Answer(GetRandomInsult())
    }

}

func GetRandomInsult() string {
    return "Arschloch!"
}

func HasUserAlreadyVoted(message *TelegramMessage) bool {
    query := `
    SELECT messageID
    FROM messages
    WHERE DATE(submitDate) = DATE(CURRENT_DATE)
        AND chatID = %d
        AND username = '%s'
    `

    rows, err := database.Query(fmt.Sprintf(query, message.GetChatID(),
        message.GetUsername()))

    if err != nil {
        log.Println(err)
    }

    defer rows.Close()

	if rows.Next() {
        log.Printf("User %s has already voted", message.GetUsername())

        return true
    }

    log.Printf("User %s has not yet voded", message.GetUsername())

    return false
}

func AddSuccessfull1337Message(message *TelegramMessage) {

    insertStatement := fmt.Sprintf("INSERT INTO messages (chatID, username) VALUES(%d, '%s')",
        message.GetChatID(), message.GetUsername());

    _, err := database.Exec(insertStatement);

    if ( err != nil ) {
        log.Println("Was not able to registers the 1337");
    } else {
        log.Println("Successfully registered the 1337");
    }
}

func IsBotAlreadyRegisteredForThisChat(message *TelegramMessage) bool {
    var amount int
    rows, err := database.Query("SELECT count(*) FROM registeredChats WHERE chatID = " + strconv.Itoa(message.GetChatID()))

	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		err = rows.Scan(&amount)

		if err != nil {
			log.Fatal(err)
		}
	}

    if amount == 0 {
        log.Println("I'm not registered to this chat")
        return false
    } else {
        log.Println("I'm already registered to this chat")
        return true
    }
}

func RegisterBotWithThisChat(message *TelegramMessage) {
    log.Printf("Registering for this chat")

    _, err := database.Exec("INSERT INTO registeredChats (chatID) VALUES(" + strconv.Itoa(message.GetChatID()) + ")");

    if ( err != nil ) {
        errorMessage := "Was not able to register to this chat";

        message.Answer(errorMessage);
        log.Println(errorMessage);
    } else {
        successMessage := "Successfully registered to this chat";

        message.Answer(successMessage);
        log.Println(successMessage);
    }
}
