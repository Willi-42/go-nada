package nada

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Loss Intervall", func() {
	Context("test", func() {
		It("outdated events", func() {
			conf := Config{}

			receiver := NewReceiver(conf)

			receiver.baseDelay = 30_000 // 30ms

			// 10 packets with same delay
			for i := uint64(1); i <= 10; i++ {
				addPacket(&receiver, i, i*10, 50)
			}

			Expect(receiver.baseDelay).To(Equal(uint64(30_000)))
			Expect(receiver.qDelay).To(Equal(uint64(20_000)))
			Expect(receiver.recvRate).To(Equal(uint64(8000 * 10 * 1_000_000 / receiver.config.LogWin)))

			// add outdated packet with higher delay
			addPacket(&receiver, 11, 20, 90)

			Expect(receiver.baseDelay).To(Equal(uint64(30_000)))

			// ignore outdated packet; already have newer measurements
			Expect(receiver.qDelay).To(Equal(uint64(20_000)))

			// outdated packet should be counted in rate
			Expect(receiver.recvRate).To(Equal(uint64(8000 * 11 * 1_000_000 / receiver.config.LogWin)))

			// add outdated packet with lower delay
			addPacket(&receiver, 12, 25, 20)

			// use outdated packet for base delay
			Expect(receiver.baseDelay).To(Equal(uint64(20_000)))
			Expect(receiver.recvRate).To(Equal(uint64(8000 * 12 * 1_000_000 / receiver.config.LogWin)))

			// ignore for qdelay; but have newer base delay
			Expect(receiver.qDelay).To(Equal(uint64(20_000))) // still 20ms; because of min filter

			// but it should have added the updated the latest delay to 30ms

			// min filter of 15; add 14 bigger so that only the 30ms is left
			for i := uint64(0); i < 14; i++ {
				receiver.delayWin.AddSample(60_000)
			}

			Expect(receiver.delayWin.MinDelay()).To(Equal(uint64(30_000))) // check if 30ms is in delay window -> qdelay was updated despite the out-of-order packet was skipped
		})
	})
})

func addPacket(r *Receiver, id uint64, sendTs uint64, owd uint64) {
	departure := time.UnixMilli(int64(sendTs))
	arrival := time.UnixMilli(int64(sendTs + owd))
	size := uint64(8000)
	r.PacketArrived(id, departure, arrival, size, false)
}

func TestAll(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "General Suite")
}
