package test

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/ethereum/go-ethereum/log"
	"reflect"
)

type EventProviderRegisterResourceSuccess struct {
	Phase            types.Phase
	AccountId        types.AccountID
	Index            types.U64
	PeerId           string
	Cpu              types.U64
	Memory           types.U64
	System           string
	CpuModel         string
	PriceHour        types.U128
	RentDurationHour types.U32
	Topics           []types.Hash
}

type EventResourceOrderCreateOrderSuccess struct {
	Phase         types.Phase
	AccountId     types.AccountID
	OrderIndex    types.U64
	ResourceIndex types.U64
	Duration      types.U32
	PublicKey     string
	Topics        []types.Hash
}

type EventResourceOrderOrderExecSuccess struct {
	Phase          types.Phase
	AccountId      types.AccountID
	OrderIndex     types.U64
	ResourceIndex  types.U64
	AgreementIndex types.U64
	Topics         []types.Hash
}

type EventResourceOrderReNewOrderSuccess struct {
	Phase          types.Phase
	AccountId      types.AccountID
	OrderIndex     types.U64
	ResourceIndex  types.U64
	AgreementIndex types.U64
	Topics         []types.Hash
}

type EventResourceOrderWithdrawLockedOrderPriceSuccess struct {
	Phase      types.Phase
	AccountId  types.AccountID
	OrderIndex types.U64
	OrderPrice types.U128
	Topics     []types.Hash
}

type EventResourceOrderFreeResourceProcessed struct {
	Phase      types.Phase
	OrderIndex types.U64
	PeerId     string
	Topics     []types.Hash
}

type EventResourceOrderFreeResourceApplied struct {
	Phase      types.Phase
	AccountId  types.AccountID
	OrderIndex types.U64
	Cpu        types.U64
	Memory     types.U64
	Duration   types.U32
	DeployType types.U32
	PublicKey  string
	Topics     []types.Hash
}

type EventMarketMoney struct {
	Phase  types.Phase
	Money  types.U128
	Topics []types.Hash
}

type EventElectionProviderMultiPhase_SignedPhaseStarted struct {
	Phase       types.Phase
	SignedPhase types.U32
	Topics      []types.Hash
}

type EventRegisterGatewayNodeSuccess struct {
	Phase       types.Phase
	AccountId   types.AccountID
	BlockNumber types.BlockNumber
	PeerId      []types.U8
	Topics      []types.Hash
}

type MyEventRecords struct {
	types.EventRecords
	Provider_RegisterResourceSuccess              []EventProviderRegisterResourceSuccess //nolint:stylecheck,golint
	ResourceOrder_CreateOrderSuccess              []EventResourceOrderCreateOrderSuccess //nolint:stylecheck,golint
	ResourceOrder_OrderExecSuccess                []EventResourceOrderOrderExecSuccess
	ResourceOrder_ReNewOrderSuccess               []EventResourceOrderReNewOrderSuccess
	ResourceOrder_WithdrawLockedOrderPriceSuccess []EventResourceOrderWithdrawLockedOrderPriceSuccess
	Gateway_RegisterGatewayNodeSuccess            []EventRegisterGatewayNodeSuccess
	ResourceOrder_FreeResourceProcessed           []EventResourceOrderFreeResourceProcessed
	ResourceOrder_FreeResourceApplied             []EventResourceOrderFreeResourceApplied
	ElectionProviderMultiPhase_SignedPhaseStarted []EventElectionProviderMultiPhase_SignedPhaseStarted
	Market_Money                                  []EventMarketMoney
}

