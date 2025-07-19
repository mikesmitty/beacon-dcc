//go:build !rp

package wavegen

import "github.com/mikesmitty/beacon-dcc/pkg/shared"

func (w *Wavegen) initPIO(sp, bp shared.Pin) error {
	_ = sp
	_ = bp
	return nil
}
