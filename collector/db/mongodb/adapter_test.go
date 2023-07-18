package mongodb

import (
	"context"
	"testing"
)

func Test_GetLastTime(t *testing.T) {
	//t.Error()
	var a Adapter
	err := a.Connect(context.TODO(), "mongodb://localhost:27017/collectorDb")
	if err != nil {
		t.Fatalf("cannot connect to the database")
	}

	_, err = a.GetLastTime(context.TODO(), "pseudo_1")
	if err != nil {
		t.Errorf("GetLastTime error: %v", err)
	}
}
