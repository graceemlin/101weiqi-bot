package bot

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"
)

var LEADERBOARDURL string

type Statistic struct {
	Correct     string
	Time        string
	Leaderboard bool
}

func get_profile_stats(message *discordgo.MessageCreate, session *discordgo.Session) {
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

		// set border string
		profile_stats_border := "===================================================="

		// start the results code block
		results := fmt.Sprintf("```diff\n%s\nSkill Test Results: %s\n%s\n", profile_stats_border, user, profile_stats_border)

		// initialize tracked stats
		var placements, perfect_scores int
		hardest_level_passed := "N/A"

		// initialize stats array
		var user_stats [23]Statistic

		// fetch stats per level
		for i := 1; i <= 22; i++ {
			var current_line string

			// converts pop level to kyu/dan ranks
			level := pop_to_level(i)

			// add current level to current line
			current_line += level + ": " + "\t"

			// maintain alignment
			if i > 6 {
				current_line += " "
			}

			// fetch leaderboard text for current level
			leaderboard_text := fetch_leaderboard(i)

			// populate Statistic struct for current user and level
			current_statistic := user_stats[i]
			found := populate_statistic(&current_statistic, user, leaderboard_text)
			if found == false {
				// if population fails, indicate the current level was not passed, continue to next level
				current_line += current_statistic.Correct + "\t\t" + current_statistic.Time + "\n"
				results += current_line
				continue
			}

			// update the  highest level passed
			hardest_level_passed = level

			// add solved count to the current line, maintaining alignment
			current_line += current_statistic.Correct + "/10 " + "\t"
			for i := len(current_statistic.Correct); i <= 2; i++ {
				current_line += " "
			}

			// add time spent to the result line, maintaining alignment
			current_line += current_statistic.Time + " seconds"

			// maintain alignment
			for i := len(current_statistic.Time); i <= 3; i++ {
				current_line += " "
			}

			// update perfect level count
			if current_statistic.Correct == "10" {
				perfect_scores += 1
			}

			// update leaderboard placement count, and print leaderboard text
			if current_statistic.Leaderboard == true {
				placements += 1

				// add leaderboard text to results line
				current_line += "\t(Global Top 100)"
			}

			// new result line for each level
			current_line += "\n"
			results += current_line
		}

		// add tracked stats to the result
		results += profile_stats_border
		results += "\n"
		results += "Highest Level Passed: " + hardest_level_passed + "\n"
		results += "Perfect Scores: " + strconv.Itoa(perfect_scores) + "\n"
		results += "Leaderboard Placements: " + strconv.Itoa(placements) + "\n"
		results += profile_stats_border
		results += "\n"
		results += "```"

		// print result
		_, print_result_error := session.ChannelMessageSend(message.ChannelID, results)
		if print_result_error != nil {
			log.Fatal("Error printing results:", print_result_error)
		}
	}

	return
}

