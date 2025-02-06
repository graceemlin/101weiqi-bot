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

var BotToken, LoginURL, HomeURL string

func Run() {
	// create a discord session
	session, session_error := discordgo.New("Bot " + BotToken)
	if session_error != nil {
		log.Fatal("Error creating session:", session_error)
	}

	// add a event handler
	session.AddHandler(newMessage)
	session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	// open session
	session.Open()
	defer session.Close() // close session, after function termination

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

	// create a cookiejar to store cookies
	jar, jar_error := cookiejar.New(nil)
	if jar_error != nil {
		log.Fatal("Error creating cookiejar:", jar_error)
	}
	
	client := &http.Client{
		Jar: jar,
	}

	//initial login
	login(client, message, session)
	
	home_url_object, home_url_object_error := url.Parse(HomeURL)
	if home_url_object_error != nil {
		log.Fatal("Error accessing 101weiqi homepage:", home_url_object_error)
	}
	
	cookies := jar.Cookies(home_url_object)
		
	switch {
	case strings.Contains(message.Content, "!profile"):
		session_active := false
		for _, cookie := range cookies {
			if cookie.Name == "sessionid" {				
				session_active = true
			}
		}

		if (session_active == false) {
			fmt.Println("new session")
			login(client, message, session)
			
			home_url_object_new, home_url_object_new_error := url.Parse(HomeURL)
			if home_url_object_new_error != nil {
				log.Fatal("Error accessing 101weiqi homepage for new session:", home_url_object_new_error)
			}

			cookies = jar.Cookies(home_url_object_new)

			login_successful := false
			for _, cookie := range cookies {
				if cookie.Name == "sessionid" {				
					login_successful = true
					fmt.Println(cookie.Value)
				}
			}

			if (login_successful == false) {
				session.ChannelMessageSend(message.ChannelID, "the cookies aren't cookie-ing, please fix me :(")
				log.Fatal(nil)
			}		
		}
		
		getProfile(client, message, session)		
	}
}
