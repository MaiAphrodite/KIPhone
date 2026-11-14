package main

// unionFind groups student IDs that share one bank account (R2).
type unionFind struct {
	parent map[uint]uint
}

func newUnionFind() *unionFind { return &unionFind{parent: map[uint]uint{}} }

func (u *unionFind) find(x uint) uint {
	p, ok := u.parent[x]
	if !ok {
		u.parent[x] = x
		return x
	}
	if p != x {
		u.parent[x] = u.find(p)
	}
	return u.parent[x]
}

func (u *unionFind) union(a, b uint) {
	ra, rb := u.find(a), u.find(b)
	if ra != rb {
		u.parent[ra] = rb
	}
}
