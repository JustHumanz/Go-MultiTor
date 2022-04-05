package main

import (
	"testing"
)

var torTest = TorStruct{
	TorList: nil,
	Port:    "2525",
	IPAddr:  "127.0.0.1",
	Country: "ID",
	City:    "Jakarta",
	Load:    100,
}

func TestChangeCountry(t *testing.T) {
	countryTest := "JP"
	t.Logf("new country %s", countryTest)
	torTest.AddCountry(countryTest)
	if torTest.Country != countryTest {
		t.Errorf("country should be changed into %s", countryTest)
	}
}

func TestChangeIP(t *testing.T) {
	ipTest := "192.168.1.254"
	t.Logf("new ip %s", ipTest)
	torTest.AddIP(ipTest)
	if torTest.IPAddr != ipTest {
		t.Errorf("ip should be changed into %s", ipTest)
	}
}

func TestTorRRLB(t *testing.T) {
	torTestSlice := []TorStruct{torTest, torTest}

	t.Logf("test round robin LB")
	torTestLb := GetTorLB(torTestSlice)

	//it's should be the first slice of torTestSlice
	if torTestLb.Load != torTest.Load+1 {
		t.Errorf("failed to counter TorStruct load")
	}
}

func TestTorLcLB(t *testing.T) {
	newTorTest := torTest
	newTorTest.Load = 10

	t.Logf("test least connetion LB")
	torTestSlice := []TorStruct{torTest, newTorTest}
	torTestLb := GetTorLBWeight(torTestSlice)

	//it's should be the secound slice of torTestSlice
	if torTestLb.Load != newTorTest.Load+1 {
		t.Errorf("failed to counter TorStruct load")
	}
}
