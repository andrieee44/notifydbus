package main

import "github.com/andrieee44/notifydbus/pkg"

func newConfig() []notifydbus.Module {
	return []notifydbus.Module{
		notifydbus.NewVol(),
	}
}
