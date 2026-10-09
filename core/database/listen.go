package database

import "github.com/mendahu/provenencia/core/database/effects"

// CommitListener hears a committed write. OnCommit applies the notice.
// Drop discards the listener's state. writes.Run calls Drop on every
// listener when any OnCommit fails, so a partial notice does not stick.
// Nothing registers a listener until the graph cache does.
type CommitListener interface {
	OnCommit(rev int64, set effects.Set) error
	Drop()
}

// Listen registers a listener. The order of OnCommit is registration order.
func (c *Catalog) Listen(l CommitListener) {
	if c == nil || l == nil {
		return
	}
	c.mu.Lock()
	c.listeners = append(c.listeners, l)
	c.mu.Unlock()
}

// Notify delivers a committed write. The first OnCommit error stops delivery
// and Drop runs for every listener, including ones already notified.
func (c *Catalog) Notify(rev int64, set effects.Set) error {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	ls := append([]CommitListener(nil), c.listeners...)
	c.mu.Unlock()
	for _, l := range ls {
		if err := l.OnCommit(rev, set); err != nil {
			for _, d := range ls {
				d.Drop()
			}
			return err
		}
	}
	return nil
}
