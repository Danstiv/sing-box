//go:build tools

package main

// Vendor-only imports. gobind loads these gomobile packages at binding
// generation time, but no sing-box source imports them, so `go mod vendor`
// would otherwise omit them and the AAR build fails with
// "no Go package in github.com/sagernet/gomobile/bind[/java]".
//
// The `tools` build tag keeps this file out of normal builds (bind/java is
// //go:build android and would not compile on the host), while `go mod vendor`
// still scans it and vendors the packages ignoring build constraints.
import (
	_ "github.com/sagernet/gomobile/bind"
	_ "github.com/sagernet/gomobile/bind/java"
)
