package nada

import "time"

type SenderOnly struct {
	sender   *Sender
	receiver *Receiver
}

func NewSenderOnly(config Config) SenderOnly {
	configPopulated := populateConfig(&config)
	sender := NewSender(*configPopulated)
	receiver := NewReceiver(*configPopulated)

	return SenderOnly{
		sender:   &sender,
		receiver: &receiver,
	}
}

func (s *SenderOnly) OnAck(rtt time.Duration, feedback []FeedbackSO) (newRate uint64) {
	// register packets
	for _, event := range feedback {
		s.receiver.PacketArrived(event.PacketNumber, uint64(event.sentTs.UnixMicro()), uint64(event.RecvTs.UnixMicro()), event.PacketSizeBit, event.Marked)
	}

	// calc new rate
	internalFeedback := s.receiver.GenerateFeedbackRLD()
	newRate = s.sender.FeedbackReport(internalFeedback, uint64(rtt.Microseconds()))

	return newRate
}
