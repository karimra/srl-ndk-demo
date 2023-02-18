package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/prototext"

	agent "github.com/karimra/srl-ndk-demo"
)

const retryInterval = 5 * time.Second

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "agent_name", "ndk-demo")
CRAGENT:
	app, err := agent.New(ctx, "ndk-demo", agent.WithRetryTimer(10*time.Second), agent.WithLogger(nil))
	if err != nil {
		log.Printf("failed to create agent: %v", err)
		log.Printf("retrying in %s", retryInterval)
		time.Sleep(retryInterval)
		goto CRAGENT
	}

	wg := new(sync.WaitGroup)

	// keepAlive
	wg.Add(1)
	go func() {
		defer wg.Done()
		app.KeepAlive(ctx, time.Minute)
	}()

	// Config notifications
	wg.Add(1)
	go func() {
		defer wg.Done()
		appIdChan := app.StartConfigNotificationStream(ctx)
		for {
			select {
			case event := <-appIdChan:
				log.Printf("Config notification: %+v", event)
				b, err := prototext.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(event)
				if err != nil {
					log.Printf("Config notification Marshal failed: %+v", err)
					continue
				}
				fmt.Printf("%s\n", string(b))
			case <-ctx.Done():
				return
			}
		}
	}()

	// AppId notifications
	wg.Add(1)
	go func() {
		defer wg.Done()
		appIdChan := app.StartAppIdNotificationStream(ctx, nil)
		for {
			select {
			case event := <-appIdChan:
				log.Printf("appID notification: %+v", event)
				b, err := prototext.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(event)
				if err != nil {
					log.Printf("appID notification Marshal failed: %+v", err)
					continue
				}
				fmt.Printf("%s\n", string(b))
			case <-ctx.Done():
				return
			}
		}
	}()

	// BFDSession notifications
	wg.Add(1)
	go func() {
		defer wg.Done()
		appIdChan := app.StartBFDSessionNotificationStream(ctx, nil)
		for {
			select {
			case event := <-appIdChan:
				log.Printf("BFDSession notification: %+v", event)
				b, err := prototext.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(event)
				if err != nil {
					log.Printf("BFDSession notification Marshal failed: %+v", err)
					continue
				}
				fmt.Printf("%s\n", string(b))
			case <-ctx.Done():
				return
			}
		}
	}()

	// NwInst notifications
	wg.Add(1)
	go func() {
		defer wg.Done()
		appIdChan := app.StartNwInstNotificationStream(ctx)
		for {
			select {
			case event := <-appIdChan:
				log.Printf("NwInst notification: %+v", event)
				b, err := prototext.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(event)
				if err != nil {
					log.Printf("NwInst notification Marshal failed: %+v", err)
					continue
				}
				fmt.Printf("%s\n", string(b))
			case <-ctx.Done():
				return
			}
		}
	}()

	// Interface notifications
	wg.Add(1)
	go func() {
		defer wg.Done()
		appIdChan := app.StartInterfaceNotificationStream(ctx, "")
		for {
			select {
			case event := <-appIdChan:
				log.Printf("Interface notification: %+v", event)
				b, err := prototext.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(event)
				if err != nil {
					log.Printf("Interface notification Marshal failed: %+v", err)
					continue
				}
				fmt.Printf("%s\n", string(b))
			case <-ctx.Done():
				return
			}
		}
	}()

	// LLDPNeighbor notifications
	wg.Add(1)
	go func() {
		defer wg.Done()
		appIdChan := app.StartLLDPNeighNotificationStream(ctx, "", "", "")
		for {
			select {
			case event := <-appIdChan:
				log.Printf("LLDPNeighbor notification: %+v", event)
				b, err := prototext.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(event)
				if err != nil {
					log.Printf("LLDPNeighbor notification Marshal failed: %+v", err)
					continue
				}
				fmt.Printf("%s\n", string(b))
			case <-ctx.Done():
				return
			}
		}
	}()

	//Route notifications
	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	appIdChan := app.StartRouteNotificationStream(ctx, "default", net.IP{0, 0, 0, 0}, 0)
	// 	for {
	// 		select {
	// 		case event := <-appIdChan:
	// 			log.Printf("Route notification: %+v", event)
	// 			b, err := prototext.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(event)
	// 			if err != nil {
	// 				log.Printf("Route notification Marshal failed: %+v", err)
	// 				continue
	// 			}
	// 			fmt.Printf("%s\n", string(b))
	// 		case <-ctx.Done():
	// 			return
	// 		}
	// 	}
	// }()

	wg.Wait()
}
