package health

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestString(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Happy path", func(t *testing.T) {
		b := newBool()

		// should be false by default
		Expect(b.v).To(BeFalse())

		// Mutex should be created
		Expect(b.mu).ToNot(BeNil())

		b.setFalse()
		Expect(b.String()).To(Equal("false"))
		Expect(b.val()).To(BeFalse())

		b.setTrue()
		Expect(b.String()).To(Equal("true"))
		Expect(b.val()).To(BeTrue())
	})
}
