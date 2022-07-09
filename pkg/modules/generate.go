package modules

//go:generate mockgen -destination ./mock_modules/mock_modules.mock.go . Module,PassiveModule

////go:generate mockgen -destination ./mock_modules/modules.mock.go -source ./modules.go
////go:generate mockgen -destination ./mock_modules/passivemodule.mock.go -source ./passivemodule.go -aux_files modules=github.com/filecoin-project/mir/pkg/modules
