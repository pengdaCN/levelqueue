package levelqueue

import (
	"context"
	"encoding/binary"
	lediscfg "github.com/ledisdb/ledisdb/config"
	"github.com/ledisdb/ledisdb/server"
	"log"
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"
	"unsafe"
)

func TestPopCh(t *testing.T) {
	ldis, err := Open("./tmp")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		cancel()
		ldis.Close()
	}()

	queue, err := NewSimpleQueue("test-1", WithOwnLedis(ldis), WithContext(ctx))
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		ch := queue.GlobalPopCh()
		for bs := range ch {
			log.Println(string(bs))
			//_len, err := queue.Len()
			//if err != nil {
			//	panic(err)
			//}
			//
			//log.Println("len", _len)
			//time.Sleep(time.Second / 10)
		}

		log.Println("OK")
	}()

	for i := 0; i < 100; i++ {
		if err := queue.Push([]byte(strconv.Itoa(i + 1))); err != nil {
			t.Fatal(err)
		}
	}

	time.Sleep(time.Second * 2)
	cancel()

	wg.Wait()

	t.Log("OK")
}

func TestSimpleQueue(t *testing.T) {
	ldis, err := Open("./tmp")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		cancel()
		ldis.Close()
	}()

	queue, err := NewSimpleQueue("test-1", WithOwnLedis(ldis), WithContext(ctx))
	if err != nil {
		t.Fatal(err)
	}

	const (
		times = 10000000
	)

	var wg sync.WaitGroup
	wg.Add(2)
	now := time.Now()
	go func() {
		defer wg.Done()

		for i := 0; i < times; i++ {
			if err := queue.Push(genValue(i+1, 1024)); err != nil {
				panic(err)
			}
		}

		log.Println("写耗时", time.Since(now))
	}()

	go func() {
		defer wg.Done()

		var (
			lastSeq   int
			readCount int
		)

		for {
			bs, err := queue.PopWithTimeout(time.Second)
			if err != nil {
				panic(err)
			}

			if len(bs) == 0 {
				break
			}

			seq := getSeq(bs)
			if lastSeq+1 != seq {
				log.Panicf("序列循序错误 本次: %d 上次: %d", seq, lastSeq)
			}

			lastSeq = seq
			readCount++
		}

		if readCount != times {
			log.Panicf("数据没读取完 读取 %d 全部 %d", readCount, times)
		}

		log.Println("读耗时", time.Since(now))

	}()

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				var m runtime.MemStats
				runtime.ReadMemStats(&m)

				log.Println(m.Sys)
			}
		}
	}()

	wg.Wait()

	t.Log("耗时", time.Since(now))
}

// 64 w 1.30m r 2.15m mem 100mb data 1000 0000
// 1024 w 2.6m r 2.6m mem 56mb data 1000 0000
func TestRun1(t *testing.T) {
	const (
		times = 10000000
	)

	l, err := Open("./tmp/queue1")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	db, err := l.Select(0)
	if err != nil {
		t.Fatal(err)
	}

	key := []byte("test-log")
	now := time.Now()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()

		for i := 0; i < times; i++ {
			if _, err := db.RPush(key, genValue(i+1, 1024)); err != nil {
				panic(err)
			}
		}

		log.Println("写耗时", time.Since(now))
	}()

	go func() {
		defer wg.Done()

		var (
			lastSeq   int
			readCount int
		)

		for {
			val, err := db.BLPop([][]byte{key}, time.Second)
			if err != nil {
				log.Println(err)
				break
			}

			if len(val) == 0 {
				break
			}

			var bs []byte
			{
				t1, ok := val[1].([]any)
				if !ok {
					panic("类型错误")
				}

				t2, ok := t1[1].([]byte)
				if !ok {
					panic("类型错误")
				}

				bs = t2
			}

			//bs := val[1].([]interface{})[1].([]interface{})[1].([]byte)
			seq := getSeq(bs)
			if lastSeq+1 != seq {
				log.Panicf("序列循序错误 本次: %d 上次: %d", seq, lastSeq)
			}

			lastSeq = seq
			readCount++
		}

		if readCount != times {
			log.Panicf("数据没读取完 读取 %d 全部 %d", readCount, times)
		}

		log.Println("读耗时", time.Since(now))

	}()

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				var m runtime.MemStats
				runtime.ReadMemStats(&m)

				log.Println(m.Sys)
			}
		}
	}()

	wg.Wait()

	t.Log("耗时", time.Since(now))
}

func TestRun2(t *testing.T) {
	cfg := lediscfg.NewConfigDefault()
	cfg.DataDir = "./tmp/queue1"

	l, err := server.NewApp(cfg)
	if err != nil {
		t.Fatal(err)
	}

	l.Run()

	defer l.Close()

	select {}
}

func TestGenKey(t *testing.T) {
	val := genValue(12, 64)
	t.Log(getSeq(val))
	t.Log(string(val[8:]))
}

func genValue(seq int, size int) []byte {
	bs := make([]byte, 8+56)
	binary.LittleEndian.PutUint64(bs[:8], uint64(seq))

	copy(bs[8:], randStr(size-8))

	return bs
}

func getSeq(bs []byte) int {
	return int(binary.LittleEndian.Uint64(bs[:8]))
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var src = rand.New(rand.NewSource(time.Now().UnixNano()))

const (
	// 6 bits to represent a letter index
	letterIdBits = 6
	// All 1-bits as many as letterIdBits
	letterIdMask = 1<<letterIdBits - 1
	letterIdMax  = 63 / letterIdBits
)

func randStr(n int) string {
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdMax letters!
	for i, cache, remain := n-1, src.Int63(), letterIdMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdMax
		}
		if idx := int(cache & letterIdMask); idx < len(letters) {
			b[i] = letters[idx]
			i--
		}
		cache >>= letterIdBits
		remain--
	}
	return *(*string)(unsafe.Pointer(&b))
}

func TestFalseSharing(t *testing.T) {
	const (
		times = 1000000000
	)

	var arr [20]int64
	now := time.Now()
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < times; i++ {
			arr[0]++
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < times; i++ {
			arr[1]++
		}
	}()

	wg.Wait()

	t.Log(time.Since(now))
}

func TestWrite(t *testing.T) {
	ch := make(chan int)

	go func() {
		for {
			select {
			case ch <- 1:
			default:

			}
			time.Sleep(time.Second)
		}
	}()

	time.Sleep(time.Second * 5)
	_ = <-ch
	close(ch)

	time.Sleep(time.Second * 10)
}
