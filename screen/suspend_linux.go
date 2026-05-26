package screen

import (
	"syscall"
)

func (screen *Screen) Suspend() {
	screen.Screen.Suspend()

	// suspend the process
	pid := syscall.Getpid()
	tid := syscall.Gettid()
	err := syscall.Tgkill(pid, tid, syscall.SIGSTOP)
	if err != nil {
		panic(err)
	}

	// reset the state so we can get back to work again
	err = screen.Resume()
	if err != nil {
		panic(err)
	}
}
