package notifydbus

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/godbus/dbus/v5"
)

type brightnessOpts struct {
	icons []string
}

type brightness struct {
	opts    *brightnessOpts
	notif   *brightnessNotif
	watcher *fsnotify.Watcher
	maxBri  int
}

type brightnessNotif struct {
	data *NotificationData
}

func (notif *brightnessNotif) Name() string {
	return "brightness"
}

func (notif *brightnessNotif) Closed(_ uint32) error {
	return nil
}

func (notif *brightnessNotif) ActionInvoked(_ string) error {
	return nil
}

func (notif *brightnessNotif) ActivationToken(_ string) error {
	return nil
}

func (notif *brightnessNotif) Data() *NotificationData {
	return notif.data
}

func (notifier *brightness) Init(_ []string) error {
	var err error

	notifier.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	err = notifier.watcher.Add("/sys/class/backlight/intel_backlight/brightness")
	if err != nil {
		return err
	}

	notifier.maxBri, err = fileAtoi("/sys/class/backlight/intel_backlight/max_brightness")
	if err != nil {
		return err
	}

	notifier.notif = &brightnessNotif{
		data: &NotificationData{
			AppName:       "notifydbus",
			Summary:       "Brightness",
			ReplacesID:    true,
			ExpireTimeout: -1,

			Hints: map[string]dbus.Variant{
				"urgency": dbus.MakeVariant(0),
			},
		},
	}

	return nil
}

func (notifier *brightness) Run() (Notification, error) {
	var (
		event fsnotify.Event
		bri   int
		perc  float64
		err   error
	)

	for {
		select {
		case event = <-notifier.watcher.Events:
			if !event.Has(fsnotify.Write) {
				continue
			}

			bri, err = fileAtoi("/sys/class/backlight/intel_backlight/brightness")
			if err != nil {
				return nil, err
			}

			perc = float64(bri) / float64(notifier.maxBri) * 100
			bri = int(perc + 0.5)

			notifier.notif.data.Body = fmt.Sprintf("%s %d%%", icon(notifier.opts.icons, 100, perc), bri)
			notifier.notif.data.Hints["value"] = dbus.MakeVariant(bri)

			return notifier.notif, nil
		case err = <-notifier.watcher.Errors:
			return nil, err
		}
	}
}

func (notifier *brightness) Sleep() error {
	return nil
}

func (notifier *brightness) Close() error {
	return nil
}

func NewBrightness(icons []string) *brightness {
	return &brightness{
		opts: &brightnessOpts{icons: icons},
	}
}
