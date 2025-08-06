# go-nada

NADA implementation in Go.

## Network-Assisted Dynamic Adaptation (NADA)
NADA is a real-time congestion control algorithm.

NADA RFC: [RFC 8698](https://www.rfc-editor.org/rfc/rfc8698)

## Versions
* **Split NADA**: NADA runs at sender and at the reciever.
* **Sender-only NADA**: NADA only runs at the sender.

### **Split** with Receiver Loss Detection
Loss are detected at the receiver side.

#### Receiver Side
* Register arrived packets with `PacketArrived`.
* Register a packet without a timestamp with `PacketArrivedWithoutTs`.
* To generate the feedback for the sender, use `GenerateFeedbackRLD`.

#### Sender Side
* Register the arrived feedback with `FeedbackReport`.
  `FeedbackReport` returns the new media rate.

### **Split** with Sender Loss Detection
Losses are detected at the sender side.

#### Receiver Side
* Register arrived packets with `PacketArrived`.
* To generate the feedback for the sender, use `GenerateFeedbackSLD`.

#### Sender Side
* Register lost packets with `LostPacket`.
* Register successfully delivered packets with `PacketDelivered`.
* Register the arrived feedback with `FeedbackReport`.
  `FeedbackReport` returns the new media rate.


### **Sender-only**
* Requires feedback at the configured interval.
* Regsitered feedback with `OnAck`. Returns the new target rate.

## Tunning Parameters

Reactiveness of Gradual Update Mode:
* **KAPPA**: General reactivness
 * KAPPA < 0.5: less reactive
 * KAPPA > 0.5: more reactive
* **ETA**: Reactiveness towards current congestion change
 * ETA < 2 -> less reactive
 * ETA > 2 -> more reactive