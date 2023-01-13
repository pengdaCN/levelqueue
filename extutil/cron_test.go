package extutil

import (
	"context"
	"github.com/pengdacn/levelqueue"
	"testing"
)

func TestAutoCompact(t *testing.T) {
	ldis, err := levelqueue.Open("./tmp")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		cancel()
		ldis.Close()
	}()

	queue, err := levelqueue.NewSimpleQueue(
		"test-1",
		levelqueue.WithOwnLedis(AutoCompact(ldis)),
		levelqueue.WithContext(ctx),
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := queue.Clear(); err != nil {
		t.Fatal(err)
	}

	t.Log("OK")
}
