package orderers

type permissiveValidityChecker struct{}

func newPermissiveValidityChecker() *permissiveValidityChecker {
	return &permissiveValidityChecker{}
}

func (pvc *permissiveValidityChecker) Check(_ []byte) error {
	return nil
}
