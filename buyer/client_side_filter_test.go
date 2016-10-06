package buyer

import (
	"testing"

	"github.com/mtlynch/gofn-prosper/interval"
	"github.com/mtlynch/gofn-prosper/prosper"
)

func TestClientSideFilter(t *testing.T) {
	var tests = []struct {
		listing prosper.Listing
		filter  ClientSideFilter
		want    bool
	}{
		{
			listing: prosper.Listing{},
			filter:  ClientSideFilter{},
			want:    true,
		},
		{
			listing: prosper.Listing{
				PriorProsperLoansLatePaymentsOneMonthPlus: 5,
			},
			filter: ClientSideFilter{
				PriorProsperLoansLatePaymentsOneMonthPlus: interval.NewInt32Range(0, 4),
			},
			want: false,
		},
		{
			listing: prosper.Listing{
				PriorProsperLoansLatePaymentsOneMonthPlus: 1,
			},
			filter: ClientSideFilter{
				PriorProsperLoansLatePaymentsOneMonthPlus: interval.NewInt32Range(2, 6),
			},
			want: false,
		},
		{
			listing: prosper.Listing{
				PriorProsperLoansLatePaymentsOneMonthPlus: 2,
			},
			filter: ClientSideFilter{
				PriorProsperLoansLatePaymentsOneMonthPlus: interval.NewInt32Range(2, 6),
			},
			want: true,
		},
		{
			listing: prosper.Listing{
				PriorProsperLoansLatePaymentsOneMonthPlus: 6,
			},
			filter: ClientSideFilter{
				PriorProsperLoansLatePaymentsOneMonthPlus: interval.NewInt32Range(2, 6),
			},
			want: true,
		},
		{
			listing: prosper.Listing{
				PriorProsperLoansBalanceOutstanding: 100.0,
			},
			filter: ClientSideFilter{
				PriorProsperLoansBalanceOutstanding: interval.NewFloat64Range(105.0, 106.0),
			},
			want: false,
		},
		{
			listing: prosper.Listing{
				CurrentDelinquencies: 3,
			},
			filter: ClientSideFilter{
				CurrentDelinquencies: interval.NewInt32Range(0, 2),
			},
			want: false,
		},
		{
			listing: prosper.Listing{
				InquiriesLast6Months: 3,
			},
			filter: ClientSideFilter{
				InquiriesLast6Months: interval.NewInt32Range(0, 2),
			},
			want: false,
		},
		{
			listing: prosper.Listing{
				EmploymentStatusDescription: "Has a great job",
			},
			filter: ClientSideFilter{
				EmploymentStatusDescriptionBlacklist: []string{"Unemployed"},
			},
			want: true,
		},
		{
			listing: prosper.Listing{
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
