package state

import (
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/voedger/voedger/pkg/appdef"
	"github.com/voedger/voedger/pkg/istructs"
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
	ret := m.Called()
	return ret.Get(0).(float64), ret.Get(1).(float64), ret.Error(2)
}

func (m *mockModbusTL) Close() {
	m.Called()
}

func TestModbusStorage_Read(t *testing.T) {
	req := require.New(t)
	mockModbusDevice := new(mockModbusTL)
	mockModbusDevice.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockModbusDevice.On("ReadData").Return(27.4, 33.5, nil)
	mockModbusDevice.On("Close")

	s := ProvideAsyncActualizerStateFactory()(context.Background(), nilAppStructsFunc, nil, nil, nil, nil, nil, nil, nil, 0, 0, WithCustomModbusClient(mockModbusDevice))

	kb, err := s.KeyBuilder(Modbus, appdef.NullQName)
	req.NoError(err)
	kb.PutString("Host", "127.0.0.1")
	kb.PutString("Port", "502")
	kb.PutInt32("SlaveID", 1)

	var v istructs.IStateValue
	err = s.Read(kb, func(_ istructs.IKey, value istructs.IStateValue) (err error) {
		v = value
		return
	})
	req.NoError(err)
	req.Equal(27.4, v.AsFloat64("T"))
	req.Equal(33.5, v.AsFloat64("H"))
}
