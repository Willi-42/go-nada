package nada

import (
	"time"
)

type SenderOnly struct {
	senderSLD *SenderSLD
	senderRLD *Sender
	receiver  *Receiver

	useNACKs bool
}

func NewSenderOnly(config Config, useNACKs bool) SenderOnly {
	receiver := NewReceiver(config)

	newSO := SenderOnly{
		senderSLD: nil,
		senderRLD: nil,
		receiver:  &receiver,
		useNACKs:  useNACKs,
	}

	if useNACKs {
		senderSLD := NewSenderSLD(config)
		newSO.senderSLD = &senderSLD
	} else {
		senderRLD := NewSender(config)
		newSO.senderRLD = &senderRLD
	}

	return newSO
}

func (s *SenderOnly) OnAcks(rtt time.Duration, acks []Acknowledgment) (newRate uint64) {
	// register packets
	for _, ack := range acks {
		if ack.Arrived {
			// allways register packet at receiver part
			s.receiver.PacketArrived(ack.SeqNr, ack.Departure, ack.Arrival, ack.SizeBit, ack.Marked)

			if s.useNACKs {
				// for NACKs also register at sender part
				s.senderSLD.PacketDelivered(ack.SeqNr, uint64(ack.Arrival.UnixMicro()), ack.SizeBit, ack.Marked)
			}

		} else if s.useNACKs {
			// only register loss if use NACKs
			s.senderSLD.LostPacket(ack.SeqNr, uint64(ack.Departure.UnixMicro()))
		}
	}

	// calc new rate
	if s.useNACKs {
		internalFeedback := s.receiver.GenerateFeedbackSLD()
		newRate = s.senderSLD.FeedbackReport(internalFeedback, rtt)

	} else {
		internalFeedback := s.receiver.GenerateFeedbackRLD()
		newRate = s.senderRLD.FeedbackReport(internalFeedback, rtt)
	}

	return newRate
}
