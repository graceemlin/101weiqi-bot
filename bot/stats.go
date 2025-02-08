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

var LEADERBOARDURL string

type Statistic struct {
	Correct string
	Time string
	Leaderboard bool
}

func get_stats(message *discordgo.MessageCreate, session *discordgo.Session) {
	if strings.HasPrefix(message.Content, "!profile") {
		// split into command and username
		parts := strings.Split(message.Content, " ")
		if len(parts) != 2 {
			session.ChannelMessageSend(message.ChannelID, "Usage: !profile <101weiqi_username>")
			return
		}

		// check if user profile exists
		user := parts[1]
		is_valid, id := valid_profile(message, session, user)
		if is_valid == false {
			return
		}
		
		// friend user
		friend(1, user, id, message, session)

		stats_border := "===================================================="
		// start the results code block
		results := fmt.Sprintf("```diff\n%s\nSkill Test Results: %s\n%s\n", stats_border, user, stats_border)

		// initialize tracked stats
		var placements, perfect int
		best := "N/A"

		// initialize stats array
		var user_stats[23] Statistic 

		// fetch stats per level
		for i := 1; i <= 22; i++ {
			var current_line string
			
			current_statistic := user_stats[i]
			found := populate_statistic(&current_statistic, user, i)
			if found == false {
				break;
			}

			// converts pop level to kyu/dan ranks
			level := pop_to_level(i)

			// update the  highest level solved
			best = level
			
			current_line += level + ": " + "\t"
			
			// maintain alignment
			if (i > 6) {
				current_line += " "
			}

			// add solved count to current line, maintaining alignment
			current_line += current_statistic.Correct + "/10 " + "\t"
			for i := len(current_statistic.Correct); i <= 2; i++ {
				current_line += " "
			}

			// add time_spent to result line, maintaining alignment
			current_line += current_statistic.Time + " seconds"
			
			// maintain alignment
			for i := len(current_statistic.Time); i <= 3; i++ {
				current_line += " "
			}

			if current_statistic.Correct == "10" {
				perfect += 1
			}
			
			if current_statistic.Leaderboard == true {
				placements += 1
				
				// add leaderboard text to results line
				current_line += "\t(Global Top 100)"
			}
			
			// new result line for each level
			current_line += "\n"
			results += current_line
		}


		// add tracked stats to result
		results += stats_border
		results += "\n"
		results += "Highest Level Passed: " + best + "\n"
		results += "Perfect Scores: " + strconv.Itoa(perfect) + "\n"
		results += "Leaderboard Placements: " + strconv.Itoa(placements) + "\n"
		results += stats_border
		results += "\n"
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


func valid_profile(message *discordgo.MessageCreate, session *discordgo.Session, user string) (exists bool, userid string) {
	// construct the profile URL
	profileURL := fmt.Sprintf("https://www.101weiqi.com/u/%s/", user)
	
	// disallow self-lookup
	if (user == USERNAME) {
		session.ChannelMessageSend(message.ChannelID, "You can see your own stats at https://www.101weiqi.com/guan/")
		return false, ""
	}
	
	// GET profile
	profile_get_response, profile_get_response_error := client.Get(profileURL)
	if profile_get_response_error != nil {
		log.Println("Error fetching URL:", profile_get_response_error)
		session.ChannelMessageSend(message.ChannelID, "Error fetching profile.")
		return false, ""
	}
	defer profile_get_response.Body.Close()
	
	if profile_get_response.StatusCode != 200 {
		log.Printf("Status code error: %d %s", profile_get_response.StatusCode, profile_get_response.Status)
		session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("101weiqi profile for the user \"%s\" not found. Please check spelling/capitalization.", user))
		return false, ""
	}
	
	// parse profile HTML
	profile_html, profile_html_error := goquery.NewDocumentFromReader(profile_get_response.Body)
	if profile_html_error != nil {
		log.Fatal("Error parsing profile HTML:", profile_html_error)
		return false, ""
	}

	// extract the userid
	id, found_id := profile_html.Find(".staruser").Attr("userid")
	if found_id == false {
		log.Fatal("Error finding profile ID:", profile_html_error)
		return false, ""
	}

	return true, id
}

func pop_to_level (pop int) string {
	if (pop <= 15) {
		return strconv.Itoa(16 - pop) + "K"
	} else {
		return strconv.Itoa(pop - 15) + "D"
	}
}


func populate_statistic (stat *Statistic, user string, i int) (bool) {
	// construct temp URL for current level
	tempURL := LEADERBOARDURL + strconv.Itoa(i) + "/"
	
	// GET temp URL
	temp_url_get_response, temp_url_get_response_error := client.Get(tempURL)
	if temp_url_get_response_error != nil {
		log.Fatal("Error fetching temp URL:", temp_url_get_response_error)
	}
	if temp_url_get_response.StatusCode != 200 {
		log.Fatal("Fetching temp URL gives status code error: %d %s", temp_url_get_response.StatusCode, temp_url_get_response.Status)
	}
	defer temp_url_get_response.Body.Close()
	
	// temp URL body as text
	temp_url_get_response_body, temp_url_get_response_body_read_error := ioutil.ReadAll(temp_url_get_response.Body)
	if temp_url_get_response_body_read_error != nil {
		log.Fatal(temp_url_get_response_body_read_error)
	}
	temp_url_get_response_body_text := string(temp_url_get_response_body)
	
	// finds user data for current level from temp URL body
	regex_for_user := regexp.MustCompile(fmt.Sprintf(`"%s",\s*(\S+)\s+(\S+)\s*(\S+)\s+(\S+)\s*(\S+)\s+(\S+)`, user))
	match := regex_for_user.FindStringSubmatch(temp_url_get_response_body_text)
	
	// checks if user's name only appears on the friends tab
	freq := regex_for_user.FindAllString(temp_url_get_response_body_text, -1)
	
	// if user's name appears more than once, user is on the leaderboard for this level
	if len(freq) > 1 {
		stat.Leaderboard = true
	}
	
	if match != nil {
		// base stats
		stat.Correct = strings.Trim(match[6], ",")
		stat.Time = strings.Trim(match[2], ",")
		return true
	} else {
		// no results found for current level, update results line
		return false
	}
}
