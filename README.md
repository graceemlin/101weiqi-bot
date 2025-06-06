<div align="center">
   <img src="https://github.com/egonelbre/gophers/blob/master/vector/computer/gamer.svg" width=30% height=30%> 

# 101weiqi-bot

A scraping tool for [Go](https://en.wikipedia.org/wiki/Go_(game)) puzzle skill test results

[![License](https://img.shields.io/badge/license-MIT-00acc1?label=license&style=flat&labelColor=282c34&logo=open-source-initiative)](LICENSE)
</div>

---

101weiqi is a Chinese language website for solving puzzles related to the board game [Go](https://en.wikipedia.org/wiki/Go_(game)). 

On 101weiqi, there is no simple way to view another user's results across all [skill test levels](https://www.101weiqi.com/guan/). This Discord bot aggregates 101weiqi skill test data and makes it trivial to compare the results of your favorite puzzle solvers.

The bot uses [goquery](https://github.com/PuerkitoBio/goquery) for parsing webpages, [discordgo](https://github.com/bwmarrin/discordgo) for interacting with the Discord API, and [gonum](https://github.com/gonum/plot) for data visualization. 

### Screenshots
<img src="https://github.com/graceemlin/101weiqi-bot/blob/main/docs/profile.webp" width=50% height=50%> <img src="https://github.com/graceemlin/101weiqi-bot/blob/main/docs/compare.webp" width=49% height=50%>

### Setting up 101weiqi-bot:
1. Clone the 101weiqi-bot repository from GitHub:\
   `git clone https://github.com/graceemlin/101weiqi-bot`
2. Navigate to the project directory on your machine:\
   `cd 101weiqi-bot`
3. Initialize the project as a Go module:\
   `go mod init 101weiqi-bot`
4. Download dependencies:\
   `go mod tidy`
5. Configure main.go:
    * Credentials: Add your 101weiqi.com username and password.
    * Discord Bot Token: Add your Discord Bot Token (obtained from the [Discord Developer Portal](https://discord.com/developers/applications)).
    * Filepath Information: Update filepath information as needed.
6. Run 101weiqi-bot:\
   `go run main.go`

### Using 101weiqi-bot:
```
!profile [user] [flags]
   - Retrieves and analyzes skill test data for any user, providing a comprehensive performance overview for each skill test level.
   - Displays relevant user statistics for measuring overall performance, including:
      * Hardest level passed
      * Perfect scores
      * Leaderboard placements

!compare [user1] [user2] [flags]
   - Compares the skill test data of two users side-by-side, highlighting key differences in performance for each skill test level.
   - Utilizes diff syntax highlighting in Discord codeblocks for clearer comparisons and added flair.
   - Tracks head-to-head results and outputs overall performance statistics.

!help
   - Displays a help message.

Currently supported flags:
    -f : Forces cache invalidation, ensuring the most up-to-date information is received.
    -t : Truncates the output to provide a concise summary.
    -g : Outputs histograms for visualizing performance trends.
```

### Changelog
 * **2025-04-29** : 101weiqi-bot is currently non-functional as its data source, 101weiqi.com, is undergoing significant updates. Suspending development of this project until the 101weiqi.com website reaches operational stability.
 * **2025-02-28** : Merged and adjusted @redreoicy's formatting PR.
 * **2025-02-26** : Added a fetch delay to avoid exceeding rate limits, created a setup section in the README, merged and adjusted @redreoicy's formatting PR. 
 * **2025-02-18** : Added simple histograms, updated regex queries to support expanded character sets.
 * **2025-02-11** : Added caching and concurrency to mitigate rate limiting concerns and improve performance.
 * **2025-02-06** : Initial commit. Added login and friend request functions for retrieving skill test results outside of the global top 100.

### To do list:
* Slash command migration.
* Improve README.

### Attributions:
* The logo is the "Gamer Gopher" designed by [Egon Elbre](https://github.com/egonelbre/gophers).
