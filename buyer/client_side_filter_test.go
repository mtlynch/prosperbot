package buyer

import (
	"testing"

	"github.com/mtlynch/gofn-prosper/types"
)

func TestClientSideFilter(t *testing.T) {
	var tests = []struct {
		listing types.Listing
		filter  ClientSideFilter
		want    bool
	}{
		{
			listing: types.Listing{},
			filter:  ClientSideFilter{},
			want:    true,
		},
		{
			listing: types.Listing{
				PriorProsperLoansLatePaymentsOneMonthPlus: 5,
			},
			filter: ClientSideFilter{
				PriorProsperLoansLatePaymentsOneMonthPlus: types.NewInt32Range(0, 4),
			},
			want: false,
		},
		{
			listing: types.Listing{
				PriorProsperLoansLatePaymentsOneMonthPlus: 1,
			},
			filter: ClientSideFilter{
				PriorProsperLoansLatePaymentsOneMonthPlus: types.NewInt32Range(2, 6),
			},
			want: false,
		},
		{
			listing: types.Listing{
				PriorProsperLoansLatePaymentsOneMonthPlus: 2,
			},
			filter: ClientSideFilter{
				PriorProsperLoansLatePaymentsOneMonthPlus: types.NewInt32Range(2, 6),
			},
			want: true,
		},
		{
			listing: types.Listing{
				PriorProsperLoansLatePaymentsOneMonthPlus: 6,
			},
			filter: ClientSideFilter{
				PriorProsperLoansLatePaymentsOneMonthPlus: types.NewInt32Range(2, 6),
			},
			want: true,
		},
		{
			listing: types.Listing{
				PriorProsperLoansBalanceOutstanding: 100.0,
			},
			filter: ClientSideFilter{
				PriorProsperLoansBalanceOutstanding: types.NewFloat64Range(105.0, 106.0),
			},
			want: false,
		},
		{
			listing: types.Listing{
				CurrentDelinquencies: 3,
			},
			filter: ClientSideFilter{
				CurrentDelinquencies: types.NewInt32Range(0, 2),
			},
			want: false,
		},
		{
			listing: types.Listing{
				InquiriesLast6Months: 3,
			},
			filter: ClientSideFilter{
				InquiriesLast6Months: types.NewInt32Range(0, 2),
			},
			want: false,
		},
		{
			listing: types.Listing{
				EmploymentStatusDescription: "Has a great job",
			},
			filter: ClientSideFilter{
				EmploymentStatusDescriptionBlacklist: []string{"Unemployed"},
			},
			want: true,
		},
		{
			listing: types.Listing{
				EmploymentStatusDescription: "Unemployed",
			},
			filter: ClientSideFilter{
				EmploymentStatusDescriptionBlacklist: []string{"Unemployed"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		got := tt.filter.Filter(tt.listing)
		if got != tt.want {
			t.Errorf("unexpected client side filter result for listing: %+v and filter: %+v. got = %v, want = %v", tt.listing, tt.filter, got, tt.want)
		}
	}
}
