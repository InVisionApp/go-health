package health

import (
	"sync"
	"testing"

	. "github.com/onsi/gomega"
)

func Test_sBool_String(t *testing.T) {
	t.Parallel()
	RegisterTestingT(t)

	t.Run("Happy path", func(t *testing.T) {
		t.Parallel()
		b := &sBool{}

		b.setFalse()
		Expect(b.String()).To(Equal("false"))
		Expect(b.val()).To(BeFalse())

		b.setTrue()
		Expect(b.String()).To(Equal("true"))
		Expect(b.val()).To(BeTrue())
	})
}

func Test_sBool_Race(t *testing.T) {
	t.Parallel()

	b := &sBool{}
	wg := &sync.WaitGroup{}
	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		go func(b *sBool, wg *sync.WaitGroup) {
			b.setTrue()
			b.setFalse()
			wg.Done()
		} (b, wg)
	}
	wg.Wait()
}
