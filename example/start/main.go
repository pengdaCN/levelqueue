package main

import (
	"context"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/pengdacn/levelqueue/extutil"

	"github.com/pengdacn/levelqueue"
)

func main() {
	// 创建ldis
	ldis, err := levelqueue.Open("./tmp")
	if err != nil {
		panic(err)
	}

	// 创建可取消的context，用于关闭queue
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer ldis.Close()

	queue, err := levelqueue.NewSimpleQueue(
		"test-1",
		levelqueue.WithOwnLedis(
			// 使用自动compact，只有leveldb 需要
			extutil.AutoCompact(ldis),
		),
		levelqueue.WithContext(
			// 传入退出context
			ctx,
		),
	)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		// BPop
		defer wg.Done()

		for {
			bs := queue.BPop()
			if len(bs) == 0 {
				break
			}

			log.Println("BPop ->", string(bs))
		}

		log.Println("BPop exit")
	}()

	wg.Add(1)
	go func() {
		// GlobalPopCh
		defer wg.Done()

		for bs := range queue.GlobalPopCh() {
			log.Println("GlobalPopCh ->", string(bs))
		}

		log.Println("GlobalPopCh exit")
	}()

	wg.Add(1)
	go func() {
		// NewPopCh
		defer wg.Done()

		popChCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		go func() {
			time.Sleep(time.Second)
			cancel()
		}()

		for bs := range queue.NewPopCh(popChCtx) {
			log.Println("NewPopCh ->", string(bs))
		}

		log.Println("NewPopCh exit")
	}()

	for i := 0; i < 1000; i++ {
		err := queue.Push([]byte(strconv.Itoa(i + 1)))
		if err != nil {
			panic(err)
		}
		time.Sleep(time.Second / 100)
	}

	go func() {
		for {
			_len, err := queue.Len()
			if err != nil {
				panic(err)
			}

			if _len == 0 {
				break
			}

			time.Sleep(time.Second)
		}

		cancel()
	}()

	wg.Wait()
	_len, err := queue.Len()
	if err != nil {
		panic(err)
	}
	log.Println("Len ->", _len)

	log.Println("Exit")
}
