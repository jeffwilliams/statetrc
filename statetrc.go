// Package statetrc implements basic state tracing. When some state in the program is entered, use Enter to
// record it. When the state is left, use Leave to remove the state. At any time List can be used to obtain a
// list of all the existing entries, which is a snapshot of the current state of the program. This can be useful
// in debugging to tell what functions or higher-level states are stuck or taking a long time to complete.
package statetrc

import (
	"bytes"
	"fmt"
	"sort"
	"sync"
	"time"
)

var (
	entries = map[string]Entry{}
	mtx     sync.Mutex
)

// Entry represents a single event in the trace. Usually used to
// represent entering some state.
type Entry struct {
	// Identifier for the state entered
	Id string
	// User-added properties
	Props interface{}
	// Time when the Entry was added
	Time time.Time
}

type EntrySlice []Entry

func (e EntrySlice) String() string {
	var buf bytes.Buffer
	now := time.Now()

	for _, e := range e {
		d := now.Sub(e.Time)
		fmt.Fprintf(&buf, "%s: %v\n", e.Id, d)
		props := fmt.Sprintf("%v", e.Props)

		// Indent each line in props by two spaces when printing
		buf.WriteString("  ")
		for _, r := range props {
			buf.WriteRune(r)
			if r == '\n' {
				buf.WriteString("  ")
			}
		}
		buf.WriteRune('\n')
	}

	return buf.String()
}

// Enter creates a new Entry with the passed id and properties,
// with the Time set to now.
// id should be of the form item/item/prop
// This allows using the package for function entry/exit (use /funcname)
// but also for items in a set (/itemtype/id1, /itemtype/id2) which is useful
// for counting how many things are there in a set, etc.
func Enter(id string, props interface{}) {
	mtx.Lock()
	defer mtx.Unlock()
	entries[id] = Entry{Id: id, Props: props, Time: time.Now()}
}

// Leave removes the entry with the specified id.
func Leave(id string) {
	mtx.Lock()
	defer mtx.Unlock()
	delete(entries, id)
}

var (
	// ById is an ordering that may be passed to List to return Entries ordered by id ascending.
	ById Order = func(l []Entry) func(i, j int) bool {
		return func(i, j int) bool {
			return l[i].Id < l[j].Id
		}
	}

	// ById is an ordering that may be passed to List to return Entries ordered by duration descending.
	ByDuration Order = func(l []Entry) func(i, j int) bool {
		return func(i, j int) bool {
			return l[i].Time.After(l[j].Time)
		}
	}
)

type Order func(l []Entry) func(i, j int) bool

// List returns a slice of all currently existing entries, ordered in the specified Order.
func List(order Order) EntrySlice {
	mtx.Lock()

	res := make([]Entry, len(entries))

	i := 0
	for _, v := range entries {
		res[i] = v
		i++
	}

	mtx.Unlock()

	if order == nil {
		order = ById
	}

	sort.Slice(res, order(res))

	return res
}

// Clear removes all entries. It clears all state.
func Clear() {
	mtx.Lock()
	defer mtx.Unlock()

	entries = map[string]Entry{}
}
