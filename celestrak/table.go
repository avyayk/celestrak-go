package celestrak

import "context"

// FetchTable fetches table data from table.php endpoint.
func (c *Client) FetchTable(ctx context.Context, q Query) ([]byte, error) {
	return c.fetch(ctx, q, "table.php")
}

