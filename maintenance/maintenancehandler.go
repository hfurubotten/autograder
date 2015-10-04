package maintenance

type MaintenaceHandler interface {
	ExecutingTime() int
	Execute() error
	RemoveAfterExecute() bool
}
