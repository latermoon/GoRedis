package rocks

import (
	"bytes"
	"errors"
	"github.com/tecbot/gorocksdb"
	"sync"
)

// SortedSet
// +key,z = ""
// z[key]m member = score
// z[key]s score member = ""
type SortedSetElement struct {
	db  *DB
	key []byte
	mu  sync.RWMutex
}

type SortedSetEnumerateFunc func(i int, score, member []byte, quit *bool)

func NewSortedSetElement(db *DB, key []byte) *SortedSetElement {
	s := &SortedSetElement{db: db, key: key}
	return s
}

// http://redis.io/commands/zadd#return-value
func (s *SortedSetElement) Add(scoreMembers ...[]byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := len(scoreMembers)
	if count < 2 || count%2 != 0 {
		return 0, errors.New("invalid score/member pairs")
	}

	batch := gorocksdb.NewWriteBatch()
	defer batch.Destroy()

	added := 0
	for i := 0; i < count; i += 2 {
		score, member := scoreMembers[i], scoreMembers[i+1]
		skey, mkey := s.scoreKey(score, member), s.memberKey(member)

		// remove old score key
		oldscore, err := s.db.RawGet(mkey)
		if err != nil {
			return 0, err
		} else if oldscore != nil {
			batch.Delete(skey)
		} else {
			added++
		}

		// put new value
		batch.Put(mkey, score)
		batch.Put(skey, nil)
	}

	return added, s.db.WriteBatch(batch)
}

func (s *SortedSetElement) Score(member []byte) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.score(member)
}

func (s *SortedSetElement) score(member []byte) ([]byte, error) {
	return s.db.RawGet(s.memberKey(member))
}

func (s *SortedSetElement) Remove(members ...[]byte) (int, error) {
	return 0, nil
}

// +key,z = ""
func (s *SortedSetElement) rawKey() []byte {
	return rawKey(s.key, SORTEDSET)
}

// z[key]
func (s *SortedSetElement) keyPrefix() []byte {
	return bytes.Join([][]byte{[]byte{SORTEDSET}, SOK, s.key, EOK}, nil)
}

func (s *SortedSetElement) memberKey(member []byte) []byte {
	return bytes.Join([][]byte{s.keyPrefix(), []byte{'m'}, member}, nil)
}

func (s *SortedSetElement) scoreKey(score, member []byte) []byte {
	return bytes.Join([][]byte{s.keyPrefix(), []byte{'s'}, score, []byte{' '}, member}, nil)
}

// split (z[key]s score member) into (score, member)
func (s *SortedSetElement) splitScoreKey(skey []byte) ([]byte, []byte, error) {
	buf := bytes.TrimPrefix(skey, s.keyPrefix())
	pairs := bytes.Split(buf[1:], []byte{' '}) // skip score mark 's'
	if len(pairs) != 2 {
		return nil, nil, errors.New("invalid score/member key: " + string(skey))
	}
	return pairs[0], pairs[1], nil
}

// split (z[key]m member) into (member)
func (s *SortedSetElement) splitMemberKey(mkey []byte) ([]byte, error) {
	buf := bytes.TrimPrefix(mkey, s.keyPrefix())
	return buf[1:], nil // skip member mark 'm'
}

func (s *SortedSetElement) RemoveByScore(min, max []byte) (int, error) {
	return 0, nil
}

func (s *SortedSetElement) RangeByScore(min, max []byte, fn SortedSetEnumerateFunc) error {
	return nil
}

func (s *SortedSetElement) RangeByMember(min, max []byte, fu SortedSetEnumerateFunc) error {
	return nil
}
