package main

import "github.com/andrieee44/notifydbus/pkg"

func panicIf(err error) {
	if err != nil {
		panic(err)
	}
}

func runModule(notifChan chan<- *notifydbus.Notification, mod notifydbus.Module) {
	var (
		notif *notifydbus.Notification
		err   error
	)

	panicIf(mod.Init())

	defer func() {
		panicIf(mod.Cleanup())
	}()

	for {
		notif, err = mod.Run()
		panicIf(err)

		notifChan <- notif
		panicIf(mod.Sleep())
	}
}

func runConfig(notifChan chan<- *notifydbus.Notification) {
	var mod notifydbus.Module

	for _, mod = range newConfig() {
		go runModule(notifChan, mod)
	}
}
