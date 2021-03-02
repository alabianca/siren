package main

import (
	"archive/tar"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

var hostFlag = flag.String("host", "", "what address to copy to")
var portFlag = flag.String("port", "4200", "what port to connect to or listen on")
var watchRootFlag = flag.String("watch", ".", "which directory to watch")
var listenFlag = flag.Bool("listen", false, "Determines if we should listen as a server")

func main() {
	flag.Parse()

	if *listenFlag {
		listen()
	} else {
		watch()
	}
}

func watch() {
	watcher, err := fsnotify.NewWatcher()
	must(err)

	defer watcher.Close()

	err = filepath.Walk(*watchRootFlag, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			fmt.Printf("Watching %s\n", path)
			watcher.Add(path)
		}

		return nil
	})

	must(err)

	// get notified whenever there is a change
	copyToDestination(notifyOnEvent(watcher, fsnotify.Write))

}

type ChangeEvent struct {
	Time time.Time
	Name string
	Op   fsnotify.Op
}

func (c ChangeEvent) String() string {
	return fmt.Sprintf("%s %s %d", c.Time, c.Name, c.Op)
}

func notifyOnEvent(watcher *fsnotify.Watcher, op fsnotify.Op) <-chan ChangeEvent {
	out := make(chan ChangeEvent)

	go func() {
		for {
			select {
			case ev := <-watcher.Events:
				if ev.Op == op {
					fmt.Println("Watcher Event ", ev)
					out <- ChangeEvent{
						Time: time.Now(),
						Name: ev.Name,
						Op:   ev.Op,
					}
				}
			}
		}
	}()

	return out
}

func copyToDestination(changes <-chan ChangeEvent) {

	for ch := range changes {
		fmt.Println("File Change ", ch)
		conn, err := net.Dial("tcp", net.JoinHostPort(*hostFlag, *portFlag))
		if err != nil {
			fmt.Println("Connection error ", err)
			continue
		}
		err = CreateTarballFrom(*watchRootFlag, conn)
		if err != nil {
			fmt.Printf("Targball Error %s\n", err)
		}

	}
}

func listen() {
	l, err := net.Listen("tcp", net.JoinHostPort("", *portFlag))
	must(err)

	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}
		fmt.Println("Received connection")
		err = handleConnection(conn)
		if err != nil {
			fmt.Printf("Handle Connection Error %s\n", err)
		}

	}
}

func handleConnection(conn net.Conn) error {
	gzipReader, err := gzip.NewReader(conn)
	if err != nil {
		return err
	}

	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			fmt.Printf("Seeing a directory with name %s\n", header.Name)
			err := os.Mkdir(header.Name, header.FileInfo().Mode())
			if err != nil {
				return err
			}
			continue
		case tar.TypeReg:
			fmt.Printf("Seeing a file with name %s\n", header.Name)
			file, err := create(header.Name)
			if err != nil {
				return err
			}
			io.Copy(file, tarReader)
			file.Close()
		}
	}

	return nil
}

func create(p string) (*os.File, error) {
	parts := strings.Split(p, string(os.PathSeparator))
	// create all the directories if they don't exist
	if len(parts) > 1 {
		err := os.MkdirAll(strings.Join(parts[:len(parts)-1], string(os.PathSeparator)), 0777)
		if err != nil {
			return nil, err
		}
	}

	return os.Create(p)
}

func must(err error) {
	if err != nil {
		fmt.Printf("Error %s\n", err)
		os.Exit(1)
	}
}
