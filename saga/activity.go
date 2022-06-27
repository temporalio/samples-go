package saga

import (
	"context"
	"errors"
	"fmt"
)

func Withdraw(ctx context.Context, transferDetails TransferDetails) error {
	fmt.Printf(
		"\nWithdrawing $%f from account %s. ReferenceId: %s\n",
		transferDetails.Amount,
		transferDetails.FromAccount,
		transferDetails.ReferenceID,
	)

	return nil
}

func WithdrawCompensation(ctx context.Context, transferDetails TransferDetails) error {
	fmt.Printf(
		"\nWithdrawing compensation $%f from account %s. ReferenceId: %s\n",
		transferDetails.Amount,
		transferDetails.FromAccount,
		transferDetails.ReferenceID,
	)

	return nil
}

func Deposit(ctx context.Context, transferDetails TransferDetails) error {
	fmt.Printf(
		"\nDepositing $%f into account %s. ReferenceId: %s\n",
		transferDetails.Amount,
		transferDetails.ToAccount,
		transferDetails.ReferenceID,
	)

	return nil
}

func DepositCompensation(ctx context.Context, transferDetails TransferDetails) error {
	fmt.Printf(
		"\nDepositing compensation $%f into account %s. ReferenceId: %s\n",
		transferDetails.Amount,
		transferDetails.ToAccount,
		transferDetails.ReferenceID,
	)

	return nil
}

func StepWithError(ctx context.Context, transferDetails TransferDetails) error {
	fmt.Printf(
		"\nSimulate failure to trigger compensation. ReferenceId: %s\n",
		transferDetails.ReferenceID,
	)

	return errors.New("some error")
}
