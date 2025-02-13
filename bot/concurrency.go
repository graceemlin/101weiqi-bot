package bot

import (
	"io/ioutil"
	"log"
	"strconv"
	"sync"
)

func concurrent_leaderboard_retrieval(cached bool) map[int]string {
	var wait_group sync.WaitGroup
	pop_to_cached_leaderboard_text := make(map[int]string)
	var leaderboard_mutex sync.Mutex

	for pop := 1; pop <= 22; pop++ {
		// track the number of goroutines
		wait_group.Add(1)

		// pass current pop to a goroutine
		go func(pop int) {
			// track the number of go routines
			defer wait_group.Done()

			// fetch leaderboard text
			leaderboard_text := fetch_leaderboard(cached, pop)

			// protect map access with mutex
			leaderboard_mutex.Lock()

			// populate pop_to_cached_leaderboard_text map
			pop_to_cached_leaderboard_text[pop] = leaderboard_text

			leaderboard_mutex.Unlock()
		}(pop)
	}

	// wait until all goroutines in the wait group finish
	wait_group.Wait()

	// return the populated leaderboard map
	return pop_to_cached_leaderboard_text
}

func fetch_leaderboard(cached bool, pop int) string {
	// check if a local copy of the leaderboard exists
	if cached == true {
		pop_local_text, ok := get_local_leaderboard(pop)
		if ok == true {
			return pop_local_text
		}
	}

	// construct pop URL
	popURL := LEADERBOARDURL + strconv.Itoa(pop) + "/"

	// GET pop URL
	pop_url_get_response, pop_url_get_response_error := client.Get(popURL)
	if pop_url_get_response_error != nil {
		log.Fatal("Error fetching pop URL:", pop_url_get_response_error)
	}
	if pop_url_get_response.StatusCode != 200 {
		log.Fatalf("Fetching pop URL gives status code error: %d %s", pop_url_get_response.StatusCode, pop_url_get_response.Status)
	}
	defer pop_url_get_response.Body.Close()

	// convert pop URL body to text and return string
	pop_url_get_response_body, pop_url_get_response_body_read_error := ioutil.ReadAll(pop_url_get_response.Body)
	if pop_url_get_response_body_read_error != nil {
		log.Fatal(pop_url_get_response_body_read_error)
	}

	pop_url_get_response_body_text := string(pop_url_get_response_body)

	// cache current leaderboard text
	add_to_skill_test_cache(pop, pop_url_get_response_body_text)

	// return leaderboard text
	return pop_url_get_response_body_text
}
