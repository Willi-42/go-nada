package nada

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LogWin", func() {
	Context("test", func() {
		It("1 packets no update", func() {
			logWin := NewLogWinQueue(10)
			logWin.NewMediaPacketRecieved(0, 0, 12, false, false)
			Expect(logWin.lastPn).To(Equal(uint64(0)))

			Expect(logWin.numberPacketArrived).To(Equal(uint64(1)))
			Expect(logWin.numberMarkedPackets).To(Equal(uint64(0)))
			Expect(logWin.accumulatedSize).To(Equal(uint64(12)))
			Expect(logWin.numberLostPackets).To(BeZero())
			Expect(logWin.numberQueueBuildup).To(BeZero())

			Expect(logWin.numberPacketsSinceLoss).To(Equal(uint64(1)))
			Expect(logWin.lastPn).To(Equal(uint64(0)))
		})

		It("3 packets no update", func() {
			logWin := NewLogWinQueue(10)

			logWin.NewMediaPacketRecieved(0, 0, 12, false, false)
			logWin.NewMediaPacketRecieved(1, 1, 8, true, false)
			logWin.NewMediaPacketRecieved(2, 5, 20, false, false)

			Expect(logWin.numberPacketArrived).To(Equal(uint64(3)))
			Expect(logWin.numberMarkedPackets).To(Equal(uint64(1)))
			Expect(logWin.accumulatedSize).To(Equal(uint64(40)))
			Expect(logWin.numberLostPackets).To(BeZero())
			Expect(logWin.numberQueueBuildup).To(BeZero())

			Expect(logWin.numberPacketsSinceLoss).To(Equal(uint64(3)))
			Expect(logWin.lastPn).To(Equal(uint64(2)))
		})

		It("3 packets with gap", func() {
			logWin := NewLogWinQueue(10)

			logWin.NewMediaPacketRecieved(0, 0, 12, false, false)
			logWin.NewMediaPacketRecieved(2, 1, 8, true, false)
			logWin.NewMediaPacketRecieved(6, 5, 20, false, false)

			Expect(logWin.numberPacketArrived).To(Equal(uint64(3)))
			Expect(logWin.numberMarkedPackets).To(Equal(uint64(1)))
			Expect(logWin.accumulatedSize).To(Equal(uint64(40)))
			Expect(logWin.numberLostPackets).To(Equal(uint64(4)))
			Expect(logWin.numberQueueBuildup).To(BeZero())

			Expect(logWin.numberPacketsSinceLoss).To(Equal(uint64(1)))
			Expect(logWin.lastPn).To(Equal(uint64(6)))
		})

		It("packets out of order", func() {
			logWin := NewLogWinQueue(10)

			logWin.NewMediaPacketRecieved(0, 0, 12, false, false)
			logWin.NewMediaPacketRecieved(2, 1, 8, true, false)
			logWin.NewMediaPacketRecieved(6, 5, 20, false, false)

			logWin.NewMediaPacketRecieved(5, 5, 20, false, false)
			logWin.NewMediaPacketRecieved(4, 5, 20, false, false)
			logWin.NewMediaPacketRecieved(6, 5, 20, false, false)

			Expect(logWin.numberPacketArrived).To(Equal(uint64(3)))
			Expect(logWin.numberMarkedPackets).To(Equal(uint64(1)))
			Expect(logWin.accumulatedSize).To(Equal(uint64(40)))
			Expect(logWin.numberLostPackets).To(Equal(uint64(4)))
			Expect(logWin.numberQueueBuildup).To(BeZero())

			Expect(logWin.numberPacketsSinceLoss).To(Equal(uint64(1)))
			Expect(logWin.lastPn).To(Equal(uint64(6)))
		})

		It("3 packets with update", func() {
			logWin := NewLogWinQueue(10)
			logWin.lastPn = 200

			Expect(logWin.sizeInMicroS).To(Equal(uint64(10)))

			logWin.NewMediaPacketRecieved(201, 100, 12, false, false)
			logWin.NewMediaPacketRecieved(202, 101, 8, true, false)
			logWin.NewMediaPacketRecieved(203, 105, 20, false, false)

			Expect(logWin.numberPacketArrived).To(Equal(uint64(3)))
			Expect(logWin.numberMarkedPackets).To(Equal(uint64(1)))
			Expect(logWin.accumulatedSize).To(Equal(uint64(40)))
			Expect(logWin.numberLostPackets).To(BeZero())
			Expect(logWin.numberQueueBuildup).To(BeZero())

			Expect(logWin.numberPacketsSinceLoss).To(Equal(uint64(3)))
			Expect(logWin.lastPn).To(Equal(uint64(203)))

			logWin.updateStats(105)

			Expect(logWin.numberPacketArrived).To(Equal(uint64(3)))
			Expect(logWin.numberMarkedPackets).To(Equal(uint64(1)))
			Expect(logWin.accumulatedSize).To(Equal(uint64(40)))
			Expect(logWin.numberLostPackets).To(BeZero())
			Expect(logWin.numberQueueBuildup).To(BeZero())

			Expect(logWin.numberPacketsSinceLoss).To(Equal(uint64(3)))
			Expect(logWin.lastPn).To(Equal(uint64(203)))
		})

		It("update with old packets", func() {
			logWin := NewLogWinQueue(10)
			logWin.lastPn = 198

			Expect(logWin.sizeInMicroS).To(Equal(uint64(10)))

			logWin.NewMediaPacketRecieved(199, 50, 5, false, false)
			logWin.NewMediaPacketRecieved(200, 90, 5, true, false)
			logWin.NewMediaPacketRecieved(201, 100, 12, false, false)
			logWin.NewMediaPacketRecieved(202, 101, 8, true, false)
			logWin.NewMediaPacketRecieved(203, 105, 20, false, false)

			Expect(logWin.numberPacketArrived).To(Equal(uint64(5)))
			Expect(logWin.numberMarkedPackets).To(Equal(uint64(2)))
			Expect(logWin.accumulatedSize).To(Equal(uint64(50)))
			Expect(logWin.numberLostPackets).To(BeZero())
			Expect(logWin.numberQueueBuildup).To(BeZero())

			Expect(logWin.numberPacketsSinceLoss).To(Equal(uint64(5)))
			Expect(logWin.lastPn).To(Equal(uint64(203)))

			logWin.updateStats(105)

			Expect(logWin.numberPacketArrived).To(Equal(uint64(3)))
			Expect(logWin.numberMarkedPackets).To(Equal(uint64(1)))
			Expect(logWin.accumulatedSize).To(Equal(uint64(40)))
			Expect(logWin.numberLostPackets).To(BeZero())
			Expect(logWin.numberQueueBuildup).To(BeZero())

			Expect(logWin.numberPacketsSinceLoss).To(Equal(uint64(5)))
			Expect(logWin.lastPn).To(Equal(uint64(203)))
		})

		It("update with old packets and gaps", func() {
			logWin := NewLogWinQueue(10)
			logWin.lastPn = 198

			Expect(logWin.sizeInMicroS).To(Equal(uint64(10)))

			logWin.NewMediaPacketRecieved(199, 50, 5, false, true)
			logWin.NewMediaPacketRecieved(200, 90, 5, true, false)
			logWin.NewMediaPacketRecieved(201, 100, 12, false, false)
			logWin.NewMediaPacketRecieved(202, 101, 8, true, true)
			logWin.NewMediaPacketRecieved(213, 105, 20, false, true)

			Expect(logWin.numberPacketArrived).To(Equal(uint64(5)))
			Expect(logWin.numberMarkedPackets).To(Equal(uint64(2)))
			Expect(logWin.accumulatedSize).To(Equal(uint64(50)))
			Expect(logWin.numberLostPackets).To(Equal(uint64(10)))
			Expect(logWin.numberQueueBuildup).To(Equal(uint64(3)))

			Expect(logWin.numberPacketsSinceLoss).To(Equal(uint64(1)))
			Expect(logWin.lastPn).To(Equal(uint64(213)))

			logWin.updateStats(105)

			Expect(logWin.numberPacketArrived).To(Equal(uint64(3)))
			Expect(logWin.numberMarkedPackets).To(Equal(uint64(1)))
			Expect(logWin.accumulatedSize).To(Equal(uint64(40)))
			Expect(logWin.numberLostPackets).To(Equal(uint64(10)))
			Expect(logWin.numberQueueBuildup).To(Equal(uint64(2)))

			Expect(logWin.numberPacketsSinceLoss).To(Equal(uint64(1)))
			Expect(logWin.lastPn).To(Equal(uint64(213)))
		})

		It("update with old packets and gaps that are removed", func() {
			logWin := NewLogWinQueue(10)
			logWin.lastPn = 188

			Expect(logWin.sizeInMicroS).To(Equal(uint64(10)))

			logWin.NewMediaPacketRecieved(189, 50, 5, false, true)
			logWin.NewMediaPacketRecieved(200, 90, 5, true, true)
			logWin.NewMediaPacketRecieved(201, 100, 12, false, false)
			logWin.NewMediaPacketRecieved(202, 101, 8, true, false)
			logWin.NewMediaPacketRecieved(203, 105, 20, false, false)

			Expect(logWin.numberPacketArrived).To(Equal(uint64(5)))
			Expect(logWin.numberMarkedPackets).To(Equal(uint64(2)))
			Expect(logWin.accumulatedSize).To(Equal(uint64(50)))
			Expect(logWin.numberLostPackets).To(Equal(uint64(10)))
			Expect(logWin.numberQueueBuildup).To(Equal(uint64(2)))

			Expect(logWin.numberPacketsSinceLoss).To(Equal(uint64(4)))
			Expect(logWin.lastPn).To(Equal(uint64(203)))

			logWin.updateStats(105)

			Expect(logWin.numberPacketArrived).To(Equal(uint64(3)))
			Expect(logWin.numberMarkedPackets).To(Equal(uint64(1)))
			Expect(logWin.accumulatedSize).To(Equal(uint64(40)))
			Expect(logWin.numberLostPackets).To(Equal(uint64(0)))
			Expect(logWin.numberQueueBuildup).To(BeZero())

			Expect(logWin.numberPacketsSinceLoss).To(Equal(uint64(4)))
			Expect(logWin.lastPn).To(Equal(uint64(203)))
		})

		It("3 packets with queue buildup with update", func() {
			logWin := NewLogWinQueue(10)
			logWin.lastPn = 200

			Expect(logWin.sizeInMicroS).To(Equal(uint64(10)))

			logWin.NewMediaPacketRecieved(201, 100, 12, false, false)
			logWin.NewMediaPacketRecieved(202, 101, 8, true, true)
			logWin.NewMediaPacketRecieved(203, 105, 20, false, true)

			Expect(logWin.numberPacketArrived).To(Equal(uint64(3)))
			Expect(logWin.numberMarkedPackets).To(Equal(uint64(1)))
			Expect(logWin.accumulatedSize).To(Equal(uint64(40)))
			Expect(logWin.numberLostPackets).To(BeZero())
			Expect(logWin.numberQueueBuildup).To(Equal(uint64(2)))

			Expect(logWin.numberPacketsSinceLoss).To(Equal(uint64(3)))
			Expect(logWin.lastPn).To(Equal(uint64(203)))

			logWin.updateStats(105)

			Expect(logWin.numberPacketArrived).To(Equal(uint64(3)))
			Expect(logWin.numberMarkedPackets).To(Equal(uint64(1)))
			Expect(logWin.accumulatedSize).To(Equal(uint64(40)))
			Expect(logWin.numberLostPackets).To(BeZero())
			Expect(logWin.numberQueueBuildup).To(Equal(uint64(2)))

			Expect(logWin.numberPacketsSinceLoss).To(Equal(uint64(3)))
			Expect(logWin.lastPn).To(Equal(uint64(203)))
		})
	})
})

func TestAll(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LogWin Suite")
}
