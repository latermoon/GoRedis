package storage

type EntryType int

const (
	EntryTypeUnknown   = 0
	EntryTypeString    = 1
	EntryTypeHash      = 2
	EntryTypeList      = 3
	EntryTypeSet       = 4
	EntryTypeSortedSet = 5
)

type Entry struct {
	Value interface{}
	Type  EntryType
	size  int
}

func (e *Entry) Size() int {
	return e.size
}

func NewEntry(value interface{}, entryType EntryType) (e *Entry) {
	e = &Entry{}
	e.Value = value
	e.Type = entryType
	return
}
