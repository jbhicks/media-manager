package assets

import (
	_ "embed"
)

//go:embed logo.png
var LogoPNG []byte

//go:embed logo.svg
var LogoSVG []byte
