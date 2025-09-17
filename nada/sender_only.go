package nada

import (
	"time"
)

type SenderOnly struct {
	sender   *Sender
	receiver *Receiver
}

func NewSenderOnly(config Config) SenderOnly {
	receiver := NewReceiver(config)

	newSO := SenderOnly{
		sender:   nil,
		receiver: &receiver,
	}

	sender := NewSender(config)
	newSO.sender = &sender

	return newSO
}

func (s *SenderOnly) OnAcks(rtt time.Duration, acks []Acknowledgment) (newRate uint64) {
	// register packets
	for _, ack := range acks {
		if ack.Arrived {
			// register packet at receiver part
			s.receiver.PacketArrived(ack.SeqNr, ack.Departure, ack.Arrival, ack.SizeBit, ack.Marked)

			// register at sender part. Necassary for correct ratios
			s.sender.PacketDelivered(ack.SeqNr, uint64(ack.Arrival.UnixMicro()), ack.SizeBit, ack.Marked)

		} else {
			// register loss
			s.sender.LostPacket(ack.SeqNr, uint64(ack.Departure.UnixMicro()))
		}
	}

	// calc new rate
	internalFeedback := s.receiver.GenerateFeedback()
	newRate = s.sender.FeedbackReport(internalFeedback, rtt)

	return newRate
}
