Package kbd is a simple package to allow one to test key state:
press or not-pressed. It works only for Linux systems and requires that
executables using it are started with `sudo` privledges. This is because
keyboard events are read directly from the device file in `/dev/input/`.
This also means that key events are read from the *entire system*, not just the
terminal in which the executable was run.

Example (obviously no error handling):

```go
   kb, _ := kbd.Open("/dev/input/event0")
   defer kb.Close()
   
   kb.Start()
   for key := range kb.Event() {
   		switch key {
   		case kbd.KeyA:
   			if kb.IsDown(key) {
  				fmt.Println("A down")
  			} else {
  				fmt.Println("A up")
   			}
   		case kbd.KeyESC:
  			if kb.IsDown(key) {
  				fmt.Println("ESC")
  				kb.Stop()
  			}
   		}
   }
   fmt.Println("Error:", kb.Err())
```