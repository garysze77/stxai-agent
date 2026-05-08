//go:build windows

package service

import (
	"context"
	"log"

	"golang.org/x/sys/windows/svc"
)

// RunWindowsService runs the agent as a Windows service.
// It blocks until the service is stopped by SCM.
func RunWindowsService(ctx context.Context, run func(context.Context) error) error {
	return svc.Run(ServiceName, &windowsHandler{ctx: ctx, run: run})
}

type windowsHandler struct {
	ctx context.Context
	run func(context.Context) error
}

func (h *windowsHandler) Execute(args []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (bool, uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown

	s <- svc.Status{State: svc.StartPending}

	// Start the agent in a goroutine
	errCh := make(chan error, 1)
	ctx, cancel := context.WithCancel(h.ctx)
	defer cancel()

	go func() {
		errCh <- h.run(ctx)
	}()

	s <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	for {
		select {
		case err := <-errCh:
			if err != nil {
				log.Printf("agent exited with error: %v", err)
			}
			s <- svc.Status{State: svc.Stopped}
			return false, 0
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				s <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				s <- svc.Status{State: svc.StopPending}
				cancel()
				<-errCh // wait for agent to stop
				s <- svc.Status{State: svc.Stopped}
				return false, 0
			default:
				log.Printf("unexpected service control: %v", c.Cmd)
			}
		}
	}
}
