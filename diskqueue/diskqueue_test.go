package diskqueue

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/pengswift/libonepiece/test"
)

func TestDiskQueue(t *testing.T) {
	dqName := "test_disk_queue" + strconv.Itoa(int(time.Now().Unix()))
	tmpDir, err := ioutil.TempDir("", fmt.Sprintf("diskqueue-test-%d", time.Now().UnixNano()))
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)
	dq := newDiskQueue(dqName, tmpDir, 1024, 4, 1<<10, 2500, 2*time.Second)
	defer dq.Close()
	NotEqual(t, dq, nil)
	Equal(t, dq.Depth(), int64(0))

	msg := []byte("test")
	err = dq.Put(msg)
	Equal(t, err, nil)
	Equal(t, dq.Depth(), int64(1))

	msgOut := <-dq.ReadChan()
	Equal(t, msgOut, msg)
}

func TestDiskQueueRoll(t *testing.T) {
	dqName := "test_disk_queue_roll" + strconv.Itoa(int(time.Now().Unix()))
	tmpDir, err := ioutil.TempDir("", fmt.Sprintf("diskqueue-test-%d", time.Now().UnixNano()))
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)
	msg := bytes.Repeat([]byte{0}, 10)
	ml := int64(len(msg))
	dq := newDiskQueue(dqName, tmpDir, 9*(ml+4), int32(ml), 1<<10, 2500, 2*time.Second)
	defer dq.Close()
	NotEqual(t, dq, nil)
	Equal(t, dq.Depth(), int64(0))

	for i := 0; i < 10; i++ {
		err := dq.Put(msg)
		Equal(t, err, nil)
		Equal(t, dq.Depth(), int64(i+1))
	}

	Equal(t, dq.(*diskQueue).writeFileNum, int64(1))
	Equal(t, dq.(*diskQueue).writePos, int64(0))
}

func assertFileNotExist(t *testing.T, fn string) {
	f, err := os.OpenFile(fn, os.O_RDONLY, 0600)
	Equal(t, f, (*os.File)(nil))
	Equal(t, os.IsNotExist(err), true)
}

