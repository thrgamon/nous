package main

type Resource struct {
  ID uint
  Link string
  Name string
  Rank int
}

type ResourceRepo struct {
  storage map[uint]Resource
}

func NewResourceRepo() *ResourceRepo{
  r1 := Resource{Link: "https://hello.com", Name: "Hello is Cool", Rank: 5, ID: 1}
  r2 := Resource{Link: "https://tomgamon.com", Name: "Wow so nice", Rank: 5, ID: 2}
  var repo ResourceRepo
  repo.storage = make(map[uint]Resource)
  repo.storage[1] = r1
  repo.storage[2] = r2
  return &repo
}

func (rr ResourceRepo) Get(id uint) Resource{
  return rr.storage[id]
}

func (rr ResourceRepo) Upvote(id uint) Resource{
  resource := rr.storage[id]
  resource.Rank = resource.Rank + 1
  rr.storage[id] = resource
  return rr.storage[id]
}

func (rr ResourceRepo) Downvote(id uint) Resource{
  resource := rr.storage[id]
  resource.Rank = resource.Rank - 1
  rr.storage[id] = resource
  return rr.storage[id]
}
