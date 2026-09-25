package hotkey

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

func init() {
	// xgb logs harmless notices (e.g. a missing ~/.Xauthority) to stderr.
	xgb.Logger = log.New(io.Discard, "", 0)
}

var errKeyTaken = errors.New("shortcut already taken by another application")

func x11ErrorMessage(err error) string {
	if errors.Is(err, errKeyTaken) {
		return "Das Tastenkürzel ist bereits von einer anderen Anwendung belegt."
	}
	return fmt.Sprintf("Tastenkürzel konnte nicht registriert werden (%v).", err)
}

// listenX11 grabs the key on the root window and calls onPress for every
// key press until ctx is cancelled.
func listenX11(ctx context.Context, acc Accelerator, onPress func()) error {
	conn, err := xgb.NewConn()
	if err != nil {
		return err
	}
	setup := xproto.Setup(conn)
	root := setup.DefaultScreen(conn).Root

	keycode, err := findKeycode(conn, setup, acc.keysym())
	if err != nil {
		conn.Close()
		return err
	}

	var mods uint16
	if acc.Shift {
		mods |= xproto.ModMaskShift
	}
	if acc.Ctrl {
		mods |= xproto.ModMaskControl
	}
	if acc.Alt {
		mods |= xproto.ModMask1
	}
	if acc.Super {
		mods |= xproto.ModMask4
	}
	// Also grab with Caps Lock and Num Lock (Mod2) active, or the shortcut
	// silently stops working when either is on.
	variants := []uint16{0, xproto.ModMaskLock, xproto.ModMask2, xproto.ModMaskLock | xproto.ModMask2}
	for _, v := range variants {
		err := xproto.GrabKeyChecked(conn, true, root, mods|v, keycode, xproto.GrabModeAsync, xproto.GrabModeAsync).Check()
		if err != nil {
			for _, u := range variants {
				xproto.UngrabKey(conn, keycode, root, mods|u)
			}
			conn.Close()
			var access xproto.AccessError
			if errors.As(err, &access) {
				return errKeyTaken
			}
			return err
		}
	}

	go func() {
		<-ctx.Done()
		for _, v := range variants {
			xproto.UngrabKey(conn, keycode, root, mods|v)
		}
		conn.Close()
	}()
	go func() {
		for {
			ev, err := conn.WaitForEvent()
			if ev == nil && err == nil {
				return // connection closed
			}
			if kp, ok := ev.(xproto.KeyPressEvent); ok && kp.Detail == keycode {
				onPress()
			}
		}
	}()
	return nil
}

func findKeycode(conn *xgb.Conn, setup *xproto.SetupInfo, keysym uint32) (xproto.Keycode, error) {
	count := byte(setup.MaxKeycode - setup.MinKeycode + 1)
	reply, err := xproto.GetKeyboardMapping(conn, setup.MinKeycode, count).Reply()
	if err != nil {
		return 0, err
	}
	per := int(reply.KeysymsPerKeycode)
	for i := 0; i < int(count); i++ {
		for j := 0; j < per; j++ {
			if uint32(reply.Keysyms[i*per+j]) == keysym {
				return setup.MinKeycode + xproto.Keycode(i), nil
			}
		}
	}
	return 0, fmt.Errorf("key not found on the keyboard layout")
}
