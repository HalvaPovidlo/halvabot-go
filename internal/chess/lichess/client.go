package lichess

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"

	phttp "github.com/HalvaPovidlo/discordBotGo/pkg/http"
)

const lichessOpenGameURL = "https://lichess.org/api/challenge/open"

type Client struct {
	client phttp.Client
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) StartOpenGame(ctx context.Context) (*OpenGameResponse, error) {
	reqBody := url.Values{}
	reqBody.Set("name", "Halva vs. Povidlo Brawl")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, lichessOpenGameURL, strings.NewReader(reqBody.Encode()))
	if err != nil {
		return nil, errors.Wrap(err, "create new post req to lichess")
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.PostRequest(req)
	if err != nil {
		return nil, errors.Wrap(err, "do post req to lichess")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Wrap(err, "resp from lichess")
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read response")
	}

	var game OpenGameResponse
	if err := json.Unmarshal(data, &game); err != nil {
		return nil, errors.Wrapf(err, "unable to unmarshal response")
	}
	return &game, nil
}

type OpenGameResponse struct {
	Challenge struct {
		ID         string      `json:"id"`
		URL        string      `json:"url"`
		Status     string      `json:"status"`
		Challenger interface{} `json:"challenger"`
		DestUser   interface{} `json:"destUser"`
		Variant    struct {
			Key   string `json:"key"`
			Name  string `json:"name"`
			Short string `json:"short"`
		} `json:"variant"`
		Rated       bool   `json:"rated"`
		Speed       string `json:"speed"`
		TimeControl struct {
			Type string `json:"type"`
		} `json:"timeControl"`
		Color      string `json:"color"`
		FinalColor string `json:"finalColor"`
		Perf       struct {
			Icon string `json:"icon"`
			Name string `json:"name"`
		} `json:"perf"`
	} `json:"challenge"`
	SocketVersion int    `json:"socketVersion"`
	URLWhite      string `json:"urlWhite"`
	URLBlack      string `json:"urlBlack"`
}
