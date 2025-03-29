package notifydbus

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/fhs/gompd/v2/mpd"
	"github.com/godbus/dbus/v5"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type music struct {
	notif                              *musicNotif
	watcher                            *mpd.Watcher
	socketPath, summary, body, artPath string
}

type musicNotif struct {
	data *NotificationData
}

func (notif *musicNotif) Name() string {
	return "MPD"
}

func (notif *musicNotif) Closed(_ uint32) error {
	return nil
}

func (notif *musicNotif) ActionInvoked(_ string) error {
	return nil
}

func (notif *musicNotif) ActivationToken(_ string) error {
	return nil
}

func (notif *musicNotif) Data() *NotificationData {
	return notif.data
}

func (notifier *music) Init(_ []string) error {
	var err error

	notifier.socketPath = filepath.Join(os.Getenv("XDG_RUNTIME_DIR"), "mpd/socket")

	notifier.artPath, err = xdg.CacheFile("notifydbus/mpd/albumArt.png")
	if err != nil {
		return err
	}

	notifier.watcher, err = mpd.NewWatcher("unix", notifier.socketPath, "", "player")
	if err != nil {
		return err
	}

	notifier.notif = &musicNotif{
		data: &NotificationData{
			AppName:       "MPD",
			ReplacesID:    true,
			ExpireTimeout: -1,

			Hints: map[string]dbus.Variant{
				"urgency": dbus.MakeVariant(0),
			},
		},
	}

	select {
	case <-notifier.watcher.Event:
		return notifier.updateOutput()
	case err = <-notifier.watcher.Error:
		return err
	}
}

func (notifier *music) Run() (Notification, error) {
	notifier.notif.data.Summary = notifier.summary
	notifier.notif.data.Body = notifier.body
	notifier.notif.data.Hints["image-path"] = dbus.MakeVariant(notifier.artPath)

	return notifier.notif, nil
}

func (notifier *music) Sleep() error {
	var err error

	select {
	case <-notifier.watcher.Event:
		return notifier.updateOutput()
	case err = <-notifier.watcher.Error:
		return err
	}
}

func (notifier *music) Close() error {
	return nil
}

func (notifier *music) albumArt(file string) error {
	return ffmpeg.Input(file).Output(notifier.artPath, ffmpeg.KwArgs{"an": "", "c:v": "copy", "update": "1"}).OverWriteOutput().Silent(true).Run()
}

func (notifier *music) updateOutput() error {
	var (
		client               *mpd.Client
		song, status, config mpd.Attrs
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

	switch status["state"] {
	case "play":
		notifier.summary = "Playing"
	case "pause":
		notifier.summary = "Paused"
	case "stop":
		notifier.summary = "Stopped"
	}

	notifier.body = fmt.Sprintf("%s - %s - %s - %s", song["AlbumArtist"], song["Track"], song["Album"], song["Title"])

	return client.Close()
}

func NewMPD() *music {
	return &music{}
}
