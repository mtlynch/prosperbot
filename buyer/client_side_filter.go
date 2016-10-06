package buyer

import (
	"github.com/mtlynch/gofn-prosper/interval"
	"github.com/mtlynch/gofn-prosper/prosper"
)

type ClientSideFilter struct {
	PriorProsperLoansLatePaymentsOneMonthPlus interval.Int32Range
	PriorProsperLoansBalanceOutstanding       interval.Float64Range
	CurrentDelinquencies                      interval.Int32Range
	InquiriesLast6Months                      interval.Int32Range
	EmploymentStatusDescriptionBlacklist      []string
}

func (csf ClientSideFilter) Filter(l prosper.Listing) bool {
	if !isInInt32Range(csf.PriorProsperLoansLatePaymentsOneMonthPlus, int32(l.PriorProsperLoansLatePaymentsOneMonthPlus)) {
		return false
	}
	if !isInFloat64Range(csf.PriorProsperLoansBalanceOutstanding, l.PriorProsperLoansBalanceOutstanding) {
		return false
	}
	if !isInInt32Range(csf.CurrentDelinquencies, int32(l.CurrentDelinquencies)) {
		return false
	}
	if !isInInt32Range(csf.InquiriesLast6Months, int32(l.InquiriesLast6Months)) {
		return false
	}
	for _, blacklisted := range csf.EmploymentStatusDescriptionBlacklist {
		if l.EmploymentStatusDescription == blacklisted {
			return false
		}
	}
	return true
}

func isInInt32Range(r interval.Int32Range, v int32) bool {
	if r.Max != nil && v > *r.Max {
		return false
	}
	if r.Min != nil && v < *r.Min {
		return false
	}
	return true
}

func isInFloat64Range(r interval.Float64Range, v float64) bool {
	if r.Max != nil && v > *r.Max {
		return false
	}
	if r.Min != nil && v < *r.Min {
		return false
	}
	return true
}
