package levelredis

type LevelElem interface {
	Type() string
	Drop() bool
}
