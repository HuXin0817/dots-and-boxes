package message

import (
	"fmt"
)

const topicNumber = 5

type RedisPartition int

func (r RedisPartition) ListKey() string {
	return fmt.Sprintf("Partition-%d", r)
}

func (r RedisPartition) OwnerKey() string {
	return fmt.Sprintf("Partition %d Owner", r)
}

func (r RedisPartition) LockName() string {
	return fmt.Sprintln(r, "Lock")
}

var RedisPartitions []RedisPartition

func init() {
	for i := range topicNumber {
		RedisPartitions = append(RedisPartitions, RedisPartition(i+1))
	}
}
