package bot

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var CACHEFILE string
var cached_friends_map map[string]string
var skill_test_caches map[int]string

func load_friend_cache() {
	// check if file exists
	_, file_retrieval_error := os.Stat(CACHEFILE)
	if os.IsNotExist(file_retrieval_error) {
		log.Println("No cache file to load.")
		return
	}

	// open cache file
	friend_file, error_opening_friend_file := os.OpenFile(CACHEFILE, os.O_RDONLY, 0644)
	if error_opening_friend_file != nil {
		log.Fatal("Error opening friend file", error_opening_friend_file)
	}
	defer friend_file.Close()

	// scan friend info from the cache file into the friends map
	scanner := bufio.NewScanner(friend_file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			username := parts[0]
			user_id := parts[1]
			if username != "" && user_id != "" {
				cached_friends_map[username] = user_id
			}
		}
	}

	// check for scanner errors
	scanner_error := scanner.Err()
	if scanner_error != nil {
		log.Fatal("Error reading friend cache file:", scanner_error)
	}
}

func save_friend_cache() {
	// create a temporary file
	temp_dir := filepath.Dir(CACHEFILE)
	temp_friend_cache, temp_friend_cache_error := os.CreateTemp(temp_dir, CACHEFILE+".tmp")
	if temp_friend_cache_error != nil {
		log.Fatal("Error creating temporary friend cache file:", temp_friend_cache)
	}

	// write updated friend information to the temp file
	friend_writer := bufio.NewWriter(temp_friend_cache)
	for username, user_id := range cached_friends_map {
		fmt.Fprintf(friend_writer, "%s:%s\n", username, user_id)
	}

	friend_write_error := friend_writer.Flush()
	if friend_write_error != nil {
		log.Fatal("Error writing to friend file", friend_write_error)
	}
	temp_friend_cache.Close()

	// replace the cache file with the temp file
	friend_rename_error := os.Rename(temp_friend_cache.Name(), CACHEFILE)
	if friend_rename_error != nil {
		log.Fatal("Error renaming temporary cache file:", friend_rename_error)
	}

}

func get_user_id_from_friend_cache(username string) (string, bool) {
	user_id, ok := cached_friends_map[username]
	return user_id, ok
}

func add_to_friend_cache(username string, user_id string) {
	cached_friends_map[username] = user_id
	save_friend_cache()
}

func load_skill_test_caches() {
	for pop := 1; pop <= 22; pop++ {
		// construct temp cache file name
		temp_cache := filepath.Join(CACHEDIR, fmt.Sprintf("pop_%d.txt", pop))

		// check if file exists
		_, temp_cache_retrieval_error := os.Stat(temp_cache)
		if os.IsNotExist(temp_cache_retrieval_error) {
			log.Printf("Cache file for pop %d does not exist.", pop)
			continue
		}

		// open temp cache file for current level copy data to local leaderboard string
		temp_cache_file, error_opening_temp_cache := os.Open(temp_cache)
		if error_opening_temp_cache != nil {
			log.Fatal("Error opening temp leaderboard cache file", error_opening_temp_cache)
		}
		defer temp_cache_file.Close()

		var local_leaderboard strings.Builder
		scanner := bufio.NewScanner(temp_cache_file)
		for scanner.Scan() {
			local_leaderboard.WriteString(scanner.Text())
			local_leaderboard.WriteString("\n")
		}

		// check for scanner errors
		scanner_error := scanner.Err()
		if scanner_error != nil {
			log.Fatal("Error reading temp cache file", scanner_error)
		}

		// populate pop to leaderboard map with cached leaderboard data
		skill_test_caches[pop] = local_leaderboard.String()
	}
}

func save_skill_test_cache(pop int, leaderboard_text string) {
	// construct temp cache file name
	temp_cache := filepath.Join(CACHEDIR, fmt.Sprintf("pop_%d.txt", pop))

	// create temp leaderboard cache file
	temp_skill_test_file, temp_skill_test_file_creation_error := os.CreateTemp(CACHEDIR, fmt.Sprintf("pop_%d.tmp", pop))
	if temp_skill_test_file_creation_error != nil {
		log.Fatal("Error creating temporary leaderboard cache file", temp_skill_test_file_creation_error)
	}

	// write to temp leaderboard cache file
	_, temp_skill_test_file_write_error := temp_skill_test_file.WriteString(leaderboard_text)
	if temp_skill_test_file_write_error != nil {
		log.Fatal("Error writing to temporary leaderboard cache file", temp_skill_test_file_write_error)
	}

	temp_skill_test_file.Close()

	// replace the cache file with the temp file
	temp_skill_test_file_rename_error := os.Rename(temp_skill_test_file.Name(), temp_cache)
	if temp_skill_test_file_rename_error != nil {
		log.Fatal("Error renaming temporary leaderboard cache file", temp_skill_test_file_rename_error)
	}
}

func get_local_leaderboard(pop int) (string, bool) {
	local_leaderboard_text, ok := skill_test_caches[pop]
	return local_leaderboard_text, ok
}

func add_to_skill_test_cache(pop int, leaderboard_text string) {
	skill_test_caches[pop] = leaderboard_text
	save_skill_test_cache(pop, leaderboard_text)
}
