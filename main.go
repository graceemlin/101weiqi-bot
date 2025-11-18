package main

import bot "path/to/bot"

func main() {
	bot.BOTTOKEN = "$TOKEN"
	bot.USERNAME = "$USERNAME"
	bot.PASSWORD = "$PASSWORD"
	
	bot.CACHEDIR = "path/to/cachedir"
	bot.CACHEFILE = "path/to/cachefile.txt"
	bot.HISTOGRAMFILE = "path/to/histogramfile.png"
  
	bot.HELPMESSAGE = "DEFAULT_HELP_MESSAGE"

	bot.ATTIONURL = "https://www.101weiqi.com/attionuser/"
	bot.HOMEURL = "https://www.101weiqi.com/home/"
	bot.LEADERBOARDURL = "https://www.101weiqi.com/guan/pop/"
	bot.LOGINURL = "https://www.101weiqi.com/login/"
	bot.LOGINWQURL = "https://www.101weiqi.com/wq/login/"
	
	bot.Init()
	bot.Run()
}
