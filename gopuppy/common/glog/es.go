package glog

import (
	"github.com/olivere/elastic"
	"github.com/olivere/elastic/config"
	"context"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap"
	"time"
	"fmt"
	"sync"
	"github.com/xiaonanln/go-xnsyncutil/xnsyncutil"
)

var (
	closeWaitGroup sync.WaitGroup
	isClose = false
	_ zapcore.WriteSyncer = &esWriter{}
	esConfig iEsConfig = nil

	bulkGuard sync.Mutex
	type2BulkProcessor = map[string]*esProcessor{}
)

const (
	infoLogType = "infolog"
)

func SetEsConfig(cfg iEsConfig) {
	esConfig = cfg
}

type iEsConfig interface {
	GetUrl() string
	GetIndex() string
}

type esLogger struct {
}

func (el *esLogger) Printf(format string, v ...interface{}) {
	errSugar.Errorf(format, v...)
}

type esProcessor struct {
	queue *xnsyncutil.SyncQueue
	closeNotify chan struct{}
	processor *elastic.BulkProcessor
}

func (ep *esProcessor) close() {
	if ep.queue != nil {
		ep.queue.Close()
	}

	go func() {
		<- ep.closeNotify
		err := ep.processor.Flush()
		if err != nil {
			fmt.Printf("esBulkProcessor Flush error %s\n", err)
		}
		ep.processor.Close()
		closeWaitGroup.Done()
	}()
}

func (ep *esProcessor) run(index, type_ string) {
	index = fmt.Sprintf("%s_%s", index, type_)

	go func() {
		defer func() {
			err := recover()
			if err != nil {
				errSugar.Errorf("esProcessor run panic: %s", err)
			}
		}()

		closeWaitGroup.Add(1)
		for {
			req := ep.queue.Pop()
			if req == nil {
				close(ep.closeNotify)
				return
			}

			req2, ok := req.(*elastic.BulkIndexRequest)
			if !ok {
				errSugar.With(zap.Time("ts", time.Now())).Error("esProcessor ep.queue.Pop not " +
					"BulkIndexRequest")
				continue
			}

			req2.Index(index)
			ep.processor.Add(req2)
		}

	}()
}

func newBulkProcessor(type_ string) *esProcessor {
	if esConfig == nil || esConfig.GetUrl() == "" || esConfig.GetIndex() == "" {
		return nil
	}

	//url := "http://kingwar_es:kingwaropenew@127.0.0.1:5003/king_war"
	//url := "http://134.175.13.251:9500?sniff=false"
	cfg, err := config.Parse(esConfig.GetUrl())
	if err != nil {
		errSugar.With(zap.Time("ts", time.Now())).Error("newBulkProcessor Parse config error %s", err)
		return nil
	}

	cli, err := elastic.NewClientFromConfig(cfg)
	if err != nil {
		errSugar.With(zap.Time("ts", time.Now())).Error("newBulkProcessor NewClientFromConfig error %s", err)
		return nil
	}

	elog := &esLogger{}
	elastic.SetErrorLog(elog)(cli)
	//elastic.SetInfoLog(elog)(cli)
	//elastic.SetTraceLog(elog)(cli)

	ps := elastic.NewBulkProcessorService(cli)
	ps.FlushInterval(time.Minute)
	p, err := ps.Do(context.Background())
	if err != nil {
		errSugar.With(zap.Time("ts", time.Now())).Error("newBulkProcessor NewBulkProcessorService error %s", err)
		return nil
	}

	processor := &esProcessor{
		queue: xnsyncutil.NewSyncQueue(),
		processor: p,
		closeNotify: make(chan struct{}),
	}
	processor.run(esConfig.GetIndex(), type_)
	return processor
}

type esWriter struct {
	type_ string
}

func (w *esWriter) Sync() error {
	return nil
}

func (w *esWriter) Write(p []byte) (n int, err error) {
	if esConfig == nil {
		return
	}

	processor, ok := type2BulkProcessor[w.type_]
	if !ok {
		bulkGuard.Lock()

		processor, ok = type2BulkProcessor[w.type_]
		if !ok {
			processor = newBulkProcessor(w.type_)
			if processor != nil {
				type2BulkProcessor[w.type_] = processor
			}
		}

		bulkGuard.Unlock()
	}

	if processor == nil {
		errSugar.Error("es onLog type=%s, message=%s", w.type_, string(p))
		return 0, nil
	}

	req := elastic.NewBulkIndexRequest()
	req.Type(w.type_)
	req.Doc(string(p))

	//errSugar.Infof("es onLog type=%s, message=%s", w.type_, p)

	if !isClose {
		processor.queue.Push(req)
	} else {
		errSugar.Error("es onLog type=%s, message=%s", w.type_, string(p))
	}

	return len(p), nil
}

func closeEs() {
	isClose = true
	for _, processor := range type2BulkProcessor {
		processor.close()
	}

	closeWaitGroup.Wait()
}
