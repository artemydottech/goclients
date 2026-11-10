package repository

import (
	"context"
	"reflect"
	"testing"
)

func TestClientTagsReplaceRollbackAndCascade(t *testing.T) {
	f := newReminderFixture(t)
	ctx := context.Background()
	repo := NewClientRepository(f.db)
	if err := repo.ReplaceClientTags(ctx, f.clientID, []string{"new", "vip"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.ReplaceClientTags(ctx, f.clientID, []string{"same", "same"}); err == nil {
		t.Fatal("expected duplicate constraint")
	}
	got, err := repo.GetClientTags(ctx, f.clientID)
	if err != nil || !reflect.DeepEqual(got, []string{"new", "vip"}) {
		t.Fatalf("rollback: %+v %v", got, err)
	}
	if err := repo.ReplaceClientTags(ctx, f.clientID, []string{"loyal"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.DeleteClientById(ctx, f.clientID); err != nil {
		t.Fatal(err)
	}
	got, err = repo.GetClientTags(ctx, f.clientID)
	if err != nil || len(got) != 0 {
		t.Fatalf("cascade: %+v %v", got, err)
	}
}
