package bot

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var USERNAME, PASSWORD string

func login() {
	// GET login page
	login_get_response, login_get_error := client.Get(LOGINURL)
	if login_get_error != nil {
		log.Fatal("Error fetching login page:", login_get_error)
	}
	defer login_get_response.Body.Close()

	// find csrftoken
	var csrftoken string
	for _, cookie := range login_get_response.Cookies() {
		if cookie.Name == "csrftoken" {
			csrftoken = cookie.Value
		}
	}

	if csrftoken == "" {
		log.Fatal("csrftoken not found on the login page.", nil)
		return
	}

	// parse login page html
	login_html, login_html_error := goquery.NewDocumentFromReader(login_get_response.Body)
	if login_html_error != nil {
		log.Fatal("Error parsing login HTML:", login_html_error)
		return
	}
	
	// extract the csrfmiddlewaretoken
	csrfmiddlewaretoken, found_csrfmiddlewaretoken := login_html.Find("[name=csrfmiddlewaretoken]").Attr("value")
	if found_csrfmiddlewaretoken == false {
		log.Fatal("Could not find csrfmiddlewaretoken on login page.")
		return
	}

	// create form data for login POST
	formData := url.Values {
		"csrfmiddlewaretoken": {csrfmiddlewaretoken},
		"source":              {"index_nav"},
		"form_username":       {USERNAME},
		"form_password":       {PASSWORD},
	}

	// construct login POST request
	login_post_request, login_post_request_creation_error := http.NewRequest(http.MethodPost, LOGINURL, strings.NewReader(formData.Encode()))
	if login_post_request_creation_error != nil {
		log.Fatal("Error creating login POST request:", login_post_request_creation_error)
	}

	login_post_request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	login_post_request.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	login_post_request.Header.Set("Accept-Language", "en-US,en;q=0.9")	
	login_post_request.Header.Set("Cache-Control", "no-cache")
	login_post_request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	login_post_request.Header.Set("Origin", "https://www.101weiqi.com")
	login_post_request.Header.Set("Pragma", "no-cache")
	login_post_request.Header.Set("Priority", "u=0, i")
	login_post_request.Header.Set("Referer", "https://www.101weiqi.com/")
	login_post_request.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	login_post_request.Header.Set("Sec-Fetch-Dest", "document")
	login_post_request.Header.Set("Sec-Fetch-Mode", "navigate")
	login_post_request.Header.Set("Sec-Fetch-Site", "same-origin")
	login_post_request.Header.Set("Sec-Fetch-User", "?1")
	login_post_request.Header.Set("Upgrade-Insecure-Requests", "1")

	// send login POST request
	login_post_response, login_post_error := client.Do(login_post_request)
	if login_post_error != nil {
                log.Fatal("Login POST request failed.", nil)           
        }
	defer login_post_response.Body.Close()
	
	if login_post_response.Request.URL.String() == HOMEURL {
		fmt.Println("Successfully logged in!")
	} else if login_post_response.Request.URL.String() == LOGINURL {
		log.Fatal("Failed to login. Check your username/password")
	} else {
		log.Fatal("Failed to login. Possible throttle on POSTs from this account")
	}
}
