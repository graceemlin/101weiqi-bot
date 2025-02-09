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

			// converts pop level to kyu/dan ranks
			level := pop_to_level(i)
			current_line += level + ": " + "\t"
			
			// maintain alignment
			if (i > 6) {
				current_line += " "
			}

			current_statistic := user_stats[i]
			found := populate_statistic(&current_statistic, user, i)
			if found == false {
				current_line += current_statistic.Correct + "\t\t" + current_statistic.Time + "\n"
				results += current_line
				continue
			}

			// update the  highest level solved
			best = level
			
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
		
	} else if strings.HasPrefix(message.Content, "!compare") {
		// split into command and usernames
		parts := strings.Split(message.Content, " ")
		if len(parts) != 3 {
			session.ChannelMessageSend(message.ChannelID, "Usage: !compare <101weiqi_username>  <101weiqi_username>")
			return
		}

		user1 := parts[1]
		user2 := parts[2]

		// check if user profiles exist
		user1_is_valid, id1 := valid_profile(message, session, user1)
		user2_is_valid, id2 := valid_profile(message, session, user2)
		if user1_is_valid == false || user2_is_valid == false {
			session.ChannelMessageSend(message.ChannelID, "Please try again with two valid 101weiqi usernames")
			return
		}

		// friend users
		friend(1, user1, id1, message, session)
		friend(1, user2, id2, message, session)

		compare_border := "========================================================="
		// start the results code block
		compare_results := fmt.Sprintf("```diff\n%s\nSkill Test Comparison: %s vs %s\n%s\n", compare_border, user1, user2, compare_border)

		// initialize tracked stats
		var placements1, perfect1 int
		best1 := "N/A"
		var placements2, perfect2 int
		best2 := "N/A"

		var wins, losses, ties, nocontest int
		
		// initialize stats arrays
		var user1_stats[23] Statistic 
		var user2_stats[23] Statistic

		// fetch stats per level
		for i := 1; i <= 22; i++ {
			var current_line string

			// converts pop level to kyu/dan ranks
			level := pop_to_level(i)
			current_line += level + ": "
			
			// maintain alignment
			if (i > 6) {
				current_line += " "
			}

			statistic1 := user1_stats[i]
			one_found := populate_statistic(&statistic1, user1, i)
			if one_found == false {
				current_line += statistic1.Correct + "   " + statistic1.Time + "            |"
			} else {
				// update the  highest level solved
				best1 = level
				
				// add solved count to current line, maintaining alignment
				current_line += statistic1.Correct + "/10"
				for i := len(statistic1.Correct); i <= 2; i++ {
					current_line += " "
				}
				
				// add time_spent to result line, maintaining alignment
				current_line += statistic1.Time + "s"
				
				// maintain alignment
				for i := len(statistic1.Time); i <= 3; i++ {
					current_line += " "
				}
				
				if statistic1.Correct == "10" {
					perfect1 += 1
				}
				
				if statistic1.Leaderboard == true {
					placements1 += 1
					
					// add leaderboard text to results line
					current_line += "(Top 100) |"
				} else {
					current_line += "          |"
				}
			}

			statistic2 := user2_stats[i]
			two_found := populate_statistic(&statistic2, user2, i)
			if two_found == false {
				current_line += " " + statistic2.Correct + "   " + statistic2.Time + "        "
			} else {
				// update the  highest level solved
				best2 = level
				
				// add solved count to current line, maintaining alignment
				current_line += " "+ statistic2.Correct + "/10"
				for i := len(statistic2.Correct); i <= 2; i++ {
					current_line += " "
				}
				
				// add time_spent to result line, maintaining alignment
				current_line += statistic2.Time + "s"
				
				// maintain alignment
				for i := len(statistic2.Time); i <= 3; i++ {
					current_line += " "
				}
				
				if statistic2.Correct == "10" {
					perfect2 += 1
				}
				
				if statistic2.Leaderboard == true {
					placements2 += 1
					
					// add leaderboard text to results line
					current_line += "(Top 100)"
				}
			}

			comparison := compare_statistics(statistic1, statistic2)

			save := false
			switch {
			case comparison == "+  ":
				wins += 1
			case comparison == "-  ":
				losses += 1
			case comparison == "***":
				nocontest += 1
				save = true
			case comparison == "!  ":
				ties += 1
			}

			current_line = comparison + " " + current_line

			if save == false {
				for z := len(current_line); z < len(compare_border); z++ {
					current_line += " "
				}
			}

			// new result line for each level
			current_line += "\n"
			
			compare_results += current_line
		}
		
		compare_results += compare_border + "\n"
		footer1 := "Highest Level Passed: " + best1
		for pad1 := len(footer1); pad1 < 30; pad1++ {
			footer1 += " "
		}
		footer1 += "| Highest Level Passed: " + best2
		
		footer2 := "Perfect Scores: " + strconv.Itoa(perfect1)
		for pad2 := len(footer2); pad2 < 30; pad2++ {
			footer2 += " "
		}
		footer2 += "| Perfect Scores: " + strconv.Itoa(perfect2)
		
		footer3 := "Leaderboard Placements: " + strconv.Itoa(placements1)
		for pad3 := len(footer3); pad3 < 30; pad3++ {
			footer3 += " "
		}
		footer3 += "| Leaderboard Placements: " + strconv.Itoa(placements2)

		compare_results += footer1 + "\n" + footer2 + "\n" + footer3 + "\n"
		compare_results += compare_border + "\n"
		
		compare_results += "+ Wins: " + strconv.Itoa(wins) + "\n"
		compare_results += "- Losses: " + strconv.Itoa(losses) + "\n"
		compare_results += "! Ties: " + strconv.Itoa(ties) + "\n"
		compare_results += "*** No Contest: " + strconv.Itoa(nocontest)
		
		compare_results += "\n"
		compare_results += "```"

		// print result
		_, print_result_error := session.ChannelMessageSend(message.ChannelID, compare_results)
		if print_result_error != nil {
			log.Fatal("Error printing results:", print_result_error)
		}

	}

	return
}

func compare_statistics(user1 Statistic, user2 Statistic) (string) {
	if (user1.Correct == "N/A" && user2.Correct == "N/A") {
		return "***"
	} else if user1.Correct == "N/A" {
		return "-  "
	} else if user2.Correct == "N/A" {
		return "+  "
	}

	time1, time_err1 := strconv.Atoi(user1.Time)
	if time_err1 != nil {
            log.Fatal("Error: the math isn't mathing", time_err1)
        }
	
	time2, time_err2 := strconv.Atoi(user2.Time)
	if time_err2 != nil {
            log.Fatal("Error: the math isn't mathing", time_err2)
        }

	correct1, correct_err1 := strconv.Atoi(user1.Correct)
	if correct_err1 != nil {
            log.Fatal("Error: the math isn't mathing", correct_err1)
        }
	
	correct2, correct_err2 := strconv.Atoi(user2.Correct)
	if correct_err2 != nil {
            log.Fatal("Error: the math isn't mathing", correct_err2)
        }
	
	if (correct1 == correct2) {
		if time1 == time2 {
			return "!  "
		} else if time1 < time2 {
			return "+  "
		} else {
			return "-  "
		}
	}

	if (correct1 > correct2) {
		return "+  "
	} else {
		return "-  "
	}
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


func populate_statistic(stat *Statistic, user string, i int) (bool) {
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
		stat.Correct = "N/A"
		stat.Time = "N/A"
		return false
	}
}
