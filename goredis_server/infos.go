package goredis_server

type ReplicationInfo struct {
	IsMaster         bool
	IsSlave          bool
	MasterHost       string
	MasterPort       string
	MasterLinkStatus string // pending|up|down
}
