package celestrak

import "context"

// FetchGP fetches GP data as raw bytes from gp.php endpoint.
// This works for TLE/XML/CSV/JSON formats.
func (c *Client) FetchGP(ctx context.Context, q Query) ([]byte, error) {
	return c.fetch(ctx, q, "gp.php")
}

// FetchGPFirst fetches first GP data available from gp-first.php endpoint.
func (c *Client) FetchGPFirst(ctx context.Context, q Query) ([]byte, error) {
	return c.fetch(ctx, q, "gp-first.php")
}

// FetchGPLast fetches last GP data available from gp-last.php endpoint.
func (c *Client) FetchGPLast(ctx context.Context, q Query) ([]byte, error) {
	return c.fetch(ctx, q, "gp-last.php")
}

