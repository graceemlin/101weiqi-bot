package bot

import (
	"sync"
)

func concurrent_leaderboard_retrieval(cached bool) map[int]string {
	var wait_group sync.WaitGroup
	pop_to_cached_leaderboard_text := make(map[int]string)
	var leaderboard_mutex sync.Mutex

	for pop := 1; pop <= 22; pop++ {
		wait_group.Add(1)
		go func(pop int) {
			defer wait_group.Done()
			leaderboard_text := fetch_leaderboard(cached, pop)
			leaderboard_mutex.Lock()
			pop_to_cached_leaderboard_text[pop] = leaderboard_text
			leaderboard_mutex.Unlock()
		}(pop)
	}

	wait_group.Wait()
	return pop_to_cached_leaderboard_text
}
