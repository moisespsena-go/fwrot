package main

import (
	"fmt"
	"time"

	logrotate "github.com/moisespsena-go/glogrotate"
)

func main() {
	r := logrotate.New("app.log",
		logrotate.Options{
			HistoryPath: "%h/%m/%s",
			Duration:    logrotate.Minutely,
		})
	defer r.Close()

	var t time.Time
	h, _ := r.History(t, t, 0)
	for _, e := range h {
		fmt.Println(e.At(), e.Path())
	}

	for {
		fmt.Fprintln(r, "hello")
		fmt.Fprintln(r, "World!")
		fmt.Fprintln(r, "fo!")
		<-time.After(time.Second * 15)
	}
}