func get_comparison_stats(message *discordgo.MessageCreate, session *discordgo.Session) {
	if strings.HasPrefix(message.Content, "!compare") {
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

		// set border string
		compare_border := "========================================================="

		// start the results code block
		compare_results := fmt.Sprintf("```diff\n%s\nSkill Test Comparison: %s vs %s\n%s\n", compare_border, user1, user2, compare_border)

		// initialize tracked stats
		var user1_leaderboard_placements, user1_perfect_scores int
		user1_hardest_level_passed := "N/A"
		var user2_leaderboard_placements, user2_perfect_scores int
		user2_hardest_level_passed := "N/A"

		// initialize comparison stats
		var wins, losses, ties, nocontest int

		// initialize stats arrays
		var user1_stats [23]Statistic
		var user2_stats [23]Statistic

		// fetch stats per level
		for i := 1; i <= 22; i++ {
			var current_line string

			// converts pop level to kyu/dan ranks
			level := pop_to_level(i)

			// add level information to the current line
			current_line += level + ": "

			// maintain alignment
			if i > 6 {
				current_line += " "
			}

			// fetch leaderboard text
			compare_leaderboard_text := fetch_leaderboard(i)

			// populate Statistic struct for user 1
			current_user1_statistic := user1_stats[i]
			first_user_found := populate_statistic(&current_user1_statistic, user1, compare_leaderboard_text)
			if first_user_found == false {
				current_line += current_user1_statistic.Correct + "   " + current_user1_statistic.Time + "            |"
			} else {
				// update the highest level solved
				user1_hardest_level_passed = level

				// add user 1 solved count to current line, maintaining alignment
				current_line += current_user1_statistic.Correct + "/10"
				for i := len(current_user1_statistic.Correct); i <= 2; i++ {
					current_line += " "
				}

				// add user 1 time spent to result line, maintaining alignment
				current_line += current_user1_statistic.Time + "s"

				// maintain alignment
				for i := len(current_user1_statistic.Time); i <= 3; i++ {
					current_line += " "
				}

				// update user 1 perfect level count
				if current_user1_statistic.Correct == "10" {
					user1_perfect_scores += 1
				}

				// update user 1 leaderboard placement count and add leaderboard text to current line
				if current_user1_statistic.Leaderboard == true {
					user1_leaderboard_placements += 1

					// add leaderboard text to results line
					current_line += "(Top 100) |"
				} else {
					// maintain alignment
					current_line += "          |"
				}
			}

			// populate Statistic struct for user 2
			current_user2_statistic := user2_stats[i]
			second_user_found := populate_statistic(&current_user2_statistic, user2, compare_leaderboard_text)
			if second_user_found == false {
				current_line += " " + current_user2_statistic.Correct + "   " + current_user2_statistic.Time + "        "
			} else {
				// update the highest level solved
				user2_hardest_level_passed = level

				// add user 2 solved count to current line, maintaining alignment
				current_line += " " + current_user2_statistic.Correct + "/10"
				for i := len(current_user2_statistic.Correct); i <= 2; i++ {
					current_line += " "
				}

				// add user 2 time spent to result line, maintaining alignment
				current_line += current_user2_statistic.Time + "s"

				// maintain alignment
				for i := len(current_user2_statistic.Time); i <= 3; i++ {
					current_line += " "
				}

				// update user 2 perfect level count
				if current_user2_statistic.Correct == "10" {
					user2_perfect_scores += 1
				}

				// update user 2 leaderboard placement count and add leaderboard text to current line
				if current_user2_statistic.Leaderboard == true {
					user2_leaderboard_placements += 1

					// add leaderboard text to results line
					current_line += "(Top 100)"
				}
			}

			// generate diff prefix
			current_level_comparison := compare_statistics(current_user1_statistic, current_user2_statistic)

			// update comparison stats
			current_line_contains_highlight := true
			switch {
			case current_level_comparison == "+  ":
				wins += 1
			case current_level_comparison == "-  ":
				losses += 1
			case current_level_comparison == "***":
				nocontest += 1
				current_line_contains_highlight = false
			case current_level_comparison == "!  ":
				ties += 1
			}

			// add diff prefix to current line
			current_line = current_level_comparison + " " + current_line

			// maintain highlight alignment
			if current_line_contains_highlight == true {
				for color_alignment := len(current_line); color_alignment < len(compare_border); color_alignment++ {
					current_line += " "
				}
			}

			// new result line for each level
			current_line += "\n"

			// add current line to result
			compare_results += current_line
		}

		// construct footer for tracked stats, add dividing column, maintain alignment
		compare_results += compare_border + "\n"
		footer1 := "Highest Level Passed: " + user1_hardest_level_passed
		for pad1 := len(footer1); pad1 < 30; pad1++ {
			footer1 += " "
		}
		footer1 += "| Highest Level Passed: " + user2_hardest_level_passed

		footer2 := "Perfect Scores: " + strconv.Itoa(user1_perfect_scores)
		for pad2 := len(footer2); pad2 < 30; pad2++ {
			footer2 += " "
		}
		footer2 += "| Perfect Scores: " + strconv.Itoa(user2_perfect_scores)

		footer3 := "Global Leaderboards: " + strconv.Itoa(user1_leaderboard_placements)
		for pad3 := len(footer3); pad3 < 30; pad3++ {
			footer3 += " "
		}
		footer3 += "| Global Leaderboards: " + strconv.Itoa(user2_leaderboard_placements)

		// combine tracked stat lines
		compare_results += footer1 + "\n" + footer2 + "\n" + footer3 + "\n"
		compare_results += compare_border + "\n"

		// print comparison stats
		compare_results += "+ Wins: " + strconv.Itoa(wins) + "\n"
		compare_results += "- Losses: " + strconv.Itoa(losses) + "\n"
		compare_results += "! Ties: " + strconv.Itoa(ties) + "\n"
		compare_results += "*** No Contest: " + strconv.Itoa(nocontest) + "\n"

		// end codeblock
		compare_results += "```"

		// print result
		_, print_result_error := session.ChannelMessageSend(message.ChannelID, compare_results)
		if print_result_error != nil {
			log.Fatal("Error printing results:", print_result_error)
		}
	}

	return
}

