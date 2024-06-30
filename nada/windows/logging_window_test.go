package windows

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LogWin", func() {
	Context("test", func() {
		It("1 packets no update", func() {
			logWin := NewLogWindow(10, 2)
			logWin.NewMediaPacketRecieved(0, 0, 12, false, false)
			Expect(logWin.lastPn).To(Equal(uint64(0)))

			Expect(logWin.arrivedPackets).To(Equal(uint64(1)))
			Expect(logWin.markedPackets).To(Equal(uint64(0)))
			Expect(logWin.receivedBits).To(Equal(uint64(12)))
			Expect(logWin.lostPackets).To(BeZero())
			Expect(logWin.queueBuildupCnt).To(BeZero())

			Expect(logWin.packetsSinceLoss).To(Equal(uint64(1)))
			Expect(logWin.lastPn).To(Equal(uint64(0)))
		})

		It("3 packets no update", func() {
			logWin := NewLogWindow(10, 2)

			logWin.NewMediaPacketRecieved(0, 0, 12, false, false)
			logWin.NewMediaPacketRecieved(1, 1, 8, true, false)
			logWin.NewMediaPacketRecieved(2, 5, 20, false, false)

			Expect(logWin.arrivedPackets).To(Equal(uint64(3)))
			Expect(logWin.markedPackets).To(Equal(uint64(1)))
			Expect(logWin.receivedBits).To(Equal(uint64(40)))
			Expect(logWin.lostPackets).To(BeZero())
			Expect(logWin.queueBuildupCnt).To(BeZero())

			Expect(logWin.packetsSinceLoss).To(Equal(uint64(3)))
			Expect(logWin.lastPn).To(Equal(uint64(2)))
		})

		It("3 packets with gap", func() {
			logWin := NewLogWindow(10, 2)

			logWin.NewMediaPacketRecieved(0, 0, 12, false, false)
			logWin.NewMediaPacketRecieved(2, 1, 8, true, false)
			logWin.NewMediaPacketRecieved(6, 5, 20, false, false)

			Expect(logWin.arrivedPackets).To(Equal(uint64(3)))
			Expect(logWin.markedPackets).To(Equal(uint64(1)))
			Expect(logWin.receivedBits).To(Equal(uint64(40)))
			Expect(logWin.lostPackets).To(Equal(uint64(4)))
			Expect(logWin.queueBuildupCnt).To(BeZero())

			Expect(logWin.packetsSinceLoss).To(Equal(uint64(1)))
			Expect(logWin.lastPn).To(Equal(uint64(6)))
		})

		It("packets out of order", func() {
			logWin := NewLogWindow(10, 2)

			logWin.NewMediaPacketRecieved(0, 0, 12, false, false)
			logWin.NewMediaPacketRecieved(2, 1, 8, true, false)
			logWin.NewMediaPacketRecieved(6, 5, 20, false, false)

			logWin.NewMediaPacketRecieved(5, 5, 20, false, false)
			logWin.NewMediaPacketRecieved(4, 5, 20, false, false)
			logWin.NewMediaPacketRecieved(6, 5, 20, false, false)

			Expect(logWin.arrivedPackets).To(Equal(uint64(3)))
			Expect(logWin.markedPackets).To(Equal(uint64(1)))
			Expect(logWin.receivedBits).To(Equal(uint64(40)))
			Expect(logWin.lostPackets).To(Equal(uint64(4)))
			Expect(logWin.queueBuildupCnt).To(BeZero())

			Expect(logWin.packetsSinceLoss).To(Equal(uint64(1)))
			Expect(logWin.lastPn).To(Equal(uint64(6)))
		})

		It("3 packets with update", func() {
			logWin := NewLogWindow(10, 2)
			logWin.lastPn = 200

			Expect(logWin.windowSize).To(Equal(uint64(10)))

			logWin.NewMediaPacketRecieved(201, 100, 12, false, false)
			logWin.NewMediaPacketRecieved(202, 101, 8, true, false)
			logWin.NewMediaPacketRecieved(203, 105, 20, false, false)

			Expect(logWin.arrivedPackets).To(Equal(uint64(3)))
			Expect(logWin.markedPackets).To(Equal(uint64(1)))
			Expect(logWin.receivedBits).To(Equal(uint64(40)))
			Expect(logWin.lostPackets).To(BeZero())
			Expect(logWin.queueBuildupCnt).To(BeZero())

			Expect(logWin.packetsSinceLoss).To(Equal(uint64(3)))
			Expect(logWin.lastPn).To(Equal(uint64(203)))

			logWin.UpdateStats(105)

			Expect(logWin.arrivedPackets).To(Equal(uint64(3)))
			Expect(logWin.markedPackets).To(Equal(uint64(1)))
			Expect(logWin.receivedBits).To(Equal(uint64(40)))
			Expect(logWin.lostPackets).To(BeZero())
			Expect(logWin.queueBuildupCnt).To(BeZero())

			Expect(logWin.packetsSinceLoss).To(Equal(uint64(3)))
			Expect(logWin.lastPn).To(Equal(uint64(203)))
		})

		It("update with old packets", func() {
			logWin := NewLogWindow(10, 2)
			logWin.lastPn = 198

			Expect(logWin.windowSize).To(Equal(uint64(10)))

			logWin.NewMediaPacketRecieved(199, 50, 5, false, false)
			logWin.NewMediaPacketRecieved(200, 90, 5, true, false)
			logWin.NewMediaPacketRecieved(201, 100, 12, false, false)
			logWin.NewMediaPacketRecieved(202, 101, 8, true, false)
			logWin.NewMediaPacketRecieved(203, 105, 20, false, false)

			Expect(logWin.arrivedPackets).To(Equal(uint64(5)))
			Expect(logWin.markedPackets).To(Equal(uint64(2)))
			Expect(logWin.receivedBits).To(Equal(uint64(50)))
			Expect(logWin.lostPackets).To(BeZero())
			Expect(logWin.queueBuildupCnt).To(BeZero())

			Expect(logWin.packetsSinceLoss).To(Equal(uint64(5)))
			Expect(logWin.lastPn).To(Equal(uint64(203)))

			logWin.UpdateStats(105)

			Expect(logWin.arrivedPackets).To(Equal(uint64(3)))
			Expect(logWin.markedPackets).To(Equal(uint64(1)))
			Expect(logWin.receivedBits).To(Equal(uint64(40)))
			Expect(logWin.lostPackets).To(BeZero())
			Expect(logWin.queueBuildupCnt).To(BeZero())

			Expect(logWin.packetsSinceLoss).To(Equal(uint64(5)))
			Expect(logWin.lastPn).To(Equal(uint64(203)))
		})

		It("update with old packets and gaps", func() {
			logWin := NewLogWindow(10, 2)
			logWin.lastPn = 198

			Expect(logWin.windowSize).To(Equal(uint64(10)))

			logWin.NewMediaPacketRecieved(199, 50, 5, false, true)
			logWin.NewMediaPacketRecieved(200, 90, 5, true, false)
			logWin.NewMediaPacketRecieved(201, 100, 12, false, false)
			logWin.NewMediaPacketRecieved(202, 101, 8, true, true)
			logWin.NewMediaPacketRecieved(213, 105, 20, false, true)

			Expect(logWin.arrivedPackets).To(Equal(uint64(5)))
			Expect(logWin.markedPackets).To(Equal(uint64(2)))
			Expect(logWin.receivedBits).To(Equal(uint64(50)))
			Expect(logWin.lostPackets).To(Equal(uint64(10)))
			Expect(logWin.queueBuildupCnt).To(Equal(uint64(3)))

			Expect(logWin.packetsSinceLoss).To(Equal(uint64(1)))
			Expect(logWin.lastPn).To(Equal(uint64(213)))

			logWin.UpdateStats(105)

			Expect(logWin.arrivedPackets).To(Equal(uint64(3)))
			Expect(logWin.markedPackets).To(Equal(uint64(1)))
			Expect(logWin.receivedBits).To(Equal(uint64(40)))
			Expect(logWin.lostPackets).To(Equal(uint64(10)))
			Expect(logWin.queueBuildupCnt).To(Equal(uint64(2)))

			Expect(logWin.packetsSinceLoss).To(Equal(uint64(1)))
			Expect(logWin.lastPn).To(Equal(uint64(213)))
		})

		It("update with old packets and gaps that are removed", func() {
			logWin := NewLogWindow(10, 2)
			logWin.lastPn = 188

			Expect(logWin.windowSize).To(Equal(uint64(10)))

			logWin.NewMediaPacketRecieved(189, 50, 5, false, true)
			logWin.NewMediaPacketRecieved(200, 90, 5, true, true)
			logWin.NewMediaPacketRecieved(201, 100, 12, false, false)
			logWin.NewMediaPacketRecieved(202, 101, 8, true, false)
			logWin.NewMediaPacketRecieved(203, 105, 20, false, false)

			Expect(logWin.arrivedPackets).To(Equal(uint64(5)))
			Expect(logWin.markedPackets).To(Equal(uint64(2)))
			Expect(logWin.receivedBits).To(Equal(uint64(50)))
			Expect(logWin.lostPackets).To(Equal(uint64(10)))
			Expect(logWin.queueBuildupCnt).To(Equal(uint64(2)))

			Expect(logWin.packetsSinceLoss).To(Equal(uint64(4)))
			Expect(logWin.lastPn).To(Equal(uint64(203)))

			logWin.UpdateStats(105)

			Expect(logWin.arrivedPackets).To(Equal(uint64(3)))
			Expect(logWin.markedPackets).To(Equal(uint64(1)))
			Expect(logWin.receivedBits).To(Equal(uint64(40)))
			Expect(logWin.lostPackets).To(Equal(uint64(0)))
			Expect(logWin.queueBuildupCnt).To(BeZero())

			Expect(logWin.packetsSinceLoss).To(Equal(uint64(4)))
			Expect(logWin.lastPn).To(Equal(uint64(203)))
		})

		It("3 packets with queue buildup with update", func() {
			logWin := NewLogWindow(10, 2)
			logWin.lastPn = 200

			Expect(logWin.windowSize).To(Equal(uint64(10)))

			logWin.NewMediaPacketRecieved(201, 100, 12, false, false)
			logWin.NewMediaPacketRecieved(202, 101, 8, true, true)
			logWin.NewMediaPacketRecieved(203, 105, 20, false, true)

			Expect(logWin.arrivedPackets).To(Equal(uint64(3)))
			Expect(logWin.markedPackets).To(Equal(uint64(1)))
			Expect(logWin.receivedBits).To(Equal(uint64(40)))
			Expect(logWin.lostPackets).To(BeZero())
			Expect(logWin.queueBuildupCnt).To(Equal(uint64(2)))

			Expect(logWin.packetsSinceLoss).To(Equal(uint64(3)))
			Expect(logWin.lastPn).To(Equal(uint64(203)))

			logWin.UpdateStats(105)

			Expect(logWin.arrivedPackets).To(Equal(uint64(3)))
			Expect(logWin.markedPackets).To(Equal(uint64(1)))
			Expect(logWin.receivedBits).To(Equal(uint64(40)))
			Expect(logWin.lostPackets).To(BeZero())
			Expect(logWin.queueBuildupCnt).To(Equal(uint64(2)))

			Expect(logWin.packetsSinceLoss).To(Equal(uint64(3)))
			Expect(logWin.lastPn).To(Equal(uint64(203)))
		})

		It("update with old packets, gaps and skipped pns", func() {
			logWin := NewLogWindow(10, 2)
			logWin.lastPn = 198

			Expect(logWin.windowSize).To(Equal(uint64(10)))

			logWin.NewMediaPacketRecieved(199, 50, 5, false, true)
			logWin.NewMediaPacketRecieved(200, 90, 5, true, false)
			logWin.NewMediaPacketRecieved(201, 100, 12, false, false)
			logWin.NewMediaPacketRecieved(202, 101, 8, true, true)

			logWin.AddEmptyPacket(203, 102)
			logWin.AddEmptyPacket(204, 103)
			logWin.AddEmptyPacket(207, 103)

			logWin.NewMediaPacketRecieved(50, 101, 8, false, false) // reordered packet

			logWin.NewMediaPacketRecieved(213, 105, 20, false, true)

			Expect(logWin.arrivedPackets).To(Equal(uint64(5)))
			Expect(logWin.markedPackets).To(Equal(uint64(2)))
			Expect(logWin.receivedBits).To(Equal(uint64(50)))
			Expect(logWin.lostPackets).To(Equal(uint64(7)))
			Expect(logWin.queueBuildupCnt).To(Equal(uint64(3)))

			Expect(logWin.packetsSinceLoss).To(Equal(uint64(1)))
			Expect(logWin.lastPn).To(Equal(uint64(213)))

			logWin.UpdateStats(105)

			Expect(logWin.arrivedPackets).To(Equal(uint64(3)))
			Expect(logWin.markedPackets).To(Equal(uint64(1)))
			Expect(logWin.receivedBits).To(Equal(uint64(40)))
			Expect(logWin.lostPackets).To(Equal(uint64(7)))
			Expect(logWin.queueBuildupCnt).To(Equal(uint64(2)))

			Expect(logWin.packetsSinceLoss).To(Equal(uint64(1)))
			Expect(logWin.lastPn).To(Equal(uint64(213)))
		})
	})
})

func TestAll(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LogWin Suite")
}
