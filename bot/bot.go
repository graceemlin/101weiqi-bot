package bot

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"github.com/bwmarrin/discordgo"
)

var BOTTOKEN, LOGINURL, HOMEURL string
var client *http.Client

func Init() {
	// create a cookiejar to store cookies
	jar, jar_error := cookiejar.New(nil)
	if jar_error != nil {
		log.Fatal("Error creating cookiejar:", jar_error)
	}

	// set Jar field of Client struct to newly created cookiejar
	client = &http.Client{
		Jar: jar,
	}
	
	// initial 101weiqi login
	login()
}

func Run() {
	// create a Discord session
	session, session_error := discordgo.New("Bot " + BOTTOKEN)
	if session_error != nil {
		log.Fatal("Error creating session:", session_error)
	}

	// add an event handler
	session.AddHandler(newMessage)
	session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	// open session
	session.Open()
	// close session, after function termination
	defer session.Close()

	// keep bot running until there is an OS interruption (ctrl + C)
	fmt.Println("Bot running....")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func newMessage(session *discordgo.Session, message *discordgo.MessageCreate) {
	// prevent bot from responding to its own message
	if message.Author.ID == session.State.User.ID {
		return
	}

	// parse home url for cookie checks
	home_url_object, home_url_object_error := url.Parse(HOMEURL)
	if home_url_object_error != nil {
		log.Fatal("Error creating 101weiqi homepage URL object", home_url_object_error)
	}
	
	// check for an active session
	cookies := client.Jar.Cookies(home_url_object)
	session_active := false
	for _, cookie := range cookies {
		if cookie.Name == "sessionid" {				
			session_active = true
		}
	}
	
	// restart session if needed
	if (session_active == false) {
		// note when new sessions are needed
		fmt.Println("new session")
		
		// login to 101weiqi
		login()
		
		// verify 101weiqi login success
		cookies = client.Jar.Cookies(home_url_object)
		login_successful := false
		for _, cookie := range cookies {
			if cookie.Name == "sessionid" {				
				login_successful = true
			}
		}
		
		// if 101weiqi login is unsuccessful, terminate
		if (login_successful == false) {
			session.ChannelMessageSend(message.ChannelID, "the cookies aren't cookie-ing, please fix me :(")
			log.Fatal(nil)
		}		
	}
	
	switch {
	case strings.HasPrefix(message.Content, "!profile"):		
		get_stats(message, session)
	case strings.HasPrefix(message.Content, "!compare"):
		get_stats(message, session)
	}
}
