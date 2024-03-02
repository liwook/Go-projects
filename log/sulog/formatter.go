package sulog

type Formatter interface {
	Format(entry *Entry) error
}
