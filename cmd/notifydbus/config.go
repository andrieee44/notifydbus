package main

import "github.com/andrieee44/notifydbus/pkg"

func config() []notifydbus.Notifier {
	return []notifydbus.Notifier{
		notifydbus.NewVol(),
	}
}
