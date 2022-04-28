package discord

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strconv"

	"github.com/HalvaPovidlo/discordBotGo/pkg/zap"
)

const (
	messageSearching = ":trumpet: **Searching** :mag_right:"
	messageFound     = "**Song found** :notes:"
)

type sendParams struct {
	S *discordgo.Session
	M *discordgo.MessageCreate
	L *zap.Logger
}

func sendSearchingMessage(p sendParams) {
	_, err := p.S.ChannelMessageSend(p.M.ChannelID, messageSearching)
	if err != nil {
		p.L.Errorw("sending message",
			"channel", p.M.ChannelID,
			"msg", messageSearching,
			"err", err)
	}
}

func sendFoundMessage(artist, title string, playbacks int, p sendParams) {
	msg := fmt.Sprintf("%s `%s - %s` %s", messageFound, artist, title, intToEmoji(playbacks))
	_, err := p.S.ChannelMessageSend(p.M.ChannelID, msg)
	if err != nil {
		p.L.Errorw("sending message",
			"channel", p.M.ChannelID,
			"msg", msg,
			"err", err)
	}
}

func channelMessageSend(msg string, p sendParams) {
	_, err := p.S.ChannelMessageSend(p.M.ChannelID, msg)
	if err != nil {
		p.L.Errorw("sending message",
			"channel", p.M.ChannelID,
			"msg", msg,
			"err", err)
	}
}

func intToEmoji(n int) string {
	number := strconv.Itoa(n)
	res := ""
	for i := range number {
		res += digitAsEmoji(string(number[i]))
	}
	return res
}

func digitAsEmoji(digit string) string {
	switch digit {
	case "1":
		return "1️⃣"
	case "2":
		return "2️⃣"
	case "3":
		return "3️⃣"
	case "4":
		return "4️⃣"
	case "5":
		return "5️⃣"
	case "6":
		return "6️⃣"
	case "7":
		return "7️⃣"
	case "8":
		return "8️⃣"
	case "9":
		return "9️⃣"
	case "0":
		return "0️⃣"
	}
	return ""
}
