package mmpool

import (
	"sort"
	"sync"
)

type Pool struct {
	pools pools
}
type pools []*ByteBuffer

func (p pools) Len() int           { return len(p) }
func (p pools) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p pools) Less(i, j int) bool { return p[i].Size < p[j].Size }

type PoolOpt struct {
	Size int64
}
type ByteBuffer struct {
	l    *sync.Mutex
	Size int64
	pool *sync.Pool
	buf  *[]byte
}

func (buf *ByteBuffer) GetBuf() *[]byte {
	return buf.buf
}
func NewPool(pools ...*PoolOpt) *Pool {
	res := &Pool{
		pools: make([]*ByteBuffer, 0),
	}
	for _, p := range pools {
		buffer := &ByteBuffer{
			pool: &sync.Pool{
				New: func() interface{} {
					res := make([]byte, p.Size)
					return &res
				},
			},
			Size: p.Size,
			l:    &sync.Mutex{},
		}
		res.pools = append(res.pools, buffer)
	}
	sort.Sort(res.pools)
	return res
}

func (p *Pool) Get(size int64) *ByteBuffer {
	return p.getLowerBoundPool(size)

}

func (p *Pool) Put(buf *ByteBuffer) {
	if buf == nil {
		return
	}
	buf.put()
}

func (p *Pool) getLowerBoundPool(size int64) *ByteBuffer {
	for _, pool := range p.pools {
		if pool.Size >= size {
			return pool
		}
	}
	return p.pools[p.pools.Len()-1]
}
func (buf *ByteBuffer) put() {
	buf.l.Lock()
	defer buf.l.Unlock()
	if buf.buf == nil {
		return
	}
	buf.pool.Put(buf.buf)
	buf.buf = nil
}

func (buf *ByteBuffer) Get() *[]byte {
	if buf.buf == nil {
		buf.l.Lock()
		defer buf.l.Unlock()
		if buf.buf == nil {
			buf.buf = buf.pool.Get().(*[]byte)
		}
	}
	return buf.buf
}
