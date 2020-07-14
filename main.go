package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	libgiphy "github.com/sanzaru/go-giphy"
)

// Gets token from the call to start
func init() {
	flag.StringVar(&discordToken, "d", "", "Bot Token")
	flag.StringVar(&giphyToken, "g", "", "Giphy Token")
	flag.StringVar(&tenorToken, "t", "", "Tenor Token")
	flag.Parse()

	log.Println("Discord: ", discordToken)
	log.Println("Giphy: ", giphyToken)
	log.Println("Tenor: ", tenorToken)

	if discordToken == "" {
		flag.Usage()
		log.Fatal("No token provided. Bot can't be authenticated.")
	}
	giphy = libgiphy.NewGiphy(giphyToken)
}

var discordToken string
var giphyToken string
var tenorToken string
var giphy *libgiphy.Giphy
var prefix string = "?"

func main() {

	var err error

	// Create an new discord session with the provided token
	dg, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		log.Fatal("Error creating discord session: ", err)
	}

	dg.AddHandler(messageCreate)

	// Open a websocket to begin listening to discord
	err = dg.Open()
	if err != nil {
		log.Fatal("Error opening Discord session. ", err)
	}

	// Wait until singal to terminate is given
	log.Println("Bot is now running. Press CTRL-C to exit")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Close down the session if terminated
	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore messages from itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	var url string
	if strings.HasPrefix(m.Content, prefix+"giphy") {
		// Isolates the search criteria
		search := strings.TrimSpace(strings.TrimPrefix(m.Content, (prefix + "giphy ")))
		if search == "random" {
			url = getRandomGiphy()
		} else {
			url = getGiphy(search)
		}

	} else if strings.HasPrefix(m.Content, prefix+"tenor") {
		search := strings.TrimSpace(strings.TrimPrefix(m.Content, (prefix + "tenor ")))
		url = getTenor(search)
	}
	s.ChannelMessageSend(m.ChannelID, url)
}

func getRandomGiphy() string {
	log.Println("Getting random gif from giphy")
	gif, err := giphy.GetRandom("")
	if err != nil {
		log.Println("Error:", err)
	}
	return gif.Data.Url

}

func getGiphy(query string) string {
	log.Println("Getting gif with tag:", query, " from giphy")
	gif, err := giphy.GetSearch(query, 1, -1, "", "", false)
	if err != nil {
		log.Println("Error:", err)
	}
	return gif.Data[0].Url
}

func getTenor(query string) string {
	log.Println("Getting gif with tag: ", query, "from tenor")
	resp, err := http.Get(("https://api.tenor.com/v1/search?q=" +
		query + "&key=" + tenorToken + "&limit=1"))

	if err != nil {
		log.Println("Error occurred on http request. ", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error occurred when reading response ", err)
	}
	fmt.Println(body)
	return "https://tenor.com/view/kstr-kochstrasse-work-progress-concept-gif-16243141"
}
