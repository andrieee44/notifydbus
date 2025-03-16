package notifydbus

import "github.com/godbus/dbus/v5"

type Notification interface {
	Name() string
	Closed(uint32) error
	ActionInvoked(string) error
	ActivationToken(string) error
	Data() *NotificationData
}

type NotificationData struct {
	AppName, AppIcon, Summary, Body string
	ReplacesID                      bool
	Actions                         []string
	Hints                           map[string]dbus.Variant
	ExpireTimeout                   int32
}
