package log

import (
	"io"
	"os"

	"github.com/tysonmote/gommap"
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
		return nil, err
	}

   // Actual size of the file be for increasing it to MaxIndexBYtes
	idx.size = uint64(fi.Size())

   // growing the size to the max size of an index, 
   // done before memory mapping, due to the fact that
   // we can't increase the mapped memory
	if err = os.Truncate(f.Name(), int64(c.Segment.MaxIndexBytes)); err != nil {
		return nil, err
	}

   // Mapping memory
   // we allow reading (PROT_READ) and writing (PROT_WRITE)
   // commit changes to file (MAP_SHARED)
   // for more infos see C Libraries mmap
   //
   // https://codebrowser.dev/glibc/glibc/sysdeps/unix/sysv/linux/bits/mman-linux.h.html
	if idx.mmap, err = gommap.Map(
		idx.file.Fd(), gommap.PROT_READ|gommap.PROT_WRITE,
		gommap.MAP_SHARED,
	); err != nil {
		return nil, err
	}
	return idx, nil
}

func (i *index) Close() error {
   // Flush mapped memory changes back to file
   // this can still mean that the content of the file
   // is not written to the disk
	if err := i.mmap.Sync(gommap.MS_SYNC); err != nil {
		return err
	}

   // Write File changes to disk(stable storage)
	if err := i.file.Sync(); err != nil {
		return nil
	}
   // Truncate the file back to the size of its content
	if err := i.file.Truncate(int64(i.size)); err != nil {
		return err
	}
	return i.file.Close()
}

func (i *index) Read(in int64) (out uint32, pos uint64, err error) {
   // no data exists
	if i.size == 0 {
		return 0, 0, io.EOF
	}

   // last entry
	if in == -1 {
		out = uint32((i.size / entWidth) - 1)
	} else {
		out = uint32(in)
	}

   // actual position
	pos = uint64(out) * entWidth
	if i.size < pos+entWidth {
		return 0, 0, io.EOF
	}

	out = enc.Uint32(i.mmap[pos : pos+offWidth])

	pos = enc.Uint64(i.mmap[pos+offWidth : pos+entWidth])
	return out, pos, nil
}

func (i *index) Write(off uint32, pos uint64) error {

   // check index has enough space
	if uint64(len(i.mmap)) < i.size+entWidth {
		return io.EOF
	}

   // write encoded off to memory slice
   // memory slice current last entry to the offset width
	enc.PutUint32(i.mmap[i.size:i.size+offWidth], off)

   // write encoded pos to memory slice
   // memory slice directly behind the off offset slice till the end
	enc.PutUint64(i.mmap[i.size+offWidth:i.size+entWidth], pos)

   // increase file size by written bytes
	i.size += uint64(entWidth)
	return nil
}

func (i *index) Name() string {
	return i.file.Name()
}
