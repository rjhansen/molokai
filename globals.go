package main

type Reading struct {
	Time        string  `json:"timestamp"`
	Temperature float32 `json:"temperature"`
}

var (
	dbUrl       string
	sensorQueue map[string][]Reading
)
