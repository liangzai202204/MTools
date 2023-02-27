package lua

import (
	_ "embed"
)

var (

	//go:embed unlock.lua
	luaUnlock string

	//go:embed refresh.lua
	luaRefresh string

	//go:embed lock.lua
	luaLock string
)
