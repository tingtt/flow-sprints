package sprint

type Sprint struct {
	Id          uint64  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Start       string  `json:"start"`
	End         string  `json:"end"`
	ProjectId   *uint64 `json:"project_id,omitempty"`
}
