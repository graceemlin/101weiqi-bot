package bot

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"
	"log"
	"net/http"
	"strings"
	"io/ioutil"
	"strconv"
	"regexp"
)

func getProfile(client *http.Client, message *discordgo.MessageCreate, session *discordgo.Session) {
	if strings.HasPrefix(message.Content, "!profile") {
		parts := strings.Split(message.Content, " ") // Split into command and username
		if len(parts) != 2 {
			session.ChannelMessageSend(message.ChannelID, "Usage: !profile <101weiqi_username>")
			return
		}
		
		user := parts[1]
		profileURL := fmt.Sprintf("https://www.101weiqi.com/u/%s/", user) // Construct the profile URL

		if (user == USERNAME) {
			session.ChannelMessageSend(message.ChannelID, "You can see your own stats at https://www.101weiqi.com/guan/")
			return
		}
		
		profile_get_response, profile_get_response_error := client.Get(profileURL)
		if profile_get_response_error != nil {
			log.Println("Error fetching URL:", profile_get_response_error)
			session.ChannelMessageSend(message.ChannelID, "Error fetching profile.")
			return
		}
		defer profile_get_response.Body.Close()
		
		if profile_get_response.StatusCode != 200 {
			log.Printf("Status code error: %d %s", profile_get_response.StatusCode, profile_get_response.Status)
			session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Profile not found or website error: %d", profile_get_response.StatusCode))
			return
		}

		profile_html, profile_html_error := goquery.NewDocumentFromReader(profile_get_response.Body)
		if profile_html_error != nil {
			log.Fatal("Error parsing profile HTML:", profile_html_error)
			return
		}

		// Extract the userid
		id, found_id := profile_html.Find(".staruser").Attr("userid")
		if found_id == false {
			log.Fatal("Error finding profile ID:", profile_html_error)
			return
		}

		// Friend user
		friend(1, user, id, client, message, session)

		// Start the results code block
		results := fmt.Sprintf("```User: %s\n\nSkill Test Results:\n", user)

		// Initialize tracked stats
		var placements, perfect int
		best := "N/A"

		// Get stats per level
		leaderboardURL := "https://www.101weiqi.com/guan/pop/"
		for i := 1; i <= 22; i++ {
			//Construct URL for current level
			tempURL := leaderboardURL + strconv.Itoa(i) + "/"
			
			temp_url_get_response, temp_url_get_response_error := client.Get(tempURL)
			if temp_url_get_response_error != nil {
				log.Fatal("Error fetching temp URL:", temp_url_get_response_error)
				return
			}
			if temp_url_get_response.StatusCode != 200 {
				log.Fatal("Fetching temp URL gives status code error: %d %s", temp_url_get_response.StatusCode, temp_url_get_response.Status)
				return
			}
			defer temp_url_get_response.Body.Close()

			temp_url_get_response_body, temp_url_get_response_body_read_error := ioutil.ReadAll(temp_url_get_response.Body)
			if temp_url_get_response_body_read_error != nil {
				log.Fatal(temp_url_get_response_body_read_error)
			}

			temp_url_get_response_body_text := string(temp_url_get_response_body)
			regex_for_user := regexp.MustCompile(fmt.Sprintf(`"%s",\s*(\S+)\s+(\S+)\s*(\S+)\s+(\S+)\s*(\S+)\s+(\S+)`, user))
			match := regex_for_user.FindStringSubmatch(temp_url_get_response_body_text)
			freq := regex_for_user.FindAllString(temp_url_get_response_body_text, -1)

			leaderboard := false
			if len(freq) > 1 {
				leaderboard = true
			}
			
			var level string
			if (i <= 15) {
				level = strconv.Itoa(16 - i) + "K"
			} else {
				level = strconv.Itoa(i - 15) + "D"
			}
			
			results += level + ": " + "\t"
			if (i > 6) {
				results += " "
			}
			
			if match != nil {
				match[6] = strings.Trim(match[6], ",")
				match[2] = strings.Trim(match[2], ",")

				results += match[6] + "/10 " + "\t"

				if match[6] == "10" {
					perfect += 1
				}
				
				for i := len(match[6]); i <= 2; i++ {
					results += " "
				}

				results += match[2] + " seconds"
				best = level
			} else {
				results += "N/A " + "\t" + "   N/A"
			}

			if leaderboard == true {
				for i := len(match[2]); i < 3; i++ {
					results += " "
				}

				placements += 1
				results += "\t(Global Top 100)"
			}
			
			results += "\n"
		}


		results += "\n"
		results += "Highest Level Passed: " + best + "\n"
		results += "Perfect Scores: " + strconv.Itoa(perfect) + "\n"
		results += "Leaderboard Placements: " + strconv.Itoa(placements) + "\n"
		results += "```"

		_, print_result_error := session.ChannelMessageSend(message.ChannelID, results)
		if print_result_error != nil {
			log.Fatal("Error printing results:", print_result_error)
		}

		// Unfriend user
		// friend(0, user, id, client, message, session)
		
	}

	return
}
