//go:build !rp

package motor

func (m *Motor) Init(trackId string) error {
	m.trackId = trackId
	return nil
}
