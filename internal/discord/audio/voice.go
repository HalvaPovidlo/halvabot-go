package audio

import "github.com/bwmarrin/discordgo"

type Client struct {
	conn    *discordgo.VoiceConnection
	session *discordgo.Session
}

func NewVoiceClient(s *discordgo.Session) *Client {
	return &Client{
		session: s,
	}
}

func (c *Client) Connection() *discordgo.VoiceConnection {
	return c.conn
}

// Connect TODO: deadlock super rare
func (c *Client) Connect(guildID, channelID string) error {
	conn, err := c.session.ChannelVoiceJoin(guildID, channelID, false, false)
	if err != nil {
		c.conn = nil
		return err
	}
	c.conn = conn
	return nil
}

func (c *Client) IsConnected() bool {
	return c.conn != nil
}

func (c *Client) Disconnect() error {
	if err := c.conn.Disconnect(); err != nil {
		return err
	}
	c.conn = nil
	return nil
}
