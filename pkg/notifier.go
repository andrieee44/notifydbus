package notifydbus

import "github.com/godbus/dbus/v5"

type Notifier struct {
	sysbus *dbus.Conn
	obj    dbus.BusObject
}

func (notifier *Notifier) Notify(notif *Notification) (uint32, error) {
	var (
		id  uint32
		err error
	)

	err = notifier.obj.Call("org.freedesktop.Notifications.Notify", 0, notif.AppName, notif.ReplacesID, notif.AppIcon, notif.Summary, notif.Body, notif.Actions, notif.Hints, notif.ExpireTimeout).Store(&id)

	return id, err
}

func (notifier *Notifier) Capabilities() ([]string, error) {
	var (
		capabilities []string
		err          error
	)

	err = notifier.obj.Call("org.freedesktop.Notifications.GetCapabilities", 0).Store(&capabilities)

	return capabilities, err
}

func (notifier *Notifier) Close() error {
	return notifier.sysbus.Close()
}

func (notifier *Notifier) matchInterface(iface string) error {
	return notifier.sysbus.AddMatchSignal(dbus.WithMatchInterface("org.freedesktop.Notifications"), dbus.WithMatchMember(iface))
}

func NewNotifier(signals chan *dbus.Signal) (*Notifier, error) {
	var (
		notifier *Notifier
		err      error
	)

	notifier = new(Notifier)
	notifier.sysbus, err = dbus.ConnectSessionBus()
	if err != nil {
		return nil, err
	}

	notifier.obj = notifier.sysbus.Object("org.freedesktop.Notifications", "/org/freedesktop/Notifications")
	notifier.sysbus.Signal(signals)

	err = notifier.matchInterface("NotificationClosed")
	if err != nil {
		return nil, err
	}

	err = notifier.matchInterface("ActionInvoked")
	if err != nil {
		return nil, err
	}

	err = notifier.matchInterface("ActivationToken")

	return notifier, err
}
