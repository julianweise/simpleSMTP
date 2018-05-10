package core

import (
	"time"
	"fmt"
)

type SessionMeasuringService struct {
	data 		map[string][]time.Duration
	startTime	time.Time
	command 	string
}

func NewSessionMeasuringService() SessionMeasuringService {
	s := SessionMeasuringService{}
	s.data = make(map[string][]time.Duration)
	return s
}

func (s *SessionMeasuringService) StartMeasuring(command string) {
	s.startTime = time.Now()
	s.command = command
}

func (s *SessionMeasuringService) CancelMeasuring() {
	s.startTime = time.Time{}
	s.command = ""
}

func (s *SessionMeasuringService) FinalizeMeasuring() {
	if s.startTime.IsZero() || s.command == "" {
		return
	}
	if s.data[s.command] != nil {
		s.data[s.command] = append(s.data[s.command], time.Since(s.startTime))
	} else {
		s.data[s.command] = make([]time.Duration, 1)
		s.data[s.command][0] = time.Since(s.startTime)
	}
	// set zero
	s.startTime = time.Time{}
}

func (s *SessionMeasuringService) PrintResults() {
	fmt.Println("### Average response times ###")
	var overallSum int64
	for k, v := range s.data {
		var sum int64
		for _, num := range v {
			sum += num.Nanoseconds()
		}
		overallSum += sum
		fmt.Printf("%s: \t\t %d ns \n", k, sum)
	}
	fmt.Println("------------------------------")
	fmt.Printf("Average: \t %f\n", float64(overallSum/int64(len(s.data))))
	fmt.Println("##############################")
}