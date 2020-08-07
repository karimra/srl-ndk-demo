package main

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/karimra/srl-ndk-demo/agent"
	"google.golang.org/grpc/metadata"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "agent_name", "ndk-demo")
	app, err := agent.NewAgent(ctx, "ndk-demo")
	if err != nil {
		log.Printf("failed to create agent: %v", err)
		os.Exit(1)
	}
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go app.KeepAlive(ctx, time.Minute)

	// Config notifications
	wg.Add(1)
	go func() {
		defer wg.Done()
		appIdChan := app.StartConfigNotificationStream(ctx)
		for {
			select {
			case event := <-appIdChan:
				log.Printf("Config notification: %+v", event)
			case <-ctx.Done():
				return
			}
		}
	}()
	//time.Sleep(time.Second)
	// AppId notifications
	wg.Add(1)
	go func() {
		defer wg.Done()
		appIdChan := app.StartAppIdNotificationStream(ctx, 0)
		for {
			select {
			case event := <-appIdChan:
				log.Printf("appID notification: %+v", event)
			case <-ctx.Done():
				return
			}
		}
	}()
	//time.Sleep(time.Second)
	// BFDSession notifications
	wg.Add(1)
	go func() {
		defer wg.Done()
		appIdChan := app.StartBFDSessionNotificationStream(ctx, nil, nil, 0)
		for {
			select {
			case event := <-appIdChan:
				log.Printf("BFDSession notification: %+v", event)
			case <-ctx.Done():
				return
			}
		}
	}()
	//time.Sleep(time.Second)
	// NwInst notifications
	wg.Add(1)
	go func() {
		defer wg.Done()
		appIdChan := app.StartNwInstNotificationStream(ctx)
		for {
			select {
			case event := <-appIdChan:
				log.Printf("NwInst notification: %+v", event)
			case <-ctx.Done():
				return
			}
		}
	}()
	//time.Sleep(time.Second)
	// Interface notifications
	wg.Add(1)
	go func() {
		defer wg.Done()
		appIdChan := app.StartInterfaceNotificationStream(ctx, "")
		for {
			select {
			case event := <-appIdChan:
				log.Printf("Interface notification: %+v", event)
			case <-ctx.Done():
				return
			}
		}
	}()
	//time.Sleep(time.Second)
	// LLDPNeighbor notifications
	wg.Add(1)
	go func() {
		defer wg.Done()
		appIdChan := app.StartLLDPNeighNotificationStream(ctx, "", "", "")
		for {
			select {
			case event := <-appIdChan:
				log.Printf("LLDPNeighbor notification: %+v", event)
			case <-ctx.Done():
				return
			}
		}
	}()
	//time.Sleep(time.Second)
	// Route notifications
	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	appIdChan := app.StartRouteNotificationStream(ctx, "", nil, 0)
	// 	for {
	// 		select {
	// 		case event := <-appIdChan:
	// 			log.Printf("Route notification: %+v", event)
	// 		case <-ctx.Done():
	// 			return
	// 		}
	// 	}
	// }()

	wg.Wait()
}
