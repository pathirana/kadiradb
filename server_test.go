package main

import (
	"bytes"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
)

const (
	DatabasePath = "/tmp/d1"
)

var (
	s  Server
	ss *server
)

func init() {
	if err := os.RemoveAll(DatabasePath); err != nil {
		panic(err)
	}

	srv, err := NewServer(&Options{Path: DatabasePath})
	if err != nil {
		panic(err)
	}

	openReq := &OpenReq{
		Database:    "test-info",
		Resolution:  60,
		Retention:   36000,
		EpochTime:   3600,
		MaxROEpochs: 2,
		MaxRWEpochs: 2,
	}

	openReqData, err := proto.Marshal(openReq)
	if err != nil {
		panic(err)
	}

	_, err = srv.Open(openReqData)
	if err != nil {
		panic(err)
	}

	s = srv
	ss = s.(*server)
}

func TestInfo(t *testing.T) {
	reqData := []byte{}
	resData, err := s.Info(reqData)
	if err != nil {
		t.Fatal(err)
	}

	res := &InfoRes{}
	if err := proto.Unmarshal(resData, res); err != nil {
		t.Fatal(err)
	}

	if len(res.Databases) != 1 {
		t.Fatal("should have db info")
	}

	dbInfo := res.Databases[0]

	if dbInfo.Database != "test-info" ||
		dbInfo.Resolution != uint32(60) {
		t.Fatal("wrong values")
	}
}

func TestOpen(t *testing.T) {
	// tested by init fn
}

func TestEdit(t *testing.T) {
	info, err := ss.databases["test-info"].Info()
	if err != nil {
		t.Fatal(err)
	}

	if info.MaxROEpochs != 2 ||
		info.MaxRWEpochs != 2 {
		t.Fatal("wrong values")
	}

	req := &EditReq{
		Database:    "test-info",
		MaxROEpochs: 3,
		MaxRWEpochs: 3,
	}

	reqData, err := proto.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}

	resData, err := s.Edit(reqData)
	if err != nil {
		t.Fatal(err)
	}

	res := &EditRes{}
	if err := proto.Unmarshal(resData, res); err != nil {
		t.Fatal(err)
	}

	info, err = ss.databases["test-info"].Info()
	if err != nil {
		t.Fatal(err)
	}

	if info.MaxROEpochs != 3 ||
		info.MaxRWEpochs != 3 {
		t.Fatal("wrong values")
	}
}

func TestPutGet(t *testing.T) {
	fld := []string{"test", "put", "get"}
	now := uint32(time.Now().Unix())
	req := &PutReq{
		Database:  "test-info",
		Fields:    fld,
		Timestamp: now,
		Count:     1,
		Value:     1.1,
	}

	reqData, err := proto.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}

	resData, err := s.Put(reqData)
	if err != nil {
		t.Fatal(err)
	}

	res := &PutRes{}
	if err := proto.Unmarshal(resData, res); err != nil {
		t.Fatal(err)
	}

	req2 := &GetReq{
		Database:  "test-info",
		Fields:    fld,
		GroupBy:   []bool{true, true, true},
		StartTime: now,
		EndTime:   now + 60,
	}

	reqData2, err := proto.Marshal(req2)
	if err != nil {
		t.Fatal(err)
	}

	resData2, err := s.Get(reqData2)
	if err != nil {
		t.Fatal(err)
	}

	res2 := &GetRes{}
	if err := proto.Unmarshal(resData2, res2); err != nil {
		t.Fatal(err)
	}

	if len(res2.Groups) != 1 {
		t.Fatal("incorrect number of results")
	}

	grp := res2.Groups[0]
	if !reflect.DeepEqual(grp.Fields, fld) {
		t.Fatal("incorrect set of fields", grp.Fields, fld)
	}

	if len(grp.Points) != 1 {
		t.Fatal("incorrect number of points")
	}

	point := grp.Points[0]
	if point.Value != 1.1 || point.Count != 1 {
		t.Fatal("incorrect values for point")
	}
}

