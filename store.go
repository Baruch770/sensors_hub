package main

import (
	"fmt"
	"sync"
	"time"
)

const layout = "2006-01-02"

type Stats struct {
	Min   int
	Max   int
	Avg   int
	Count int
	Total int
}

type SensorInfo struct {
	Time        string `json:"timestamp"`
	Temperature int    `json:"temperature"`
}

type InMemoryStore struct {
	mu sync.RWMutex
	// map[sensor]map[date][]temperature
	Sensors map[string]map[string][]int
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		Sensors: make(map[string]map[string][]int),
	}
}

func (s *InMemoryStore) AddInfo(sensorName string, info SensorInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.Sensors[sensorName]
	if !ok {
		sensor := make(map[string][]int)
		s.Sensors[sensorName] = sensor
	}

	s.Sensors[sensorName][info.Time] = append(s.Sensors[sensorName][info.Time], info.Temperature)
}

func (s *InMemoryStore) GetWeeklyStatsForSensors(date string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info := []string{}

	if len(s.Sensors) == 0 {
		return nil, fmt.Errorf("no data for any sensor")
	}

	var minAll, maxAll, totalAll, countAll int

	t, err := time.Parse(layout, date)
	if err != nil {
		return nil, fmt.Errorf("invalid date: %s", date)
	}

	initIterationG := true
	for sensorName, _ := range s.Sensors {
		var min, max, total, count int

		initIteration := true
		for i := 7; i >= 0; i-- {
			date = t.AddDate(0, 0, -i).Format(layout)

			st, err := s.GetDateStatsForSensor(date, sensorName)
			if err != nil {
				continue
			}

			if initIterationG {
				minAll = st.Min
				maxAll = st.Max
				initIterationG = false
			}

			if initIteration {
				min = st.Min
				max = st.Max
				initIteration = false
			}

			total += st.Total
			count += st.Count

			if st.Min < min {
				min = st.Min
			}

			if st.Max > max {
				max = st.Max
			}

			totalAll += st.Total
			countAll += st.Count

			if st.Min < minAll {
				minAll = st.Min
			}

			if st.Max > maxAll {
				maxAll = st.Max
			}
		}

		if initIteration {
			info = append(info, fmt.Sprintf("no data for sensor %s for %s\n", sensorName, date))
			continue
		}

		info = append(info, fmt.Sprintf("sensor %s weekly stats is: min: %d. max: %d avg: %d\n", sensorName, min, max, total/count))
	}

	if countAll != 0 {
		info = append(info, fmt.Sprintf("all sensors weekly stats are: min: %d. max: %d avg: %d\n", minAll, maxAll, totalAll/countAll))
	}

	return info, nil
}

func (s *InMemoryStore) GetDailyStatsForSensor(date string) ([]string, error) {
	info := []string{}

	t, err := time.Parse(layout, date)
	if err != nil {
		return nil, fmt.Errorf("invalid date: %s", date)
	}

	for i := 7; i >= 0; i-- {
		var minAll, maxAll, totalAll, countAll int

		initIteration := true
		for sensorName, _ := range s.Sensors {

			date = t.AddDate(0, 0, -i).Format(layout)

			st, err := s.GetDateStatsForSensor(date, sensorName)
			if err != nil {
				res := fmt.Sprintf(err.Error() + "\n")
				info = append(info, res)
				continue
			}

			if initIteration {
				minAll = st.Min
				maxAll = st.Max
				initIteration = false
			}

			totalAll += st.Total
			countAll += st.Count

			if st.Min < minAll {
				minAll = st.Min
			}

			if st.Max > maxAll {
				maxAll = st.Max
			}

			res := fmt.Sprintf("date %s stats for sensor %s are: min: %d. max: %d avg: %d\n", date, sensorName, st.Min, st.Max, st.Avg)

			info = append(info, res)
		}

		if countAll != 0 {
			info = append(info, fmt.Sprintf("all sensors stats for %s are: min: %d. max: %d avg: %d\n", date, minAll, maxAll, totalAll/countAll))
		}
	}

	return info, nil
}

func (s *InMemoryStore) GetDateStatsForSensor(date, sensorName string) (Stats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.Sensors[sensorName][date]) == 0 {
		return Stats{}, fmt.Errorf("date %s has no data for sensor: %s", date, sensorName)
	}

	temperatures := s.Sensors[sensorName][date]

	var min = temperatures[0]
	var max = temperatures[0]

	count := len(temperatures)

	var total int

	for _, temperature := range temperatures {

		total += temperature

		if max < temperature {
			max = temperature
		}
		if min > temperature {
			min = temperature
		}
	}

	avg := total / count

	st := Stats{min, max, avg, count, total}

	return st, nil
}
