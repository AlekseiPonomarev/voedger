/*
 * Copyright (c) 2024-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package schedulers

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/voedger/voedger/pkg/appdef"
	"github.com/voedger/voedger/pkg/goutils/testingu/require"
	"github.com/voedger/voedger/pkg/istructs"
)

func TestSchedulersWaitTimeout(t *testing.T) {
	appName := istructs.AppQName_test1_app1
	partCnt := istructs.NumAppPartitions(2)
	wsCnt := istructs.NumAppWorkspaces(10)
	partID := istructs.PartitionID(1)
	jobNames := appdef.MustParseQNames("test.j1", "test.j2")

	appDef := func() appdef.IAppDef {
		adb := appdef.New()
		adb.AddPackage("test", "test.com/test")
		wsb := adb.AddWorkspace(appdef.NewQName("test", "workspace"))
		for _, name := range jobNames {
			wsb.AddJob(name).SetCronSchedule("@every 5s")
		}
		return adb.MustBuild()
	}

	require := require.New(t)

	t.Run("should ok to wait for all actualizers finished", func(t *testing.T) {
		ctx, stop := context.WithCancel(context.Background())

		schedulers := New(appName, partCnt, wsCnt, partID)

		app := appDef()

		runCalls := sync.Map{}
		runKey := func(j appdef.QName, ws istructs.AppWorkspaceNumber) string {
			return fmt.Sprintf("%s[%d]", j, ws)
		}
		for _, name := range jobNames {
			for ws := istructs.AppWorkspaceNumber(0); ws < istructs.AppWorkspaceNumber(wsCnt); ws++ {
				if ws%2 == 1 {
					runCalls.Store(runKey(name, ws), 1)
				}
			}
		}
		schedulers.Deploy(ctx, app,
			func(ctx context.Context, app appdef.AppQName, partID istructs.PartitionID, wsNum istructs.AppWorkspaceNumber, wsID istructs.WSID, name appdef.QName) {
				key := runKey(name, wsNum)
				require.True(runCalls.CompareAndDelete(key, 1), "scheduler %s was run more than once", key)

				require.Equal(appName, app)
				require.Equal(partID, partID)
				require.Contains(jobNames, name)
				require.Equal(
					istructs.NewWSID(istructs.CurrentClusterID(), istructs.WSID(wsNum)+istructs.FirstBaseAppWSID),
					wsID,
					"wsID for %s", key,
				)

				<-ctx.Done()
			})

		require.Equal(
			map[appdef.QName][]istructs.WSID{
				jobNames[0]: {
					istructs.NewWSID(istructs.CurrentClusterID(), istructs.WSID(1)+istructs.FirstBaseAppWSID),
					istructs.NewWSID(istructs.CurrentClusterID(), istructs.WSID(3)+istructs.FirstBaseAppWSID),
					istructs.NewWSID(istructs.CurrentClusterID(), istructs.WSID(5)+istructs.FirstBaseAppWSID),
					istructs.NewWSID(istructs.CurrentClusterID(), istructs.WSID(7)+istructs.FirstBaseAppWSID),
					istructs.NewWSID(istructs.CurrentClusterID(), istructs.WSID(9)+istructs.FirstBaseAppWSID),
				},
				jobNames[1]: {
					istructs.NewWSID(istructs.CurrentClusterID(), istructs.WSID(1)+istructs.FirstBaseAppWSID),
					istructs.NewWSID(istructs.CurrentClusterID(), istructs.WSID(3)+istructs.FirstBaseAppWSID),
					istructs.NewWSID(istructs.CurrentClusterID(), istructs.WSID(5)+istructs.FirstBaseAppWSID),
					istructs.NewWSID(istructs.CurrentClusterID(), istructs.WSID(7)+istructs.FirstBaseAppWSID),
					istructs.NewWSID(istructs.CurrentClusterID(), istructs.WSID(9)+istructs.FirstBaseAppWSID),
				},
			},
			schedulers.Enum())

		// stop vvm from context, wait actualizers finished
		stop()

		const timeout = 1 * time.Second
		require.True(schedulers.WaitTimeout(timeout))
		require.Empty(schedulers.Enum())

		runCalls.Range(func(key, value any) bool {
			require.Fail("scheduler %v was not run", key)
			return true
		})
	})

	t.Run("should timeout to wait infinite run schedulers", func(t *testing.T) {

		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		ctx, stop := context.WithCancel(context.Background())

		schedulers := New(appName, partCnt, wsCnt, partID)

		app := appDef()

		schedulers.Deploy(ctx, app,
			func(context.Context, appdef.AppQName, istructs.PartitionID, istructs.AppWorkspaceNumber, istructs.WSID, appdef.QName) {
				for {
					time.Sleep(time.Millisecond) // infinite loop
				}
			})

		// stop vvm from context, wait actualizers finished
		stop()

		const timeout = 1 * time.Second
		require.False(schedulers.WaitTimeout(timeout))
		require.Len(schedulers.Enum(), 2)
	})
}