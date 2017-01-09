package packet

import (
	"bytes"
	"testing"
	"time"

	"github.com/twitchyliquid64/nugget"
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

func TestTransieverEncodesDecodesPongCorrectly(t *testing.T) {
	var dataChannel bytes.Buffer
	transiever := MakeTransiever(&dataChannel, &dataChannel)

	err := transiever.WritePong(&PingPong{Sent: time.Date(2006, 2, 2, 4, 1, 0, 0, time.UTC)})
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
	if pktType != PktPong {
		t.Error("Expected PktPong packet type")
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

func TestTransieverEncodesDecodesLookupRPCCorrectly(t *testing.T) {
	var dataChannel bytes.Buffer
	transiever := MakeTransiever(&dataChannel, &dataChannel)

	err := transiever.WriteLookupReq(&LookupReq{ID: 455243, Path: "/lol"})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if dataChannel.Len() <= 0 {
		t.Error("Expected data to be written")
	}

	var out LookupReq
	pktType, err := transiever.Decode()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if pktType != PktLookup {
		t.Error("Expected PktLookup packet type")
	}

	err = transiever.GetLookupReq(&out)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if out.ID != 455243 || out.Path != "/lol" {
		t.Error("Incorrect packet value")
	}
}

func TestTransieverEncodesDecodesLookupRPCResponseCorrectly(t *testing.T) {
	var dataChannel bytes.Buffer
	transiever := MakeTransiever(&dataChannel, &dataChannel)

	err := transiever.WriteLookupResp(&LookupResp{ID: 455243})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if dataChannel.Len() <= 0 {
		t.Error("Expected data to be written")
	}

	var out LookupResp
	pktType, err := transiever.Decode()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if pktType != PktLookupResp {
		t.Error("Expected PktLookupResp packet type")
	}

	err = transiever.GetLookupResp(&out)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if out.ID != 455243 {
		t.Error("Incorrect packet value")
	}
}

func TestTransieverEncodesDecodesReadMetaRPCCorrectly(t *testing.T) {
	var dataChannel bytes.Buffer
	transiever := MakeTransiever(&dataChannel, &dataChannel)

	err := transiever.WriteReadMetaReq(&ReadMetaReq{ID: 455243, EntryID: nugget.EntryID{1, 2, 3}})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if dataChannel.Len() <= 0 {
		t.Error("Expected data to be written")
	}

	var out ReadMetaReq
	pktType, err := transiever.Decode()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if pktType != PktReadMeta {
		t.Error("Expected PktReadMeta packet type")
	}

	err = transiever.GetReadMetaReq(&out)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if (out.ID != 455243 || out.EntryID != nugget.EntryID{1, 2, 3}) {
		t.Error("Incorrect packet value")
	}
}

func TestTransieverEncodesDecodesReadMetaRPCResponseCorrectly(t *testing.T) {
	var dataChannel bytes.Buffer
	transiever := MakeTransiever(&dataChannel, &dataChannel)

	err := transiever.WriteReadMetaResp(&ReadMetaResp{ID: 455243})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if dataChannel.Len() <= 0 {
		t.Error("Expected data to be written")
	}

	var out ReadMetaResp
	pktType, err := transiever.Decode()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if pktType != PktReadMetaResp {
		t.Error("Expected PktReadMetaResp packet type")
	}

	err = transiever.GetReadMetaResp(&out)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if out.ID != 455243 {
		t.Error("Incorrect packet value")
	}
}

func TestTransieverEncodesDecodesListRPCCorrectly(t *testing.T) {
	var dataChannel bytes.Buffer
	transiever := MakeTransiever(&dataChannel, &dataChannel)

	err := transiever.WriteListReq(&ListReq{ID: 455243, Path: "/cat"})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if dataChannel.Len() <= 0 {
		t.Error("Expected data to be written")
	}

	var out ListReq
	pktType, err := transiever.Decode()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if pktType != PktList {
		t.Error("Expected PktList packet type")
	}

	err = transiever.GetListReq(&out)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if out.ID != 455243 || out.Path != "/cat" {
		t.Error("Incorrect packet value")
	}
}

func TestTransieverEncodesDecodesListRPCResponseCorrectly(t *testing.T) {
	var dataChannel bytes.Buffer
	transiever := MakeTransiever(&dataChannel, &dataChannel)

	err := transiever.WriteListResp(&ListResp{ID: 455243})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if dataChannel.Len() <= 0 {
		t.Error("Expected data to be written")
	}

	var out ListResp
	pktType, err := transiever.Decode()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if pktType != PktListResp {
		t.Error("Expected PktListResp packet type")
	}

	err = transiever.GetListResp(&out)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if out.ID != 455243 {
		t.Error("Incorrect packet value")
	}
}

func TestTransieverEncodesDecodesFetchRPCCorrectly(t *testing.T) {
	var dataChannel bytes.Buffer
	transiever := MakeTransiever(&dataChannel, &dataChannel)

	err := transiever.WriteFetchReq(&FetchReq{ID: 455243, Path: "/cat2"})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if dataChannel.Len() <= 0 {
		t.Error("Expected data to be written")
	}

	var out FetchReq
	pktType, err := transiever.Decode()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if pktType != PktFetch {
		t.Error("Expected PktFetch packet type")
	}

	err = transiever.GetFetchReq(&out)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if out.ID != 455243 || out.Path != "/cat2" {
		t.Error("Incorrect packet value")
	}
}

func TestTransieverEncodesDecodesFetchRPCResponseCorrectly(t *testing.T) {
	var dataChannel bytes.Buffer
	transiever := MakeTransiever(&dataChannel, &dataChannel)

	err := transiever.WriteFetchResp(&FetchResp{ID: 4552, Data: []byte("ab"), EntryID: nugget.EntryID{1}})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if dataChannel.Len() <= 0 {
		t.Error("Expected data to be written")
	}

	var out FetchResp
	pktType, err := transiever.Decode()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if pktType != PktFetchResp {
		t.Error("Expected PktFetchResp packet type")
	}

	err = transiever.GetFetchResp(&out)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if (out.ID != 4552 || bytes.Compare(out.Data, []byte("ab")) != 0 || out.EntryID != nugget.EntryID{1}) {
		t.Error("Incorrect packet value")
	}
}
