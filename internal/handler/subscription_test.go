package handler

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestParseMonthYear(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     string
		wantYear  int
		wantMonth time.Month
		wantDay   int
		wantErr   bool
	}{
		{
			name:      "valid month year",
			value:     "07-2025",
			wantYear:  2025,
			wantMonth: time.July,
			wantDay:   1,
			wantErr:   false,
		},
		{
			name:    "invalid format",
			value:   "2025-07",
			wantErr: true,
		},
		{
			name:    "empty value",
			value:   "",
			wantErr: true,
		},
		{
			name:    "invalid month",
			value:   "13-2025",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseMonthYear(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Errorf("want error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("want no error, got %v", err)
			}

			if got.Year() != tt.wantYear {
				t.Errorf("expected year %d, got %d", tt.wantYear, got.Year())
			}

			if got.Month() != tt.wantMonth {
				t.Errorf("expected month %d, got %d", tt.wantMonth, got.Month())
			}

			if got.Day() != tt.wantDay {
				t.Errorf("expected day %d, got %d", tt.wantDay, got.Day())
			}
		})
	}
}

func TestParseOptionalMonthYear(t *testing.T) {
	t.Parallel()

	t.Run("empty value returns false", func(t *testing.T) {
		t.Parallel()

		got, ok, err := parseOptionalMonthYear("")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if ok {
			t.Fatal("expected ok to be false")
		}

		if !got.IsZero() {
			t.Fatalf("expected zero time, got %v", got)
		}
	})

	t.Run("valid value returns date", func(t *testing.T) {
		t.Parallel()

		got, ok, err := parseOptionalMonthYear("12-2025")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !ok {
			t.Fatal("expected ok to be true")
		}

		if got.Year() != 2025 {
			t.Fatalf("expected year 2025, got %d", got.Year())
		}

		if got.Month() != time.December {
			t.Fatalf("expected month December, got %v", got.Month())
		}

		if got.Day() != 1 {
			t.Fatalf("expected day 1, got %d", got.Day())
		}
	})

	t.Run("invalid value returns error", func(t *testing.T) {
		t.Parallel()

		got, ok, err := parseOptionalMonthYear("2025-12")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if ok {
			t.Fatal("expected ok to be false")
		}

		if !got.IsZero() {
			t.Fatalf("expected zero time, got %v", got)
		}
	})
}

func TestParseID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rawID   string
		wantID  int64
		wantErr bool
	}{
		{
			name:    "valid id",
			rawID:   "1",
			wantID:  1,
			wantErr: false,
		},
		{
			name:    "zero id",
			rawID:   "0",
			wantErr: true,
		},
		{
			name:    "negative id",
			rawID:   "-1",
			wantErr: true,
		},
		{
			name:    "empty id",
			rawID:   "",
			wantErr: true,
		},
		{
			name:    "not a number",
			rawID:   "hello",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseID(tt.rawID)
			if tt.wantErr {
				if err == nil {
					t.Errorf("want error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("want no error, got %v", err)
			}

			if got != tt.wantID {
				t.Errorf("expected id %d, got %d", tt.wantID, got)
			}
		})
	}
}

func TestBuildCreateSubscriptionParams(t *testing.T) {
	t.Parallel()

	userID := "60601fee-2bf1-4721-ae6f-7636e79a0cba"

	t.Run("valid request", func(t *testing.T) {
		t.Parallel()

		req := CreateSubscriptionRequest{
			ServiceName: "  Yandex Plus",
			Price:       400,
			UserID:      userID,
			StartDate:   "07-2025",
			EndDate:     "12-2025",
		}

		got, err := buildCreateSubscriptionParams(req)
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}

		if got.ServiceName != "Yandex Plus" {
			t.Errorf("want Yandex Plus, got %s", got.ServiceName)
		}

		if got.Price != 400 {
			t.Errorf("want 400, got %d", got.Price)
		}

		if got.UserID != uuid.MustParse(userID) {
			t.Errorf("want %s, got %s", userID, got.UserID)
		}

		if got.StartDate.Format("01-2006") != "07-2025" {
			t.Errorf("want 07-2025, got %s", got.StartDate.Format("01-2006"))
		}

		if got.EndDate == nil {
			t.Errorf("want end date, got nil")
		}

		if got.EndDate.Format("01-2006") != "12-2025" {
			t.Errorf("want 12-2025, got %s", got.EndDate.Format("01-2006"))
		}
	})

	t.Run("valid request with no end date", func(t *testing.T) {
		t.Parallel()

		req := CreateSubscriptionRequest{
			ServiceName: "Netflix",
			Price:       999,
			UserID:      userID,
			StartDate:   "07-2025",
		}

		got, err := buildCreateSubscriptionParams(req)
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}

		if got.EndDate != nil {
			t.Errorf("want nil end date, got %v", got.EndDate)
		}
	})
}

