package batch_sliding_window

import (
	"fmt"
)

type RecordLoader struct {
	RecordCount int
}

type SingleRecord struct {
	Id string
}

func (p *RecordLoader) GetRecords(pageSize, offset int) (result []SingleRecord, err error) {
	if offset < p.RecordCount {
		size := offset + pageSize
		if size > p.RecordCount {
			size = p.RecordCount
		}
		for i := 0; i < size; i++ {
			recordId := fmt.Sprintf("%d", i)
			result = append(result, SingleRecord{Id: recordId})
		}
	}
	return
}
