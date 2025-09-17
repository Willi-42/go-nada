package windows

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Delay window test", func() {
	Context("test", func() {
		It("simple test", func() {
			delayWin := NewDelayWindow(5)

			for i := 0; i < 4; i++ {
				delayWin.AddSample(100)
			}

			res := delayWin.MinDelay()
			Expect(res).To(Equal(uint64(100)))

			delayWin.AddSample(50)
			res = delayWin.MinDelay()
			Expect(res).To(Equal(uint64(50)))

			for i := 0; i < 4; i++ {
				delayWin.AddSample(200)
			}

			res = delayWin.MinDelay()
			Expect(res).To(Equal(uint64(50)))

			delayWin.AddSample(300)
			res = delayWin.MinDelay()
			Expect(res).To(Equal(uint64(200)))
		})
	})
})
