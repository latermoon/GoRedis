package goredis_server

import (
	. "GoRedis/libs/goredis"
	"GoRedis/libs/iotool"
	"GoRedis/libs/levelredis"
	"GoRedis/libs/rdb"
	"GoRedis/libs/stdlog"
	"bufio"
	"errors"
	"fmt"
	"github.com/latermoon/levigo"
	"github.com/latermoon/msgpackgo/codec"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

type SlaveStatus int

type SlaveClientCallback interface {
	RdbSizeCallback(client *SlaveClient, totalsize int64)
	RdbRecvFinishCallback(client *SlaveClient, r *bufio.Reader)
	RdbRecvProcessCallback(client *SlaveClient, size int64, rate int)
	IdleCallback(client *SlaveClient)
	CommandRecvCallback(client *SlaveClient, cmd *Command)
}

var slavelog = stdlog.Log("slaveof")
var msgpack_handle codec.MsgpackHandle

/**

client := NewSlaveClient(...)
client.Sync(uid)
client.Cancel()

*/
type SlaveClient struct {
	session    *Session
	server     *GoRedisServer
	callback   SlaveClientCallback
	levelredis *levelredis.LevelRedis
}

func NewSlaveClient(server *GoRedisServer, session *Session) (s *SlaveClient) {
	s = &SlaveClient{}
	s.server = server
	s.session = session
	return
}

func (s *SlaveClient) SetCallback(callback SlaveClientCallback) {
	s.callback = callback
}

func (s *SlaveClient) RemoteAddr() net.Addr {
	return s.session.RemoteAddr()
}

func (s *SlaveClient) directory() string {
	return s.server.directory + "slaveof_" + fmt.Sprint(s.session.RemoteAddr()) + "/"
}

func (s *SlaveClient) rdbfilename() string {
	return s.directory() + "dump.rdb"
}

// 开始同步
func (s *SlaveClient) Sync(uid string) (err error) {
	isgoredis, version, e1 := s.masterInfo()
	if e1 != nil {
		return e1
	}
	if isgoredis {
		slavelog.Printf("[M %s] slaveof %s GoRedis:%s\n", s.RemoteAddr(), s.RemoteAddr(), version)
	} else {
		slavelog.Printf("[M %s] slaveof %s Redis:%s\n", s.RemoteAddr(), s.RemoteAddr(), version)
	}

	args := [][]byte{[]byte("SYNC")}
	if isgoredis && len(uid) > 0 {
		args = append(args, []byte(uid))
	}
	s.session.WriteCommand(NewCommand(args...))

	rdbsaved := false
	for {
		var c byte
		c, err = s.session.PeekByte()
		if !rdbsaved && c == '$' {
			err = s.recvRdb()
			if err != nil {
				slavelog.Printf("[M %s] recv rdb error:%s\n", s.RemoteAddr(), err)
				break
			}
			rdbsaved = true
			s.initLeveldb() // init for command sync
		} else if c == '\n' {
			s.session.ReadByte()
			s.callback.IdleCallback(s)
		} else {
			var cmd *Command
			cmd, err = s.session.ReadCommand()
			if err != nil {
				break
			}
			s.callback.CommandRecvCallback(s, cmd)
		}
	}
	return
}

func (s *SlaveClient) initLeveldb() {
	opts := levigo.NewOptions()
	opts.SetCache(levigo.NewLRUCache(32 * 1024 * 1024))
	opts.SetCompression(levigo.SnappyCompression)
	opts.SetBlockSize(32 * 1024)
	opts.SetWriteBufferSize(128 * 1024 * 1024)
	opts.SetCreateIfMissing(true)
	db, e1 := levigo.Open(s.directory()+"/db0", opts)
	if e1 != nil {
		panic(e1)
	}
	s.levelredis = levelredis.NewLevelRedis(db)
}

func (s *SlaveClient) recvRdb() (err error) {
	var f *os.File
	os.Mkdir(s.directory(), os.ModePerm)
	f, err = os.OpenFile(s.rdbfilename(), os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return
	}
	slavelog.Printf("[M %s] create rdb:%s\n", s.RemoteAddr(), s.rdbfilename())
	defer func() {
		filename := f.Name()
		f.Close()
		os.Remove(filename)
	}()

	s.session.ReadByte()
	var size int64
	size, err = s.session.ReadLineInteger()
	if err != nil {
		return
	}
	s.callback.RdbSizeCallback(s, size)

	// read
	w := bufio.NewWriter(f)
	// var written int64
	_, err = iotool.RateLimitCopy(w, io.LimitReader(s.session, size), 40*1024*1024, func(written int64, rate int) {
		s.callback.RdbRecvProcessCallback(s, written, rate)
	})
	// _, err = io.CopyN(w, s.session, size)
	if err != nil {
		return
	}
	w.Flush()
	f.Seek(0, 0)
	// callback
	s.callback.RdbRecvFinishCallback(s, bufio.NewReader(f))
	return
}

// 清空本地的同步状态
func (s *SlaveClient) Destory() (err error) {
	return
}

func (s *SlaveClient) rdbFileWriter() (w *bufio.Writer, err error) {
	var file *os.File
	file, err = os.OpenFile(fmt.Sprintf("/tmp/%s.rdb", s.session.RemoteAddr()), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	w = bufio.NewWriter(file)
	return
}

func (s *SlaveClient) masterInfo() (isgoredis bool, version string, err error) {
	cmdinfo := NewCommand([]byte("info"), []byte("server"))
	s.session.WriteCommand(cmdinfo)
	var reply *Reply
	reply, err = s.session.ReadReply()
	if err != nil {
		return
	}
	if reply.Value == nil {
		err = errors.New("reply nil")
		return
	}

	var info string
	switch reply.Value.(type) {
	case string:
		info = reply.Value.(string)
	case []byte:
		info = string(reply.Value.([]byte))
	default:
		info = reply.String()
	}

	// 切分info返回的数据，存放到map里
	kv := make(map[string]string)
	lines := strings.Split(info, "\n")
	for _, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		line = strings.TrimPrefix(line, " ")
		if strings.HasPrefix(line, "#") {
			continue
		}
		pairs := strings.Split(line, ":")
		if len(pairs) != 2 {
			continue
		}
		// done
		kv[pairs[0]] = pairs[1]
	}

	_, isgoredis = kv["goredis_version"]
	if isgoredis {
		version = kv["goredis_version"]
	} else {
		version = kv["redis_version"]
	}

	return
}

// ==============================
// 处理获得的数据
// ==============================
type slaveCallback struct {
	SlaveClientCallback
	server *GoRedisServer
}

func newSlaveCallback(server *GoRedisServer) (s *slaveCallback) {
	s = &slaveCallback{}
	s.server = server
	return
}

func (s *slaveCallback) RdbSizeCallback(client *SlaveClient, totalsize int64) {
	slavelog.Printf("[M %s] rdb size: %d\n", client.RemoteAddr(), totalsize)
	// create leveldb
}

func (s *slaveCallback) RdbRecvFinishCallback(client *SlaveClient, r *bufio.Reader) {
	slavelog.Printf("[M %s] rdb recv finish \n", client.RemoteAddr())

	// init levellist

	// decode
	dec := newRdbDecoder(client)
	err := rdb.Decode(r, dec)
	if err != nil {
		// must cancel
		slavelog.Printf("[M %s] decode error %s\n", client.RemoteAddr(), err)
	}
	return
}

func (s *SlaveClient) rdbDecodeCommand(client *SlaveClient, cmd *Command) {
	slavelog.Printf("[M %s] rdb decode %s\n", client.RemoteAddr(), cmd)
	s.server.On(client.session, cmd)
}

func (s *SlaveClient) rdbDecodeFinish(client *SlaveClient, n int64) {
	slavelog.Printf("[M %s] rdb decode finish, items: %d\n", client.RemoteAddr(), n)
}

func (s *slaveCallback) RdbRecvProcessCallback(client *SlaveClient, size int64, rate int) {
	slavelog.Printf("[M %s] rdb recv: %d, rate:%d\n", client.RemoteAddr(), size, rate)
}

func (s *slaveCallback) IdleCallback(client *SlaveClient) {
	slavelog.Printf("[M %s] slaveof waiting\n", client.RemoteAddr())
}

func (s *slaveCallback) CommandRecvCallback(client *SlaveClient, cmd *Command) {
	slavelog.Printf("[M %s] recv: %s\n", client.RemoteAddr(), cmd)
	if cmd.Name() == "PING" {
		return
	}
	lst := client.levelredis.GetList("cmdlist_0")
	out, err := s.encodeCommand(cmd)
	if err == nil {
		lst.RPush(out)
	} else {
		slavelog.Printf("[M %s] encode error %s, %s\n", client.RemoteAddr(), cmd, err)
	}
}

func (s *slaveCallback) encodeCommand(cmd *Command) (out []byte, err error) {
	enc := codec.NewEncoderBytes(&out, &msgpack_handle)
	err = enc.Encode(cmd.Args)
	return
}

func (s *slaveCallback) decodeCommand(in []byte) (cmd *Command, err error) {
	dec := codec.NewDecoderBytes(in, &msgpack_handle)
	var v interface{}
	err = dec.Decode(&v)
	if err == nil {
		objs, ok := v.([]interface{})
		if !ok {
			err = errors.New("bad command bytes")
			return
		}
		args := make([][]byte, 0, len(objs))
		for i := 0; i < len(objs); i++ {
			args = append(args, objs[i].([]byte))
		}
		cmd = NewCommand(args...)
	}
	return
}

// =============================================
// 第三方rdb解释函数
// =============================================
type rdbDecoder struct {
	rdb.NopDecoder
	db       int
	i        int
	keyCount int64
	bufsize  int
	client   *SlaveClient
	// 数据缓冲
	hashEntry [][]byte
	setEntry  [][]byte
	listEntry [][]byte
	zsetEntry [][]byte
}

func newRdbDecoder(s *SlaveClient) (dec *rdbDecoder) {
	dec = &rdbDecoder{}
	dec.client = s
	dec.keyCount = 0
	dec.bufsize = 200
	return
}

func (p *rdbDecoder) StartDatabase(n int) {
	p.db = n
}

func (p *rdbDecoder) EndDatabase(n int) {
}

func (p *rdbDecoder) EndRDB() {
	p.client.rdbDecodeFinish(p.client, p.keyCount)
}

// Set
func (p *rdbDecoder) Set(key, value []byte, expiry int64) {
	cmd := NewCommand([]byte("SET"), key, value)
	p.client.rdbDecodeCommand(p.client, cmd)
	p.keyCount++
}

func (p *rdbDecoder) StartHash(key []byte, length, expiry int64) {
	if int(length) < p.bufsize {
		p.hashEntry = make([][]byte, 0, length+2)
	} else {
		p.hashEntry = make([][]byte, 0, p.bufsize)
	}
	p.hashEntry = append(p.hashEntry, []byte("HSET"))
	p.hashEntry = append(p.hashEntry, key)
	p.keyCount++
}

func (p *rdbDecoder) Hset(key, field, value []byte) {
	p.hashEntry = append(p.hashEntry, field)
	p.hashEntry = append(p.hashEntry, value)
	if len(p.hashEntry) >= p.bufsize {
		cmd := NewCommand(p.hashEntry...)
		p.client.rdbDecodeCommand(p.client, cmd)
		p.hashEntry = make([][]byte, 0, p.bufsize)
		p.hashEntry = append(p.hashEntry, []byte("HSET"))
		p.hashEntry = append(p.hashEntry, key)
	}
}

// Hash
func (p *rdbDecoder) EndHash(key []byte) {
	if len(p.hashEntry) > 2 {
		cmd := NewCommand(p.hashEntry...)
		p.client.rdbDecodeCommand(p.client, cmd)
	}
}

func (p *rdbDecoder) StartSet(key []byte, cardinality, expiry int64) {
	if int(cardinality) < p.bufsize {
		p.setEntry = make([][]byte, 0, cardinality+2)
	} else {
		p.setEntry = make([][]byte, 0, p.bufsize)
	}
	p.setEntry = append(p.setEntry, []byte("SADD"))
	p.setEntry = append(p.setEntry, key)
	p.keyCount++
}

func (p *rdbDecoder) Sadd(key, member []byte) {
	p.setEntry = append(p.setEntry)
	if len(p.setEntry) >= p.bufsize {
		cmd := NewCommand(p.setEntry...)
		p.client.rdbDecodeCommand(p.client, cmd)
		p.setEntry = make([][]byte, 0, p.bufsize)
		p.setEntry = append(p.setEntry, []byte("SADD"))
		p.setEntry = append(p.setEntry, key)
	}
}

// Set
func (p *rdbDecoder) EndSet(key []byte) {
	if len(p.setEntry) > 2 {
		cmd := NewCommand(p.setEntry...)
		p.client.rdbDecodeCommand(p.client, cmd)
	}
}

func (p *rdbDecoder) StartList(key []byte, length, expiry int64) {
	if int(length) < p.bufsize {
		p.listEntry = make([][]byte, 0, length+2)
	} else {
		p.listEntry = make([][]byte, 0, p.bufsize)
	}
	p.listEntry = append(p.listEntry, []byte("RPUSH"))
	p.listEntry = append(p.listEntry, key)
	p.keyCount++
	p.i = 0
}

func (p *rdbDecoder) Rpush(key, value []byte) {
	p.listEntry = append(p.listEntry, value)
	if len(p.listEntry) >= p.bufsize {
		cmd := NewCommand(p.listEntry...)
		p.client.rdbDecodeCommand(p.client, cmd)
		p.listEntry = make([][]byte, 0, p.bufsize)
		p.listEntry = append(p.listEntry, []byte("RPUSH"))
		p.listEntry = append(p.listEntry, key)
	}
	p.i++
}

// List
func (p *rdbDecoder) EndList(key []byte) {
	if len(p.listEntry) > 2 {
		cmd := NewCommand(p.listEntry...)
		p.client.rdbDecodeCommand(p.client, cmd)
	}
}

func (p *rdbDecoder) StartZSet(key []byte, cardinality, expiry int64) {
	if int(cardinality) > p.bufsize {
		p.zsetEntry = make([][]byte, 0, cardinality)
	} else {
		p.zsetEntry = make([][]byte, 0, p.bufsize)
	}
	p.zsetEntry = append(p.zsetEntry, []byte("ZADD"))
	p.zsetEntry = append(p.zsetEntry, key)
	p.keyCount++
	p.i = 0
}

func (p *rdbDecoder) Zadd(key []byte, score float64, member []byte) {
	p.zsetEntry = append(p.zsetEntry, []byte(strconv.FormatInt(int64(score), 10)))
	p.zsetEntry = append(p.zsetEntry, member)
	if len(p.zsetEntry) >= p.bufsize {
		cmd := NewCommand(p.zsetEntry...)
		p.client.rdbDecodeCommand(p.client, cmd)
		p.zsetEntry = make([][]byte, 0, p.bufsize)
		p.zsetEntry = append(p.zsetEntry, []byte("ZADD"))
		p.zsetEntry = append(p.zsetEntry, key)
	}
	p.i++
}

// ZSet
func (p *rdbDecoder) EndZSet(key []byte) {
	if len(p.zsetEntry) > 2 {
		cmd := NewCommand(p.zsetEntry...)
		p.client.rdbDecodeCommand(p.client, cmd)
	}
}
