package util

type State struct {
	Alert bool
	AlertRooms []string
	Playing bool
	Test bool
}

//Alert and Connected here are used when drawing - our state keeps track of 
//if we need to play sound and what rooms need to be displayed as needing assistance
type Room struct {
	Room string
	Ip string
	Alert bool
	Connected bool
}

type Update struct {
	Index int
	Type string
	Connected bool
	Room string
}