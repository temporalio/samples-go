package batch_sliding_window

import "fmt"

type (
	// RecordLoader activities structure.
	RecordLoader struct {
		RecordCount int
	}

	GetRecordsInput struct {
		PageSize  int
		Offset    int
		MaxOffset int
	}

	SingleRecord struct {
		Id int
	}

	GetRecordsOutput struct {
		Records []SingleRecord
	}
)

// GetRecordCount activity returns the total record count.
// Used to partition processing across parallel sliding windows.
// The sample implementation just returns a fake value passed during worker initialization.
func (p *RecordLoader) GetRecordCount() (int, error) {
	return p.RecordCount, nil
}

// GetRecords activity returns records loaded from an external data source. The sample returns fake records.
func (p *RecordLoader) GetRecords(input GetRecordsInput) (output GetRecordsOutput, err error) {
	if input.MaxOffset > p.RecordCount {
		panic(fmt.Sprintf("maxOffset(%d)>recordCount(%d", input.MaxOffset, p.RecordCount))
	}
	var records []SingleRecord
	limit := input.Offset + input.PageSize
	if limit > input.MaxOffset {
		limit = input.MaxOffset
	}
	for i := input.Offset; i < limit; i++ {
		records = append(records, SingleRecord{Id: i})
	}
	return GetRecordsOutput{Records: records}, nil
}