func TestDiskQueueEmpty(t *testing.T) {
	dqName := "test_disk_queue_empty" + strconv.Itoa(int(time.Now().Unix()))
	tmpDir, err := ioutil.TempDir("", fmt.Sprintf("diskqueue-test-%d", time.Now().UnixNano()))
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)
	msg := bytes.Repeat([]byte{0}, 10)
	dq := newDiskQueue(dqName, tmpDir, 100, 0, 1<<10, 2500, 2*time.Second)
	defer dq.Close()
	NotEqual(t, dq, nil)
	Equal(t, dq.Depth(), int64(0))

	for i := 0; i < 100; i++ {
		err := dq.Put(msg)
		Equal(t, err, nil)
		Equal(t, dq.Depth(), int64(i+1))
	}

	for i := 0; i < 3; i++ {
		<-dq.ReadChan()
	}

	for {
		if dq.Depth() == 97 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	Equal(t, dq.Depth(), int64(97))

	numFiles := dq.(*diskQueue).writeFileNum
	dq.Empty()

	assertFileNotExist(t, dq.(*diskQueue).metaDataFileName())
	for i := int64(0); i <= numFiles; i++ {
		assertFileNotExist(t, dq.(*diskQueue).fileName(i))
	}
	Equal(t, dq.Depth(), int64(0))
	Equal(t, dq.(*diskQueue).readFileNum, dq.(*diskQueue).writeFileNum)
	Equal(t, dq.(*diskQueue).readPos, dq.(*diskQueue).writePos)
	Equal(t, dq.(*diskQueue).nextReadPos, dq.(*diskQueue).readPos)
	Equal(t, dq.(*diskQueue).nextReadFileNum, dq.(*diskQueue).readFileNum)

	for i := 0; i < 100; i++ {
		err := dq.Put(msg)
		Equal(t, err, nil)
		Equal(t, dq.Depth(), int64(i+1))
	}

	for i := 0; i < 100; i++ {
		<-dq.ReadChan()
	}

	for {
		if dq.Depth() == 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	Equal(t, dq.Depth(), int64(0))
	Equal(t, dq.(*diskQueue).readFileNum, dq.(*diskQueue).writeFileNum)
	Equal(t, dq.(*diskQueue).readPos, dq.(*diskQueue).writePos)
	Equal(t, dq.(*diskQueue).nextReadPos, dq.(*diskQueue).readPos)
}

func TestDiskQueueCorruption(t *testing.T) {
	dqName := "test_disk_queue_corruption" + strconv.Itoa(int(time.Now().Unix()))
	tmpDir, err := ioutil.TempDir("", fmt.Sprintf("diskqueue-test-%d", time.Now().UnixNano()))
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	dq := newDiskQueue(dqName, tmpDir, 1000, 10, 1<<10, 5, 2*time.Second)
	defer dq.Close()

	//消息长度 123 + 4 字节head = 127
	// 每个文件最多可存放8条消息
	msg := make([]byte, 123) // 127 bytes per message, 8 (1016 bytes) message perfile
	for i := 0; i < 25; i++ {
		dq.Put(msg)
	}

	// 存放25条消息， 占用4个文件 8, 8, 8, 1
	Equal(t, dq.Depth(), int64(25))

	// corrupt the 2nd file
	// 第2个文件长度设置为500, 第二个文件后5个文件丢失,  目前总的可用消息数量为20
	dqFn := dq.(*diskQueue).fileName(1)
	// 跟新文件长度
	os.Truncate(dqFn, 500) // 3 valid messages, 5 corrupted

	// 8 + 3 + 8
	// 取19个消息，第4个文件耽搁消息暂不取
	for i := 0; i < 19; i++ {
		Equal(t, <-dq.ReadChan(), msg) // 1 message leftover in 4th file
	}

	// corrupt the 4th (current) file
	// 设置第四个文件长度为100, 读消息出错，会跳到下一个文件
	// 如果此事设置 time.sleep 超过一定时长， 会先读出数据，单不会报错
	dqFn = dq.(*diskQueue).fileName(3)
	os.Truncate(dqFn, 100)

	// 第四个文件损坏了，直接跳到第5个文件
	dq.Put(msg) // in 5th file

	Equal(t, <-dq.ReadChan(), msg)

	// 写入0长度到文件5
	dq.(*diskQueue).writeFile.Write([]byte{0, 0, 0, 0})

	// 写数据到文件5
	dq.Put(msg)
	// 写到文件6
	dq.Put(msg)

	Equal(t, <-dq.ReadChan(), msg)
}

func TestDiskQueueTorture(t *testing.T) {
	var wg sync.WaitGroup
	dqName := "test_disk_queue_torture" + strconv.Itoa(int(time.Now().Unix()))
	tmpDir, err := ioutil.TempDir("", fmt.Sprintf("diskqueue-test-%d", time.Now().UnixNano()))
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)
	dq := newDiskQueue(dqName, tmpDir, 262144, 0, 1<<10, 2500, 2*time.Second)
	NotEqual(t, dq, nil)
	Equal(t, dq.Depth(), int64(0))

	msg := []byte("aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffffffffff")

	numWriters := 4
	numReaders := 4
	readExitChan := make(chan int)
	writeExitChan := make(chan int)

	var depth int64
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				time.Sleep(100000 * time.Nanosecond)
				select {
				case <-writeExitChan:
					return
				default:
					err := dq.Put(msg)
					if err == nil {
						atomic.AddInt64(&depth, 1)
					}
				}
			}
		}()
	}

	time.Sleep(1 * time.Second)

	dq.Close()

	t.Logf("closing writeExitChan")
	close(writeExitChan)
	wg.Wait()

	t.Logf("restarting diskqueue")

	dq = newDiskQueue(dqName, tmpDir, 262144, 0, 1<<10, 2500, 2*time.Second)
	defer dq.Close()
	NotEqual(t, dq, nil)

	var read int64
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				time.Sleep(100000 * time.Nanosecond)
				select {
				case m := <-dq.ReadChan():
					Equal(t, msg, m)
					atomic.AddInt64(&read, 1)
				case <-readExitChan:
					return
				}
			}
		}()
	}

	t.Logf("waiting for depth 0")
	for {
		if dq.Depth() == 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	t.Logf("closing readExitChan")
	close(readExitChan)
	wg.Wait()

	Equal(t, read, depth)
}

