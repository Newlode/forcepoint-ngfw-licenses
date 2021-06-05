package ngfwlicenses

import (
	"github.com/mattn/go-colorable"
	"github.com/mbndr/logo"
)

var (
	Logger *logo.Logger
)

func init() {
	vrai := true
	colorable.EnableColorsStdout(&vrai)
}
