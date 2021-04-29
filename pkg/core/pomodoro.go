package core

import (
	"sync"
	"time"

	"github.com/TensRoses/iris/internal/datastore"
)

// Pomodoro defines a single state of a pomodoro sessions.
// Usage of channel to handle cancel signal.
type Pomodoro struct {
	workDuration time.Duration
	onWorkEnd    TaskCallback
	notifyInfo   NotifyInfo
	cancelChan   chan struct{}
	cancel       sync.Once
}

// NotifyInfo defines notification message for users.
type NotifyInfo struct {
	TitleID string
	User    *datastore.User
}

// TaskCallback receives NotifyInfo and a boolean to define whether the task is completed or not.
type TaskCallback func(info NotifyInfo, finished bool)

// UserPomodoroMap is a map-like structure to init a single pomodoro in a channel. It has goroutine safe ops.
type UserPomodoroMap struct {
	mutex     sync.Mutex
	userToPom map[string]*Pomodoro
}

// NewUserPomodoroMap creates a ChannelPomMap and prepares it to be used.
func NewUserPomodoroMap() UserPomodoroMap {
	return UserPomodoroMap{userToPom: make(map[string]*Pomodoro)}
}

// NewPom create a new pomodoro and start it using time.NewTimer. onWorkEnd will be called after the goroutine.
func NewPom(workDuration time.Duration, onWorkEnd TaskCallback, notify NotifyInfo) *Pomodoro {
	pom := &Pomodoro{
		workDuration: workDuration,
		onWorkEnd:    onWorkEnd,
		notifyInfo:   notify,
		cancelChan:   make(chan struct{}),
		cancel:       sync.Once{},
	}

	go pom.startPom()
	return pom
}

// Cancel is used to cancel the current state of the goroutine. sync.Once to prevent panic.
func (pom *Pomodoro) Cancel() {
	pom.cancel.Do(func() {
		close(pom.cancelChan)
	})
}

func (pom *Pomodoro) startPom() {
	workTimer := time.NewTimer(pom.workDuration)

	select {
	case <-workTimer.C:
		go pom.onWorkEnd(pom.notifyInfo, true)
	case <-pom.cancelChan:
		go pom.onWorkEnd(pom.notifyInfo, false)
	}
}

// CreateIfEmpty will create a new Pomodoro for given user according to their discordID if user has none.
// The pomodoro will then be removed from the mapping once completed or canceled.
func (u *UserPomodoroMap) CreateIfEmpty(duration time.Duration, onWorkEnd TaskCallback, notify NotifyInfo) bool {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	wasCreated := false
	if _, exists := u.userToPom[notify.User.DiscordID]; !exists {
		doneInMap := func(notifs NotifyInfo, completed bool) {
			// only called when it is done then we can use the mutex
			// cancellation won't trigger onWorkEnd since startPom is already done at this point
			u.RemoveIfExists(notifs.User.DiscordID)
			onWorkEnd(notifs, completed)
		}

		u.userToPom[notify.User.DiscordID] = NewPom(duration, doneInMap, notify)
		wasCreated = true
	}
	return wasCreated
}

// RemoveIfExists will remove a Pomodoro from given channel i one already exists.
func (u *UserPomodoroMap) RemoveIfExists(discordID string) bool {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	wasRemoved := false
	if p, exists := u.userToPom[discordID]; exists {
		delete(u.userToPom, discordID)
		p.Cancel()
		wasRemoved = true
	}

	return wasRemoved
}

// Count counts the number of current Pomodoro being tracked.
func (u *UserPomodoroMap) Count() int {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	return len(u.userToPom)
}
