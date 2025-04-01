package main

import "github.com/andrieee44/notifydbus/pkg"

func config() []notifydbus.Notifier {
	var briIcons, volIcons []string

	briIcons = []string{"󰃞", "󰃟", "󰃝", "󰃠"}
	volIcons = []string{"󰝟", "󰕿", "󰖀", "󰕾"}

	return []notifydbus.Notifier{
		notifydbus.NewPipeWire(volIcons),
		notifydbus.NewMPD(),
		notifydbus.NewBrightness(briIcons),
	}
}
