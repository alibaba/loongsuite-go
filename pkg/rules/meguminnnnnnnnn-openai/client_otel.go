//go:build ignore

package openai

func (c *Client) GetClientBaseURL() string {
	return c.config.BaseURL
}
