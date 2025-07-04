# go-nada

NADA implementation in Go.

## Network-Assisted Dynamic Adaptation (NADA)
NADA is a real-time congestion control algorithm.

NADA RFC: [RFC 8698](https://www.rfc-editor.org/rfc/rfc8698)

## How to

### Default / Receiver Loss Detection
Loss are detected at the receiver side.

#### Receiver Side
* Register arrived packets with `PacketArrived`.
* Register a packet without a timestamp with `PacketArrivedWithoutTs`.
* To generate the feedback for the sender, use `GenerateFeedbackRLD`.

#### Sender Side
* Register the arrived feedback with `FeedbackReport`.
  `FeedbackReport` returns the new media rate.

### Sender Loss Detection
Losses are detected at the sender side.

#### Receiver Side
* Register arrived packets with `PacketArrived`.
* To generate the feedback for the sender, use `GenerateFeedbackSLD`.

#### Sender Side
* Register lost packets with `LostPacket`.
* Register successfully delivered packets with `PacketDelivered`.
* Register the arrived feedback with `FeedbackReport`.
  `FeedbackReport` returns the new media rate.