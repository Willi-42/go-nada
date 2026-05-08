package nada

import (
	"time"
)

type SenderOnly struct {
	sender   *Sender
	receiver *Receiver
	lastRTT  time.Duration
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
			s.OnAck(ack.SeqNr, ack.Departure, ack.Arrival, ack.SizeBit, ack.Marked)
		} else {
			s.OnLoss(ack.SeqNr, ack.Departure)
		}
	}
	s.UpdateRTT(rtt)
	return s.UpdateTargetRate()
}

func (s *SenderOnly) OnAck(seqNr uint64, departure, arrival time.Time, sizeBit uint64, marked bool) {
	// register packet at receiver part
	s.receiver.PacketArrived(seqNr, departure, arrival, sizeBit, marked)

	// register at sender part. Necassary for correct ratios
	s.sender.PacketDelivered(seqNr, uint64(arrival.UnixMicro()), sizeBit, marked)
}

func (s *SenderOnly) OnLoss(seqNr uint64, departure time.Time) {
	s.sender.LostPacket(seqNr, uint64(departure.UnixMicro()))
}

func (s *SenderOnly) UpdateRTT(rtt time.Duration) {
	s.lastRTT = rtt
}

func (s *SenderOnly) UpdateTargetRate() uint64 {
	internalFeedback := s.receiver.GenerateFeedback()
	return s.sender.FeedbackReport(internalFeedback, s.lastRTT)
}
