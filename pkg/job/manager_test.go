package job

import (
	"context"
	"errors"
	"testing"
)

// initManagerContext initiates Manager and the context used for Manager functions.
func initManagerContext(role string) (*Manager, context.Context) {
	m := NewManager()
	ctx := WithUserInfo(context.Background(), "testdummy", role)

	return m, ctx
}

func TestInvalidJobID(t *testing.T) {
	m, ctx := initManagerContext(Admin)

	_, err := m.GetStatus(ctx, "dummy_12345")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("GetStatus() expected error:  %s, got: %s", ErrNotFound, err.Error())
	}

	_, _, err = m.GetOutput(ctx, "dummy_12345")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("GetOutput() expected error: %s, got: %s", ErrNotFound, err.Error())
	}
}

func TestUnauthorized(t *testing.T) {
	m, ctx := initManagerContext(Admin)
	jobID, err := m.Start(ctx, shortCmd[0], shortCmd[1:])
	if err != nil {
		t.Errorf("Start() error: %s", err)
	}

	// get status with wrong user; user should not see the job
	newCtx := WithUserInfo(context.Background(), "falsedummy", User)
	_, err = m.GetStatus(newCtx, jobID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("GetStatus() expected error: %s: %s", ErrNotFound, err.Error())
	}

	// get status with right user
	_, err = m.GetStatus(ctx, jobID)
	if err != nil {
		t.Errorf("GetStatus() error: %s", err.Error())
	}
}
