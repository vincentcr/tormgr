package main

import (
	"fmt"
	"log"
	"time"

	"github.com/OneOfOne/xxhash"
	"github.com/satori/go.uuid"
	"gopkg.in/redis.v3"
)

type cacheHint struct {
	userID   RecordID
	table    string
	recordID RecordID
}

func cacheMakeKeyFromQuery(query string, args []interface{}) string {
	h := xxhash.NewS64(0XBABE)
	h.Write([]byte(query))
	for _, arg := range args {
		h.Write([]byte{0})
		h.Write([]byte(fmt.Sprintf("%v", arg)))
	}
	return fmt.Sprintf("q.%v", h.Sum64())
}

func cacheGet(key string) ([]byte, string, error) {
	hash, err := services.redis.HMGet(key, "data", "etag").Result()
	if err == redis.Nil || hash[0] == nil {
		return nil, "", nil
	} else if err != nil {
		return nil, "", fmt.Errorf("error fetching from cache with key %v: %v", key, err)
	} else {
		data := hash[0].(string)
		etag := hash[1].(string)
		return []byte(data), etag, nil
	}
}

func cacheSet(key string, data []byte, expiration time.Duration, hint cacheHint) (string, error) {
	etag := cacheMakeEtag()
	_, err := services.redis.Pipelined(func(pipe *redis.Pipeline) error {
		pipe.HMSet(key, "data", string(data), "etag", etag)
		pipe.Expire(key, expiration)
		rkey := cacheMakeInvalidateKey(hint)
		pipe.SAdd(rkey, key)
		return nil
	})
	return etag, err
}

func cacheMakeEtag() string {
	return uuid.NewV4().String()
}

func cacheMakeInvalidateKey(hint cacheHint) string {
	return fmt.Sprintf("rkeys|%s|%s|%s", hint.table, hint.userID, hint.recordID)
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
	return []cacheHint{
		hint,
		cacheHint{table: hint.table, userID: hint.userID},
	}
}
