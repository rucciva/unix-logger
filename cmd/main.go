package main

import (
	"bufio"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"git.rucciva.one/rucciva/unix-logger/writer"
)

func getEnvOrDefault(name, def string) string {
	v := os.Getenv(name)
	if v == "" {
		return def
	}
	return v
}

func getEnvOrDefaultInt(name string, def int) int {
	v, err := strconv.Atoi(os.Getenv(name))
	if err != nil {
		return def
	}
	return v
}

type unixListener struct {
	path    string
	queue   chan bool
	writter writer.Writer

	listener net.Listener
}

func newUnixListener() *unixListener {
	path := getEnvOrDefault("UNIX_LOGGER_PATH", "/var/run/unix_logger.sock")
	queueSize := getEnvOrDefaultInt("UNIX_LOGGER_MAX_CONNECTION", 1024)
	w, _ := writer.NewStdoutWriter()
	return &unixListener{
		path:    path,
		queue:   make(chan bool, queueSize),
		writter: w,
	}
}

func (u *unixListener) dispatch(c net.Conn) {
	select {
	case u.queue <- true:
	default:
		c.Write([]byte("Queue Full\r\n"))
		c.Close()
		return
	}

	go func() {
		defer func() { <-u.queue }()
		s := bufio.NewScanner(c)
		s.Split(bufio.ScanLines)
		for s.Scan() {
			u.writter.WriteLine(s.Text())
		}
		c.Close()
	}()

}

func (u *unixListener) ListenAndServe() (err error) {
	os.Remove(u.path)

	u.listener, err = net.Listen("unix", u.path)
	if err != nil {
		return err
	}

	for {
		c, err := u.listener.Accept()
		if err != nil {
			log.Fatal("accept error: ", err)
			continue
		}
		u.dispatch(c)
	}
}

func (u *unixListener) Close() {
	u.listener.Close()
	os.Remove(u.path)
}

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	u := newUnixListener()
	go func() {
		<-sigs
		u.Close()
		os.Exit(1)
	}()

	log.Fatal(u.ListenAndServe())
}
