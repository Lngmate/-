package services

import (
	"sort"
	"sync"
)

var equipmentLocks sync.Map

// LockEquipments shares locks with single reservations. Stable order avoids deadlocks.
// This guards a single backend process; multiple instances need database-level locking.
func LockEquipments(ids []uint) func() {
	ordered := append([]uint(nil), ids...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })
	locks := []*sync.Mutex{}
	for i, id := range ordered {
		if i > 0 && ordered[i-1] == id {
			continue
		}
		value, _ := equipmentLocks.LoadOrStore(id, &sync.Mutex{})
		lock := value.(*sync.Mutex)
		lock.Lock()
		locks = append(locks, lock)
	}
	return func() {
		for i := len(locks) - 1; i >= 0; i-- {
			locks[i].Unlock()
		}
	}
}
