package notifydbus

type Notifier interface {
	Init([]string) error
	Run() (Notification, error)
	Sleep() error
	Close() error
}
