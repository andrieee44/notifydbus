package main

import "github.com/andrieee44/notifydbus/pkg"

func config() []notifydbus.Notifier {
	return []notifydbus.Notifier{
		notifydbus.NewPipeWire(),
		notifydbus.NewMPD(),
	}
}
