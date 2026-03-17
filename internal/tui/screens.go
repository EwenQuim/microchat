package tui

// screen identifies which screen is currently active.
type screen int

const (
	screenIdentity screen = iota
	screenIdentities
	screenServers
	screenRooms
	screenContacts
)

// navigateMsg is sent by sub-models to trigger a screen transition.
type navigateMsg struct{ to screen }
