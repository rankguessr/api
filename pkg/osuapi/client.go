package osuapi

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
)

type Client struct {
	ClientID     string
	ClientSecret string
	AppURL       string
}

func NewClient(clientID, clientSecret, appURL string) *Client {
	return &Client{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AppURL:       appURL,
	}
}

func (c *Client) ExchangeToken(ctx context.Context, code string) (ExchangeTokenResponse, error) {
	vals := url.Values{}
	vals.Set("code", code)
	vals.Set("client_id", c.ClientID)
	vals.Set("client_secret", c.ClientSecret)
	vals.Set("grant_type", "authorization_code")
	vals.Set("redirect_uri", fmt.Sprintf("%s/auth/callback", c.AppURL))

	body := strings.NewReader(vals.Encode())
	req, err := NewWebRequest(ctx, "POST", "/oauth/token", body)
	if err != nil {
		return ExchangeTokenResponse{}, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return DoAndParse[ExchangeTokenResponse](req)
}

func (c *Client) GetMe(ctx context.Context, accessToken string) (UserExtended, error) {
	req, err := NewAPIv2Request(ctx, "GET", "/me/osu", nil)
	if err != nil {
		return UserExtended{}, err
	}

	SetAuthHeader(req, accessToken)

	return DoAndParse[UserExtended](req)
}

func (c *Client) TokenRefresh(ctx context.Context, refreshToken string) (ExchangeTokenResponse, error) {
	vals := url.Values{}
	vals.Set("client_id", c.ClientID)
	vals.Set("client_secret", c.ClientSecret)
	vals.Set("grant_type", "refresh_token")
	vals.Set("refresh_token", refreshToken)

	body := strings.NewReader(vals.Encode())
	req, err := NewWebRequest(ctx, "POST", "/oauth/token", body)
	if err != nil {
		return ExchangeTokenResponse{}, err
	}

	return DoAndParse[ExchangeTokenResponse](req)
}

func (c *Client) GetClientAccessToken(ctx context.Context) (ClientToken, error) {
	vals := url.Values{}
	vals.Set("client_id", c.ClientID)
	vals.Set("client_secret", c.ClientSecret)
	vals.Set("grant_type", "client_credentials")
	vals.Set("scope", "public")

	data := strings.NewReader(vals.Encode())
	req, err := NewWebRequest(ctx, "POST", "/oauth/token", data)
	if err != nil {
		return ClientToken{}, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return DoAndParse[ClientToken](req)
}

func (c *Client) GetUser(ctx context.Context, accessToken string, userId int) (UserExtended, error) {
	req, err := NewAPIv2Request(ctx, "GET", fmt.Sprintf("/users/%d/osu", userId), nil)
	if err != nil {
		return UserExtended{}, err
	}

	SetAuthHeader(req, accessToken)

	return DoAndParse[UserExtended](req)
}

func (c *Client) GetUserScores(ctx context.Context, accessToken string, userId int) ([]Score, error) {
	vals := url.Values{}
	vals.Set("mode", "osu")
	vals.Set("limit", "20")
	vals.Set("offset", "0")

	path := fmt.Sprintf("/users/%d/scores/best?%s", userId, vals.Encode())
	req, err := NewAPIv2Request(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	SetAuthHeader(req, accessToken)
	return DoAndParse[[]Score](req)
}

func (c *Client) GetScore(ctx context.Context, accessToken string, scoreId int) (Score, error) {
	req, err := NewAPIv2Request(ctx, "GET", fmt.Sprintf("/scores/%d", scoreId), nil)
	if err != nil {
		return Score{}, err
	}

	SetAuthHeader(req, accessToken)
	return DoAndParse[Score](req)
}

func (c *Client) DownloadReplay(ctx context.Context, accessToken string, scoreId int) ([]byte, error) {
	req, err := NewAPIv2Request(ctx, "GET", fmt.Sprintf("/scores/%d/download", scoreId), nil)
	if err != nil {
		return nil, err
	}

	SetAuthHeader(req, accessToken)

	resp, err := Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (c *Client) GetMultiRooms(ctx context.Context, accessToken string) ([]MultiRoom, error) {
	req, err := NewAPIv2Request(ctx, "GET", "/rooms", nil)
	if err != nil {
		return nil, err
	}
	SetAuthHeader(req, accessToken)
	return DoAndParse[[]MultiRoom](req)
}

func (c *Client) GetRankings(ctx context.Context, accessToken string, cursor *Cursor) (Rankings, error) {
	data := url.Values{}
	if cursor != nil {
		data.Set("cursor[page]", strconv.Itoa(cursor.Page))
	}

	path := fmt.Sprintf("/rankings/osu/global?%s", data.Encode())
	req, err := NewAPIv2Request(ctx, "GET", path, nil)
	if err != nil {
		return Rankings{}, err
	}

	SetAuthHeader(req, accessToken)

	return DoAndParse[Rankings](req)
}
