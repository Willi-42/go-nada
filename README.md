# go-nada

NADA implementation in Go.

## Network-Assisted Dynamic Adaptation (NADA)
NADA is a real-time congestion control algorithm.

NADA RFC: [RFC 8698](https://www.rfc-editor.org/rfc/rfc8698)

## Differences to RFC 8689
* Calculation 5 of [Section 4.3](https://www.rfc-editor.org/rfc/rfc8698#name-sender-side-algorithm) has been changed. We removed the dependence on the max target rate (RMAX).
  * Can be reactivated by setting `UseDefaultGradualUpdates` to true.
* Maximal increment of the target rate in Gradual Update Mode is clipped to an increase of `MaxGradualUpdateFactor`.

### Loss detection
* This NADA implementation registers losses at the sender side. For example, through NACKs. In comparison, RFC 8689 detects losses purely through sequence number skips.

### Clock skew
If NADA observes a packet whose receive timestamp is earlier than its send timestamp, it treats this as clock skew between sender and receiver.
Since packets cannot arrive before they are sent, the receiver applies a fixed positive offset to all subsequent delay estimats, so they remains non-negative.

## Versions
* **Split NADA**: NADA runs at sender and at the reciever.
* **Sender-only NADA**: NADA only runs at the sender.

### **Split NADA** 

#### Receiver Side
* Register arrived packets with `PacketArrived`.
* To generate the feedback for the sender, use `GenerateFeedbackSLD`.

#### Sender Side
* Register lost packets with `LostPacket`.
* Register successfully delivered packets with `PacketDelivered`.
* Register the arrived feedback with `FeedbackReport`.
  `FeedbackReport` returns the new media rate.

#### Feedback
The feedback contains three values. Namely, the current receive rate, the newest delay measurement, and a boolean that indicates delay build-up.


### **Sender-only NADA**
* Requires feedback at the configured interval.
* Regsitered feedback with `OnAck`. Returns the new target rate.

#### Feedback
The feedback is a list of Acknowledgments. This list also contains the NACKs.

## Tunning Parameters
The `RefCongLevel` affects the stability of the target rate.
Default is 10ms. Increasing it can help to achieve a more stable target rate.

Reactiveness of Gradual Update Mode:
* **KAPPA**: General reactivness
  * KAPPA < 0.5: less reactive
  * KAPPA > 0.5: more reactive
* **ETA**: Reactiveness towards current congestion change
  * ETA < 2 -> less reactive
  * ETA > 2 -> more reactive

### Effects of pacing
NADA requires a paced media stream to achieve a stable target rate.
If the delay measurements contain latency spikes caused by bursts, NADA might produce a very unstable target rate.