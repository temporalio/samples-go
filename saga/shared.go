package saga

const TransferMoneyTaskQueue = "TRANSFER_MONEY_TASK_QUEUE"

type TransferDetails struct {
	Amount      float32
	FromAccount string
	ToAccount   string
	ReferenceID string
}
