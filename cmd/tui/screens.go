package main

// screen identifies which screen is currently active.
type screen int

const (
	screenIdentity screen = iota
	screenServers
	screenRooms
)

// navigateMsg is sent by sub-models to trigger a screen transition.
type navigateMsg struct{ to screen }
