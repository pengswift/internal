/*
	UUID  int64 ( 41bit timestamp + 10 bit machine-id + 12bit sn )
	毫秒级时间41位+机器ID 10位+内序列12位

*/
package guid

import (
	"errors"
	"time"
)

const (
	workerIDBits   = uint64(10)
	sequeuceBits   = uint64(12)
	workerIDShift  = sequeuceBits
	timestampShift = sequeuceBits + workerIDBits
	sequenceMask   = int64(-1) ^ (int64(-1) << sequeuceBits)

	// ( 2012-10-28 16:23:42 UTC ).UnixNano() >> 20
	twepoch = int64(1288834974288)
)

var ErrTimeBackwards = errors.New("time has gone backwards")
var ErrSequenceExpired = errors.New("sequence expired")
var ErrIDBackwards = errors.New("ID went backward")

type guid int64

type guidFactory struct {
	sequence      int64
	lastTimestamp int64
	lastID        guid
}

func (f *guidFactory) NewGUID(workerID int64) (guid, error) {
	ts := time.Now().UnixNano() >> 20

	if ts < f.lastTimestamp {
		return 0, ErrTimeBackwards
	}

	if f.lastTimestamp == ts {
		f.sequence = (f.sequence + 1) & sequenceMask
		if f.sequence == 0 {
			return 0, ErrSequenceExpired
		}
	} else {
		f.sequence = 0
	}

	f.lastTimestamp = ts

	id := guid(((ts - twepoch) << timestampShift) |
		(workerID << workerIDShift) |
		f.sequence)

	if id <= f.lastID {
		return 0, ErrIDBackwards
	}

	f.lastID = id

	return id, nil
}
