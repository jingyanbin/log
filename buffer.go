package log

import (
	"sync"
)

const minBufferSize = 128
const bufferPoolNumber = 20

type Buffer []byte

func (my *Buffer) AppendBytes(bs ...byte) {
	*my = append(*my, bs...)
}

func (my *Buffer) AppendString(s string) {
	*my = append(*my, s...)
}

func (my *Buffer) AppendStrings(ss... string) {
	for _, s := range ss{
		*my = append(*my, s...)
	}
}

func (my *Buffer) AppendInt(n, w int) {
	itoa((*[]byte)(my), n, w)
}

func (my *Buffer) Bytes()[]byte{
	return *my
}

func (my *Buffer) free(){
	*my = (*my)[:0]
	size := cap(*my)
	index := size / minBufferSize
	if index >= 0 && index < bufferPoolNumber{
		bufferPools[index].Put(my)
	}
}

func newBuffer(size int) *Buffer {
	buf := make(Buffer, 0, size)
	return &buf
}





type logBufferPools [bufferPoolNumber]sync.Pool

func (my *logBufferPools) init(index int) {
	my[index].New = func() interface{} {
		if index > 0{
			return newBuffer(index * minBufferSize)
		} else{
			return newBuffer(minBufferSize / 2)
		}
	}
}

func (my *logBufferPools)Init(){
	for i := 0 ; i < len(my); i++{
		my.init(i)
	}
}

var bufferPools logBufferPools


func getBuffer(size int) *Buffer{
	index := size / minBufferSize
	if index < 0 || index >= bufferPoolNumber{
		return newBuffer(size)
	} else{
		return bufferPools[index].Get().(*Buffer)
	}
}


func init() {
	bufferPools.Init()
}