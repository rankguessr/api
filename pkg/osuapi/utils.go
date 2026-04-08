package osuapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	osuBaseURL = "https://osu.ppy.sh"
)

var (
	apiV2BaseURL = fmt.Sprintf("%s/api/v2", osuBaseURL)
)

func SetDefaultHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json")
}

func NewWebRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%s%s", osuBaseURL, path), body)
	if err != nil {
		return nil, err
	}

	SetDefaultHeaders(req)
	return req, nil
}

func NewAPIv2Request(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%s%s", apiV2BaseURL, path), body)
	if err != nil {
		return nil, err
	}

	SetDefaultHeaders(req)

	return req, nil
}

func SetAuthHeader(req *http.Request, accessToken string) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
}

func Do(req *http.Request) (*http.Response, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}

func DoAndParse[T any](req *http.Request) (val T, err error) {
	resp, err := Do(req)
	if err != nil {
		return val, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&val)
	if err != nil {
		return val, fmt.Errorf("failed to decode API response: %w", err)
	}

	return val, nil
}
