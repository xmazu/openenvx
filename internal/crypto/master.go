package crypto

const (
	masterKeySize = 32 // 256 bits
)

type MasterKey struct {
	key []byte
}

func (m *MasterKey) Bytes() []byte {
	return m.key
}
