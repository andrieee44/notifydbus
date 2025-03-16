package notifydbus

import (
	"fmt"
	"time"

	"github.com/godbus/dbus/v5"
)

type vol struct {
	notif *volNotif
}

type volNotif struct {
	counter int
}

func (notif *volNotif) Name() string {
	return "volume"
}

func (notif *volNotif) Closed(_ uint32) error {
	return nil
}

func (notif *volNotif) ActionInvoked(_ string) error {
	return nil
}

func (notif *volNotif) ActivationToken(_ string) error {
	return nil
}

func (notif *volNotif) Data() *NotificationData {
	return &NotificationData{
		AppName:       "PipeWire",
		Summary:       "Volume",
		Body:          fmt.Sprintf("%d", notif.counter),
		ReplacesID:    true,
		Actions:       []string{},
		Hints:         map[string]dbus.Variant{},
		ExpireTimeout: -1,
	}
}

func (notifier *vol) Init(_ []string) error {
	notifier.notif = &volNotif{}

	return nil
}

func (notifier *vol) Run() (Notification, error) {
	return notifier.notif, nil
}

func (notifier *vol) Sleep() error {
	time.Sleep(time.Second)
	notifier.notif.counter++

	return nil
}

func (notifier *vol) Close() error {
	return nil
}

func NewVol() *vol {
	return &vol{}
}
