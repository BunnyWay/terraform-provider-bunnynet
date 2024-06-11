package api

type Client struct {
	apiKey string
	apiUrl string
}

func NewClient(apiKey string, apiUrl string) *Client {
	return &Client{
		apiKey: apiKey,
		apiUrl: apiUrl,
	}
}
