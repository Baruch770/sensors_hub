package main

type Message struct {
	SensorName string `json:"sensorname"`
	Type        string `json:"type"`
	Temperature int    `json:"temperature"`
	Date        string `json:"date"`
}
