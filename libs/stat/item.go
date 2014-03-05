package stat

type textItem struct {
	Item
	title   string
	padding int
	value   interface{}
}

func (s *textItem) Title() string {
	return s.title
}

func (s *textItem) Padding() int {
	return s.padding
}

func (s *textItem) Value() interface{} {
	return s.value
}

type incrItem struct {
	Item
	title     string
	padding   int
	fn        func() int64
	lastCount int64
}

func (c *incrItem) Title() string {
	return c.title
}

func (c *incrItem) Padding() int {
	return c.padding
}

func (c *incrItem) Value() interface{} {
	var chg int64
	cur := c.fn()
	chg, c.lastCount = cur-c.lastCount, cur
	return chg
}

func TextItem(title string, padding int, value interface{}) Item {
	return &textItem{
		title:   title,
		padding: padding,
		value:   value,
	}
}

func IncrItem(title string, padding int, fn func() int64) Item {
	return &incrItem{
		title:   title,
		padding: padding,
		fn:      fn,
	}
}