func compare_statistics(stat1 Statistic, stat2 Statistic) string {
	// non-numeric comparison handling
	if stat1.Correct == "N/A" && stat2.Correct == "N/A" {
		return "***"
	} else if stat1.Correct == "N/A" {
		return "-  "
	} else if stat2.Correct == "N/A" {
		return "+  "
	}

	// convert compared stats from strings to ints for numeric comparisons
	time1, time_err1 := strconv.Atoi(stat1.Time)
	if time_err1 != nil {
		log.Fatal("Error: the math isn't mathing", time_err1)
	}

	time2, time_err2 := strconv.Atoi(stat2.Time)
	if time_err2 != nil {
		log.Fatal("Error: the math isn't mathing", time_err2)
	}

	correct1, correct_err1 := strconv.Atoi(stat1.Correct)
	if correct_err1 != nil {
		log.Fatal("Error: the math isn't mathing", correct_err1)
	}

	correct2, correct_err2 := strconv.Atoi(stat2.Correct)
	if correct_err2 != nil {
		log.Fatal("Error: the math isn't mathing", correct_err2)
	}

	// compare first by correct count, then time if result is unclear
	if correct1 > correct2 {
		return "+  "
	} else if correct1 < correct2 {
		return "-  "
	} else {
		if time1 == time2 {
			return "!  "
		} else if time1 < time2 {
			return "+  "
		} else {
			return "-  "
		}
	}
}

func valid_profile(message *discordgo.MessageCreate, session *discordgo.Session, user string) (exists bool, userid string) {
	// construct the profile URL
	profileURL := fmt.Sprintf("https://www.101weiqi.com/u/%s/", user)

	// disallow self-lookup
	if user == USERNAME {
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

func pop_to_level(pop int) string {
	if pop <= 15 {
		return strconv.Itoa(16-pop) + "K"
	} else {
		return strconv.Itoa(pop-15) + "D"
	}
}

func fetch_leaderboard(pop int) string {
	// construct temp URL for current level
	tempURL := LEADERBOARDURL + strconv.Itoa(pop) + "/"

	// GET temp URL
	temp_url_get_response, temp_url_get_response_error := client.Get(tempURL)
	if temp_url_get_response_error != nil {
		log.Fatal("Error fetching temp URL:", temp_url_get_response_error)
	}
	if temp_url_get_response.StatusCode != 200 {
		log.Fatal("Fetching temp URL gives status code error: %d %s", temp_url_get_response.StatusCode, temp_url_get_response.Status)
	}
	defer temp_url_get_response.Body.Close()

	// convert temp URL body to text and return string
	temp_url_get_response_body, temp_url_get_response_body_read_error := ioutil.ReadAll(temp_url_get_response.Body)
	if temp_url_get_response_body_read_error != nil {
		log.Fatal(temp_url_get_response_body_read_error)
	}

	temp_url_get_response_body_text := string(temp_url_get_response_body)
	return temp_url_get_response_body_text
}

func populate_statistic(stat *Statistic, user string, text string) bool {
	// finds user data for current level from input text
	regex_for_user := regexp.MustCompile(fmt.Sprintf(`"%s",\s*(\S+)\s+(\S+)\s*(\S+)\s+(\S+)\s*(\S+)\s+(\S+)`, user))
	match := regex_for_user.FindStringSubmatch(text)

	// checks if user's name only appears on the friends tab
	freq := regex_for_user.FindAllString(text, -1)

	// if user's name appears more than once, user is on the leaderboard for this level
	if len(freq) > 1 {
		stat.Leaderboard = true
	}

	if match != nil {
		// if user if found, populate Statistic with base stats
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
