package wire

type Packet struct {
	PacketType     string `cbor:"ptype"`
	HttpRequestId  uint32 `cbor:"http_reqid"` // make it string / nanoId
	FragmentId     uint32 `cbor:"fid"`
	TotalFragments uint32 `cbor:"total_frags"`
	Data           []byte `cbor:"data"`
	Direction      uint8  `cbor:"direction"`
}

const (
	FragmentSend   = "FRAG_SEND"   // sending 10 fragments for a request id 123, send the fragment packet
	FragmentTally  = "FRAG_TALLY"  // when done sending all fragments, send this packet
	FragmentResend = "FRAG_RESEND" // and another side will request any missing fragments
	FragmentEnd    = "FRAG_END"    // when all fragments are received, send this packet that ends that request or response
	PING           = "PING"
	PONG           = "PONG"
)

const (
	DirectionRequest = iota
	DirectionResponse
)

const (
	PacketFragmentationSize = 1024 * 10
)
