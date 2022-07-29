package music

import (
	"context"
	"os"
	"time"
)

const (
	add       = "add"    // modify song (toPlay +-)
	clear     = "clear"  // delete marked songs
	wait      = "wait"   // do not delete or modify any (loop case)
	deleteAll = "delete" // mark all songs as deleted (bot disconnect)
)

type command struct {
	name      string
	filepath  string
	increment bool
}

type LoadMaster struct {
	files    map[string]int
	commands chan *command
	waitFlag bool
}

func NewLoadMaster(ctx context.Context, clearTime time.Duration) *LoadMaster {
	lm := &LoadMaster{
		files: make(map[string]int),
	}
	lm.commands = lm.process(ctx)

	ticker := time.NewTicker(clearTime)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lm.softClear()
			}
		}
	}()
	return lm
}

func (m *LoadMaster) Add(path string) {
	m.commands <- &command{
		name:      add,
		filepath:  path,
		increment: true,
	}
}

func (m *LoadMaster) Remove(path string) {
	m.commands <- &command{
		name:      add,
		filepath:  path,
		increment: false,
	}
}

func (m *LoadMaster) DeleteAll() {
	m.commands <- &command{
		name: deleteAll,
	}
}

func (m *LoadMaster) softClear() {
	m.commands <- &command{
		name: clear,
	}
}

func (m *LoadMaster) process(ctx context.Context) chan *command {
	commands := make(chan *command)
	go func() {
		defer close(commands)
		for {
			select {
			case c := <-commands:
				switch c.name {
				case wait:
					m.waitFlag = !m.waitFlag
				case add:
					m.modify(c)
				case clear:
					m.clear()
				case deleteAll:
					m.deleteAll()
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return commands
}

func (m *LoadMaster) modify(c *command) {
	if m.waitFlag {
		return
	}
	if c.increment {
		m.files[c.filepath]++
	} else {
		if _, ok := m.files[c.filepath]; ok {
			m.files[c.filepath]--
		}
	}
}

func (m *LoadMaster) clear() {
	for path, toPlay := range m.files {
		if toPlay <= 0 {
			_ = os.Remove(path)
			delete(m.files, path)
		}
	}
}

func (m *LoadMaster) deleteAll() {
	for k := range m.files {
		m.files[k] = 0
	}
}
