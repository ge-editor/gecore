package tree

func NewViews() *ViewsStruct {
	return &ViewsStruct{names: map[string]int{}}
}

type ViewsStruct struct {
	views []View
	names map[string]int
}

// Register View
func (vs *ViewsStruct) Register(view View) {
	index := len(vs.views)
	vs.views = append(vs.views, view)
	vs.names[view.Name()] = index // Allow view to be retrieved from view name
}

func (vs ViewsStruct) GetViews() []View {
	return vs.views
}

// Return *View by View Name
func (vs *ViewsStruct) GetViewByName(name string) *View {
	return &(vs.views[vs.names[name]])
}

// Return the first registered *View as default view
func (vs *ViewsStruct) GetDefaultView() *View {
	return &(vs.views[0])
}
