package bot

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"
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

		// handle cache invalidation flag
		force_invalidation := false
		if len(parts) == 3 {
			if parts[2] == "-f" {
				force_invalidation = true
			} else {
				session.ChannelMessageSend(message.ChannelID, "Usage: !profile <101weiqi_username> <flag>")
				return
			}
		} else if len(parts) != 2 {
			session.ChannelMessageSend(message.ChannelID, "Usage: !profile <101weiqi_username> <flag>")
			return
		}

		// user from message
		user := parts[1]

		var user_id string
		cached := false
		cached_id, user_id_is_cached := get_user_id_from_friend_cache(user)
		if user_id_is_cached == true {
			// if id is cached, set user_id to cached id
			user_id = cached_id

			// set cached check to true
			cached = true
		} else {
			// check if user profile exists
			is_valid, id := valid_profile(message, session, user)
			if is_valid == false {
				return
			}

			// set user_id to fetched id
			user_id = id

			// friend user
			friend(add_friend, user, id, message, session)

			// add user info to cache
			add_to_friend_cache(user, user_id)
		}

		// print loading message
		if force_invalidation == true {
			invalidation_message := fmt.Sprintf("Fetching updated Skill Test Results for %s.", user)
			session.ChannelMessageSend(message.ChannelID, invalidation_message)
			cached = false
		} else if cached == false {
			loading_message := fmt.Sprintf("Fetching Skill Test Results for %s.", user)
			session.ChannelMessageSend(message.ChannelID, loading_message)
		}

		// set border string
		profile_stats_border := "===================================================="

		// start the results code block
		results := fmt.Sprintf("```diff\n%s\nSkill Test Results: %s\n%s\n", profile_stats_border, user, profile_stats_border)

		// initialize tracked stats
		var placements, perfect_scores int
		hardest_level_passed := "N/A"

		// initialize stats array
		var user_stats [23]Statistic

		// load leaderboard information
		pop_to_cached_leaderboard_text := concurrent_leaderboard_retrieval(cached)
		regex_for_user := regexp.MustCompile(fmt.Sprintf(`"%s",\s*(\S+)\s+(\S+)\s*(\S+)\s+(\S+)\s*(\S+)\s+(\S+)`, user))

		// fetch stats per level
		for pop := 1; pop <= 22; pop++ {
			var current_line string

			// converts pop level to kyu/dan ranks
			level := pop_to_level(pop)

			// add current level to current line
			current_line += level + ": " + "\t"

			// maintain alignment
			if pop > 6 {
				current_line += " "
			}

			// fetch leaderboard text for current level
			leaderboard_text := pop_to_cached_leaderboard_text[pop]

			// populate Statistic struct for current user and level
			current_statistic := user_stats[pop]
			found := populate_statistic(&current_statistic, regex_for_user, leaderboard_text)
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

		// handle cache invalidation flag
		force_invalidation := false
		if len(parts) == 4 {
			if parts[3] == "-f" {
				force_invalidation = true
			} else {
				session.ChannelMessageSend(message.ChannelID, "Usage: !compare <101weiqi_username>  <101weiqi_username> <flag>")
				return
			}
		} else if len(parts) != 3 {
			session.ChannelMessageSend(message.ChannelID, "Usage: !compare <101weiqi_username>  <101weiqi_username> <flag>")
			return
		}

		user1 := parts[1]
		user2 := parts[2]

		var user1_id string
		var user2_id string

		cached_id1, user1_id_is_cached := get_user_id_from_friend_cache(user1)
		if user1_id_is_cached == true {
			// if id is cached, set user_id1 to cached_id1
			user1_id = cached_id1
		} else {
			// check if user profiles exist
			user1_is_valid, id1 := valid_profile(message, session, user1)

			if user1_is_valid == false {
				session.ChannelMessageSend(message.ChannelID, "Please try again with two valid 101weiqi usernames")
				return
			}

			// update user 1 id
			user1_id = id1

			// friend user 2
			friend(add_friend, user1, user1_id, message, session)

			// add user 2 info to cache
			add_to_friend_cache(user1, user1_id)
		}

		cached_id2, user2_id_is_cached := get_user_id_from_friend_cache(user2)
		if user2_id_is_cached == true {
			// if id is cached, set user_id to cached id
			user2_id = cached_id2
		} else {
			user2_is_valid, id2 := valid_profile(message, session, user2)
			if user2_is_valid == false {
				session.ChannelMessageSend(message.ChannelID, "Please try again with two valid 101weiqi usernames")
				return
			}

			// update user 2 id
			user2_id = id2

			// friend user 2
			friend(add_friend, user2, user2_id, message, session)

			// add user 2 info to cache
			add_to_friend_cache(user2, user2_id)
		}

		// cached check is true if both user profiles were previously queried
		cached := user1_id_is_cached && user2_id_is_cached

		// send loading messages
		if force_invalidation == true {
			invalidation_message := fmt.Sprintf("Fetching updated Skill Test Comparison for %s vs %s.", user1, user2)
			session.ChannelMessageSend(message.ChannelID, invalidation_message)
			cached = false
		} else if cached == false {
			loading_message := fmt.Sprintf("Fetching Skill Test Comparison for %s vs %s.", user1, user2)
			session.ChannelMessageSend(message.ChannelID, loading_message)
		}

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

		pop_to_cached_leaderboard_text := concurrent_leaderboard_retrieval(cached)
		regex_for_user1 := regexp.MustCompile(fmt.Sprintf(`"%s",\s*(\S+)\s+(\S+)\s*(\S+)\s+(\S+)\s*(\S+)\s+(\S+)`, user1))
		regex_for_user2 := regexp.MustCompile(fmt.Sprintf(`"%s",\s*(\S+)\s+(\S+)\s*(\S+)\s+(\S+)\s*(\S+)\s+(\S+)`, user2))

		// fetch stats per level
		for pop := 1; pop <= 22; pop++ {
			var current_line string

			// converts pop level to kyu/dan ranks
			level := pop_to_level(pop)

			// add level information to the current line
			current_line += level + ": "

			// maintain alignment
			if pop > 6 {
				current_line += " "
			}

			// fetch leaderboard text
			compare_leaderboard_text := pop_to_cached_leaderboard_text[pop]

			// populate Statistic struct for user 1
			current_user1_statistic := user1_stats[pop]
			first_user_found := populate_statistic(&current_user1_statistic, regex_for_user1, compare_leaderboard_text)
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
			current_user2_statistic := user2_stats[pop]
			second_user_found := populate_statistic(&current_user2_statistic, regex_for_user2, compare_leaderboard_text)
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

func populate_statistic(stat *Statistic, regex_for_user *regexp.Regexp, text string) bool {
	// finds user data for current level from input text
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
