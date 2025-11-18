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

	// parse login page HTML
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
	formData := url.Values{
		"csrfmiddlewaretoken": {csrfmiddlewaretoken},
		"username":            {USERNAME},
		"password":            {PASSWORD},
	}

	// construct login POST request
	login_post_request, login_post_request_creation_error := http.NewRequest(http.MethodPost, LOGINWQURL, strings.NewReader(formData.Encode()))
	if login_post_request_creation_error != nil {
		log.Fatal("Error creating login POST request:", login_post_request_creation_error)
	}

	login_post_request.Header.Set("Accept", "application/json, text/plain, */*")
	login_post_request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	login_post_request.Header.Set("Origin", "https://www.101weiqi.com")
	login_post_request.Header.Set("Referer", "https://www.101weiqi.com/login/")
	login_post_request.Header.Set("X-Requested-With", "XMLHttpRequest")

	// send login POST request
	login_post_response, login_post_error := client.Do(login_post_request)
	if login_post_error != nil {
		log.Fatal("Login POST request failed.", nil)
	}
	defer login_post_response.Body.Close()

	login_successful := false
	for _, cookie := range login_post_response.Cookies() {
		if cookie.Name == "sessionid" {
			login_successful = true
		}
	}

	if login_successful {
		fmt.Println("Successfully logged in!")
	} else {
		log.Fatal("Failed to login. Possible throttle on POSTs from this account.")
	}
}
