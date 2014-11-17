package goredis

type SortCommand struct {
	redis  *Redis
	key    string
	by     string
	limit  bool
	offset int
	count  int
	get    []string
	order  string
	alpha  bool
	store  string
}

// http://redis.io/commands/sort
// SORT key [BY pattern] [LIMIT offset count] [GET pattern [GET pattern ...]] [ASC|DESC] [ALPHA] [STORE destination]
func (r *Redis) Sort(key string) *SortCommand {
	return &SortCommand{redis: r, key: key}
}

// The BY option can also take a non-existent key, which causes SORT to skip the sorting operation.
func (s *SortCommand) By(pattern string) *SortCommand {
	s.by = pattern
	return s
}

// This modifier takes the offset argument,
// specifying the number of elements to skip and the count argument,
// specifying the number of elements to return from starting at offset.
func (s *SortCommand) Limit(offset, count int) *SortCommand {
	s.limit = true
	s.offset = offset
	s.count = count
	return s
}

func (s *SortCommand) Get(patterns ...string) *SortCommand {
	s.get = patterns
	return s
}

func (s *SortCommand) ASC() *SortCommand {
	s.order = "ASC"
	return s
}

func (s *SortCommand) DESC() *SortCommand {
	s.order = "DESC"
	return s
}

func (s *SortCommand) Alpha(b bool) *SortCommand {
	s.alpha = b
	return s
}

func (s *SortCommand) Store(destination string) *SortCommand {
	s.store = destination
	return s
}

func (s *SortCommand) Run() (*Reply, error) {
	args := packArgs("SORT", s.key)
	if s.by != "" {
		args = append(args, "BY", s.by)
	}
	if s.limit {
		args = append(args, "LIMIT", s.offset, s.count)
	}
	if s.get != nil && len(s.get) > 0 {
		for _, pattern := range s.get {
			args = append(args, "GET", pattern)
		}
	}
	if s.order != "" {
		args = append(args, s.order)
	}
	if s.alpha {
		args = append(args, "ALPHA")
	}
	if s.store != "" {
		args = append(args, "STORE", s.store)
	}
	return s.redis.ExecuteCommand(args...)
}
