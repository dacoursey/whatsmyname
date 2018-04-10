package actions

import "github.com/gobuffalo/buffalo"

// SearchSearch default implementation.
func SearchHandler(c buffalo.Context) error {
	return c.Render(200, r.HTML("search/search.html"))
}
