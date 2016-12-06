package packet

import (
	"bytes"
	"testing"
	"time"
)

func TestTransieverEncodesDecodesPingCorrectly(t *testing.T) {
	var dataChannel bytes.Buffer
	transiever := MakeTransiever(&dataChannel, &dataChannel)

	err := transiever.WritePing(&PingPong{Sent: time.Date(2006, 2, 2, 4, 1, 0, 0, time.UTC)})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if dataChannel.Len() <= 0 {
		t.Error("Expected data to be written")
	}

	var out PingPong
	pktType, err := transiever.Decode()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if pktType != PktPing {
		t.Error("Expected PktPing packet type")
	}

	err = transiever.GetPing(&out)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if out.Sent.Unix() != time.Date(2006, 2, 2, 4, 1, 0, 0, time.UTC).Unix() {
		t.Error("Incorrect packet value")
	}
}