func DecodeEventRecordsWithIgnoreError(e types.EventRecordsRaw, m *types.Metadata, t interface{}) error {
	fmt.Println(fmt.Sprintf("will decode event records from raw hex: %#x", e))

	// ensure t is a pointer
	ttyp := reflect.TypeOf(t)
	if ttyp.Kind() != reflect.Ptr {
		return errors.New("target must be a pointer, but is " + fmt.Sprint(ttyp))
	}
	// ensure t is not a nil pointer
	tval := reflect.ValueOf(t)
	if tval.IsNil() {
		return errors.New("target is a nil pointer")
	}
	val := tval.Elem()
	typ := val.Type()
	// ensure val can be set
	if !val.CanSet() {
		return fmt.Errorf("unsettable value %v", typ)
	}
	// ensure val points to a struct
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("target must point to a struct, but is " + fmt.Sprint(typ))
	}

	decoder := scale.NewDecoder(bytes.NewReader(e))

	// determine number of events
	n, err := decoder.DecodeUintCompact()
	if err != nil {
		return err
	}

	fmt.Println(fmt.Sprintf("found %v events", n))

	// iterate over events
	for i := uint64(0); i < n.Uint64(); i++ {
		fmt.Println(fmt.Sprintf("decoding event #%v", i))

		// decode Phase
		phase := types.Phase{}
		err := decoder.Decode(&phase)
		if err != nil {
			return fmt.Errorf("unable to decode Phase for event #%v: %v", i, err)
		}

		// decode EventID
		id := types.EventID{}
		err = decoder.Decode(&id)
		if err != nil {
			return fmt.Errorf("unable to decode EventID for event #%v: %v", i, err)
		}

		fmt.Println(fmt.Sprintf("event #%v has EventID %v", i, id))

		// ask metadata for method & event name for event
		moduleName, eventName, err := m.FindEventNamesForEventID(id)
		// moduleName, eventName, err := "System", "ExtrinsicSuccess", nil
		if err != nil {
			//return fmt.Errorf("unable to find event with EventID %v in metadata for event #%v: %s", id, i, err)
			log.Warn("unable to find event with EventID %v in metadata for event #%v: %s", id, i, err)

			return err
		}

		fmt.Println(fmt.Sprintf("event #%v is in module %v with event name %v", i, moduleName, eventName))

		// check whether name for eventID exists in t
		field := val.FieldByName(fmt.Sprintf("%v_%v", moduleName, eventName))
		var holder reflect.Value
		if !field.IsValid() {
			fmt.Println(fmt.Sprintf("unable to find field %v_%v for event #%v with EventID %v ", moduleName, eventName, i, id))
			holder, err = getNewStructWithEventInfo(m, moduleName, eventName)
			if err != nil {
				return err
			}
		} else {
			// create a pointer to with the correct type that will hold the decoded event
			holder = reflect.New(field.Type().Elem())
		}

		// ensure first field is for Phase, last field is for Topics
		numFields := holder.Elem().NumField()
		if numFields < 2 {
			return fmt.Errorf("expected event #%v with EventID %v, field %v_%v to have at least 2 fields "+
				"(for Phase and Topics), but has %v fields", i, id, moduleName, eventName, numFields)
		}
		phaseField := holder.Elem().FieldByIndex([]int{0})
		if phaseField.Type() != reflect.TypeOf(phase) {
			return fmt.Errorf("expected the first field of event #%v with EventID %v, field %v_%v to be of type "+
				"types.Phase, but got %v", i, id, moduleName, eventName, phaseField.Type())
		}
		topicsField := holder.Elem().FieldByIndex([]int{numFields - 1})
		if topicsField.Type() != reflect.TypeOf([]types.Hash{}) {
			return fmt.Errorf("expected the last field of event #%v with EventID %v, field %v_%v to be of type "+
				"[]types.Hash for Topics, but got %v", i, id, moduleName, eventName, topicsField.Type())
		}

		// set the phase we decoded earlier
		phaseField.Set(reflect.ValueOf(phase))

		// set the remaining fields
		for j := 1; j < numFields; j++ {
			err = decoder.Decode(holder.Elem().FieldByIndex([]int{j}).Addr().Interface())
			if err != nil {
				return fmt.Errorf("unable to decode field %v event #%v with EventID %v, field %v_%v: %v", j, i, id, moduleName,
					eventName, err)
			}
		}

		if field.IsValid() {
			// add the decoded event to the slice
			field.Set(reflect.Append(field, holder.Elem()))

			fmt.Println(fmt.Sprintf("decoded event #%v", i))
		}
	}
	return nil
}

func getNewStructWithEventInfo(m *types.Metadata, moduleName types.Text, eventName types.Text) (reflect.Value, error) {
	fmt.Println(m.Version)
	switch m.Version {
	case 4:
		return getNewStructWithVersion(m.AsMetadataV4, string(moduleName), string(eventName))
	case 7:
		return getNewStructWithVersion(m.AsMetadataV7, string(moduleName), string(eventName))
	case 8:
		return getNewStructWithVersion(m.AsMetadataV8, string(moduleName), string(eventName))
	case 9:
		return getNewStructWithVersion(m.AsMetadataV9, string(moduleName), string(eventName))
	case 10:
		return getNewStructWithVersion(m.AsMetadataV10, string(moduleName), string(eventName))
	case 11:
		return getNewStructWithVersion(m.AsMetadataV11, string(moduleName), string(eventName))
	case 12:
		return getNewStructWithVersion(m.AsMetadataV12, string(moduleName), string(eventName))
	case 13:
		return getNewStructWithVersion(m.AsMetadataV13, string(moduleName), string(eventName))
	case 14:
		return getNewStructWithV14(m.AsMetadataV14, moduleName, eventName)
	default:
		return reflect.Value{}, fmt.Errorf("unsupported metadata version %v", m.Version)
	}
}

func makeStruct(rTypes ...reflect.Type) reflect.Value {
	var sfs []reflect.StructField
	for i, t := range rTypes {
		sf := reflect.StructField{
			Name: fmt.Sprintf("F%d", i+1),
			Type: t,
		}
		sfs = append(sfs, sf)
	}
	return reflect.New(reflect.StructOf(sfs))
}

