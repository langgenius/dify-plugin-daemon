package entities

type FromMapper interface {
	FromMap(map[string]any) error
}
