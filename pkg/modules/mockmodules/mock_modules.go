package mockmodules

import (
	"github.com/filecoin-project/mir/stdtypes"
	"github.com/golang/mock/gomock"

	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/modules/mockmodules/internal/mock_internal"
)

// MockPassiveModule is a slightly more user-friendly wrapper around gomock_modules.MockPassiveModule.
type MockPassiveModule struct {
	impl *mock_internal.MockModuleImpl
}

func NewMockPassiveModule(ctrl *gomock.Controller) *MockPassiveModule {
	return &MockPassiveModule{mock_internal.NewMockModuleImpl(ctrl)}
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPassiveModule) EXPECT() *mock_internal.MockModuleImplMockRecorder {
	return m.impl.EXPECT()
}

// ApplyEvents applies a list of input events to the module, making it advance its state
// and returns a (potentially empty) list of output events that the application of the input events results in.
func (m *MockPassiveModule) ApplyEvents(events *stdtypes.EventList) (*stdtypes.EventList, error) {
	return modules.ApplyEventsSequentially(events, m.impl.Event)
}

// ImplementsModule only serves the purpose of indicating that this is a Module and must not be called.
func (m *MockPassiveModule) ImplementsModule() {}
