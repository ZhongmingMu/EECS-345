package libkademlia

import (
	"container/list"
)

type K_Buckets struct {
	bucket 		*list.List
	distance 	int				//0 based
//	num			int
	size  		int 
}

func NewKBuckets(k int, dis int) *K_Buckets {
	nb := new(K_Buckets)			// create a new k-bucket
	nb.distance = dis  				// set distance of this bucket
	l := list.New()
	nb.bucket = l                   // set the list 
	nb.size = k                     // set the bucket size

	return nb
}

// add the contact in the tail
func (buc * K_Buckets) AddTail(c *Contact) {
	buc.bucket.PushBack(c)
}

// check whether the list is full
func (buc * K_Buckets) CheckFull() bool {
	num := buc.Count()
	return num >= buc.size
}

//remove the head of the list
func (buc * K_Buckets) RemoveHead() {
	buc.bucket.Remove(buc.bucket.Front())
}

// return the current size of the bucket
func (buc * K_Buckets) Count() int {
	return buc.bucket.Len()
}

//move to tail
func (buc * K_Buckets) MoveToTail(c *Contact) {
	ele_ptr := new (list.Element)
	ele_ptr.Value = *c 
	buc.bucket.MoveToBack(ele_ptr)
}

//get first element
func (buc * K_Buckets) GetHead() *Contact {
	return buc.bucket.Front().Value.(*Contact)
}

