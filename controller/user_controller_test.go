package controller_test

import(
	"github.com/elias-gill/poliplanner2/controller"
	"github.com/go-chi/chi/v5"
	"testing"
)

func TestNewUsersRouter(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		want func(r chi.Router)
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := controller.NewUsersRouter()
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("NewUsersRouter() = %v, want %v", got, tt.want)
			}
		})
	}
}

