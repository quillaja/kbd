// Package kbd is a simple package to allow one to test key state:
// press or not-pressed. It works only for Linux systems and requires that
// executables using it are started with `sudo` privledges. This is because
// keyboard events are read directly from the device file in `/dev/input/`.
// This also means that key events are read from the *entire system*, not just the
// terminal in which the executable was run.
//
// Example (obviously no error handling):
//
//    kb, _ := kbd.Open("/dev/input/event0")
//    defer kb.Close()
//
//    kb.Start()
//    for key := range kb.Event() {
//    		switch key {
//    		case kbd.KeyA:
//    			if kb.IsDown(key) {
//   				fmt.Println("A down")
//   			} else {
//   				fmt.Println("A up")
//    			}
//
//    		case kbd.KeyESC:
//   			if kb.IsDown(key) {
//   				fmt.Println("ESC")
//   				kb.Stop()
//   			}
//    		}
//    }
//    fmt.Println("Error:", kb.Err())
package kbd

import (
	"encoding/binary"
	"os"
	"sync"

	"github.com/pkg/term"
)

type inputEvent struct {
	Timeval [16]byte
	Kind    uint16
	Code    uint16
	Value   uint32
}

// Keyboard allows access to key states.
type Keyboard struct {
	mu      sync.Mutex
	keys    map[KeyCode]bool
	kbfile  *os.File
	tty     *term.Term
	events  chan KeyCode
	running bool
	err     error
}

// Open will attempt to open the device at path as well as the terminal at
// `/dev/tty`. An error is returned if either of these fails.
func Open(path string) (*Keyboard, error) {
	var err error
	kb := &Keyboard{
		keys: map[KeyCode]bool{},
	}

	kb.tty, err = term.Open("/dev/tty")
	if err != nil {
		return nil, err
	}
	kb.kbfile, err = os.Open(path)
	if err != nil {
		return nil, err
	}

	return kb, err
}

// Start puts the terminal in "cbreak" mode (to prevent key echo) and kicks off
// a gofunc to read keyboard events. An error is returned if the terminal can't
// be put into cbreak mode. Errors affecting (ending) the keyboard event reading loop
// can be examined with Err().
func (kb *Keyboard) Start() error {
	err := term.CBreakMode(kb.tty)
	if err != nil {
		return err
	}
	kb.running = true
	kb.events = make(chan KeyCode)

	// kb.mu.Lock()
	// kb.keys = make(map[uint16]bool)
	// kb.mu.Unlock()

	go func() {
		var event inputEvent
		var err error
		for kb.running && err == nil {

			err = binary.Read(kb.kbfile, binary.LittleEndian, &event)
			if err != nil {
				continue // go to top of loop and end loop
			}

			if event.Kind == eventKEY {
				if event.Value != repeat { // don't change state for repeat codes

					kb.mu.Lock()
					kb.keys[KeyCode(event.Code)] = event.Value == press // set "true" when key is pressed
					kb.mu.Unlock()

					select { // non-blocking channel recieve to "drain" channel
					case <-kb.events:
					default:
					}
					select { // non-blocking channel send
					case kb.events <- KeyCode(event.Code):
					default:
					}
				}
			}
			err = kb.tty.Flush() // remove keypress(es) from stream
		}
		close(kb.events)
		if err != nil {
			kb.Stop() // restore the terminal if there's an error
			kb.err = err
		}
	}()

	return nil
}

// Stop restores the terminal state and stops reading keyboard events.
func (kb *Keyboard) Stop() error {
	err := kb.tty.Restore()
	kb.running = false
	return err
}

// Close calls Stop() and also closes files used by the Keyboard.
func (kb *Keyboard) Close() error {
	err := kb.Stop()
	err = kb.kbfile.Close()
	err = kb.tty.Close()
	return err
}

// Err reads the error that ended the keyboard event reading loop.
func (kb *Keyboard) Err() error {
	return kb.err
}

// IsDown checks if the key is pressed or held (aka repeat).
func (kb *Keyboard) IsDown(key KeyCode) bool {
	kb.mu.Lock()
	defer kb.mu.Unlock()
	return kb.keys[key]
}

