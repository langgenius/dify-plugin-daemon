package routine

import (
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/panjf2000/ants"
)

var (
	p *ants.Pool
)

func InitPool(size int) {
	log.Info("init routine pool, size: %d", size)
	p, _ = ants.NewPool(size, ants.WithNonblocking(false))
}

func Submit(f func()) {
	p.Submit(f)
}
