package ips

import (
	"fmt"
	"math/rand"
	"time"
)

type VoseRandom struct {
	dist        map[string]float64
	table_prob  map[string]float64
	table_alias map[string]string
	rnd         *rand.Rand
}

type worklist []string

func (w *worklist) push(i string) {
	*w = append(*w, i)
}

func (w *worklist) pop() string {
	l := len(*w) - 1
	n := (*w)[l]
	(*w) = (*w)[:l]
	return n
}

func FlattenProbs(dist map[string]int) map[string]float64 {
	m := make(map[string]float64, len(dist))
	totW := 0
	for _, prob := range dist {
		totW += prob
	}

	for r, prob := range dist {
		m[r] = float64(prob) / float64(totW)
	}

	return m
}

func NewVoseRandom(dist map[string]float64) *VoseRandom {
	v := &VoseRandom{
		dist:        dist,
		table_prob:  make(map[string]float64, len(dist)),
		table_alias: make(map[string]string, 0),
	}

	v.rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

	lenfloat := float64(len(dist))
	average := 1.0 / lenfloat
	scaled_prob := make(map[string]float64, len(dist))
	small := make(worklist, 0)
	large := make(worklist, 0)
	for value, prob := range v.dist {
		scaled_prob[value] = prob
		if prob > average {
			large.push(value)
		} else {
			small.push(value)
		}
	}

	for len(large) > 0 && len(small) > 0 {
		s := small.pop()
		l := large.pop()
		v.table_prob[s] = scaled_prob[s]
		v.table_alias[s] = l

		v.table_prob[s] = scaled_prob[s] * lenfloat
		v.table_alias[s] = l
		scaled_prob[l] = (scaled_prob[l] + scaled_prob[s]) - average
		if scaled_prob[l] < average {
			small.push(l)
		} else {
			large.push(l)
		}
		fmt.Printf("small: %v\nlarge:%v\n", small, large)
	}

	for len(large) > 0 {
		g := large.pop()
		v.table_prob[g] = 1
	}

	for len(small) > 0 {
		l := small.pop()
		v.table_prob[l] = 1
	}

	return v
}

func (v *VoseRandom) Next() string {

	fmt.Printf("%v\n%v\n", v.table_prob, v.table_alias)

	i := v.rnd.Intn(len(v.table_prob))
	var k string
	for k = range v.table_prob {
		if i == 0 {
			break
		}
		i--
	}

	if v.rnd.Float64() <= v.table_prob[k] {
		return k
	}

	return v.table_alias[k]
}
