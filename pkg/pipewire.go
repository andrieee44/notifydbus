package notifydbus

import (
	"fmt"

	"github.com/andrieee44/pwmon/pkg"
	"github.com/godbus/dbus/v5"
)

type pipeWireOpts struct {
	icons []string
}

type pipeWire struct {
	opts     *pipeWireOpts
	notif    *pipeWireNotif
	infoChan <-chan *pwmon.Info
	errChan  <-chan error
	info     *pwmon.Info
}

type pipeWireNotif struct {
	data *NotificationData
}

func (notif *pipeWireNotif) Name() string {
	return "PipeWire"
}

func (notif *pipeWireNotif) Closed(_ uint32) error {
	return nil
}

func (notif *pipeWireNotif) ActionInvoked(_ string) error {
	return nil
}

func (notif *pipeWireNotif) ActivationToken(_ string) error {
	return nil
}

func (notif *pipeWireNotif) Data() *NotificationData {
	return notif.data
}

func (notifier *pipeWire) Init(_ []string) error {
	var err error

	notifier.notif = &pipeWireNotif{
		data: &NotificationData{
			AppName:       "notifydbus",
			Summary:       "PipeWire Volume",
			ReplacesID:    true,
			ExpireTimeout: -1,

			Hints: map[string]dbus.Variant{
				"urgency": dbus.MakeVariant(0),
			},
		},
	}

	notifier.infoChan, notifier.errChan, err = pwmon.Monitor()
	if err != nil {
		return err
	}

	select {
	case notifier.info = <-notifier.infoChan:
		return nil
	case err = <-notifier.errChan:
		return err
	}
}

func (notifier *pipeWire) Run() (Notification, error) {
	var err error

	select {
	case notifier.info = <-notifier.infoChan:
		if notifier.info.Mute {
			notifier.notif.data.Body = fmt.Sprintf("%s muted", notifier.opts.icons[0])
			notifier.notif.data.Hints["value"] = dbus.MakeVariant(0)

			return notifier.notif, nil
		}

		notifier.notif.data.Body = fmt.Sprintf("%s %d%%", icon(notifier.opts.icons[1:], 100, float64(notifier.info.Volume)), notifier.info.Volume)
		notifier.notif.data.Hints["value"] = dbus.MakeVariant(notifier.info.Volume)

		return notifier.notif, nil
	case err = <-notifier.errChan:
		return nil, err
	}
}

func (notifier *pipeWire) Sleep() error {
	return nil
}

func (notifier *pipeWire) Close() error {
	return nil
}

func NewPipeWire(icons []string) *pipeWire {
	return &pipeWire{
		opts: &pipeWireOpts{icons: icons},
	}
}