var TypeMap = map[string]reflect.Type{
	"u8":        reflect.TypeOf(types.U8(0)),
	"u16":       reflect.TypeOf(types.U16(0)),
	"u32":       reflect.TypeOf(types.U32(0)),
	"u64":       reflect.TypeOf(types.U64(0)),
	"u128":      reflect.TypeOf(types.U128{}),
	"u256":      reflect.TypeOf(types.U256{}),
	"i8":        reflect.TypeOf(types.I8(0)),
	"i16":       reflect.TypeOf(types.I16(0)),
	"i32":       reflect.TypeOf(types.I32(0)),
	"i64":       reflect.TypeOf(types.I64(0)),
	"i128":      reflect.TypeOf(types.I128{}),
	"i256":      reflect.TypeOf(types.I256{}),
	"bool":      reflect.TypeOf(types.Bool(false)),
	"text":      reflect.TypeOf(types.Text("")),
	"hash":      reflect.TypeOf(types.Hash{}),
	"address":   reflect.TypeOf(types.Address{}),
	"AccountId": reflect.TypeOf(types.AccountID{}),
}

func getNewStructWithVersion(mv interface{}, moduleName string, eventName string) (reflect.Value, error) {
	v := reflect.ValueOf(mv)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	rModules := v.FieldByName("Modules")
	if !rModules.IsValid() {
		return reflect.Value{}, errors.New("unable to find 'Modules' field in metadata")
	}
	if rModules.Kind() != reflect.Slice {
		return reflect.Value{}, errors.New("field 'Modules' in metadata is not a slice")
	}
	return findModule(rModules, moduleName, eventName)
}

func getNewStructWithV14(mv types.MetadataV14, moduleName types.Text, eventName types.Text) (reflect.Value, error) {
	for _, p := range mv.Pallets {
		if p.Name == moduleName {
			if p.HasEvents {
				tIndex := p.Events.Type.Int64()
				for _, t := range mv.Lookup.Types {
					if t.ID.Int64() == tIndex {
						if t.Type.Def.IsVariant {
							for _, v := range t.Type.Def.Variant.Variants {
								if v.Name == eventName {
									fmt.Println(fmt.Sprintf("found event %v", v.Name))
									for _, f := range v.Fields {
										fmt.Println(fmt.Sprintf("found field %v", f.TypeName))
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return reflect.Value{}, fmt.Errorf("unable to find event %s_%s", moduleName, eventName)
}

func getNewEventStruct(rArgs reflect.Value) (reflect.Value, error) {
	rTypes := make([]reflect.Type, 0, rArgs.Len()+2)
	rTypes = append(rTypes, reflect.TypeOf(types.Phase{}))
	if rArgs.Len() != 0 {
		for k := 0; k < rArgs.Len(); k++ {
			rArg := rArgs.Index(k)
			if !rArg.IsValid() {
				return reflect.Value{}, errors.New("unable to find arg in event")
			}
			arg := rArg.String()
			a, ok := TypeMap[arg]
			if !ok {
				return reflect.Value{}, fmt.Errorf("unable to find type for arg %v", arg)
			}
			rTypes = append(rTypes, a)
		}
	}
	rTypes = append(rTypes, reflect.TypeOf([]types.Hash{}))
	return makeStruct(rTypes...), nil
}

func findModule(rModules reflect.Value, moduleName string, eventName string) (reflect.Value, error) {
	for i := 0; i < rModules.Len(); i++ {
		rModuleName := rModules.Index(i).FieldByName("Name")
		if !rModuleName.IsValid() {
			return reflect.Value{}, errors.New("unable to find 'Name' field in module")
		}
		if rModuleName.String() == moduleName {
			rEvents := rModules.Index(i).FieldByName("Events")
			if !rEvents.IsValid() {
				return reflect.Value{}, errors.New("unable to find 'Events' field in module")
			}
			if rEvents.Kind() != reflect.Slice {
				return reflect.Value{}, errors.New("field 'Events' in module is not a slice")
			}
			return findEvent(rEvents, eventName)
		}
	}
	return reflect.Value{}, fmt.Errorf("unable to find module '%s'", moduleName)
}

func findEvent(rEvents reflect.Value, eventName string) (reflect.Value, error) {
	for i := 0; i < rEvents.Len(); i++ {
		rEventName := rEvents.Index(i).FieldByName("Name")
		if !rEventName.IsValid() {
			return reflect.Value{}, errors.New("unable to find 'Name' field in event")
		}
		if rEventName.String() == eventName {
			rArgs := rEvents.Index(i).FieldByName("Args")
			if !rArgs.IsValid() {
				return reflect.Value{}, errors.New("unable to find 'Args' field in event")
			}
			if rArgs.Kind() != reflect.Slice {
				return reflect.Value{}, errors.New("field 'Args' in event is not a slice")
			}
			return getNewEventStruct(rArgs)
		}
	}
	return reflect.Value{}, fmt.Errorf("unable to find event '%s'", eventName)
}
