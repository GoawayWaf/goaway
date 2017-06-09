package obj

//DataSet is an array of LogLine to be analyzed
type Evidence struct {
	lines []LogLine
}

func NewEvidence() *Evidence {
	return &Evidence{}
}

func (d *Evidence) Add(line LogLine) {
	d.lines = append(d.lines, line)
}

func (d *Evidence) Get() []LogLine {
	return d.lines
}

func (d *Evidence) Count() int {
	return len(d.lines)
}
