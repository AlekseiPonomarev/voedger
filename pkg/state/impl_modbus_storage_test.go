package state

import (
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
	"testing"
)

type mockModbusTL struct {
	mock.Mock
}

func (m *mockModbusTL) Connect(ip string, port string, slaveID byte) error {
	ret := m.Called(ip, port, slaveID)

	return ret.Error(0)
}

func (m *mockModbusTL) ReadData() (float64, float64, error) {
	panic("implement me")
}

func (m *mockModbusTL) Close() {
	m.Called()
}

func TestModbusStorage_Read(t *testing.T) {
	mockModbusDevice := new(mockModbusTL)
	s := ProvideAsyncActualizerStateFactory()(context.Background(), nilAppStructsFunc, nil, nil, nil, nil, nil, nil, nil, 0, 0)

	mockModbusDevice.On("Connect", "127.0.0.1", "502", byte(1)).Return(nil)

}
