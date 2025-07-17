package nada

import (
	"time"
)

type SenderOnly struct {
	sender   *Sender
	receiver *Receiver
}

func NewSenderOnly(config Config) SenderOnly {
	sender := NewSender(config)
	receiver := NewReceiver(config)

	return SenderOnly{
		sender:   &sender,
		receiver: &receiver,
	}
}

func (s *SenderOnly) OnAcks(rtt time.Duration, acks []Acknowledgment) (newRate uint64) {
	// register packets
	for _, ack := range acks {
		s.receiver.PacketArrived(ack.SeqNr, ack.Departure, ack.Arrival, ack.SizeBit, ack.Marked)
	}

	// calc new rate
	internalFeedback := s.receiver.GenerateFeedbackRLD()
	newRate = s.sender.FeedbackReport(internalFeedback, rtt)

	return newRate
}
