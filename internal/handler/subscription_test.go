package handler

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/spoddub/subscription-aggregator/internal/model"
)

const userIDConst = "60601fee-2bf1-4721-ae6f-7636e79a0cba"

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

	userID := userIDConst

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

	validUserID := userIDConst

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

	userID := uuid.MustParse(userIDConst)
	startDate, err := parseMonthYear("07-2025")
	if err != nil {
		t.Fatalf("failed to parse start date: %v", err)
	}

	current := model.Subscription{
		ID:          42,
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     nil,
	}

	t.Run("updates only price", func(t *testing.T) {
		t.Parallel()

		req := UpdateSubscriptionRequest{
			Price: intPtr(500),
		}

		got, err := buildUpdateSubscriptionParams(42, current, req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if got.ID != 42 {
			t.Fatalf("expected id 42, got %d", got.ID)
		}

		if got.ServiceName != current.ServiceName {
			t.Fatalf("expected service name %q, got %q", current.ServiceName, got.ServiceName)
		}

		if got.Price != 500 {
			t.Fatalf("expected price 500, got %d", got.Price)
		}

		if got.UserID != current.UserID {
			t.Fatalf("expected user id %s, got %s", current.UserID, got.UserID)
		}

		if !got.StartDate.Equal(current.StartDate) {
			t.Fatalf("expected start date %v, got %v", current.StartDate, got.StartDate)
		}

		if got.EndDate != nil {
			t.Fatalf("expected nil end date, got %v", got.EndDate)
		}
	})

	t.Run("updates several fields", func(t *testing.T) {
		t.Parallel()

		newUserID := "2e0d4f8d-3d7c-4b22-8d6b-296f6c4a22b2"

		req := UpdateSubscriptionRequest{
			ServiceName: stringPtr("  Netflix  "),
			Price:       intPtr(999),
			UserID:      stringPtr(newUserID),
			StartDate:   stringPtr("08-2025"),
			EndDate:     stringPtr("12-2025"),
		}

		got, err := buildUpdateSubscriptionParams(42, current, req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if got.ServiceName != "Netflix" {
			t.Fatalf("expected service name Netflix, got %q", got.ServiceName)
		}

		if got.Price != 999 {
			t.Fatalf("expected price 999, got %d", got.Price)
		}

		if got.UserID != uuid.MustParse(newUserID) {
			t.Fatalf("expected user id %s, got %s", newUserID, got.UserID)
		}

		if got.StartDate.Format("01-2006") != "08-2025" {
			t.Fatalf("expected start date 08-2025, got %s", got.StartDate.Format("01-2006"))
		}

		if got.EndDate == nil {
			t.Fatal("expected end date, got nil")
		}

		if got.EndDate.Format("01-2006") != "12-2025" {
			t.Fatalf("expected end date 12-2025, got %s", got.EndDate.Format("01-2006"))
		}
	})

	t.Run("clears end date with empty string", func(t *testing.T) {
		t.Parallel()

		endDate, err := parseMonthYear("12-2025")
		if err != nil {
			t.Fatalf("failed to parse end date: %v", err)
		}

		currentWithEndDate := current
		currentWithEndDate.EndDate = &endDate

		req := UpdateSubscriptionRequest{
			EndDate: stringPtr(""),
		}

		got, err := buildUpdateSubscriptionParams(42, currentWithEndDate, req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if got.EndDate != nil {
			t.Fatalf("expected nil end date, got %v", got.EndDate)
		}
	})
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

func TestParseLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		rawLimit  string
		wantLimit int
		wantErr   bool
	}{
		{
			name:      "empty limit returns default",
			rawLimit:  "",
			wantLimit: defaultListLimit,
			wantErr:   false,
		},
		{
			name:      "valid limit",
			rawLimit:  "10",
			wantLimit: 10,
			wantErr:   false,
		},
		{
			name:     "zero limit",
			rawLimit: "0",
			wantErr:  true,
		},
		{
			name:     "negative limit",
			rawLimit: "-1",
			wantErr:  true,
		},
		{
			name:     "limit too large",
			rawLimit: "101",
			wantErr:  true,
		},
		{
			name:     "not a number",
			rawLimit: "abc",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseLimit(tt.rawLimit)
			if tt.wantErr {
				if err == nil {
					t.Errorf("want error, got nil")
				}

				return
			}

			if err != nil {
				t.Errorf("want no error, got %v", err)
			}

			if got != tt.wantLimit {
				t.Errorf("want %d, got %d", tt.wantLimit, got)
			}
		})
	}
}

func TestParseOffset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		rawOffset  string
		wantOffset int
		wantErr    bool
	}{
		{
			name:       "empty offset returns default",
			rawOffset:  "",
			wantOffset: defaultOffset,
			wantErr:    false,
		},
		{
			name:       "valid offset",
			rawOffset:  "10",
			wantOffset: 10,
			wantErr:    false,
		},
		{
			name:       "zero offset",
			rawOffset:  "0",
			wantOffset: 0,
			wantErr:    false,
		},
		{
			name:      "negative offset",
			rawOffset: "-1",
			wantErr:   true,
		},
		{
			name:      "not a number",
			rawOffset: "abc",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseOffset(tt.rawOffset)
			if tt.wantErr {
				if err == nil {
					t.Errorf("want error, got nil")
				}

				return
			}

			if err != nil {
				t.Errorf("want no error, got %v", err)
			}

			if got != tt.wantOffset {
				t.Errorf("want %d, got %d", tt.wantOffset, got)
			}
		})
	}
}

func TestBuildUpdateSubscriptionParamsErrors(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse(userIDConst)
	startDate, err := parseMonthYear("07-2025")
	if err != nil {
		t.Fatalf("failed to parse start date: %v", err)
	}

	current := model.Subscription{
		ID:          42,
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     nil,
	}

	tests := []struct {
		name string
		req  UpdateSubscriptionRequest
	}{
		{
			name: "empty request",
			req:  UpdateSubscriptionRequest{},
		},
		{
			name: "empty service name",
			req: UpdateSubscriptionRequest{
				ServiceName: stringPtr("   "),
			},
		},
		{
			name: "zero price",
			req: UpdateSubscriptionRequest{
				Price: intPtr(0),
			},
		},
		{
			name: "negative price",
			req: UpdateSubscriptionRequest{
				Price: intPtr(-100),
			},
		},
		{
			name: "invalid user id",
			req: UpdateSubscriptionRequest{
				UserID: stringPtr("invalid user id"),
			},
		},
		{
			name: "invalid start date",
			req: UpdateSubscriptionRequest{
				StartDate: stringPtr("2025-07"),
			},
		},
		{
			name: "end date before start date",
			req: UpdateSubscriptionRequest{
				EndDate: stringPtr("06-2025"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := buildUpdateSubscriptionParams(42, current, tt.req)
			if err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
