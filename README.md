# 101weiqi-bot
101weiqi is a Chinese language website for solving puzzles related to the board game [Go](https://en.wikipedia.org/wiki/Go_(game)). 

On 101weiqi, there is no simple way to view another user's results across all [skill test levels](https://www.101weiqi.com/guan/). This Discord bot aggregates 101weiqi skill test data and makes it trivial to compare the results of your favorite puzzle solvers.

The bot uses [goquery](https://github.com/PuerkitoBio/goquery) for parsing webpages, [discordgo](https://github.com/bwmarrin/discordgo) for interacting with the Discord API, and [gonum](https://github.com/gonum/plot) for data visualization.

## Screenshots
<img src="https://github.com/graceemlin/101weiqi-bot/blob/main/docs/profile.webp" width=50% height=50%> <img src="https://github.com/graceemlin/101weiqi-bot/blob/main/docs/compare.webp" width=49% height=50%>

## Setting up 101weiqi-bot:
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

## Using 101weiqi-bot:
The following commands have been implemented:

### `!profile [user] [flags]`
* Scrapes skill test stats for a specified user.
* Tracks and displays relevant user stats (hardest level passed, perfect scores, leaderboard placements).
* Currently supported flags:
    - `-f : forces cache invalidation`
    - `-t : truncates output`
    - `-g : outputs histogram`
    
### `!compare [user1] [user2] [flags]`
* Scrapes skill test stats for both users.
* Tracks and displays skill test stats for both users (hardest level passed, perfect scores, leaderboard placements), as well as comparison stats.
* Utilizes diff syntax highlighting in Discord codeblocks for clearer comparisons and added flair.
* Currently supported flags:
    - `-f : forces cache invalidation`
    - `-t : truncates output`
    - `-g : outputs histograms for both users`
      
### `!help`
* Displays help message.

## Changelog
 * **2025-02-18** : Added simple histograms, updated regex queries to support expanded character sets.
 * **2025-02-11** : Added caching and concurrency to mitigate rate limiting concerns and improve performance.
 * **2025-02-06** : Initial commit. Added login and friend request functions for retrieving skill test results outside of the global top 100.

## To do list:
* Slash command migration.
* Improve README.
