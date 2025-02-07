package bot

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
	"io/ioutil"
	"strconv"
	"regexp"
)

func getProfile(message *discordgo.MessageCreate, session *discordgo.Session) {
	if strings.HasPrefix(message.Content, "!profile") {
		// split into command and username
		parts := strings.Split(message.Content, " ")
		if len(parts) != 2 {
			session.ChannelMessageSend(message.ChannelID, "Usage: !profile <101weiqi_username>")
			return
		}

		// construct the profile URL
		user := parts[1]
		profileURL := fmt.Sprintf("https://www.101weiqi.com/u/%s/", user)

		// disallow self-lookup
		if (user == USERNAME) {
			session.ChannelMessageSend(message.ChannelID, "You can see your own stats at https://www.101weiqi.com/guan/")
			return
		}

		// get profile
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

		// parse profile html
		profile_html, profile_html_error := goquery.NewDocumentFromReader(profile_get_response.Body)
		if profile_html_error != nil {
			log.Fatal("Error parsing profile HTML:", profile_html_error)
			return
		}

		// extract the userid
		id, found_id := profile_html.Find(".staruser").Attr("userid")
		if found_id == false {
			log.Fatal("Error finding profile ID:", profile_html_error)
			return
		}

		// friend user
		friend(1, user, id, message, session)

		// start the results code block
		results := fmt.Sprintf("```User: %s\n\nSkill Test Results:\n", user)

		// initialize tracked stats
		var placements, perfect int
		best := "N/A"

		// get stats per level
		leaderboardURL := "https://www.101weiqi.com/guan/pop/"
		for i := 1; i <= 22; i++ {
			// construct URL for current level
			tempURL := leaderboardURL + strconv.Itoa(i) + "/"

			// get temp urls
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

			// temp url body as text
			temp_url_get_response_body, temp_url_get_response_body_read_error := ioutil.ReadAll(temp_url_get_response.Body)
			if temp_url_get_response_body_read_error != nil {
				log.Fatal(temp_url_get_response_body_read_error)
			}
			temp_url_get_response_body_text := string(temp_url_get_response_body)

			// finds user data for current level from temp url body
			regex_for_user := regexp.MustCompile(fmt.Sprintf(`"%s",\s*(\S+)\s+(\S+)\s*(\S+)\s+(\S+)\s*(\S+)\s+(\S+)`, user))
			match := regex_for_user.FindStringSubmatch(temp_url_get_response_body_text)

			// checks if user's name only appears on the friends tab
			freq := regex_for_user.FindAllString(temp_url_get_response_body_text, -1)

			// if user's name appears more than once, user is on the leaderboard for this level
			leaderboard := false
			if len(freq) > 1 {
				leaderboard = true
			}

			// converts pop level to kyu/dan ranks
			var level string
			if (i <= 15) {
				level = strconv.Itoa(16 - i) + "K"
			} else {
				level = strconv.Itoa(i - 15) + "D"
			}

			results += level + ": " + "\t"

			// maintain alignment
			if (i > 6) {
				results += " "
			}
			
			if match != nil {
				// base stats
				solved_correctly := strings.Trim(match[6], ",")
				time_spent := strings.Trim(match[2], ",")

				// add solved count to result line, maintaining alignment
				results += solved_correctly + "/10 " + "\t"
				for i := len(solved_correctly); i <= 2; i++ {
					results += " "
				}

				// update perfect count
				if solved_correctly == "10" {
					perfect += 1
				}

				// add time_spent to result line, maintaining alignment
				results += time_spent + " seconds"

				// update the  highest level solved
				best = level

				// leaderboard updates
				if leaderboard == true {
					// increment leaderboard placement count
					placements += 1
					
					// maintain alignment
					for i := len(time_spent); i < 3; i++ {
						results += " "
					}

					// add leaderboard text to results line
					results += "\t(Global Top 100)"
				}
			} else {
				// no results found for current level, update results line
				results += "N/A " + "\t" + "   N/A"
			}

			// new result line for each level 
			results += "\n"
		}


		// add tracked stats to result
		results += "\n"
		results += "Highest Level Passed: " + best + "\n"
		results += "Perfect Scores: " + strconv.Itoa(perfect) + "\n"
		results += "Leaderboard Placements: " + strconv.Itoa(placements) + "\n"
		results += "```"

		// print result
		_, print_result_error := session.ChannelMessageSend(message.ChannelID, results)
		if print_result_error != nil {
			log.Fatal("Error printing results:", print_result_error)
		}

		// unfriend user
		// friend(0, user, id, message, session)
		
	} else {
		session.ChannelMessageSend(message.ChannelID, "!profile goes at the start of the command :(")
		return
	}

	return
}
