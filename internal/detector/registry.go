package detector

import (
	"sort"
	"sync"
)

var (
	registry []Detector
	mu       sync.RWMutex
)

func Register(d Detector) {
	mu.Lock()
	defer mu.Unlock()
	registry = append(registry, d)
}

func GetAll() []Detector {
	mu.RLock()
	defer mu.RUnlock()
	cp := make([]Detector, len(registry))
	copy(cp, registry)
	return cp
}

func DetectAll(execPath string) []*Result {
	detectors := GetAll()
	var results []*Result
	for _, d := range detectors {
		if !d.IsAvailable() {
			continue
		}
		r := d.Detect(execPath)
		if r != nil && r.Found {
			results = append(results, r)
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return priorityOrder(results[i].Manager) < priorityOrder(results[j].Manager)
	})
	return results
}

func DetectFirst(execPath string) *Result {
	detectors := GetAll()
	for _, d := range detectors {
		if !d.IsAvailable() {
			continue
		}
		r := d.Detect(execPath)
		if r != nil && r.Found {
			return r
		}
	}
	return nil
}

func priorityOrder(m Manager) int {
	priorities := map[Manager]int{
		ManagerPacman:   1,
		ManagerParu:     2,
		ManagerYay:      3,
		ManagerFlatpak:  4,
		ManagerSnap:     5,
		ManagerBrew:     6,
		ManagerNix:      7,
		ManagerBun:      10,
		ManagerNpm:      11,
		ManagerPnpm:     12,
		ManagerYarn:     13,
		ManagerPip:      20,
		ManagerPipx:     21,
		ManagerUv:       22,
		ManagerCargo:    30,
		ManagerGo:       31,
		ManagerGem:      32,
		ManagerComposer: 33,
		ManagerDocker:   40,
		ManagerPodman:   41,
		ManagerDistrobox: 42,
		ManagerToolbox:  43,
		ManagerAppImage: 50,
		ManagerManual:   99,
	}
	if p, ok := priorities[m]; ok {
		return p
	}
	return 50
}
