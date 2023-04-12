/*
 * Copyright (c) 2022-present unTill Pro, Ltd.
 */

package state

import (
	"context"
	"encoding/json"

	"github.com/untillpro/voedger/pkg/istructs"
	coreutils "github.com/untillpro/voedger/pkg/utils"
)

type wLogStorage struct {
	ctx         context.Context
	eventsFunc  eventsFunc
	schemasFunc schemasFunc
	wsidFunc    WSIDFunc
}

func (s *wLogStorage) NewKeyBuilder(istructs.QName, istructs.IStateKeyBuilder) istructs.IStateKeyBuilder {
	return &wLogKeyBuilder{
		logKeyBuilder: logKeyBuilder{
			offset: istructs.FirstOffset,
			count:  1,
		},
		wsid: s.wsidFunc(),
	}
}
func (s *wLogStorage) GetBatch(items []GetBatchItem) (err error) {
	for i := range items {
		skip := false
		err = s.Read(items[i].key, func(_ istructs.IKey, value istructs.IStateValue) (err error) {
			if skip {
				return
			}
			items[i].value = value
			skip = true
			return
		})
		if err != nil {
			break
		}
	}
	return err
}
func (s *wLogStorage) Read(kb istructs.IStateKeyBuilder, callback istructs.ValueCallback) (err error) {
	k := kb.(*wLogKeyBuilder)
	cb := func(wlogOffset istructs.Offset, event istructs.IWLogEvent) (err error) {
		offs := int64(wlogOffset)
		return callback(
			&key{data: map[string]interface{}{Field_Offset: offs}},
			&wLogStorageValue{
				event:      event,
				offset:     offs,
				toJSONFunc: s.toJSON,
			})
	}
	return s.eventsFunc().ReadWLog(s.ctx, k.wsid, k.offset, k.count, cb)
}
func (s *wLogStorage) toJSON(sv istructs.IStateValue, _ ...interface{}) (string, error) {
	value := sv.(*wLogStorageValue)
	obj := make(map[string]interface{})
	obj["QName"] = value.event.QName().String()
	obj["ArgumentObject"] = coreutils.ObjectToMap(value.event.ArgumentObject(), s.schemasFunc())
	cc := make([]map[string]interface{}, 0)
	_ = value.event.CUDs(func(rec istructs.ICUDRow) (err error) { //no error returns
		cudRowMap := cudRowToMap(rec, s.schemasFunc)
		cc = append(cc, cudRowMap)
		return
	})
	obj["CUDs"] = cc
	obj[Field_RegisteredAt] = value.event.RegisteredAt()
	obj["Synced"] = value.event.Synced()
	obj[Field_DeviceID] = value.event.DeviceID()
	obj[Field_SyncedAt] = value.event.SyncedAt()
	errObj := make(map[string]interface{})
	errObj["ErrStr"] = value.event.Error().ErrStr()
	errObj["QNameFromParams"] = value.event.Error().QNameFromParams().String()
	errObj["ValidEvent"] = value.event.Error().ValidEvent()
	errObj["OriginalEventBytes"] = value.event.Error().OriginalEventBytes()
	obj["Error"] = errObj
	obj[Field_Offset] = value.offset
	bb, err := json.Marshal(&obj)
	return string(bb), err
}