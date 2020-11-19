package storage

import (
	"github.com/orcaman/concurrent-map"
	"time"
)

// ////////////////////////////////////////////////////////////////////////////////// //

type Store struct {
	info         cmap.ConcurrentMap
	maxStoreTime int64
}

type Info struct {
	Topic     string `json:"-"`
	Metrics   string `json:"metrics"`
	Value     string `json:"value"`
	Timestamp int64  `json:"updated_at"`
}

// ////////////////////////////////////////////////////////////////////////////////// //

// NewStore returns Store struct
func NewStore(maxStoreTime time.Duration) *Store {
	return &Store{
		info:         cmap.New(),
		maxStoreTime: int64(maxStoreTime.Seconds()),
	}
}

// Add adds new data
func (s *Store) Add(info *Info) {
	info.Timestamp = time.Now().Unix()
	s.info.Set(info.Topic, info)
}

// Get return latest info data for metric
func (s *Store) Get(metric string) *Info {
	if s.info.IsEmpty() {
		return nil
	}

	info, _ := s.info.Get(metric)

	return info.(*Info)
}

// MarshalJSON returns serialized concurrent map struct to JSON byte slice
func (s *Store) MarshalJSON() ([]byte, error) {
	return s.info.MarshalJSON()
}

// Clean remove old data from store
func (s *Store) Clean() {
	now := time.Now().Unix()

	metrics := s.info.Keys()

	for _, m := range metrics {
		lastInfo, ok := s.info.Get(m)

		if !ok {
			continue
		}

		if lastInfo.(*Info).Timestamp < now-s.maxStoreTime {
			s.info.Remove(m)
		}
	}
}
