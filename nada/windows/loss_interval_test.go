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

			lossInt.addLoss()
			res = lossInt.avgLossInt()

			Expect(res).NotTo(Equal(float64(0)))
		})

		It("loss test with drop of old interval", func() {
			lossInt := newLossIntervall(2)

			for i := 0; i < 5; i++ {
				lossInt.addPacket()
			}

			Expect(lossInt.intervals).To(Equal([]uint64{}))

			lossInt.addLoss()
			Expect(lossInt.intervals).To(Equal([]uint64{1}))

			lossInt.addPacket()
			Expect(lossInt.intervals).To(Equal([]uint64{2}))

			lossInt.addLoss()
			Expect(lossInt.intervals).To(Equal([]uint64{1, 2}))

			for i := 0; i < 5; i++ {
				lossInt.addPacket()
			}

			lossInt.addLoss()
			Expect(lossInt.intervals).To(Equal([]uint64{1, 6})) // oldest interval dropped

			res := lossInt.avgLossInt()

			Expect(res).NotTo(Equal(float64(0)))
		})
	})
})
