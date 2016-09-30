package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/nobonobo/p2pfw/signaling"
	"github.com/nobonobo/p2pfw/signaling/client"
)

// PROMPT ...
const PROMPT = "> "

var (
	noneVal = struct{}{}
	// None ...
	None = &noneVal
)

// Text ...
type Text struct {
	Message string
}

// Kind ...
func (t *Text) Kind() string { return "text" }

func init() {
	signaling.Register(func() signaling.Kinder { return new(Text) })
}

func main() {
	create := ""
	join := ""
	user := "unknown"
	secret := ""
	flag.StringVar(&create, "c", create, "create room id")
	flag.StringVar(&join, "j", join, "join room id")
	flag.StringVar(&user, "u", user, "user id")
	flag.StringVar(&secret, "s", secret, "preshared secret")
	flag.Parse()
	if len(create) > 0 && len(join) > 0 {
		log.Fatalln("only create or join")
	}
	room := create + join
	config := new(client.Config)
	config.RoomID = room
	config.UserID = user
	config.Preshared = secret
	node, err := client.NewNode(config)
	if err != nil {
		log.Fatalln(err)
	}
	receiver := client.DispatcherFunc(func(events []*signaling.Event) {
		text := []string{}
		for _, ev := range events {
			switch v := ev.Get().(type) {
			case *Text:
				text = append(text, v.Message)
			}
		}
		if len(text) > 0 {
			fmt.Printf("\n%s\n", strings.Join(text, "\n"))
			fmt.Print(PROMPT)
		}
	})
	if err := node.Start(len(create) > 0, receiver); err != nil {
		log.Fatalln(err)
	}
	defer node.Stop()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		defer close(sig)
		for {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print(PROMPT)
			text, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					continue
				}
				log.Println(err)
			}
			text = strings.TrimSpace(text)
			if len(text) > 0 {
				if err := node.Send(signaling.New(user, "", &Text{Message: text})); err != nil {
					log.Println(err)
				}
			}
		}
	}()
	<-sig
}
