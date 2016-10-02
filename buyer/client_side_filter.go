package buyer

import "github.com/mtlynch/gofn-prosper/types"

type ClientSideFilter struct {
	PriorProsperLoansLatePaymentsOneMonthPlus types.Int32Range
	PriorProsperLoansBalanceOutstanding       types.Float64Range
	CurrentDelinquencies                      types.Int32Range
	InquiriesLast6Months                      types.Int32Range
	EmploymentStatusDescriptionBlacklist      []string
}

func (csf ClientSideFilter) Filter(l types.Listing) bool {
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

func isInInt32Range(r types.Int32Range, v int32) bool {
	if r.Max != nil && v > *r.Max {
		return false
	}
	if r.Min != nil && v < *r.Min {
		return false
	}
	return true
}

func isInFloat64Range(r types.Float64Range, v float64) bool {
	if r.Max != nil && v > *r.Max {
		return false
	}
	if r.Min != nil && v < *r.Min {
		return false
	}
	return true
}
