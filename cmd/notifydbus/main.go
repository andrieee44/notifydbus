package main

import (
	"github.com/andrieee44/notifydbus/pkg"
	"github.com/godbus/dbus/v5"
)

func main() {
	var (
		notifier                   *notifydbus.Notifier
		signals                    chan *dbus.Signal
		notifChan                  chan *notifydbus.Notification
		notifs                     map[uint32]*notifydbus.Notification
		notif                      *notifydbus.Notification
		signal                     *dbus.Signal
		id, reason                 uint32
		actionKey, activationToken string
		ok                         bool
		err                        error
	)

	signals = make(chan *dbus.Signal)
	notifier, err = notifydbus.NewNotifier(signals)
	panicIf(err)

	defer func() {
		panicIf(notifier.Close())
	}()

	notifChan = make(chan *notifydbus.Notification)
	runConfig(notifChan)
	notifs = make(map[uint32]*notifydbus.Notification)

	for {
		select {
		case notif = <-notifChan:
			id, err = notifier.Notify(notif)
			panicIf(err)

			notif.Id <- id
			notifs[id] = notif
		case signal = <-signals:
			switch signal.Name {
			case "org.freedesktop.Notifications.NotificationClosed":
				err = dbus.Store(signal.Body, &id, &reason)
				panicIf(err)

				notif, ok = notifs[id]
				if !ok {
					continue
				}

				notifs[id].Closed <- reason
				delete(notifs, id)
			case "org.freedesktop.Notifications.ActionInvoked":
				err = dbus.Store(signal.Body, &actionKey)
				panicIf(err)

				notif, ok = notifs[id]
				if !ok {
					continue
				}

				notifs[id].ActionInvoked <- actionKey
			case "org.freedesktop.Notifications.ActivationToken":
				err = dbus.Store(signal.Body, &activationToken)
				panicIf(err)

				notif, ok = notifs[id]
				if !ok {
					continue
				}

				notifs[id].ActivationToken <- activationToken
			}
		}
	}
}
