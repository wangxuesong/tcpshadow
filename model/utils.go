package model

type Interceptor interface {
	Match(savePackage *SavePackage)
	Intercept(savePackage *SavePackage) []byte
}

type SelectTabFilter struct {
	command SqliCommand
	skill   []SqliTransmission
}

func (f *SelectTabFilter) Match(savePackage *SavePackage) {
	return
}

func (f *SelectTabFilter) Intercept(savePackage *SavePackage) []byte {
	if savePackage.Number == 104 {
		tuple := &SqliTuple{
			Warnings: 0,
		}
		value := &LVarcharTupleValue{Value: "sweethui"}
		tuple.Values = append(tuple.Values, value)
		done := &SqliDone{
			Warning:  16,
			Rows:     1,
			RowID:    257,
			SerialID: 0,
		}
		cost := &SqliCost{
			EstimatedRows: 1,
			EstimatedIO:   2,
		}
		eot := &SqliEot{}

		trans := SqliTransmission{}
		trans.Append(tuple)
		trans.Append(done)
		trans.Append(cost)
		trans.Append(eot)
		buf, err := trans.Pack()
		if err != nil {
			return savePackage.Buffer
		}
		return buf
	}
	return savePackage.Buffer
}

func NewInterceptor() Interceptor {
	filter := new(SelectTabFilter)
	prepare := SqliPrepare{
		QMarks: 0,
		Sql:    "select * from test:tab",
	}
	filter.command = &prepare
	return filter
}
