package loan_test

import (
	"context"
	"testing"

	"task109-loanamort/internal/loan"
	"task109-loanamort/internal/store"
)

func TestRateChangePreservesRemainingTermBoundary(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(t.TempDir() + "/loan.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	svc := loan.New(db)
	b, err := svc.CreateBorrower(ctx, loan.CreateBorrowerRequest{Name: "borrower"})
	if err != nil {
		t.Fatal(err)
	}
	l, err := svc.CreateLoan(ctx, loan.CreateLoanRequest{BorrowerID: b.BorrowerID, PrincipalCents: 1000000, AnnualPercent: 12, Periods: 12, Type: loan.EqualInstallment})
	if err != nil {
		t.Fatal(err)
	}
	s, err := svc.Schedule(ctx, l.LoanID)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = svc.RecordPayment(ctx, l.LoanID, loan.RecordPaymentRequest{AmountCents: s.Periods[0].Payment}); err != nil {
		t.Fatal(err)
	}
	if _, err = svc.ChangeRate(ctx, l.LoanID, loan.ChangeRateRequest{AnnualPercent: 24}); err != nil {
		t.Fatal(err)
	}
	p, err := svc.Projection(ctx, l.LoanID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if p.Remaining != 11 {
		t.Fatalf("remaining=%d, want 11", p.Remaining)
	}
	if len(p.Periods) == 0 || p.Periods[0].Payment <= s.Periods[1].Payment {
		t.Fatalf("new payment=%v, want greater than old %d", p.Periods, s.Periods[1].Payment)
	}
}
