package main

import (
	"fmt"

	"github.com/rjeczalik/notify"
)

func handle(ei notify.EventInfo) {
	fmt.Println(ei)

	//if err := eval("/work/run.sh", "ubuntu:trusty"); err != nil {
	//	fmt.Println(err)
	//}
}

func watch() {
	c := make(chan notify.EventInfo, 1)

	if err := notify.Watch("./...", c, notify.All); err != nil {
		panic(err)
	}
	defer notify.Stop(c)

	for {
		ei := <-c
		handle(ei)
	}
}
