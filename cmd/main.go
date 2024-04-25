package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
)

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("[INFO] Connected to MQTT broker")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Println("[ERROR] Connection to MQTT broker lost:", err.Error())
}

type LockerWall struct {
	Tenant string
	Name   string
	Client mqtt.Client
}

func (lw *LockerWall) Init(d time.Duration, changeAt int) {
	go func() {
		tick := time.NewTicker(d)
		cycle := 1
		for {
			cycle++
			dt := time.Now().UTC()
			msg := fmt.Sprintf("%s,%s,CL,OE,CL,OE,CL,OE,CL,OE,CL,OE,CL,OE,CL,OE,CL,OE", dt.Format(time.ANSIC), lw.Name)
			if changeAt <= cycle && cycle <= changeAt*2 {
				msg = fmt.Sprintf("%s,%s,OE,CL,OE,CL,OE,CL,OE,CL,OE,CL,OE,CL,OE,CL,OE,CL", dt.Format(time.ANSIC), lw.Name)
			}
			if cycle > changeAt*2 {
				cycle = 1
			}
			lw.Client.Publish(lw.Tenant+"/status", 0, false, msg)
			<-tick.C
		}
	}()
}

func main() {
	godotenv.Load()
	log.SetFlags(0)
	opts := mqtt.NewClientOptions()

	opts.AddBroker(os.Getenv("MQTT_URL"))
	opts.SetClientID(os.Getenv("MQTT_CLIENT_ID"))
	opts.SetUsername(os.Getenv("MQTT_USER"))
	opts.SetPassword(os.Getenv("MQTT_PASS"))

	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	client := mqtt.NewClient(opts)
	if session := client.Connect(); session.Wait() && session.Error() != nil {
		log.Fatalln("[ERROR] Could not connect to MQTT broker:", session.Error())
	}

	spawnNodes(1000, client, 15, 1*time.Second)

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("[INFO] Locker Walls initialized, press Ctrl + C to terminate the program ...")
	<-done
}

func spawnNodes(n int, client mqtt.Client, timing int, interval time.Duration) {
	for i := range n {
		node := LockerWall{
			Tenant: "tester",
			Name:   fmt.Sprintf("spurdo%v", i),
			Client: client,
		}
		node.Init(interval, timing)
	}
}