// Event returns a channel from which the most recently read KeyCode
// can be obtained.
func (kb *Keyboard) Event() <-chan KeyCode {
	return kb.events
}

// Values for key events.
const (
	release = 0
	press   = 1
	repeat  = 2
)

// Types of events available from /dev/input/... files.
// We're only interested in eventKEY (EV_KEY)
const (
	eventSYN       = 0x00
	eventKEY       = 0x01
	eventREL       = 0x02
	eventABS       = 0x03
	eventMSC       = 0x04
	eventSW        = 0x05
	eventLED       = 0x11
	eventSND       = 0x12
	eventREP       = 0x14
	eventFF        = 0x15
	eventPWR       = 0x16
	eventFF_STATUS = 0x17
	eventMAX       = 0x1f
	eventCNT       = eventMAX + 1
)

// KeyCode is a code from "input-event-codes.h"
type KeyCode uint16

// KeyCodes for many keys.
const (
	KeyRESERVED   KeyCode = 0
	KeyESC        KeyCode = 1
	Key1          KeyCode = 2
	Key2          KeyCode = 3
	Key3          KeyCode = 4
	Key4          KeyCode = 5
	Key5          KeyCode = 6
	Key6          KeyCode = 7
	Key7          KeyCode = 8
	Key8          KeyCode = 9
	Key9          KeyCode = 10
	Key0          KeyCode = 11
	KeyMINUS      KeyCode = 12
	KeyEQUAL      KeyCode = 13
	KeyBACKSPACE  KeyCode = 14
	KeyTAB        KeyCode = 15
	KeyQ          KeyCode = 16
	KeyW          KeyCode = 17
	KeyE          KeyCode = 18
	KeyR          KeyCode = 19
	KeyT          KeyCode = 20
	KeyY          KeyCode = 21
	KeyU          KeyCode = 22
	KeyI          KeyCode = 23
	KeyO          KeyCode = 24
	KeyP          KeyCode = 25
	KeyLEFTBRACE  KeyCode = 26
	KeyRIGHTBRACE KeyCode = 27
	KeyENTER      KeyCode = 28
	KeyLEFTCTRL   KeyCode = 29
	KeyA          KeyCode = 30
	KeyS          KeyCode = 31
	KeyD          KeyCode = 32
	KeyF          KeyCode = 33
	KeyG          KeyCode = 34
	KeyH          KeyCode = 35
	KeyJ          KeyCode = 36
	KeyK          KeyCode = 37
	KeyL          KeyCode = 38
	KeySEMICOLON  KeyCode = 39
	KeyAPOSTROPHE KeyCode = 40
	KeyGRAVE      KeyCode = 41
	KeyLEFTSHIFT  KeyCode = 42
	KeyBACKSLASH  KeyCode = 43
	KeyZ          KeyCode = 44
	KeyX          KeyCode = 45
	KeyC          KeyCode = 46
	KeyV          KeyCode = 47
	KeyB          KeyCode = 48
	KeyN          KeyCode = 49
	KeyM          KeyCode = 50
	KeyCOMMA      KeyCode = 51
	KeyDOT        KeyCode = 52
	KeySLASH      KeyCode = 53
	KeyRIGHTSHIFT KeyCode = 54
	KeyKPASTERISK KeyCode = 55
	KeyLEFTALT    KeyCode = 56
	KeySPACE      KeyCode = 57
	KeyCAPSLOCK   KeyCode = 58
	KeyF1         KeyCode = 59
	KeyF2         KeyCode = 60
	KeyF3         KeyCode = 61
	KeyF4         KeyCode = 62
	KeyF5         KeyCode = 63
	KeyF6         KeyCode = 64
	KeyF7         KeyCode = 65
	KeyF8         KeyCode = 66
	KeyF9         KeyCode = 67
	KeyF10        KeyCode = 68
	KeyNUMLOCK    KeyCode = 69
	KeySCROLLLOCK KeyCode = 70
)
