package main

import (
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"github.com/OneOfOne/xxhash"
	"github.com/satori/go.uuid"
	"gopkg.in/redis.v3"
)

type cacheHint struct {
	userID RecordID
	table  string
	params map[string]string
}
type CacheHinter interface {
	cacheHint() cacheHint
}

func cacheHintMake(table string, userID RecordID, params map[string]interface{}) cacheHint {
	paramsStrs := make(map[string]string, len(params))
	for k, v := range params {
		strv := fmt.Sprint("%v", v)
		if strv != "" {
			paramsStrs[k] = strv
		}
	}
	return cacheHint{table: table, userID: userID, params: paramsStrs}
}

type ETag string

type Cacheable struct {
	Bytes []byte
	ETag  ETag
}

const ETagNil ETag = ""

func cacheMakeKeyFromQuery(query string, args []interface{}) string {
	h := xxhash.NewS64(0XBABE)
	h.Write([]byte(query))
	for _, arg := range args {
		h.Write([]byte{0})
		h.Write([]byte(fmt.Sprintf("%v", arg)))
	}
	return fmt.Sprintf("q.%v", h.Sum64())
}

func cacheGet(key string) (Cacheable, error) {
	hash, err := services.redis.HMGet(key, "data", "etag").Result()
	if err == redis.Nil || hash[0] == nil {
		return Cacheable{}, ErrNotFound
	} else if err != nil {
		return Cacheable{}, fmt.Errorf("error fetching from cache with key %v: %v", key, err)
	}

	bytes := []byte(hash[0].(string))
	etag := ETag(hash[1].(string))
	return Cacheable{bytes, etag}, nil
}

func cacheSet(key string, bytes []byte, expiration time.Duration, hint cacheHint) (ETag, error) {

	etag := cacheMakeEtag()
	rkey := cacheMakeInvalidateKey(hint)
	_, err := services.redis.Pipelined(func(pipe *redis.Pipeline) error {
		pipe.HMSet(key, "data", string(bytes), "etag", string(etag))
		pipe.Expire(key, expiration)
		pipe.SAdd(rkey, key)
		return nil
	})
	return etag, err
}

func cacheMakeEtag() ETag {
	return ETag(uuid.NewV4().String())
}

func cacheMakeInvalidateKey(hint cacheHint) string {
	rkey := fmt.Sprintf("rkeys|%s|%s|%s", hint.table, hint.userID)

	for _, k := range sortedKeyList(hint.params) {
		v := hint.params[k]
		rkey += "|" + k + "=" + v
	}
	return rkey
}

func sortedKeyList(m map[string]string) []string {
	sortedKeys := make([]string, 0, len(m))
	for k := range m {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	return sortedKeys
}

func cacheInvalidate(cacheHint cacheHint) error {
	relatedCacheHints := cacheMakeRelatedCacheHints(cacheHint)
	rkeys := []string{}
	for _, h := range relatedCacheHints {
		rkeys = append(rkeys, cacheMakeInvalidateKey(h))
	}
	script := `
		local num_deleted = 0;
		for i=1, #KEYS do
	    local keys = redis.call('smembers', KEYS[i]);
	    if table.getn(keys) > 0 then
	      num_deleted = num_deleted + redis.call('del', unpack(keys));
	      redis.call('del', KEYS[i]);
	    else
	    end
		end
	  return num_deleted;
	`
	ret, err := services.redis.Eval(script, rkeys, nil).Result()
	if err != nil {
		return fmt.Errorf("failed to delete cached keys in %v: %v", rkeys, err)
	} else {
		log.Printf("invalidated %v cache entries for %v", ret, cacheHint)
	}
	return nil
}

func cacheMakeRelatedCacheHints(hint cacheHint) []cacheHint {
	allSubparams := submaps(hint.params)
	related := make([]cacheHint, 1+len(allSubparams))
	related[0] = hint

	for i, subparams := range allSubparams {
		related[i+1] = cacheHint{userID: hint.userID, table: hint.table, params: subparams}
	}
	return related
}

func submaps(m map[string]string) []map[string]string {
	keys := sortedKeyList(m)
	subsets := subsets(keys)
	submaps := make([]map[string]string, len(subsets))
	for i, subset := range subsets {
		submap := make(map[string]string, len(subset))
		for _, k := range subset {
			submap[k] = m[k]
		}
		submaps[i] = submap
	}

	return submaps
}

const subsetsMaxSize = 4 // number of subsets is 2^size so we must put the limit really low; in reality it should not be higher than 1 or 2
func subsets(list []string) [][]string {
	size := len(list)
	if size > subsetsMaxSize {
		panic(fmt.Sprintf("subsets: list is bigger than max size %v: %#v", size, list))
	}
	num_sets := int(math.Pow(float64(2), float64(size)))
	subsets := make([][]string, 1, num_sets) //1st elem will be empty set

	for i := 1; i < num_sets; i++ {
		max_bits := int(math.Floor(math.Log2(float64(i))))
		subset := make([]string, 0, max_bits)
		for j := 0; j <= max_bits; j++ {
			bit := (i >> uint(j)) & 1 // bit value at position j
			if bit == 1 {
				subset = append(subset, list[j])
			}
		}
		subsets = append(subsets, subset)
	}

	return subsets
}

//
