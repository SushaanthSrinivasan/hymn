package player

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"
)

type Client interface {
	Loadfile(url string) error
	LoadAppend(url string) error
	PlayIndex(i int) error
	TogglePause() error
	Pause(p bool) error
	Seek(deltaSec float64) error
	SetVolume(v int) error
	SetProperty(name string, value any) error
	Stop() error
	Events() <-chan Event
	Close() error
}

type client struct {
	cmd    *exec.Cmd
	conn   conn
	w      *bufio.Writer
	wmu    sync.Mutex
	nextID atomic.Int64
	pend   sync.Map // request_id -> chan response
	events chan Event
	done   chan struct{}
	closed atomic.Bool
}

type response struct {
	RequestID int64           `json:"request_id"`
	Error     string          `json:"error"`
	Data      json.RawMessage `json:"data"`
}

type rawEvent struct {
	Event string          `json:"event"`
	Name  string          `json:"name"`
	Data  json.RawMessage `json:"data"`
}

// Spawn starts mpv, dials its IPC, subscribes to a baseline set of properties,
// and returns a ready Client.
func Spawn(mpvPath string) (Client, error) {
	cmd, socket, err := spawnMPV(mpvPath)
	if err != nil {
		return nil, err
	}
	c, err := pollDial(socket, 5*time.Second)
	if err != nil {
		_ = cmd.Process.Kill()
		return nil, err
	}
	cl := &client{
		cmd:    cmd,
		conn:   c,
		w:      bufio.NewWriter(c),
		events: make(chan Event, 256),
		done:   make(chan struct{}),
	}
	go cl.reader()

	props := []string{"time-pos", "pause", "duration", "media-title", "eof-reached", "volume", "core-idle"}
	for i, p := range props {
		if err := cl.command([]any{"observe_property", i + 1, p}); err != nil {
			_ = cl.Close()
			return nil, fmt.Errorf("observe %s: %w", p, err)
		}
	}
	return cl, nil
}

func (c *client) reader() {
	defer close(c.events)
	defer close(c.done)
	scan := bufio.NewScanner(c.conn)
	scan.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scan.Scan() {
		line := scan.Bytes()
		if len(line) == 0 {
			continue
		}
		// Probe: response or event. Cheap path: check for "event" key.
		var probe struct {
			Event     string `json:"event"`
			RequestID *int64 `json:"request_id"`
		}
		if err := json.Unmarshal(line, &probe); err != nil {
			continue
		}
		if probe.Event != "" {
			var ev rawEvent
			if err := json.Unmarshal(line, &ev); err != nil {
				continue
			}
			c.dispatchEvent(ev)
			continue
		}
		if probe.RequestID != nil {
			var resp response
			if err := json.Unmarshal(line, &resp); err != nil {
				continue
			}
			if v, ok := c.pend.LoadAndDelete(resp.RequestID); ok {
				ch := v.(chan response)
				select {
				case ch <- resp:
				default:
				}
			}
		}
	}
}

func (c *client) dispatchEvent(ev rawEvent) {
	out := Event{}
	switch ev.Event {
	case "property-change":
		switch ev.Name {
		case "time-pos":
			out.Kind = EventTimePos
			_ = json.Unmarshal(ev.Data, &out.Float)
		case "duration":
			out.Kind = EventDuration
			_ = json.Unmarshal(ev.Data, &out.Float)
		case "pause":
			out.Kind = EventPause
			_ = json.Unmarshal(ev.Data, &out.Bool)
		case "volume":
			out.Kind = EventVolume
			_ = json.Unmarshal(ev.Data, &out.Float)
		case "media-title":
			out.Kind = EventMediaTitle
			_ = json.Unmarshal(ev.Data, &out.String)
		case "eof-reached":
			// EOF property fires true at end of file.
			var b bool
			_ = json.Unmarshal(ev.Data, &b)
			if !b {
				return
			}
			out.Kind = EventEOF
		default:
			return
		}
	case "start-file":
		out.Kind = EventStartFile
	case "file-loaded":
		out.Kind = EventFileLoaded
	case "end-file":
		out.Kind = EventEOF
	case "idle":
		out.Kind = EventIdle
	default:
		return
	}
	select {
	case c.events <- out:
	default:
		// drop on full to keep reader unblocked
	}
}

func (c *client) command(cmd []any) error {
	id := c.nextID.Add(1)
	ch := make(chan response, 1)
	c.pend.Store(id, ch)
	payload := map[string]any{"command": cmd, "request_id": id}
	b, err := json.Marshal(payload)
	if err != nil {
		c.pend.Delete(id)
		return err
	}
	c.wmu.Lock()
	_, werr := c.w.Write(append(b, '\n'))
	if werr == nil {
		werr = c.w.Flush()
	}
	c.wmu.Unlock()
	if werr != nil {
		c.pend.Delete(id)
		return werr
	}
	select {
	case resp := <-ch:
		if resp.Error != "" && resp.Error != "success" {
			return errors.New("mpv: " + resp.Error)
		}
		return nil
	case <-time.After(5 * time.Second):
		c.pend.Delete(id)
		return errors.New("mpv command timeout")
	case <-c.done:
		return errors.New("mpv connection closed")
	}
}

func (c *client) Loadfile(url string) error    { return c.command([]any{"loadfile", url, "replace"}) }
func (c *client) LoadAppend(url string) error  { return c.command([]any{"loadfile", url, "append-play"}) }
func (c *client) PlayIndex(i int) error        { return c.command([]any{"playlist-play-index", i}) }
func (c *client) TogglePause() error           { return c.command([]any{"cycle", "pause"}) }
func (c *client) Pause(p bool) error           { return c.command([]any{"set_property", "pause", p}) }
func (c *client) Seek(d float64) error         { return c.command([]any{"seek", d, "relative"}) }
func (c *client) SetVolume(v int) error        { return c.command([]any{"set_property", "volume", v}) }
func (c *client) Stop() error                  { return c.command([]any{"stop"}) }

func (c *client) SetProperty(name string, value any) error {
	return c.command([]any{"set_property", name, value})
}

func (c *client) Events() <-chan Event { return c.events }

func (c *client) Close() error {
	if !c.closed.CompareAndSwap(false, true) {
		return nil
	}
	_ = c.command([]any{"quit"})
	if c.conn != nil {
		_ = c.conn.Close()
	}
	if c.cmd != nil && c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
		_, _ = c.cmd.Process.Wait()
	}
	return nil
}
