package main

import (
	"errors"
	"flag"
	"fmt"
	"lift/gsmap/monitor"
	"math/rand"
	"net/http"
	"time"

	libuuid "github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type DummyParams struct {
	Uuid    string
	Address string
	Port    string
}

func (p *DummyParams) RawUuid() libuuid.UUID {
	return libuuid.MustParse(p.Uuid)
}

type DummyConnectionHandle struct {
	conn *websocket.Conn
}

func (h *DummyConnectionHandle) SendMonitoringMessage(param *DummyParams) {
	ticker := time.Tick(time.Second)
	rawUuid := param.RawUuid()

	for range ticker {
		if rand.Intn(100)%7 == 0 {
			panic("error of 7")
		}

		msg := monitor.MonitoringMessage{
			GuidRaw:            rawUuid[:],
			ConnectionCount:    rand.Intn(100),
			SessionCount:       rand.Intn(100),
			ActiveSessionCount: rand.Intn(100),
		}
		if err := h.conn.WriteJSON(&msg); err != nil {
			panic(err)
		}
		fmt.Println("sent a monitoring message")
	}
}

func serverURL(uuid string) string {
	return fmt.Sprintf("ws://127.0.0.1:9990/connect/%s", uuid)
}

func parseFlags() *DummyParams {
	address := flag.String("a", "127.0.0.1", "listening address")
	port := flag.String("p", "7777", "listening port")
	uuid := flag.String("u", "00000000-0000-0000-0000-000000000000", "client uuid")

	flag.Parse()
	return &DummyParams{
		Uuid:    *uuid,
		Address: *address,
		Port:    *port,
	}
}

func connect(param *DummyParams) (*DummyConnectionHandle, error) {
	conn, res, err := websocket.DefaultDialer.Dial(serverURL(param.Uuid), nil)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusSwitchingProtocols {
		return nil, errors.New("ws switching does not work")
	}
	defer res.Body.Close()
	return &DummyConnectionHandle{
		conn: conn,
	}, nil
}

func main() {
	params := parseFlags()
	fmt.Printf("dummy is starting at %s:%s as [%s]",
		params.Address,
		params.Port,
		params.Uuid,
	)

	handle, err := connect(params)
	if err != nil {
		panic(err)
	}
	defer handle.conn.Close()

	handle.SendMonitoringMessage(params)
}
