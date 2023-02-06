package batch_sliding_window

type RecordLoader struct {
	RecordCount int
}

type SingleRecord struct {
	Id int
}

func (p *RecordLoader) GetRecords(pageSize, offset int) (result []SingleRecord, err error) {
	if offset < p.RecordCount {
		size := offset + pageSize
		if size > p.RecordCount {
			size = p.RecordCount
		}
		for i := 0; i < size; i++ {
			result = append(result, SingleRecord{Id: i})
		}
	}
	return
}
