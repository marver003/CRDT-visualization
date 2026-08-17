package simulator

type Operation interface {
	Execute(*Simulator) error
}
