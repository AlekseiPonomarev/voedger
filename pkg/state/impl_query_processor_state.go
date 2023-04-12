/*
 * Copyright (c) 2022-present unTill Pro, Ltd.
 */

package state

import (
	"context"

	"github.com/untillpro/voedger/pkg/isecrets"
	"github.com/untillpro/voedger/pkg/istructs"
)

func implProvideQueryProcessorState(ctx context.Context, appStructs istructs.IAppStructs, partitionIDFunc PartitionIDFunc, wsidFunc WSIDFunc,
	secretReader isecrets.ISecretReader, principalsFunc PrincipalsFunc, tokenFunc TokenFunc) IHostState {
	bs := newHostState("QueryProcessor", 0)

	bs.addStorage(ViewRecordsStorage, &viewRecordsStorage{
		ctx:             ctx,
		viewRecordsFunc: func() istructs.IViewRecords { return appStructs.ViewRecords() },
		schemasFunc:     func() istructs.ISchemas { return appStructs.Schemas() },
		wsidFunc:        wsidFunc,
	}, S_GET_BATCH|S_READ)

	bs.addStorage(RecordsStorage, &recordsStorage{
		recordsFunc: func() istructs.IRecords { return appStructs.Records() },
		schemasFunc: func() istructs.ISchemas { return appStructs.Schemas() },
		wsidFunc:    wsidFunc,
	}, S_GET_BATCH)

	bs.addStorage(WLogStorage, &wLogStorage{
		ctx:         ctx,
		eventsFunc:  func() istructs.IEvents { return appStructs.Events() },
		schemasFunc: func() istructs.ISchemas { return appStructs.Schemas() },
		wsidFunc:    wsidFunc,
	}, S_GET_BATCH|S_READ)

	bs.addStorage(PLogStorage, &pLogStorage{
		ctx:             ctx,
		eventsFunc:      func() istructs.IEvents { return appStructs.Events() },
		schemasFunc:     func() istructs.ISchemas { return appStructs.Schemas() },
		partitionIDFunc: partitionIDFunc,
	}, S_GET_BATCH|S_READ)

	bs.addStorage(HTTPStorage, &httpStorage{}, S_READ)

	bs.addStorage(AppSecretsStorage, &appSecretsStorage{secretReader: secretReader}, S_GET_BATCH)

	bs.addStorage(SubjectStorage, &subjectStorage{
		principalsFunc: principalsFunc,
		tokenFunc:      tokenFunc,
	}, S_GET_BATCH)

	return bs
}