func BenchmarkDiskQueuePut16(b *testing.B) {
	benchmarkDiskQueuePut(16, b)
}
func BenchmarkDiskQueuePut64(b *testing.B) {
	benchmarkDiskQueuePut(64, b)
}
func BenchmarkDiskQueuePut256(b *testing.B) {
	benchmarkDiskQueuePut(256, b)
}
func BenchmarkDiskQueuePut1024(b *testing.B) {
	benchmarkDiskQueuePut(1024, b)
}
func BenchmarkDiskQueuePut4096(b *testing.B) {
	benchmarkDiskQueuePut(4096, b)
}
func BenchmarkDiskQueuePut16384(b *testing.B) {
	benchmarkDiskQueuePut(16384, b)
}
func BenchmarkDiskQueuePut65536(b *testing.B) {
	benchmarkDiskQueuePut(65536, b)
}
func BenchmarkDiskQueuePut262144(b *testing.B) {
	benchmarkDiskQueuePut(262144, b)
}
func BenchmarkDiskQueuePut1048576(b *testing.B) {
	benchmarkDiskQueuePut(1048576, b)
}

func benchmarkDiskQueuePut(size int64, b *testing.B) {
	b.StopTimer()
	dqName := "bench_disk_queue_put" + strconv.Itoa(b.N) + strconv.Itoa(int(time.Now().Unix()))
	tmpDir, err := ioutil.TempDir("", fmt.Sprintf("diskqueue-test-%d", time.Now().UnixNano()))
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)
	dq := newDiskQueue(dqName, tmpDir, 1024768*100, 0, 1<<20, 2500, 2*time.Second)
	defer dq.Close()
	b.SetBytes(size)
	data := make([]byte, size)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		err := dq.Put(data)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkDiskWrite16(b *testing.B) {
	benchmarkDiskWrite(16, b)
}
func BenchmarkDiskWrite64(b *testing.B) {
	benchmarkDiskWrite(64, b)
}
func BenchmarkDiskWrite256(b *testing.B) {
	benchmarkDiskWrite(256, b)
}
func BenchmarkDiskWrite1024(b *testing.B) {
	benchmarkDiskWrite(1024, b)
}
func BenchmarkDiskWrite4096(b *testing.B) {
	benchmarkDiskWrite(4096, b)
}
func BenchmarkDiskWrite16384(b *testing.B) {
	benchmarkDiskWrite(16384, b)
}
func BenchmarkDiskWrite65536(b *testing.B) {
	benchmarkDiskWrite(65536, b)
}
func BenchmarkDiskWrite262144(b *testing.B) {
	benchmarkDiskWrite(262144, b)
}
func BenchmarkDiskWrite1048576(b *testing.B) {
	benchmarkDiskWrite(1048576, b)
}

func benchmarkDiskWrite(size int64, b *testing.B) {
	b.StopTimer()
	fileName := "bench_disk_queue_put" + strconv.Itoa(b.N) + strconv.Itoa(int(time.Now().Unix()))
	tmpDir, err := ioutil.TempDir("", fmt.Sprintf("diskqueue-test-%d", time.Now().UnixNano()))
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)
	f, _ := os.OpenFile(path.Join(tmpDir, fileName), os.O_RDWR|os.O_CREATE, 0600)
	b.SetBytes(size)
	data := make([]byte, size)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		f.Write(data)
	}
	f.Sync()
}

