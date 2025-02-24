# 101weiqi-bot
101weiqi is a Chinese language website for solving puzzles related to the board game [Go](https://en.wikipedia.org/wiki/Go_(game)). 

On 101weiqi, there is no simple way to view another user's results across all [skill test levels](https://www.101weiqi.com/guan/). This Discord bot aggregates 101weiqi skill test data and makes it trivial to compare the results of your favorite puzzle solvers.

The bot uses [goquery](https://github.com/PuerkitoBio/goquery) for parsing webpages, [discordgo](https://github.com/bwmarrin/discordgo) for interacting with the Discord API, and [gonum](https://github.com/gonum/plot) for data visualization.

# Screenshots
<img src="https://github.com/graceemlin/101weiqi-bot/blob/main/docs/profile.webp" width=50% height=50%> <img src="https://github.com/graceemlin/101weiqi-bot/blob/main/docs/compare.webp" width=49% height=50%>

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
