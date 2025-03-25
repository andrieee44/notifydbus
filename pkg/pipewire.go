package notifydbus

import (
	"fmt"

	"github.com/andrieee44/pwmon/pkg"
	"github.com/godbus/dbus/v5"
)

type pipeWire struct {
	notif    *pipeWireNotif
	infoChan <-chan *pwmon.Info
	errChan  <-chan error
	info     *pwmon.Info
}

type pipeWireNotif struct {
	data *NotificationData
}

func (notif *pipeWireNotif) Name() string {
	return "volume"
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
			AppName:       "PipeWire",
			Summary:       "Volume",
			Body:          "",
			ReplacesID:    true,
			Actions:       []string{},
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
	var (
		info *pwmon.Info
		err  error
	)

	for {
		select {
		case info = <-notifier.infoChan:
			if info.Volume != notifier.info.Volume || info.Mute != notifier.info.Mute {
				notifier.info = info
				notifier.notif.data.Hints["value"] = dbus.MakeVariant(info.Volume)
				notifier.notif.data.Body = fmt.Sprintf("Volume: %d%%, Mute: %t", info.Volume, info.Mute)

				return notifier.notif, nil
			}
		case err = <-notifier.errChan:
			return nil, err
		}
	}
}

func (notifier *pipeWire) Sleep() error {
	return nil
}

func (notifier *pipeWire) Close() error {
	return nil
}

func NewPipeWire() *pipeWire {
	return &pipeWire{}
}