func BenchmarkDiskWriteBuffered16(b *testing.B) {
	benchmarkDiskWriteBuffered(16, b)
}
func BenchmarkDiskWriteBuffered64(b *testing.B) {
	benchmarkDiskWriteBuffered(64, b)
}
func BenchmarkDiskWriteBuffered256(b *testing.B) {
	benchmarkDiskWriteBuffered(256, b)
}
func BenchmarkDiskWriteBuffered1024(b *testing.B) {
	benchmarkDiskWriteBuffered(1024, b)
}
func BenchmarkDiskWriteBuffered4096(b *testing.B) {
	benchmarkDiskWriteBuffered(4096, b)
}
func BenchmarkDiskWriteBuffered16384(b *testing.B) {
	benchmarkDiskWriteBuffered(16384, b)
}
func BenchmarkDiskWriteBuffered65536(b *testing.B) {
	benchmarkDiskWriteBuffered(65536, b)
}
func BenchmarkDiskWriteBuffered262144(b *testing.B) {
	benchmarkDiskWriteBuffered(262144, b)
}
func BenchmarkDiskWriteBuffered1048576(b *testing.B) {
	benchmarkDiskWriteBuffered(1048576, b)
}

func benchmarkDiskWriteBuffered(size int64, b *testing.B) {
	b.StopTimer()
	fileName := "bench_disk_queue_put" + strconv.Itoa(b.N) + strconv.Itoa(int(time.Now().Unix()))
	tmpDir, err := ioutil.TempDir("", fmt.Sprintf("diskqueue-test-%d", time.Now().UnixNano()))
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)
	f, _ := os.OpenFile(path.Join(tmpDir, fileName), os.O_RDWR|os.O_CREATE, 0600)
	b.SetBytes(size)
	data := make([]byte, size)
	w := bufio.NewWriterSize(f, 1024*4)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		w.Write(data)
		if i%1024 == 0 {
			w.Flush()
		}
	}
	w.Flush()
	f.Sync()
}

func BenchmarkDiskQueueGet16(b *testing.B) {
	benchmarkDiskQueueGet(16, b)
}
func BenchmarkDiskQueueGet64(b *testing.B) {
	benchmarkDiskQueueGet(64, b)
}
func BenchmarkDiskQueueGet256(b *testing.B) {
	benchmarkDiskQueueGet(256, b)
}
func BenchmarkDiskQueueGet1024(b *testing.B) {
	benchmarkDiskQueueGet(1024, b)
}
func BenchmarkDiskQueueGet4096(b *testing.B) {
	benchmarkDiskQueueGet(4096, b)
}
func BenchmarkDiskQueueGet16384(b *testing.B) {
	benchmarkDiskQueueGet(16384, b)
}
func BenchmarkDiskQueueGet65536(b *testing.B) {
	benchmarkDiskQueueGet(65536, b)
}
func BenchmarkDiskQueueGet262144(b *testing.B) {
	benchmarkDiskQueueGet(262144, b)
}
func BenchmarkDiskQueueGet1048576(b *testing.B) {
	benchmarkDiskQueueGet(1048576, b)
}

func benchmarkDiskQueueGet(size int64, b *testing.B) {
	b.StopTimer()
	dqName := "bench_disk_queue_get" + strconv.Itoa(b.N) + strconv.Itoa(int(time.Now().Unix()))
	tmpDir, err := ioutil.TempDir("", fmt.Sprintf("diskqueue-test-%d", time.Now().UnixNano()))
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)
	dq := newDiskQueue(dqName, tmpDir, 1024768, 0, 1<<10, 2500, 2*time.Second)
	defer dq.Close()
	b.SetBytes(size)
	data := make([]byte, size)
	for i := 0; i < b.N; i++ {
		dq.Put(data)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		<-dq.ReadChan()
	}
}
