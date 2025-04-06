package notifydbus

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/adrg/xdg"
	"github.com/fhs/gompd/v2/mpd"
	"github.com/godbus/dbus/v5"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type musicOpts struct {
	icons  []string
	format string
}

type music struct {
	notif                      Notification
	opts                       *musicOpts
	player                     *musicPlayerNotif
	mixer                      *musicMixerNotif
	watcher                    *mpd.Watcher
	socketPath, event, artPath string
}

type musicPlayerNotif struct {
	data *NotificationData
}

type musicMixerNotif struct {
	data *NotificationData
}

func (notif *musicPlayerNotif) Name() string {
	return "MPD Player"
}

func (notif *musicPlayerNotif) Closed(_ uint32) error {
	return nil
}

func (notif *musicPlayerNotif) ActionInvoked(_ string) error {
	return nil
}

func (notif *musicPlayerNotif) ActivationToken(_ string) error {
	return nil
}

func (notif *musicPlayerNotif) Data() *NotificationData {
	return notif.data
}

func (notif *musicMixerNotif) Name() string {
	return "MPD Mixer"
}

func (notif *musicMixerNotif) Closed(_ uint32) error {
	return nil
}

func (notif *musicMixerNotif) ActionInvoked(_ string) error {
	return nil
}

func (notif *musicMixerNotif) ActivationToken(_ string) error {
	return nil
}

func (notif *musicMixerNotif) Data() *NotificationData {
	return notif.data
}

func (notifier *music) Init(_ []string) error {
	var err error

	notifier.socketPath, err = xdg.SearchRuntimeFile("mpd/socket")
	if err != nil {
		return err
	}

	notifier.artPath, err = xdg.CacheFile("notifydbus/mpd/albumArt.png")
	if err != nil {
		return err
	}

	notifier.watcher, err = mpd.NewWatcher("unix", notifier.socketPath, "", "player", "mixer")
	if err != nil {
		return err
	}

	notifier.player = &musicPlayerNotif{
		data: &NotificationData{
			AppName:       "notifydbus",
			ReplacesID:    true,
			ExpireTimeout: -1,

			Hints: map[string]dbus.Variant{
				"urgency": dbus.MakeVariant(0),
			},
		},
	}

	notifier.mixer = &musicMixerNotif{
		data: &NotificationData{
			AppName:       "notifydbus",
			Summary:       "MPD Volume",
			ReplacesID:    true,
			ExpireTimeout: -1,

			Hints: map[string]dbus.Variant{
				"urgency": dbus.MakeVariant(0),
			},
		},
	}

	select {
	case <-notifier.watcher.Event:
		return notifier.updatePlayer()
	case err = <-notifier.watcher.Error:
		return err
	}
}

func (notifier *music) Run() (Notification, error) {
	return notifier.notif, nil
}

func (notifier *music) Sleep() error {
	var (
		event string
		err   error
	)

	select {
	case event = <-notifier.watcher.Event:
		switch event {
		case "player":
			return notifier.updatePlayer()
		case "mixer":
			return notifier.updateMixer()
		default:
			return fmt.Errorf("%s: unexpected MPD event", event)
		}
	case err = <-notifier.watcher.Error:
		return err
	}
}

func (notifier *music) Close() error {
	return notifier.watcher.Close()
}

func (notifier *music) albumArt(file string) error {
	return ffmpeg.Input(file).Output(notifier.artPath, ffmpeg.KwArgs{"an": "", "c:v": "copy", "update": "1"}).OverWriteOutput().Silent(true).Run()
}

func (notifier *music) updatePlayer() error {
	var (
		client               *mpd.Client
		song, status, config mpd.Attrs
		data                 *NotificationData
		err                  error
	)

	client, err = mpd.Dial("unix", notifier.socketPath)
	if err != nil {
		return err
	}

	song, err = client.CurrentSong()
	if err != nil {
		return err
	}

	status, err = client.Status()
	if err != nil {
		return err
	}

	config, err = client.Command("config").Attrs()
	if err != nil {
		return err
	}

	err = notifier.albumArt(filepath.Join(config["music_directory"], song["file"]))
	if err != nil {
		return err
	}

	data = notifier.player.data

	switch status["state"] {
	case "play":
		data.Summary = "Playing"
	case "pause":
		data.Summary = "Paused"
	case "stop":
		data.Summary = "Stopped"
	}

	data.Body = regexp.MustCompilePOSIX("%[A-Za-z]+%").ReplaceAllStringFunc(notifier.opts.format, func(key string) string {
		return song[key[1:len(key)-1]]
	})

	data.Hints["image-path"] = dbus.MakeVariant(notifier.artPath)
	notifier.notif = notifier.player

	return client.Close()
}

func (notifier *music) updateMixer() error {
	var (
		client *mpd.Client
		status mpd.Attrs
		volume int
		err    error
	)

	client, err = mpd.Dial("unix", notifier.socketPath)
	if err != nil {
		return err
	}

	status, err = client.Status()
	if err != nil {
		return err
	}

	volume, err = strconv.Atoi(status["volume"])
	if err != nil {
		return err
	}

	notifier.mixer.data.Body = fmt.Sprintf("%s %d%%", icon(notifier.opts.icons[1:], 100, float64(volume)), volume)
	notifier.mixer.data.Hints["value"] = dbus.MakeVariant(volume)
	notifier.notif = notifier.mixer

	return client.Close()
}

func NewMPD(icons []string, format string) *music {
	return &music{
		opts: &musicOpts{
			icons:  icons,
			format: format,
		},
	}
}
