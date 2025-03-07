package notifydbus

import "github.com/godbus/dbus/v5"

type Notification struct {
	AppName, AppIcon, Summary, Body string
	ReplacesID                      uint32
	Actions                         []string
	Hints                           map[string]dbus.Variant
	ExpireTimeout                   int32
	Id, Closed                      chan uint32
	ActionInvoked, ActivationToken  chan string
}

func NewNotification(body string) *Notification {
	return &Notification{
		Body:            body,
		ExpireTimeout:   -1,
		Id:              make(chan uint32),
		Closed:          make(chan uint32),
		ActionInvoked:   make(chan string),
		ActivationToken: make(chan string),
	}
}
