package windows

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Loss Intervall", func() {
	Context("test", func() {
		It("simple loss int", func() {
			lossInt := newLossIntervall(2)

			for i := 0; i < 5; i++ {
				lossInt.addPacket()
			}

			res := lossInt.avgLossInt()
			Expect(res).To(Equal(float64(0)))

			lossInt.addLoss(2)
			res = lossInt.avgLossInt()

			Expect(res).NotTo(Equal(float64(0)))
		})

		It("loss test with drop of old interval", func() {
			lossInt := newLossIntervall(2)

			for i := 0; i < 5; i++ {
				lossInt.addPacket()
			}

			Expect(lossInt.intervals).To(Equal([]uint64{}))

			lossInt.addLoss(2)
			Expect(lossInt.intervals).To(Equal([]uint64{2}))

			lossInt.addLoss(32)
			Expect(lossInt.intervals).To(Equal([]uint64{32, 2}))

			for i := 0; i < 5; i++ {
				lossInt.addPacket()
			}

			lossInt.addLoss(3)
			Expect(lossInt.intervals).To(Equal([]uint64{3, 37}))

			res := lossInt.avgLossInt()

			Expect(res).NotTo(Equal(float64(0)))
		})
	})
})
