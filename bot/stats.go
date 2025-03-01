package bot

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"
)

var LEADERBOARDURL string

type Statistic struct {
	Correct     string
	Time        string
	Leaderboard bool
	Top100      string
}

type full_data struct {
	user_stats           [23]Statistic
	hardest_level_passed string
	placements           int
	perfect_scores       int
}

func get_profile_stats(message *discordgo.MessageCreate, session *discordgo.Session) {
	if strings.HasPrefix(message.Content, "!profile") {
		// split into command and username
		parts := strings.Split(message.Content, " ")

		// handle cache invalidation, truncate, and graph flags
		force_invalidation := false
		truncate_results := false
		make_graph := false

		if len(parts) == 3 {
			flag := parts[2]
			if flag[0] != '-' {
				session.ChannelMessageSend(message.ChannelID, "Usage: !profile <101weiqi_username> <flag>")
				return
			}

			// create flag map for duplicate checks
			flag_used := make(map[string]bool)

			dupe_error := false
			invalid_error := false
			for i := 0; i < len(flag); i++ {
				flag_char := string(flag[i])
				if flag_used[flag_char] == true {
					dupe_error = true
					break
				}

				flag_used[flag_char] = true
				if flag_char == "f" {
					force_invalidation = true
				} else if flag_char == "t" {
					truncate_results = true
				} else if flag_char == "g" {
					make_graph = true
				} else if flag_char != "-" {
					invalid_error = true
				}
			}

			if invalid_error == true || dupe_error == true {
				session.ChannelMessageSend(message.ChannelID, "Usage: !profile <101weiqi_username> <flag>\n")

				if invalid_error == true {
					session.ChannelMessageSend(message.ChannelID, "Valid flags only please.\n")
				}

				if dupe_error == true {
					session.ChannelMessageSend(message.ChannelID, "Duplicate flags are not allowed!\n")
				}

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
			// if id is cached, set user_id to cached_id
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
		var results strings.Builder
		results.WriteString(fmt.Sprintf("```diff\n%s\nSkill Test Results: %s\n%s\n", profile_stats_border, user, profile_stats_border))

		// initialize full data
		var user_data full_data
		user_data.hardest_level_passed = "N/A"

		// load leaderboard information
		pop_to_cached_leaderboard_text := concurrent_leaderboard_retrieval(cached)

		// if username contains characters outside of the standard ASCII range, adjust regex query
		regex_user := encoded_user(user)
		regex_for_user := regexp.MustCompile(fmt.Sprintf(`"%s",\s*(\S+)\s+(\S+)\s*(\S+)\s+(\S+)\s*(\S+)\s+(\S+)`, regex_user))

		// fetch stats per level
		for pop := 1; pop <= 22; pop++ {
			// converts pop level to kyu/dan ranks
			level := pop_to_level(pop)

			// fetch leaderboard text for current level
			leaderboard_text := pop_to_cached_leaderboard_text[pop]

			// populate Statistic struct for current user and level
			found := populate_statistic(&user_data.user_stats[pop], regex_for_user, leaderboard_text)
			current_statistic := user_data.user_stats[pop]

			// update statistics and strings
			if found {
				current_statistic.Correct += "/10"
				current_statistic.Time += " seconds"
				user_data.hardest_level_passed = level

				// update perfect level count
				if current_statistic.Correct == "10/10" {
					user_data.perfect_scores += 1
				}

				// update leaderboard placement count and leaderboard top 100 text
				if current_statistic.Leaderboard == true {
					user_data.placements += 1
					current_statistic.Top100 = "(Global Top 100)"
				}
			}

			var current_line strings.Builder
			// add statistics to current line
			current_line.WriteString(fmt.Sprintf("%-5s\t%-7s\t%-11s\t %16s\n", level+":", current_statistic.Correct, current_statistic.Time, current_statistic.Top100))

			// new result line for each level
			if truncate_results == false {
				results.WriteString(current_line.String())
			}
		}

		// add tracked stats to the result
		if truncate_results == false {
			results.WriteString(profile_stats_border)
			results.WriteString("\n")
		}

		results.WriteString(fmt.Sprintf("Highest Level Passed: %s\nPerfect Scores: %s\nLeaderboard Placements: %s\n%s\n```", user_data.hardest_level_passed, strconv.Itoa(user_data.perfect_scores), strconv.Itoa(user_data.placements), profile_stats_border))

		// print result
		_, print_result_error := session.ChannelMessageSend(message.ChannelID, results.String())
		if print_result_error != nil {
			log.Fatal("Error printing results:", print_result_error)
		}

		if make_graph == true {
			graph_create(&user_data.user_stats, user, standard)
			graph_print(user, message, session)
		}
	}

	return
}

func get_comparison_stats(message *discordgo.MessageCreate, session *discordgo.Session) {
	if strings.HasPrefix(message.Content, "!compare") {
		// split into command and usernames
		parts := strings.Split(message.Content, " ")

		// handle cache invalidation, truncate, and graph flags
		make_graph := false
		truncate_results := false
		force_invalidation := false
		if len(parts) == 4 {
			flag := parts[3]
			if flag[0] != '-' {
				session.ChannelMessageSend(message.ChannelID, "Usage: !compare <101weiqi_username> <flag>")
				return
			}

			flag_used := make(map[string]bool)
			dupe_error := false
			invalid_error := false
			for i := 0; i < len(flag); i++ {
				flag_char := string(flag[i])
				if flag_used[flag_char] == true {
					dupe_error = true
					break
				}

				flag_used[flag_char] = true
				if flag_char == "f" {
					force_invalidation = true
				} else if flag_char == "t" {
					truncate_results = true
				} else if flag_char == "g" {
					make_graph = true
				} else if flag_char != "-" {
					invalid_error = true
				}
			}

			if invalid_error == true || dupe_error == true {
				session.ChannelMessageSend(message.ChannelID, "Usage: !compare <101weiqi_username> <101weiqi_username> <flag>")
				if invalid_error == true {
					session.ChannelMessageSend(message.ChannelID, "Valid flags only please.\n")
				}

				if dupe_error == true {
					session.ChannelMessageSend(message.ChannelID, "Duplicate flags are not allowed!\n")
				}

				return
			}

		} else if len(parts) != 3 {
			session.ChannelMessageSend(message.ChannelID, "Usage: !compare <101weiqi_username> <101weiqi_username> <flag>")
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

			// friend user 1
			friend(add_friend, user1, user1_id, message, session)

			// add user 1 info to cache
			add_to_friend_cache(user1, user1_id)
		}

		cached_id2, user2_id_is_cached := get_user_id_from_friend_cache(user2)
		if user2_id_is_cached == true {
			// if id is cached, set user2_id to cached_id2
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
		var compare_results strings.Builder
		compare_results.WriteString(fmt.Sprintf("```diff\n%s\nSkill Test Comparison: %s vs %s\n%s\n", compare_border, user1, user2, compare_border))

		// initialize full data
		var user1_data full_data
		user1_data.hardest_level_passed = "N/A"
		var user2_data full_data
		user2_data.hardest_level_passed = "N/A"

		// initialize comparison stats
		var wins, losses, ties, nocontest int
		pop_to_cached_leaderboard_text := concurrent_leaderboard_retrieval(cached)

		// if username contains characters outside of the standard ASCII range, adjust regex query
		regex_user1 := encoded_user(user1)
		regex_user2 := encoded_user(user2)
		regex_for_user1 := regexp.MustCompile(fmt.Sprintf(`"%s",\s*(\S+)\s+(\S+)\s*(\S+)\s+(\S+)\s*(\S+)\s+(\S+)`, regex_user1))
		regex_for_user2 := regexp.MustCompile(fmt.Sprintf(`"%s",\s*(\S+)\s+(\S+)\s*(\S+)\s+(\S+)\s*(\S+)\s+(\S+)`, regex_user2))

		// fetch stats per level
		for pop := 1; pop <= 22; pop++ {
			var current_line strings.Builder

			// converts pop level to kyu/dan ranks
			level := pop_to_level(pop)

			// fetch leaderboard text
			compare_leaderboard_text := pop_to_cached_leaderboard_text[pop]

			// populate Statistic struct for user 1
			first_user_found := populate_statistic(&user1_data.user_stats[pop], regex_for_user1, compare_leaderboard_text)
			current_user1_statistic := user1_data.user_stats[pop]

			// populate Statistic struct for user 2
			second_user_found := populate_statistic(&user2_data.user_stats[pop], regex_for_user2, compare_leaderboard_text)
			current_user2_statistic := user2_data.user_stats[pop]

			// generate diff prefix
			current_level_comparison := compare_statistics(current_user1_statistic, current_user2_statistic)

			// update comparison stats
			switch {
			case current_level_comparison == "+  ":
				wins += 1
			case current_level_comparison == "-  ":
				losses += 1
			case current_level_comparison == "***":
				nocontest += 1
			case current_level_comparison == "!  ":
				ties += 1
			}

			if first_user_found != false {
				//update strings and hardest level passed for user 1
				current_user1_statistic.Correct += "/10"
				current_user1_statistic.Time += "s"
				user1_data.hardest_level_passed = level

				// update user 1 perfect level count for user 1
				if current_user1_statistic.Correct == "10/10" {
					user1_data.perfect_scores += 1
				}

				// update leaderboard placement count and leaderboard top 100 text for user 1
				if current_user1_statistic.Leaderboard == true {
					user1_data.placements += 1
					current_user1_statistic.Top100 = "(Global Top 100)"
				}
			}

			if second_user_found != false {
				//update strings and hardest level passed for user 2
				current_user2_statistic.Correct += "/10"
				current_user2_statistic.Time += "s"
				user2_data.hardest_level_passed = level

				// update user 2 perfect level count
				if current_user2_statistic.Correct == "10/10" {
					user2_data.perfect_scores += 1
				}

				// update leaderboard placement count and leaderboard top 100 text for user 2
				if current_user2_statistic.Leaderboard == true {
					user2_data.placements += 1
					current_user2_statistic.Top100 = "(Global Top 100)"
				}
			}

			var leaderboard_user1 strings.Builder
			var leaderboard_user2 strings.Builder

			if current_user1_statistic.Leaderboard {
				leaderboard_user1.WriteString(" (Top 100)")
			} else {
				leaderboard_user1.WriteString("")
			}
			if current_user2_statistic.Leaderboard {
				leaderboard_user2.WriteString(" (Top 100)")
			} else {
				leaderboard_user2.WriteString("")
			}

			current_line.WriteString(fmt.Sprintf("%-4s%-5s%-7s%-4s%-11s| %-7s%-4s%-13s\n", current_level_comparison, level+":", current_user1_statistic.Correct, current_user1_statistic.Time, leaderboard_user1.String(), current_user2_statistic.Correct, current_user2_statistic.Time, leaderboard_user2.String()))

			if truncate_results == false {
				// add current line to result
				compare_results.WriteString(current_line.String())
			}
		}

		if truncate_results == false {
			compare_results.WriteString(compare_border)
			compare_results.WriteString("\n")
		}

		// construct footer for tracked stats, add dividing column, maintain alignment
		compare_results.WriteString(fmt.Sprintf("%-31s| Highest Level Passed: %s\n", "Highest Level Passed: "+user1_data.hardest_level_passed, user2_data.hardest_level_passed))
		compare_results.WriteString(fmt.Sprintf("%-31s| Perfect Scores: %s\n", "Perfect Scores: "+strconv.Itoa(user1_data.perfect_scores), strconv.Itoa(user2_data.perfect_scores)))
		compare_results.WriteString(fmt.Sprintf("%-31s| Global Leaderboards: %s\n", "Global Leaderboards: "+strconv.Itoa(user1_data.placements), strconv.Itoa(user2_data.placements)))

		// add comparison stats
		comparison_stats := fmt.Sprintf("%s\n+ Wins: %s\n- Losses: %s\n! Ties: %s\n*** No Contest: %s\n```", compare_border, strconv.Itoa(wins), strconv.Itoa(losses), strconv.Itoa(ties), strconv.Itoa(nocontest))

		// append to compare_results
		compare_results.WriteString(comparison_stats)

		// print result
		_, print_result_error := session.ChannelMessageSend(message.ChannelID, compare_results.String())
		if print_result_error != nil {
			log.Fatal("Error printing results:", print_result_error)
		}

		if make_graph == true {
			graph_create(&user1_data.user_stats, user1, standard)
			graph_print(user1, message, session)

			graph_create(&user2_data.user_stats, user2, standard)
			graph_print(user2, message, session)
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

func encoded_user(username string) string {
	var encoded strings.Builder
	for _, curr_rune := range username {
		if curr_rune > 127 {
			// double escape unicode conversion if outside of ASCII range
			fmt.Fprintf(&encoded, "\\\\u%04x", curr_rune)
		} else {
			// if in ASCII range, keep character the same
			encoded.WriteRune(curr_rune)
		}

	}
	return encoded.String()
}
