package tui

func SetMock(m *MockPrompts) {
	activeMock = m
}

func ClearMock() {
	activeMock = nil
}
