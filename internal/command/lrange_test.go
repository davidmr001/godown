package command

import (
	"errors"
	"testing"
	"time"

	"github.com/gojuno/minimock"

	"github.com/namreg/godown/internal/storage"
	"github.com/namreg/godown/internal/storage/memory"
	"github.com/stretchr/testify/assert"
)

func TestLrange_Name(t *testing.T) {
	cmd := new(Lrange)
	assert.Equal(t, "LRANGE", cmd.Name())
}

func TestLrange_Help(t *testing.T) {
	cmd := new(Lrange)
	expected := `Usage: LRANGE key start stop
Returns the specified elements of the list stored at key. 
The offsets start and stop are zero-based indexes, 
with 0 being the first element of the list (the head of the list), 1 being the next element and so on.`
	assert.Equal(t, expected, cmd.Help())
}

func TestLrange_Execute(t *testing.T) {
	expired := storage.NewList([]string{"val"})
	expired.SetTTL(time.Now().Add(-1 * time.Second))

	strg := memory.New(map[storage.Key]*storage.Value{
		"string":  storage.NewString("value"),
		"list":    storage.NewList([]string{"val1", "val2"}),
		"expired": expired,
	})

	tests := []struct {
		name string
		args []string
		want Reply
	}{
		{"0:0", []string{"list", "0", "0"}, SliceReply{Value: []string{"val1"}}},
		{"0:1", []string{"list", "0", "1"}, SliceReply{Value: []string{"val1", "val2"}}},
		{"-1:-100", []string{"list", "-1", "-100"}, SliceReply{Value: []string{"val1", "val2"}}},
		{"-2:0", []string{"list", "-2", "0"}, SliceReply{Value: []string{"val1"}}},
		{"0:-1", []string{"list", "0", "-1"}, SliceReply{Value: []string{"val1", "val2"}}},
		{"1:0", []string{"list", "1", "0"}, NilReply{}},
		{"0:100500", []string{"list", "0", "100500"}, SliceReply{Value: []string{"val1", "val2"}}},
		{"100500:100501", []string{"list", "100500", "100501"}, NilReply{}},
		{"expired_key", []string{"expired", "0", "1"}, NilReply{}},
		{"not_existing_key", []string{"not_existing_key", "0", "1"}, NilReply{}},
		{"wrong_type_op", []string{"string", "0", "1"}, ErrReply{Value: ErrWrongTypeOp}},
		{"wrong_args_number/1", []string{}, ErrReply{Value: ErrWrongArgsNumber}},
		{"wrong_args_number/2", []string{"key", "0"}, ErrReply{Value: ErrWrongArgsNumber}},
		{"start_is_not_integer", []string{"list", "start", "1"}, ErrReply{Value: errors.New("start should be an integer")}},
		{"stop_is_not_integer", []string{"list", "0", "stop"}, ErrReply{Value: errors.New("stop should be an integer")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := Lrange{strg: strg}
			res := cmd.Execute(tt.args...)
			assert.Equal(t, tt.want, res)
		})
	}
}

func TestLrange_Execute_StorageErr(t *testing.T) {
	mc := minimock.NewController(t)
	defer mc.Finish()

	err := errors.New("error")

	strg := NewdataStoreMock(mc)
	strg.GetMock.Return(nil, err)

	cmd := Lrange{strg: strg}
	res := cmd.Execute("key", "0", "1")

	assert.Equal(t, ErrReply{Value: err}, res)
}
