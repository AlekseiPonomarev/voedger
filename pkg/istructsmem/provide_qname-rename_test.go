/*
 * Copyright (c) 2021-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package istructsmem

import (
	"encoding/binary"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/untillpro/voedger/pkg/iratesce"
	"github.com/untillpro/voedger/pkg/istorage"
	"github.com/untillpro/voedger/pkg/istructs"
)

func TestRenameQName(t *testing.T) {

	require := require.New(t)

	old := istructs.NewQName("test", "old")
	new := istructs.NewQName("test", "new")

	other := istructs.NewQName("test", "other")

	testStorage := func() istorage.IAppStorage {
		storage := newTestStorage()
		storageProvider := newTestStorageProvider(storage)

		cfgs := make(AppConfigsType, 1)
		cfg := cfgs.AddConfig(istructs.AppQName_test1_app1)

		_ = cfg.Schemas.Add(old, istructs.SchemaKind_Object)
		_ = cfg.Schemas.Add(other, istructs.SchemaKind_Object)

		provider, err := Provide(cfgs, iratesce.TestBucketsFactory, testTokensFactory(), storageProvider)
		require.NoError(err, err)

		_, err = provider.AppStructs(istructs.AppQName_test1_app1)
		require.NoError(err, err)

		return storage
	}

	t.Run("basic usage", func(t *testing.T) {
		storage := testStorage()

		err := RenameQName(storage, old, new)
		require.NoError(err, err)

		t.Run("check result", func(t *testing.T) {
			pKey := toBytes(uint16(QNameIDSysQNames), uint16(verSysQNames01))

			t.Run("check old is null", func(t *testing.T) {
				data := make([]byte, 0)
				ok, err := storage.Get(pKey, []byte(old.String()), &data)
				require.True(ok)
				require.NoError(err, err)
				id := QNameID(binary.BigEndian.Uint16(data))
				require.EqualValues(id, NullQNameID)
			})

			t.Run("check new is not null", func(t *testing.T) {
				data := make([]byte, 0)
				ok, err := storage.Get(pKey, []byte(new.String()), &data)
				require.True(ok)
				require.NoError(err, err)
				id := QNameID(binary.BigEndian.Uint16(data))
				require.Greater(id, QNameIDSysLast)
			})
		})
	})

	t.Run("test user level errors", func(t *testing.T) {
		t.Run("must error if old and new QNames are equals", func(t *testing.T) {
			storage := testStorage()

			err := RenameQName(storage, old, old)
			require.ErrorContains(err, "equals")
		})

		t.Run("must error if twice rename", func(t *testing.T) {
			storage := testStorage()

			err := RenameQName(storage, old, new)
			require.NoError(err)

			err = RenameQName(storage, old, new)
			require.ErrorContains(err, "already deleted")

			t.Run("but must ok reverse rename", func(t *testing.T) {
				storage := testStorage()

				err := RenameQName(storage, old, new)
				require.NoError(err)

				err = RenameQName(storage, new, old)
				require.NoError(err)
			})
		})

		t.Run("must error if old name not found", func(t *testing.T) {
			storage := testStorage()

			err := RenameQName(storage, istructs.NewQName("test", "unknown"), new)
			require.ErrorContains(err, "old QName ID not found")
		})

		t.Run("must error if new name is already exists", func(t *testing.T) {
			storage := testStorage()

			err := RenameQName(storage, old, other)
			require.ErrorContains(err, "exists")
		})
	})

	t.Run("test system level errors", func(t *testing.T) {
		t.Run("must error if no QNames system view", func(t *testing.T) {
			storage := newTestStorage()

			err := RenameQName(storage, old, new)
			require.ErrorContains(err, "read version")
		})

		t.Run("must error if unsupported version of QNames system view", func(t *testing.T) {
			storage := newTestStorage()
			data := toBytes(uint16(verSysQNamesLastest + 1)) // future version
			storage.Put(toBytes(uint16(QNameIDSysVesions)), toBytes(uint16(verSysQNames)), data)

			err := RenameQName(storage, old, new)
			require.ErrorContains(err, "unsupported version")
		})

		t.Run("must error if storage put failed for old qname", func(t *testing.T) {
			testError := errors.New("test error")

			storage := testStorage()
			storage.(*testStorageType).shedulePutError(testError, nil, []byte(old.String()))

			err := RenameQName(storage, old, new)
			require.ErrorIs(err, testError)
		})

		t.Run("must error if storage put failed for new qname", func(t *testing.T) {
			testError := errors.New("test error")

			storage := testStorage()
			storage.(*testStorageType).shedulePutError(testError, nil, []byte(new.String()))

			err := RenameQName(storage, old, new)
			require.ErrorIs(err, testError)
		})
	})
}