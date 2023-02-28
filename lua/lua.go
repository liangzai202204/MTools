package lua

import (
	_ "embed"
)

var (

	//go:embed unlock.lua
	LuaUnlock string

	//go:embed refresh.lua
	LuaRefresh string

	//go:embed lock.lua
	LuaLock string
)
