package graph

type Field struct {
	FieldType string `json:"field_type,omitempty"`
	IntLen    int64  `json:"int_len,omitempty"`
	Length    int64  `json:"length,omitempty"`
	LogicId   int64  `json:"logic_id,omitempty"`
	Name      string `json:"name,omitempty"`
	RecOffset int64  `json:"rec_offset,omitempty"`
	Reverse   bool   `json:"reverse,omitempty"`
}
type Fs struct {
	Device    string  `json:"device,omitempty"`
	Extension string  `json:"extension,omitempty"`
	FileType  string  `json:"file_type,omitempty"`
	Location  string  `json:"location,omitempty"`
	Name      string  `json:"name,omitempty"`
	Offsets   []int64 `json:"offsets,omitempty"`
	Path      string  `json:"path,omitempty"`
	Share     string  `json:"share,omitempty"`
	Spare     []int64 `json:"spare,omitempty"`
}
type FLogic struct {
	FieldType string `json:"field_type,omitempty"`
	IntLen    int64  `json:"int_len,omitempty"`
	Length    int64  `json:"length,omitempty"`
	LogicId   int64  `json:"logic_id,omitempty"`
	Name      string `json:"name,omitempty"`
	RecOffset int64  `json:"rec_offset,omitempty"`
	Reverse   bool   `json:"reverse,omitempty"`
}
type Indexe struct {
	Audit    int64     `json:"audit,omitempty"`
	FLogic   []*FLogic `json:"f_logic,omitempty"`
	GLogic   int64     `json:"g_logic,omitempty"`
	Global   int64     `json:"global,omitempty"`
	ILogic   int64     `json:"i_logic,omitempty"`
	Ignore   int64     `json:"ignore,omitempty"`
	IndexNum int64     `json:"index_num,omitempty"`
	Length   []int64   `json:"length,omitempty"`
	Name     string    `json:"name,omitempty"`
	Nonblank int64     `json:"nonblank,omitempty"`
	Primary  int64     `json:"primary,omitempty"`
	Reverse  int64     `json:"reverse,omitempty"`
	Shared   int64     `json:"shared,omitempty"`
	Spare    int64     `json:"spare,omitempty"`
	Start    []int64   `json:"start,omitempty"`
	Unique   int64     `json:"unique,omitempty"`
}
type Table struct {
	Fields   []*Field  `json:"fields,omitempty"`
	Fs       *Fs       `json:"fs,omitempty"`
	FstIdx   int64     `json:"fst_idx,omitempty"`
	Id       int64     `json:"id,omitempty"`
	Indexes  []*Indexe `json:"indexes,omitempty"`
	Level    int64     `json:"level,omitempty"`
	Name     string    `json:"name,omitempty"`
	ParentId int64     `json:"parent_id,omitempty"`
	RecLen   int64     `json:"rec_len,omitempty"`
}
