package celestrak

import "context"

// FetchTable fetches table data from table.php endpoint.
// TableFlags in the query will be applied if set.
func (c *Client) FetchTable(ctx context.Context, q Query) ([]byte, error) {
	return c.fetch(ctx, q, "table.php")
}