func TestPutGetRes(t *testing.T) {
	fld := []string{"test", "put", "get"}
	now := uint32(time.Now().Unix())

	for i := 0; i < 20; i++ {
		req := &PutReq{
			Database:  "test-info",
			Fields:    fld,
			Timestamp: now - uint32(i)*60,
			Count:     1,
			Value:     1.1,
		}

		reqData, err := proto.Marshal(req)
		if err != nil {
			t.Fatal(err)
		}

		resData, err := s.Put(reqData)
		if err != nil {
			t.Fatal(err)
		}

		res := &PutRes{}
		if err := proto.Unmarshal(resData, res); err != nil {
			t.Fatal(err)
		}
	}

	req2 := &GetReq{
		Database:   "test-info",
		Fields:     fld,
		GroupBy:    []bool{true, true, true},
		StartTime:  now - 720,
		EndTime:    now,
		Resolution: 240,
	}

	reqData2, err := proto.Marshal(req2)
	if err != nil {
		t.Fatal(err)
	}

	resData2, err := s.Get(reqData2)
	if err != nil {
		t.Fatal(err)
	}

	res2 := &GetRes{}
	if err := proto.Unmarshal(resData2, res2); err != nil {
		t.Fatal(err)
	}

	if len(res2.Groups) != 1 {
		t.Fatal("incorrect number of results")
	}

	grp := res2.Groups[0]
	if !reflect.DeepEqual(grp.Fields, fld) {
		t.Fatal("incorrect set of fields", grp.Fields, fld)
	}

	if len(grp.Points) != 3 {
		t.Fatal("incorrect number of points")
	}

	point := grp.Points[0]
	if point.Value != 4.4 || point.Count != 4 {
		t.Fatal("incorrect values for point")
	}

	req3 := &GetReq{
		Database:   "test-info",
		Fields:     fld,
		GroupBy:    []bool{true, true, true},
		StartTime:  now - 720,
		EndTime:    now,
		Resolution: 360,
	}

	reqData3, err := proto.Marshal(req3)
	if err != nil {
		t.Fatal(err)
	}

	resData3, err := s.Get(reqData3)
	if err != nil {
		t.Fatal(err)
	}

	res3 := &GetRes{}
	if err := proto.Unmarshal(resData3, res3); err != nil {
		t.Fatal(err)
	}

	if len(res3.Groups) != 1 {
		t.Fatal("incorrect number of results")
	}

	grp = res3.Groups[0]
	if !reflect.DeepEqual(grp.Fields, fld) {
		t.Fatal("incorrect set of fields", grp.Fields, fld)
	}

	if len(grp.Points) != 2 {
		t.Fatal("incorrect number of points")
	}

	point = grp.Points[0]
	if point.Value != 6.6 || point.Count != 6 {
		t.Fatal("incorrect values for point")
	}
}

func TestIncGet(t *testing.T) {
	fld := []string{"test", "inc", "get"}
	now := uint32(time.Now().Unix())
	req := &IncReq{
		Database:  "test-info",
		Fields:    fld,
		Timestamp: now,
		Count:     1,
		Value:     1.1,
	}

	reqData, err := proto.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}

	resData, err := s.Inc(reqData)
	if err != nil {
		t.Fatal(err)
	}

	res := &IncRes{}
	if err := proto.Unmarshal(resData, res); err != nil {
		t.Fatal(err)
	}

	req2 := &GetReq{
		Database:  "test-info",
		Fields:    fld,
		GroupBy:   []bool{true, true, true},
		StartTime: now,
		EndTime:   now + 60,
	}

	reqData2, err := proto.Marshal(req2)
	if err != nil {
		t.Fatal(err)
	}

	resData2, err := s.Get(reqData2)
	if err != nil {
		t.Fatal(err)
	}

	res2 := &GetRes{}
	if err := proto.Unmarshal(resData2, res2); err != nil {
		t.Fatal(err)
	}

	if len(res2.Groups) != 1 {
		t.Fatal("incorrect number of results")
	}

	grp := res2.Groups[0]
	if !reflect.DeepEqual(grp.Fields, fld) {
		t.Fatal("incorrect set of fields", grp.Fields, fld)
	}

	if len(grp.Points) != 1 {
		t.Fatal("incorrect number of points")
	}

	point := grp.Points[0]
	if point.Value != 1.1 || point.Count != 1 {
		t.Fatal("incorrect values for point")
	}
}

func TestBatch(t *testing.T) {
	req := &ReqBatch{
		Batch: []*Request{
			&Request{InfoReq: &InfoReq{}},
			&Request{InfoReq: &InfoReq{}},
		},
	}

	reqData, err := proto.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}

	resData, err := s.Batch(reqData)
	if err != nil {
		t.Fatal(err)
	}

	res := &ResBatch{}
	if err := proto.Unmarshal(resData, res); err != nil {
		t.Fatal(err)
	}

	if len(res.Batch) != 2 {
		t.Fatal("should have 2 results")
	}

	if res.Batch[0].InfoRes == nil ||
		res.Batch[1].InfoRes == nil {
		t.Fatal("should have 2 info results")
	}
}

func TestEncode(t *testing.T) {
	p := valToPld(1, 2)
	if !bytes.Equal(p, []byte{0, 0, 0, 0, 0, 0, 240, 63, 2, 0, 0, 0}) {
		t.Fatal("wrong value")
	}
}

func TestDecode(t *testing.T) {
	x, y := pldToVal([]byte{0, 0, 0, 0, 0, 0, 240, 63, 2, 0, 0, 0})
	if x != 1 || y != 2 {
		t.Fatal("wrong value")
	}
}

func BenchmarkEncode(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = valToPld(1, 2)
	}
}

func BenchmarkDecode(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = pldToVal([]byte{0, 0, 0, 0, 0, 0, 240, 63, 2, 0, 0, 0})
	}
}
