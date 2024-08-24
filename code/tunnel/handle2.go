package tunnel

import (
	"github.com/libp2p/go-libp2p/core/network"
)

func (ht *HttpTunnel) streamHandleHttp2(stream network.Stream) {
	// maddr, err := stream.Conn().RemoteMultiaddr().MarshalJSON()
	// if err != nil {
	// 	panic(err)
	// }

	// unmarsheler := cbor.NewUnmarshaller(cbor.DecodeOptions{
	// 	CoerceUndefToNull: true,
	// }, stream)

	// packet := wire.Packet{}

	// err = unmarsheler.Unmarshal(&packet)
	// if err != nil {
	// 	panic(err)
	// }

}

/*

maddr, err := stream.Conn().RemoteMultiaddr().MarshalJSON()
	if err != nil {
		panic(err)
	}

	pp.Println("@new_http_from", string(maddr))

	defer stream.Close()

	req, err := http.ReadRequest(bufio.NewReader(stream))
	if err != nil {
		panic(err)
	}

	req.URL.Host = fmt.Sprintf("localhost:%d", ht.tunnelToPort)
	req.URL.Scheme = "http"
	req.RequestURI = ""

	pp.Println("@connecting_to", req.URL.String())

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		pp.Println("@req", req)
		panic(err)
	}

	defer resp.Body.Close()

	bodyBackup := resp.Body
	resp.Body = nil

	out, err := httputil.DumpResponse(resp, false)
	if err != nil {
		panic(err)
	}

	pp.Println("@resp", string(out))

	pp.Print("@write_head")
	pp.Println(stream.Write(out))

	pp.Print("@write_body")
	pp.Println(io.Copy(stream, (bodyBackup)))

*/
