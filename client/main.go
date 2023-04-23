package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// type Room struct {
// 	Connections[1_000_000]
// }

func srv(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		fmt.Println(r.Header)
		panic(err)
	}
	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
	defer cancel()

	var v any
	pkts := 0
	for {
		err = wsjson.Read(ctx, c, &v)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
		}
		fmt.Printf("received: %v #%d\n", v, pkts)
		pkts++
	}
	c.Close(websocket.StatusNormalClosure, "")
	elapsed := time.Since(now)
	secs := int(elapsed.Seconds())
	fmt.Printf("\x1b[35m%d\x1b[0m seconds have elapsed per %d messages\n", secs, pkts)
}

func doServer(args []string) {
	addr := fmt.Sprintf("http://127.0.0.1:%s", args[0])
	http.HandleFunc("/", srv)
	http.ListenAndServe(addr, nil)
}

func main() {
	args := os.Args[1:]
	doServer := false

	switch args[0] {
	case "-s":
		args = args[1:]
		doServer = true
	}

	if doServer {
		port := fmt.Sprintf(":%s", args[0])
		fmt.Println("listening on \x1b[32m", port, "\x1b[0m")
		http.HandleFunc("/", srv)
		http.ListenAndServe(port, nil)

	} else {
		port := fmt.Sprintf("ws://localhost:%s", args[0])
		fmt.Println("running on port:", port)

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		c, _, err := websocket.Dial(ctx, port, nil)
		if err != nil {
			panic(err)
		}
		defer c.Close(websocket.StatusInternalError, "the sky is falling")

		for i := 0; i < 10000; i++ {
			err = wsjson.Write(ctx, c, "whatever")
			if err != nil {
				panic(err)
			}
		}

		c.Close(websocket.StatusNormalClosure, "")
	}
}
