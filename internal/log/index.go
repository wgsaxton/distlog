package log

import (
	"fmt"
	"io"
	"os"

	"github.com/tysonmote/gommap"
	"github.com/wgsaxton/distlog/internal/common"
)

var (
	offWidth uint64 = 4
	posWidth uint64 = 8
	entWidth        = offWidth + posWidth
)

type index struct {
	file *os.File
	mmap gommap.MMap
	size uint64
}

func newIndex(f *os.File, c Config) (*index, error) {
	idx := &index{
		file: f,
	}
	fi, err := os.Stat(f.Name())
	if err != nil {
		common.Gslog.Println("Error given here. err:", err)
		return nil, err
	}
	fmt.Printf("index.go.newIndex() f.Name() value: %+v\n", f.Name())
	fmt.Println("index.go.newIndex() c.Segment.MaxIndexBytes", c.Segment.MaxIndexBytes)
	idx.size = uint64(fi.Size())
	if err = os.Truncate(
		f.Name(), int64(c.Segment.MaxIndexBytes),
	); err != nil {
		common.Gslog.Println("Error given here. err:", err)
		return nil, err
	}
	fmt.Printf("index.go.newIndex() idx struct: %+v\n", idx)
	fmt.Println("idx.file.Fd():", idx.file.Fd(), "and name", idx.file.Name())
	if idx.mmap, err = gommap.Map(
		idx.file.Fd(),
		gommap.PROT_READ|gommap.PROT_WRITE,
		gommap.MAP_PRIVATE,
	); err != nil {
		fmt.Printf("In Error, index.go.newIndex() idx struct: %+v\n", idx)
		common.Gslog.Println("Error given here. err:", err)
		return nil, err
	}
	fmt.Printf("index.go.newIndex() No errors idx struct: %+v\n", idx)
	fmt.Println("log.go.newIndex() returned with no error")
	return idx, nil
}

func (i *index) Read(in int64) (out uint32, pos uint64, err error) {
	if i.size == 0 {
		// fmt.Println("first eof")
		return 0, 0, io.EOF
	}
	if in == -1 {
		out = uint32((i.size / entWidth) - 1)
	} else {
		out = uint32(in)
	}
	pos = uint64(out) * entWidth
	if i.size < pos+entWidth {
		fmt.Println(i.size, pos+entWidth)
		fmt.Println("2nd eof")
		return 0, 0, io.EOF
	}
	out = enc.Uint32(i.mmap[pos : pos+offWidth])
	pos = enc.Uint64(i.mmap[pos+offWidth : pos+entWidth])
	return out, pos, nil
}

func (i *index) Write(off uint32, pos uint64) error {
	if uint64(len(i.mmap)) < i.size+entWidth {
		return io.EOF
	}
	enc.PutUint32(i.mmap[i.size:i.size+offWidth], off)
	enc.PutUint64(i.mmap[i.size+offWidth:i.size+entWidth], pos)
	i.size += uint64(entWidth)
	return nil
}

func (i *index) Close() error {
	if err := i.mmap.Sync(gommap.MS_SYNC); err != nil {
		return err
	}
	if err := i.file.Sync(); err != nil {
		return err
	}
	if err := i.file.Truncate(int64(i.size)); err != nil {
		return err
	}
	return i.file.Close()
}

func (i *index) Name() string {
	return i.file.Name()
}
