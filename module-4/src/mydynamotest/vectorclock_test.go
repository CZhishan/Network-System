package mydynamotest

import (
	"mydynamo"
	"testing"
)

func TestBasicVectorClock(t *testing.T) {
	t.Logf("Starting TestBasicVectorClock")

	//create two vector clocks
	clock1 := mydynamo.NewVectorClock()
	clock2 := mydynamo.NewVectorClock()

	//Test for equality
	if !clock1.Equals(clock2) {
		t.Fail()
		t.Logf("Vector Clocks were not equal")
	}

}

func TestConcurrentVectorClock(t *testing.T) {
	t.Logf("Starting Concurrency Test")

	//create two vector clocks
	clock1 := mydynamo.NewVectorClock()
	l := map[string]int{
		"1": 2,
		"3":1,
	}
	k :=map[string]int{
		"1":2,
		"2":1,
	}
	clock1.TimeList = l
	clock2 := mydynamo.NewVectorClock()
	clock2.TimeList = k

	//Test for equality
	if !clock1.Concurrent(clock2) {
		t.Fail()
		t.Logf("Vector Clocks were not concurrent!")
	}

}
