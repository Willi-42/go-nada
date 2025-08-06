package nada

import (
	"time"
)

type SenderOnly struct {
	sender   *SenderSLD
	receiver *Receiver
}

func NewSenderOnly(config Config) SenderOnly {
	sender := NewSenderSLD(config)
	receiver := NewReceiver(config)

	return SenderOnly{
		sender:   &sender,
		receiver: &receiver,
	}
}

func (s *SenderOnly) OnAcks(rtt time.Duration, acks []Acknowledgment) (newRate uint64) {
	// register packets
	for _, ack := range acks {
		if ack.Arrived {
			s.receiver.PacketArrived(ack.SeqNr, ack.Departure, ack.Arrival, ack.SizeBit, ack.Marked)
			s.sender.PacketDelivered(ack.SeqNr, uint64(ack.Departure.UnixMicro()), ack.SizeBit, ack.Marked)
		} else {
			s.sender.LostPacket(ack.SeqNr, uint64(ack.Departure.UnixMicro()))
		}
	}

	// calc new rate
	internalFeedback := s.receiver.GenerateFeedbackSLD()
	newRate = s.sender.FeedbackReport(internalFeedback, rtt)

	return newRate
}
