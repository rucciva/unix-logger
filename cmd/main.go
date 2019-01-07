package main

import (
	"bufio"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"os/user"
	"strconv"
	"strings"
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
	path     string
	fileMode os.FileMode
	uid      int
	gid      int
	queue    chan bool
	writter  writer.Writer

	listener net.Listener
}

func newUnixListener() (*unixListener, error) {
	path := getEnvOrDefault("UNIX_LOGGER_PATH", "/var/run/unix_logger.sock")

	user, err := user.Current()
	if err != nil {
		return nil, err
	}
	owner := getEnvOrDefault("UNIX_LOGGER_FILE_OWNER", user.Uid+":"+user.Gid)
	id := strings.Split(owner, ":")
	if len(id) != 2 {
		return nil, errors.New("invalid UNIX_LOGGER_FILE_OWNER env. Expected <uid>:<gid>")
	}
	uid, _ := strconv.Atoi(id[0])
	gid, _ := strconv.Atoi(id[1])

	mode := getEnvOrDefault("UNIX_LOGGER_FILE_MODE", "0700")
	m, err := strconv.ParseUint(mode, 0, 32)
	if err != nil {
		return nil, err
	}

	queueSize := getEnvOrDefaultInt("UNIX_LOGGER_MAX_CONNECTION", 1024)
	w, _ := writer.NewStdoutWriter()
	return &unixListener{
		path:     path,
		fileMode: os.FileMode(m),
		uid:      uid,
		gid:      gid,
		queue:    make(chan bool, queueSize),
		writter:  w,
	}, nil
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
	if err := os.Chmod(u.path, u.fileMode); err != nil {
		return err
	}
	if err := os.Chown(u.path, u.uid, u.gid); err != nil {
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

	u, err := newUnixListener()
	if err != nil {
		log.Fatal("new_unix_listener_failed", err)
	}
	go func() {
		<-sigs
		u.Close()
		os.Exit(1)
	}()

	log.Fatal(u.ListenAndServe())
}
