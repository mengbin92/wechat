package main

import (
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

var (
	log *zap.SugaredLogger
	gid int
)

func init() {
	log = defaultLogger().Sugar()
}

func main() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("NewWatcher failed: ", err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case _, ok := <-watcher.Events:
				if !ok {
					return
				}
				stopWork()
				startWork()
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Errorf("error: %s", err)
			}
		}
	}()

	go startWork()

	err = watcher.Add("./conf")
	if err != nil {
		log.Fatal("Add failed:", err)
	}

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutdown watcher")
}

func startWork() {
	cmd := exec.Command("./wechat")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Start(); err != nil {
		log.Errorf("start wechat error: %s", err.Error())
		os.Exit(-1)
	}
	gid = cmd.Process.Pid
}

func stopWork() {
	cmd := exec.Command("kill", "-9", strconv.Itoa(gid))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		log.Errorf("start wechat error: %s", err.Error())
		os.Exit(-1)
	}
}
