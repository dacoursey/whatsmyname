package grifts

import (
	"github.com/dacoursey/whatsmyname/actions"
	"github.com/gobuffalo/buffalo"
)

func init() {
	buffalo.Grifts(actions.App())
}
