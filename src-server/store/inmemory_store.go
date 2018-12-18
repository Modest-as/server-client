package store

import (
	"encoding/binary"
	"errors"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
)

const checksumMod = 123456
const pollingTime = 30

// InMemoryStore session store
// we might want to use concurrent maps
// because we are currently locking the
// map every time we want to update state
// for specific uuid
type InMemoryStore struct {
	lock              *sync.RWMutex
	idStore           map[uuid.UUID]int
	dataSentCountByID map[uuid.UUID]int
	idSeedStore       map[uuid.UUID]int64
	checksumByID      map[uuid.UUID]uint32
	lastAccessByID    map[uuid.UUID]time.Time
}

// MakeInMemoryStore creates a valid store object
func MakeInMemoryStore() InMemoryStore {
	lock := sync.RWMutex{}

	store := InMemoryStore{
		lock:              &lock,
		idStore:           map[uuid.UUID]int{},
		dataSentCountByID: map[uuid.UUID]int{},
		idSeedStore:       map[uuid.UUID]int64{},
		checksumByID:      map[uuid.UUID]uint32{},
		lastAccessByID:    map[uuid.UUID]time.Time{},
	}

	return store
}

// Remove session given client id
func (s InMemoryStore) Remove(id uuid.UUID) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.idStore, id)
	delete(s.dataSentCountByID, id)
	delete(s.idSeedStore, id)
	delete(s.checksumByID, id)
	delete(s.lastAccessByID, id)

	return nil
}

// Exists check if client id exists
func (s InMemoryStore) Exists(id uuid.UUID) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if _, ok := s.idStore[id]; ok {
		return true
	}

	return false
}

// Add ads client session data
func (s InMemoryStore) Add(id uuid.UUID, n int) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	seed := int64(binary.BigEndian.Uint64(id[:8]))

	s.idStore[id] = n
	s.dataSentCountByID[id] = 0
	s.idSeedStore[id] = seed
	s.checksumByID[id] = 0
	s.lastAccessByID[id] = time.Now()

	go s.checkAbandon(id)

	return nil
}

// GetSentCount returns the number of values already sent
func (s InMemoryStore) GetSentCount(id uuid.UUID) int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.dataSentCountByID[id]
}

// GetCurrentChecksum returns current checksum so the client can continue with validation
func (s InMemoryStore) GetCurrentChecksum(id uuid.UUID) uint32 {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.checksumByID[id]
}

// HasNext checks if session needs to send more data
func (s InMemoryStore) HasNext(id uuid.UUID) (bool, error) {
	s.lock.RLock()

	if s.idStore[id]+1 > s.dataSentCountByID[id] {
		s.lock.RUnlock()
		return true, nil
	}

	s.lock.RUnlock()

	s.Remove(id)

	return false, nil
}

// GetNext will return next int we need to send in case HasNext is true
func (s InMemoryStore) GetNext(id uuid.UUID) uint32 {
	s.lock.RLock()
	n := s.idStore[id]
	count := s.dataSentCountByID[id]
	seed := s.idSeedStore[id]
	checksum := s.checksumByID[id]
	s.lock.RUnlock()

	var next uint32

	if count == n {
		next = checksum
	} else {
		next = getNthRandom(seed, count+1)
	}

	return next
}

// Update values once we confirm that the data was sent
func (s InMemoryStore) Update(id uuid.UUID, next uint32) {
	s.lock.RLock()
	count := s.dataSentCountByID[id]
	checksum := s.checksumByID[id]
	s.lock.RUnlock()

	s.lock.Lock()
	defer s.lock.Unlock()

	s.dataSentCountByID[id] = count + 1
	s.checksumByID[id] = (checksum + (next % checksumMod)) % checksumMod
	s.lastAccessByID[id] = time.Now()
}

// SyncState in case client reconnects and we want to know what he received
func (s InMemoryStore) SyncState(id uuid.UUID, lastNumber uint32) error {
	s.lock.RLock()
	seed := s.idSeedStore[id]
	total := s.idStore[id]
	s.lock.RUnlock()

	count := 0

	var checksum uint32

	for {
		count++
		n := getNthRandom(seed, count)

		// check if the number is the checksum that was sent
		if checksum == lastNumber && count == total+1 {
			break
		}

		checksum = (checksum + (n % checksumMod)) % checksumMod

		// check if the number matches nth random
		if n == lastNumber && count <= total {
			break
		}

		if count > total+1 {
			return errors.New("invalid last number")
		}
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.dataSentCountByID[id] = count
	s.checksumByID[id] = checksum
	s.lastAccessByID[id] = time.Now()

	return nil
}

func (s InMemoryStore) checkAbandon(id uuid.UUID) {
	for {
		time.Sleep(pollingTime * time.Second)

		if !s.Exists(id) {
			break
		}

		s.lock.RLock()
		t := s.lastAccessByID[id]
		s.lock.RUnlock()

		if t.Add(pollingTime * time.Second).Before(time.Now()) {
			s.Remove(id)
			log.Printf("Removed abandoned session %s", id)
			break
		}
	}
}

// this is not performant, we could store
// this a random generator and maintain it's
// state in memory or maybe find a PRNG that
// lets us compute n-th number more easily
func getNthRandom(seed int64, n int) uint32 {
	rand.Seed(seed)

	for i := 0; i < n-1; i++ {
		rand.Uint32()
	}

	return rand.Uint32()
}
