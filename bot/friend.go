package bot

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func friend(action int, username string, id string, message *discordgo.MessageCreate, session *discordgo.Session) {
	// construct urls
	friendURL := "https://www.101weiqi.com/u/" + username + "/"
	attionURL := "https://www.101weiqi.com/attionuser/"

	// get friend profile
	friend_get_response, friend_get_error := client.Get(friendURL)
	if friend_get_error != nil {
		log.Fatal("Error fetching login page:", friend_get_error)
	}
	defer friend_get_response.Body.Close()

	if friend_get_response.StatusCode != 200 {
		log.Fatalf("Status code error: %d %s", 	friend_get_response.StatusCode, friend_get_response.Status)
	}

	// find csrftoken
	var csrftoken string
	for _, cookie := range friend_get_response.Cookies() {
		if cookie.Name == "csrftoken" {
			csrftoken = cookie.Value
		}
	}
	
	if csrftoken == "" {
		log.Fatal("csrftoken not found on friend's profile page", nil)
	}

	// parse friend profile html
	friend_html, friend_html_error := goquery.NewDocumentFromReader(friend_get_response.Body)
	if friend_html_error != nil {
		log.Fatal("Error parsing HTML:", friend_html_error)
	}
	
	// extract the csrfmiddlewaretoken
	csrfmiddlewaretoken, found_csrfmiddlewaretoken := friend_html.Find("[name=csrfmiddlewaretoken]").Attr("value")
	if  found_csrfmiddlewaretoken == false {
		log.Fatal("Could not parse csrfmiddlewaretoken.", nil)
	}

	// create form data for POST
	formData := url.Values {
		"userid":              {id},
		"attion":              {string(action + '0')}, 
		"csrfmiddlewaretoken": {csrfmiddlewaretoken},
	}

	// construct friend POST request
	friend_post_request, friend_post_error := http.NewRequest(http.MethodPost, attionURL, strings.NewReader(formData.Encode()))
	if friend_post_error != nil {
		log.Fatal("Error creating friend POST request:", friend_post_error)
	}

	friend_post_request.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	friend_post_request.Header.Set("Accept-Language", "en-US,en;q=0.9")	
	friend_post_request.Header.Set("Cache-Control", "no-cache")
	friend_post_request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	friend_post_request.Header.Set("Origin", "https://www.101weiqi.com")
	friend_post_request.Header.Set("Referer", "https://www.101weiqi.com/")
	friend_post_request.Header.Set("Pragma", "no-cache")
	friend_post_request.Header.Set("Priority", "u=1, i")
	friend_post_request.Header.Set("Sec-Fetch-Mode", "cors")
	friend_post_request.Header.Set("Sec-Fetch-Site", "same-origin")
	friend_post_request.Header.Set("X-Requested-With", "XMLHttpRequest")

	// send friend POST request
	friend_post_response, friend_post_request_error := client.Do(friend_post_request)
	if friend_post_request_error != nil {
                log.Fatal("Friend POST request failed.")           
        } else {
		if (action == 1) {
			fmt.Println("Successfully friended!")
		} else {
			fmt.Println("Successfully unfriended!")
		}
	}
	
	defer friend_post_response.Body.Close()
}
