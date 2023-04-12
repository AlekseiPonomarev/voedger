/*
 * Copyright (c) 2022-present unTill Pro, Ltd.
 */

package state

import (
	"context"
	"testing"

	"github.com/untillpro/voedger/pkg/istructs"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBundledHostState_BasicUsage(t *testing.T) {
	require := require.New(t)
	factory := ProvideAsyncActualizerStateFactory()
	n10nFn := func(view istructs.QName, wsid istructs.WSID, offset istructs.Offset) {}
	appStructs := mockedAppStructs()

	// Create instance of async actualizer state
	aaState := factory(context.Background(), appStructs, nil, SimpleWSIDFunc(istructs.WSID(1)), n10nFn, nil, 2, 1)

	// Declare simple extension
	extension := func(state istructs.IState, intents istructs.IIntents) {
		//Create key
		kb, err := state.KeyBuilder(ViewRecordsStorage, testViewRecordQName1)
		require.NoError(err)
		kb.PutString("pkFld", "pkVal")

		// Create new value
		eb, _ := intents.NewValue(kb)
		eb.PutInt64("vFld", 10)
		eb.PutInt64(ColOffset, 45)
	}

	// Run extension
	extension(aaState, aaState)

	// Apply intents
	readyToFlush, _ := aaState.ApplyIntents()
	require.True(readyToFlush)

	_ = aaState.FlushBundles()
}

func mockedAppStructs() istructs.IAppStructs {
	mv := &mockValue{}
	mv.
		On("AsInt64", "vFld").Return(int64(10)).
		On("AsInt64", ColOffset).Return(int64(45))
	mvb1 := &mockValueBuilder{}
	mvb1.
		On("PutInt64", "vFld", int64(10)).
		On("PutInt64", ColOffset, int64(45)).
		On("Build").Return(mv)
	mvb2 := &mockValueBuilder{}
	mvb2.
		On("PutInt64", "vFld", int64(10)).Once().
		On("PutInt64", ColOffset, int64(45)).Once().
		On("PutInt64", "vFld", int64(17)).Once().
		On("PutInt64", ColOffset, int64(46)).Once()
	mkb := &mockKeyBuilder{}
	mkb.
		On("PutString", "pkFld", "pkVal")
	viewRecords := &mockViewRecords{}
	viewRecords.
		On("KeyBuilder", testViewRecordQName1).Return(mkb).
		On("NewValueBuilder", testViewRecordQName1).Return(mvb1).Once().
		On("NewValueBuilder", testViewRecordQName1).Return(mvb2).Once().
		On("PutBatch", istructs.WSID(1), mock.AnythingOfType("[]istructs.ViewKV")).Return(nil)
	pkSchema := &mockSchema{}
	pkSchema.
		On("Fields", mock.Anything).
		Run(func(args mock.Arguments) {
			cb := args.Get(0).(func(fieldName string, kind istructs.DataKindType))
			cb("pkFld", istructs.DataKind_string)
		})
	vSchema := &mockSchema{}
	vSchema.
		On("Fields", mock.Anything).
		Run(func(args mock.Arguments) {
			cb := args.Get(0).(func(fieldName string, kind istructs.DataKindType))
			cb("vFld", istructs.DataKind_int64)
			cb(ColOffset, istructs.DataKind_int64)
		})
	schema := &mockSchema{}
	schema.
		On("Containers", mock.AnythingOfType("func(string, istructs.QName)")).
		Run(func(args mock.Arguments) {
			cb := args.Get(0).(func(string, istructs.QName))
			cb(istructs.SystemContainer_ViewPartitionKey, testViewRecordPkQName)
			cb(istructs.SystemContainer_ViewValue, testViewRecordVQName)
		})
	schemas := &mockSchemas{}
	schemas.
		On("Schema", testViewRecordQName1).Return(schema).
		On("Schema", testViewRecordPkQName).Return(pkSchema).
		On("Schema", testViewRecordVQName).Return(vSchema)
	appStructs := &mockAppStructs{}
	appStructs.
		On("ViewRecords").Return(viewRecords).
		On("Schemas").Return(schemas).
		On("Events").Return(&nilEvents{}).
		On("Records").Return(&nilRecords{})
	return appStructs
}

func TestAsyncActualizerState_BasicUsage_Old(t *testing.T) {
	require := require.New(t)
	touched := false
	n10nFn := func(view istructs.QName, wsid istructs.WSID, offset istructs.Offset) {
		touched = true
		require.Equal(testViewRecordQName1, view)
		require.Equal(istructs.WSID(1), wsid)
		require.Equal(istructs.Offset(46), offset)
	}
	mv := &mockValue{}
	mv.
		On("AsInt64", "vFld").Return(int64(10)).
		On("AsInt64", ColOffset).Return(int64(45))
	mvb1 := &mockValueBuilder{}
	mvb1.
		On("PutInt64", "vFld", int64(10)).
		On("PutInt64", ColOffset, int64(45)).
		On("Build").Return(mv)
	mvb2 := &mockValueBuilder{}
	mvb2.
		On("PutInt64", "vFld", int64(17)).
		On("PutInt64", ColOffset, int64(46))
	mkb := &mockKeyBuilder{}
	mkb.
		On("PutString", "pkFld", "pkVal").
		On("Equals", mock.Anything).Return(true)
	viewRecords := &mockViewRecords{}
	viewRecords.
		On("KeyBuilder", testViewRecordQName1).Return(mkb).
		On("NewValueBuilder", testViewRecordQName1).Return(mvb1).
		On("UpdateValueBuilder", testViewRecordQName1, mock.Anything).Return(mvb2).
		On("PutBatch", istructs.WSID(1), mock.AnythingOfType("[]istructs.ViewKV")).Return(nil)
	appStructs := &mockAppStructs{}
	appStructs.
		On("ViewRecords").Return(viewRecords).
		On("Schemas").Return(&nilSchemas{}).
		On("Events").Return(&nilEvents{}).
		On("Records").Return(&nilRecords{})
	s := ProvideAsyncActualizerStateFactory()(context.Background(), appStructs, nil, SimpleWSIDFunc(istructs.WSID(1)), n10nFn, nil, 2, 1)

	//Create key
	kb, err := s.KeyBuilder(ViewRecordsStorage, testViewRecordQName1)
	require.NoError(err)
	kb.PutString("pkFld", "pkVal")

	//Create new value and put it to bundle
	eb, _ := s.NewValue(kb)
	eb.PutInt64("vFld", 10)
	eb.PutInt64(ColOffset, 45)
	readyToFlush, _ := s.ApplyIntents()
	require.True(readyToFlush)

	//Get value from bundle by key
	el, _ := s.MustExist(kb)
	eb, _ = s.UpdateValue(kb, el)
	eb.PutInt64("vFld", el.AsInt64("vFld")+int64(7))
	eb.PutInt64(ColOffset, 46)

	//Store key-value pair in under laying storage
	readyToFlush, _ = s.ApplyIntents()
	require.True(readyToFlush)
	_ = s.FlushBundles()

	require.True(touched)
	mvb1.AssertExpectations(t)
	mvb2.AssertExpectations(t)
	mkb.AssertExpectations(t)
}
func TestAsyncActualizerState_CanExist(t *testing.T) {
	t.Run("Should be ok", func(t *testing.T) {
		require := require.New(t)
		stateStorage := &mockStorage{}
		stateStorage.
			On("NewKeyBuilder", testViewRecordQName1, nil).Return(newKeyBuilder(testStorage, testViewRecordQName1)).
			On("GetBatch", mock.AnythingOfType("[]state.GetBatchItem")).
			Return(nil).
			Run(func(args mock.Arguments) {
				args.Get(0).([]GetBatchItem)[0].value = &mockStateValue{}
			})
		s := asyncActualizerStateWithTestStateStorage(stateStorage)
		kb, _ := s.KeyBuilder(testStorage, testViewRecordQName1)

		_, ok, _ := s.CanExist(kb)

		require.True(ok)
	})
	t.Run("Should return error when error occurred on get batch", func(t *testing.T) {
		require := require.New(t)
		stateStorage := &mockStorage{}
		stateStorage.
			On("NewKeyBuilder", testViewRecordQName1, nil).Return(newKeyBuilder(testStorage, testViewRecordQName1)).
			On("GetBatch", mock.AnythingOfType("[]state.GetBatchItem")).Return(errTest)
		s := asyncActualizerStateWithTestStateStorage(stateStorage)
		kb, _ := s.KeyBuilder(testStorage, testViewRecordQName1)

		_, _, err := s.CanExist(kb)

		require.ErrorIs(err, errTest)
	})
}
func TestAsyncActualizerState_CanExistAll(t *testing.T) {
	t.Run("Should be ok", func(t *testing.T) {
		require := require.New(t)
		times := 0
		stateStorage := &mockStorage{}
		stateStorage.
			On("NewKeyBuilder", testViewRecordQName1, nil).Return(newKeyBuilder(testStorage, testViewRecordQName1)).
			On("GetBatch", mock.AnythingOfType("[]state.GetBatchItem")).Return(nil)
		s := asyncActualizerStateWithTestStateStorage(stateStorage)
		kb1, _ := s.KeyBuilder(testStorage, testViewRecordQName1)
		kb2, _ := s.KeyBuilder(testStorage, testViewRecordQName1)

		_ = s.CanExistAll([]istructs.IStateKeyBuilder{kb1, kb2}, func(istructs.IKeyBuilder, istructs.IStateValue, bool) error {
			times++
			return nil
		})

		require.Equal(2, times)
	})
	t.Run("Should return error when error occurred on can exist", func(t *testing.T) {
		require := require.New(t)
		stateStorage := &mockStorage{}
		stateStorage.
			On("NewKeyBuilder", testViewRecordQName1, nil).Return(newKeyBuilder(testStorage, testViewRecordQName1)).
			On("GetBatch", mock.AnythingOfType("[]state.GetBatchItem")).Return(errTest)
		s := asyncActualizerStateWithTestStateStorage(stateStorage)
		kb1, _ := s.KeyBuilder(testStorage, testViewRecordQName1)
		kb2, _ := s.KeyBuilder(testStorage, testViewRecordQName1)

		err := s.CanExistAll([]istructs.IStateKeyBuilder{kb1, kb2}, nil)

		require.ErrorIs(err, errTest)
	})
}
func TestAsyncActualizerState_MustExist(t *testing.T) {
	t.Run("Should be ok", func(t *testing.T) {
		require := require.New(t)
		stateStorage := &mockStorage{}
		stateStorage.
			On("NewKeyBuilder", testViewRecordQName1, nil).Return(newKeyBuilder(testStorage, testViewRecordQName1)).
			On("GetBatch", mock.AnythingOfType("[]state.GetBatchItem")).
			Return(nil).
			Run(func(args mock.Arguments) {
				args.Get(0).([]GetBatchItem)[0].value = &mockStateValue{}
			})
		s := asyncActualizerStateWithTestStateStorage(stateStorage)
		kb, _ := s.KeyBuilder(testStorage, testViewRecordQName1)

		_, err := s.MustExist(kb)

		require.NoError(err)
	})
	t.Run("Should return error when entity not exists", func(t *testing.T) {
		require := require.New(t)
		stateStorage := &mockStorage{}
		stateStorage.
			On("NewKeyBuilder", testViewRecordQName1, nil).Return(newKeyBuilder(testStorage, testViewRecordQName1)).
			On("GetBatch", mock.AnythingOfType("[]state.GetBatchItem")).
			Return(nil).
			Run(func(args mock.Arguments) {
				args.Get(0).([]GetBatchItem)[0].value = nil
			})
		s := asyncActualizerStateWithTestStateStorage(stateStorage)
		kb, _ := s.KeyBuilder(testStorage, testViewRecordQName1)

		_, err := s.MustExist(kb)

		require.ErrorIs(err, ErrNotExists)
	})
	t.Run("Should return error when error occurred on can exist", func(t *testing.T) {
		require := require.New(t)
		stateStorage := &mockStorage{}
		stateStorage.
			On("NewKeyBuilder", testViewRecordQName1, nil).Return(newKeyBuilder(testStorage, testViewRecordQName1)).
			On("GetBatch", mock.AnythingOfType("[]state.GetBatchItem")).Return(errTest)
		s := asyncActualizerStateWithTestStateStorage(stateStorage)
		kb, _ := s.KeyBuilder(testStorage, testViewRecordQName1)

		_, err := s.MustExist(kb)

		require.ErrorIs(err, errTest)
	})
}
func TestAsyncActualizerState_MustExistAll(t *testing.T) {
	t.Run("Should be ok", func(t *testing.T) {
		require := require.New(t)
		stateStorage := &mockStorage{}
		stateStorage.
			On("NewKeyBuilder", testViewRecordQName1, nil).Return(newKeyBuilder(testStorage, testViewRecordQName1)).
			On("GetBatch", mock.AnythingOfType("[]state.GetBatchItem")).
			Return(nil).
			Run(func(args mock.Arguments) {
				args.Get(0).([]GetBatchItem)[0].value = &mockStateValue{}
			})
		s := asyncActualizerStateWithTestStateStorage(stateStorage)
		k1, _ := s.KeyBuilder(testStorage, testViewRecordQName1)
		k2, _ := s.KeyBuilder(testStorage, testViewRecordQName1)
		kk := make([]istructs.IKeyBuilder, 0, 2)

		_ = s.MustExistAll([]istructs.IStateKeyBuilder{k1, k2}, func(key istructs.IKeyBuilder, value istructs.IStateValue, ok bool) (err error) {
			kk = append(kk, key)
			require.True(ok)
			return
		})

		require.Equal(k1, kk[0])
		require.Equal(k1, kk[1])
	})
	t.Run("Should return error when entity not exists", func(t *testing.T) {
		require := require.New(t)
		stateStorage := &mockStorage{}
		stateStorage.
			On("NewKeyBuilder", testViewRecordQName1, nil).Return(newKeyBuilder(testStorage, testViewRecordQName1)).
			On("GetBatch", mock.AnythingOfType("[]state.GetBatchItem")).
			Return(nil).
			Run(func(args mock.Arguments) {
				args.Get(0).([]GetBatchItem)[0].value = nil
			})
		s := asyncActualizerStateWithTestStateStorage(stateStorage)
		k1, _ := s.KeyBuilder(testStorage, testViewRecordQName1)
		k2, _ := s.KeyBuilder(testStorage, testViewRecordQName1)

		err := s.MustExistAll([]istructs.IStateKeyBuilder{k1, k2}, nil)

		require.ErrorIs(err, ErrNotExists)
	})
}
func TestAsyncActualizerState_MustNotExist(t *testing.T) {
	t.Run("Should be ok", func(t *testing.T) {
		require := require.New(t)
		stateStorage := &mockStorage{}
		stateStorage.
			On("NewKeyBuilder", testViewRecordQName1, nil).Return(newKeyBuilder(testStorage, testViewRecordQName1)).
			On("GetBatch", mock.AnythingOfType("[]state.GetBatchItem")).
			Return(nil).
			Run(func(args mock.Arguments) {
				args.Get(0).([]GetBatchItem)[0].value = nil
			})
		s := asyncActualizerStateWithTestStateStorage(stateStorage)
		k, _ := s.KeyBuilder(testStorage, testViewRecordQName1)

		err := s.MustNotExist(k)

		require.NoError(err)
	})
	t.Run("Should return error when entity exists", func(t *testing.T) {
		require := require.New(t)
		stateStorage := &mockStorage{}
		stateStorage.
			On("NewKeyBuilder", testViewRecordQName1, nil).Return(newKeyBuilder(testStorage, testViewRecordQName1)).
			On("GetBatch", mock.AnythingOfType("[]state.GetBatchItem")).
			Return(nil).
			Run(func(args mock.Arguments) {
				args.Get(0).([]GetBatchItem)[0].value = &mockStateValue{}
			})
		s := asyncActualizerStateWithTestStateStorage(stateStorage)
		k, _ := s.KeyBuilder(testStorage, testViewRecordQName1)

		err := s.MustNotExist(k)

		require.ErrorIs(err, ErrExists)
	})
	t.Run("Should return error when error occurred on must exist", func(t *testing.T) {
		require := require.New(t)
		stateStorage := &mockStorage{}
		stateStorage.
			On("NewKeyBuilder", testViewRecordQName1, nil).Return(newKeyBuilder(testStorage, testViewRecordQName1)).
			On("GetBatch", mock.AnythingOfType("[]state.GetBatchItem")).Return(errTest)
		s := asyncActualizerStateWithTestStateStorage(stateStorage)
		kb, _ := s.KeyBuilder(testStorage, testViewRecordQName1)

		err := s.MustNotExist(kb)

		require.ErrorIs(err, errTest)
	})
}
func TestAsyncActualizerState_MustNotExistAll(t *testing.T) {
	t.Run("Should be ok", func(t *testing.T) {
		require := require.New(t)
		stateStorage := &mockStorage{}
		stateStorage.
			On("NewKeyBuilder", testViewRecordQName1, nil).Return(newKeyBuilder(testStorage, testViewRecordQName1)).
			On("GetBatch", mock.AnythingOfType("[]state.GetBatchItem")).
			Return(nil).
			Run(func(args mock.Arguments) {
				args.Get(0).([]GetBatchItem)[0].value = nil
			})
		s := asyncActualizerStateWithTestStateStorage(stateStorage)
		k1, _ := s.KeyBuilder(testStorage, testViewRecordQName1)
		k2, _ := s.KeyBuilder(testStorage, testViewRecordQName1)

		err := s.MustNotExistAll([]istructs.IStateKeyBuilder{k1, k2})

		require.NoError(err)
	})
	t.Run("Should return error when entity exists", func(t *testing.T) {
		require := require.New(t)
		stateStorage := &mockStorage{}
		stateStorage.
			On("NewKeyBuilder", testViewRecordQName1, nil).Return(newKeyBuilder(testStorage, testViewRecordQName1)).
			On("GetBatch", mock.AnythingOfType("[]state.GetBatchItem")).
			Return(nil).
			Run(func(args mock.Arguments) {
				args.Get(0).([]GetBatchItem)[0].value = &mockStateValue{}
			})
		s := asyncActualizerStateWithTestStateStorage(stateStorage)
		k1, _ := s.KeyBuilder(testStorage, testViewRecordQName1)
		k2, _ := s.KeyBuilder(testStorage, testViewRecordQName1)

		err := s.MustNotExistAll([]istructs.IStateKeyBuilder{k1, k2})

		require.ErrorIs(err, ErrExists)
	})
}
func TestAsyncActualizerState_Read(t *testing.T) {
	t.Run("Should flush bundle before read", func(t *testing.T) {
		require := require.New(t)
		touched := false
		schemas := &mockSchemas{}
		schemas.On("Schema", testViewRecordQName1).Return(&nilSchema{})
		viewRecords := &mockViewRecords{}
		viewRecords.
			On("KeyBuilder", testViewRecordQName1).Return(&nilKeyBuilder{}).
			On("KeyBuilder", testViewRecordQName2).Return(&nilKeyBuilder{}).
			On("NewValueBuilder", testViewRecordQName1).Return(&nilValueBuilder{}).
			On("NewValueBuilder", testViewRecordQName2).Return(&nilValueBuilder{}).
			On("PutBatch", istructs.WSID(1), mock.AnythingOfType("[]istructs.ViewKV")).
			Return(nil).
			Run(func(args mock.Arguments) {
				require.Len(args.Get(1).([]istructs.ViewKV), 2)
			}).
			On("Read", context.Background(), istructs.WSID(1), mock.Anything, mock.AnythingOfType("istructs.ValuesCallback")).
			Return(nil).
			Run(func(args mock.Arguments) {
				_ = args.Get(3).(istructs.ValuesCallback)(&nilKey{}, &nilValue{})
			})
		appStructs := &mockAppStructs{}
		appStructs.
			On("ViewRecords").Return(viewRecords).
			On("Schemas").Return(schemas).
			On("Records").Return(&nilRecords{}).
			On("Events").Return(&nilEvents{})
		s := ProvideAsyncActualizerStateFactory()(context.Background(), appStructs, nil, SimpleWSIDFunc(istructs.WSID(1)), nil, nil, 10, 10)
		kb1, _ := s.KeyBuilder(ViewRecordsStorage, testViewRecordQName1)
		kb2, _ := s.KeyBuilder(ViewRecordsStorage, testViewRecordQName2)

		_, _ = s.NewValue(kb1)
		_, _ = s.NewValue(kb2)

		readyToFlush, err := s.ApplyIntents()
		require.False(readyToFlush)
		require.NoError(err)

		_ = s.Read(kb1, func(key istructs.IKey, value istructs.IStateValue) (err error) {
			touched = true
			return
		})

		require.True(touched)
	})
	t.Run("Should return error when error occurred on apply batch", func(t *testing.T) {
		require := require.New(t)
		touched := false
		schemas := &mockSchemas{}
		schemas.On("Schema", testViewRecordQName1).Return(&nilSchema{})
		viewRecords := &mockViewRecords{}
		viewRecords.
			On("KeyBuilder", testViewRecordQName1).Return(&nilKeyBuilder{}).
			On("KeyBuilder", testViewRecordQName2).Return(&nilKeyBuilder{}).
			On("NewValueBuilder", testViewRecordQName1).Return(&nilValueBuilder{}).
			On("NewValueBuilder", testViewRecordQName2).Return(&nilValueBuilder{}).
			On("PutBatch", istructs.WSID(1), mock.AnythingOfType("[]istructs.ViewKV")).Return(errTest)
		appStructs := &mockAppStructs{}
		appStructs.
			On("ViewRecords").Return(viewRecords).
			On("Schemas").Return(schemas).
			On("Records").Return(&nilRecords{}).
			On("Events").Return(&nilEvents{})
		s := ProvideAsyncActualizerStateFactory()(context.Background(), appStructs, nil, SimpleWSIDFunc(istructs.WSID(1)), nil, nil, 10, 10)
		kb1, _ := s.KeyBuilder(ViewRecordsStorage, testViewRecordQName1)
		kb2, _ := s.KeyBuilder(ViewRecordsStorage, testViewRecordQName2)

		_, _ = s.NewValue(kb1)
		_, _ = s.NewValue(kb2)

		readyToFlush, err := s.ApplyIntents()
		require.False(readyToFlush)
		require.NoError(err)

		err = s.Read(kb1, func(key istructs.IKey, value istructs.IStateValue) (err error) {
			touched = true
			return err
		})

		require.ErrorIs(err, errTest)
		require.False(touched)
	})
}
func asyncActualizerStateWithTestStateStorage(s *mockStorage) istructs.IState {
	as := ProvideAsyncActualizerStateFactory()(context.Background(), &nilAppStructs{}, nil, nil, nil, nil, 10, 10)
	as.(*bundledHostState).addStorage(testStorage, s, S_GET_BATCH|S_READ)
	return as
}