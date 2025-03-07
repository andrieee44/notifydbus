package notifydbus

type Module interface {
	Init() error
	Run() (*Notification, error)
	Sleep() error
	Cleanup() error
}
