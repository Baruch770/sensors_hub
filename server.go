package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

type server struct {
	Port     string
	InMemory *InMemoryStore
}

func NewServer(prot string) *server {
	return &server{
		Port:     prot,
		InMemory: NewInMemoryStore(),
	}
}

func (s *server) start() {
	listener, err := net.Listen("tcp", ":"+s.Port)
	if err != nil {
		log.Fatal("error listening: ", err)
	}
	defer listener.Close()

	log.Println("Listening ..")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("error accepting conn: ", err)
		}

		go s.handleRequest(conn)
	}
}

func (s *server) handleRequest(conn net.Conn) {
	defer conn.Close()

	msg := &Message{}

	err := s.decodeMessage(msg, conn)
	if err != nil {
		conn.Write([]byte("wrong format sent\n"))
		return
	}

	switch msg.Type {
	case "info":
		if msg.SensorName == "" || msg.Temperature == 0 {
			conn.Write([]byte("please provide a valid sensorname and temperature\n"))
			return
		}
		s.handleInfoMsg(msg)
		conn.Write([]byte("inforation recived\n"))
	case "daily_stats":
		s.handleDailyStatsReq(msg, conn)
	case "weekly_stats":
		s.handleWeeklyStatReq(msg, conn)
	default:
		conn.Write([]byte("wrong type\n"))
	}

}

func (s *server) handleWeeklyStatReq(msg *Message, w io.Writer) {
	statsSlice, err := s.InMemory.GetWeeklyStatsForSensors(time.Now().Format(layout))
	if err != nil {
		w.Write([]byte(fmt.Sprintf("error getting stats: %s\n", err)))
		return
	}

	for _, res := range statsSlice {
		w.Write([]byte(res))
	}
}

func (s *server) handleDailyStatsReq(msg *Message, w io.Writer) {
	info, err := s.InMemory.GetDailyStatsForSensor(time.Now().Format(layout))
	if err != nil {
		w.Write([]byte(fmt.Sprintf("error getting stats: %s\n", err)))
		return
	}
	for _, res := range info {
		w.Write([]byte(res))
	}
}

func (s *server) handleInfoMsg(msg *Message) {
	info := SensorInfo{Time: time.Now().Format(layout), Temperature: msg.Temperature}

	s.InMemory.AddInfo(msg.SensorName, info)
}

func (s *server) decodeMessage(msg *Message, r io.Reader) error {
	decoder := json.NewDecoder(r)
	err := decoder.Decode(msg)
	if err != nil {
		return err
	}

	return nil
}
