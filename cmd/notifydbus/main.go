package main

import "github.com/andrieee44/notifydbus/pkg"

func main() {
	var (
		server *notifydbus.Server
		err    error
	)

	server, err = notifydbus.NewServer(config())
	if err != nil {
		panic(err)
	}

	defer func() {
		err = server.Close()
		if err != nil {
			panic(err)
		}
	}()

	panic(<-server.ErrChan())
}
