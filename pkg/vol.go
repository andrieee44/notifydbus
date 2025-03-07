package notifydbus

type vol struct {
	notif *Notification
}

func (mod *vol) Init() error {
	mod.notif = NewNotification("body")

	return nil
}

func (mod *vol) Run() (*Notification, error) {
	mod.notif.AppName = "PipeWire"
	mod.notif.Summary = "summary"

	return mod.notif, nil
}

func (mod *vol) Sleep() error {
	for {
		select {
		case <-mod.notif.Id:
		case <-mod.notif.Closed:
			return nil
		case <-mod.notif.ActionInvoked:
		case <-mod.notif.ActivationToken:
		}
	}
}

func (mod *vol) Cleanup() error {
	return nil
}

func NewVol() *vol {
	return &vol{}
}