func TestBuildCreateSubscriptionParamsErrors(t *testing.T) {
	t.Parallel()

	validUserID := "60601fee-2bf1-4721-ae6f-7636e79a0cba"

	tests := []struct {
		name string
		req  CreateSubscriptionRequest
	}{
		{
			name: "empty service name",
			req: CreateSubscriptionRequest{
				ServiceName: "",
				Price:       400,
				UserID:      validUserID,
				StartDate:   "07-2025",
			},
		},
		{
			name: "blank service name",
			req: CreateSubscriptionRequest{
				ServiceName: "  ",
				Price:       400,
				UserID:      validUserID,
				StartDate:   "07-2025",
			},
		},
		{
			name: "zero price",
			req: CreateSubscriptionRequest{
				ServiceName: "Yandex Plus",
				Price:       0,
				UserID:      validUserID,
				StartDate:   "07-2025",
			},
		},
		{
			name: "negative price",
			req: CreateSubscriptionRequest{
				ServiceName: "Yandex Plus",
				Price:       -400,
				UserID:      validUserID,
				StartDate:   "07-2025",
			},
		},
		{
			name: "invalid start date",
			req: CreateSubscriptionRequest{
				ServiceName: "Yandex Plus",
				Price:       400,
				UserID:      validUserID,
				StartDate:   "2025-07",
			},
		},
		{
			name: "invalid end date",
			req: CreateSubscriptionRequest{
				ServiceName: "Yandex Plus",
				Price:       400,
				UserID:      validUserID,
				StartDate:   "07-2025",
				EndDate:     "2025-12",
			},
		},
		{
			name: "end date before start date",
			req: CreateSubscriptionRequest{
				ServiceName: "Yandex Plus",
				Price:       400,
				UserID:      validUserID,
				StartDate:   "07-2025",
				EndDate:     "06-2025",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := buildCreateSubscriptionParams(tt.req)
			if err == nil {
				t.Errorf("want error, got nil")
			}
		})
	}
}

func TestBuildUpdateSubscriptionParams(t *testing.T) {
	t.Parallel()

	userID := "60601fee-2bf1-4721-ae6f-7636e79a0cba"

	req := UpdateSubscriptionRequest{
		ServiceName: "Yandex Plus",
		Price:       500,
		UserID:      userID,
		StartDate:   "07-2025",
		EndDate:     "12-2025",
	}

	got, err := buildUpdateSubscriptionParams(42, req)
	if err != nil {
		t.Errorf("want no error, got %v", err)
	}

	if got.ID != 42 {
		t.Errorf("want id 42, got %d", got.ID)
	}

	if got.ServiceName != "Yandex Plus" {
		t.Errorf("want Yandex Plus, got %s", got.ServiceName)
	}

	if got.Price != 500 {
		t.Errorf("want 500, got %d", got.Price)
	}

	if got.UserID != uuid.MustParse(userID) {
		t.Errorf("want user id %s, got %s", userID, got.UserID)
	}

	if got.StartDate.Format("01-2006") != "07-2025" {
		t.Errorf("want date 07-2025, got %s", got.StartDate.Format("01-2006"))
	}

	if got.EndDate == nil {
		t.Errorf("want end date, got nil")
	}

	if got.EndDate.Format("01-2006") != "12-2025" {
		t.Errorf("want date 12-2025, got %s", got.EndDate.Format("01-2006"))
	}
}

func TestFormatMonthYear(t *testing.T) {
	t.Parallel()

	date := time.Date(2025, time.July, 1, 0, 0, 0, 0, time.UTC)

	got := formatMonthYear(date)
	want := "07-2025"

	if got != want {
		t.Errorf("want date %s, got %s", want, got)
	}
}
