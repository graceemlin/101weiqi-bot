package main

import bot "path/to/bot"

func main() {
	bot.BOTTOKEN = "$TOKEN"
	bot.USERNAME = "$USERNAME"
	bot.PASSWORD = "$PASSWORD"

  bot.CACHEDIR = "path/to/cachedir"
  bot.CACHEFILE = "path/to/cachefile"
	bot.HISTOGRAMFILE = "path/to/histogramfile"
  
	bot.HELPMESSAGE = "DEFAULT_HELP_MESSAGE"
	
	bot.LOGINURL = "https://www.101weiqi.com/login/"
	bot.HOMEURL = "https://www.101weiqi.com/home/"
	bot.ATTIONURL = "https://www.101weiqi.com/attionuser/"
	bot.LEADERBOARDURL = "https://www.101weiqi.com/guan/pop/"
	
	bot.Init()
	bot.Run()
}
