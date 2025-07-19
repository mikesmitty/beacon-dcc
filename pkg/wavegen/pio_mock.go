//go:build !rp

package wavegen

import "github.com/mikesmitty/beacon-dcc/pkg/shared"

func (w *Wavegen) initPIO(pioNum int, p shared.Pin) error {
	_ = pioNum
	_ = p
	return nil
}
