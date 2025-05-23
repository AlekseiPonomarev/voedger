/*
 * Copyright (c) 2022-present Sigma-Soft, Ltd.
 * @author: Dmitry Molchanovsky
 */

package bbolt

import (
	"context"
	"sync"

	"github.com/voedger/voedger/pkg/goutils/timeu"
	"github.com/voedger/voedger/pkg/istorage"
)

func Provide(params ParamsType, iTime timeu.ITime) istorage.IAppStorageFactory {
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	return &appStorageFactory{
		bboltParams: params,
		iTime:       iTime,
		ctx:         ctx,
		cancel:      cancel,
		wg:          wg,
	}
}
