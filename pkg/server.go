package notifydbus

import (
	"github.com/godbus/dbus/v5"
)

type Server struct {
	sysbus  *dbus.Conn
	errChan chan error
}

func registerIfaces(sysbus *dbus.Conn, ifaces []string) error {
	var (
		iface string
		err   error
	)

	for _, iface = range ifaces {
		err = sysbus.AddMatchSignal(dbus.WithMatchInterface("org.freedesktop.Notifications"), dbus.WithMatchMember(iface))
		if err != nil {
			return err
		}
	}

	return nil
}

func runNotifier(notifChan chan<- Notification, errChan chan<- error, capabilities []string, notifier Notifier) {
	var (
		notif Notification
		err   error
	)

	err = notifier.Init(capabilities)
	if err != nil {
		errChan <- err

		return
	}

	defer func() {
		err = notifier.Close()
		if err != nil {
			errChan <- err

			return
		}
	}()

	for {
		notif, err = notifier.Run()
		if err != nil {
			errChan <- err

			return
		}

		notifChan <- notif

		err = notifier.Sleep()
		if err != nil {
			errChan <- err

			return
		}
	}
}

func signalHandler(signal *dbus.Signal, notifs map[uint32]Notification) error {
	var (
		notif                      Notification
		id, reason                 uint32
		actionKey, activationToken string
		err                        error
		ok                         bool
	)

	switch signal.Name {
	case "org.freedesktop.Notifications.NotificationClosed":
		err = dbus.Store(signal.Body, &id, &reason)
		if err != nil {
			return err
		}

		notif, ok = notifs[id]
		if !ok {
			return nil
		}

		err = notif.Closed(reason)
		if err != nil {
			return err
		}
	case "org.freedesktop.Notifications.ActionInvoked":
		err = dbus.Store(signal.Body, &id, &actionKey)
		if err != nil {
			return err
		}

		notif, ok = notifs[id]
		if !ok {
			return nil
		}

		err = notif.ActionInvoked(actionKey)
		if err != nil {
			return err
		}
	case "org.freedesktop.Notifications.ActivationToken":
		err = dbus.Store(signal.Body, &id, &activationToken)
		if err != nil {
			return err
		}

		notif, ok = notifs[id]
		if !ok {
			return nil
		}

		err = notif.ActivationToken(activationToken)
		if err != nil {
			return err
		}
	}

	return nil
}

func notify(notifChan <-chan Notification, errChan chan<- error, signalChan <-chan *dbus.Signal, obj dbus.BusObject) {
	var (
		notifs     map[uint32]Notification
		notifNames map[string]uint32
		notif      Notification
		data       *NotificationData
		id         uint32
		name       string
		signal     *dbus.Signal
		err        error
	)

	notifs = make(map[uint32]Notification)
	notifNames = make(map[string]uint32)

	for {
		select {
		case notif = <-notifChan:
			data = notif.Data()

			if data.ReplacesID {
				name = notif.Name()
				id = notifNames[name]
			}

			println("before:", id)
			err = obj.Call("org.freedesktop.Notifications.Notify", 0, data.AppName, id, data.AppIcon, data.Summary, data.Body, data.Actions, data.Hints, data.ExpireTimeout).Store(&id)
			if err != nil {
				errChan <- err

				return
			}
			println("after:", id)

			notifs[id] = notif

			if data.ReplacesID {
				notifNames[name] = id
			}
		case signal = <-signalChan:
			err = signalHandler(signal, notifs)
			if err != nil {
				errChan <- err

				return
			}
		}
	}
}

func (server *Server) ErrChan() <-chan error {
	return server.errChan
}

func (server *Server) Close() error {
	return server.sysbus.Close()
}

func NewServer(notifiers []Notifier) (*Server, error) {
	var (
		server       *Server
		signalChan   chan *dbus.Signal
		notifChan    chan Notification
		obj          dbus.BusObject
		capabilities []string
		notifier     Notifier
		err          error
	)

	server = new(Server)
	notifChan = make(chan Notification)
	signalChan = make(chan *dbus.Signal)
	server.errChan = make(chan error)

	server.sysbus, err = dbus.ConnectSessionBus()
	if err != nil {
		return nil, err
	}

	obj = server.sysbus.Object("org.freedesktop.Notifications", "/org/freedesktop/Notifications")
	server.sysbus.Signal(signalChan)

	err = registerIfaces(server.sysbus, []string{"NotificationClosed", "ActionInvoked", "ActivationToken"})
	if err != nil {
		return nil, err
	}

	err = obj.Call("org.freedesktop.Notifications.GetCapabilities", 0).Store(&capabilities)
	if err != nil {
		return nil, err
	}

	for _, notifier = range notifiers {
		go runNotifier(notifChan, server.errChan, capabilities, notifier)
	}

	go notify(notifChan, server.errChan, signalChan, obj)

	return server, nil
}
