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

var TypeMap = map[string]reflect.Type{
	"u8":                      reflect.TypeOf(types.U8(0)),
	"u16":                     reflect.TypeOf(types.U16(0)),
	"u32":                     reflect.TypeOf(types.U32(0)),
	"u64":                     reflect.TypeOf(types.U64(0)),
	"u128":                    reflect.TypeOf(types.U128{}),
	"u256":                    reflect.TypeOf(types.U256{}),
	"i8":                      reflect.TypeOf(types.I8(0)),
	"i16":                     reflect.TypeOf(types.I16(0)),
	"i32":                     reflect.TypeOf(types.I32(0)),
	"i64":                     reflect.TypeOf(types.I64(0)),
	"i128":                    reflect.TypeOf(types.I128{}),
	"i256":                    reflect.TypeOf(types.I256{}),
	"bool":                    reflect.TypeOf(types.Bool(false)),
	"text":                    reflect.TypeOf(types.Text("")),
	"hash":                    reflect.TypeOf(types.Hash{}),
	"address":                 reflect.TypeOf(types.Address{}),
	"AccountId":               reflect.TypeOf(types.AccountID{}),
	"T::AccountId":            reflect.TypeOf(types.AccountID{}),
	"Balance":                 reflect.TypeOf(types.BalanceStatus(0)),
	"T::Balance":              reflect.TypeOf(types.U128{}),
	"ElectionCompute":         reflect.TypeOf(types.ElectionCompute(0)),
	"Option<ElectionCompute>": reflect.TypeOf(types.NewOptionU8(types.U8(0))),
	"Vec<u8>":                 reflect.TypeOf([]types.U8{}),
	"BoundedVec<u8, frame_support::traits::ConstU32<64>>": reflect.TypeOf([]types.U8{}),
	"<<T as Config>::Time as Time>::Moment":               reflect.TypeOf(types.U64(0)),
}

func DecodeEventRecordsWithIgnoreError(e types.EventRecordsRaw, m *types.Metadata, t interface{}) error {
	log.Debug(fmt.Sprintf("will decode event records from raw hex: %#x", e))

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

	log.Debug(fmt.Sprintf("found %v events", n))

	// iterate over events
	for i := uint64(0); i < n.Uint64(); i++ {
		log.Debug(fmt.Sprintf("decoding event #%v", i))

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

		log.Debug(fmt.Sprintf("event #%v has EventID %v", i, id))

		// ask metadata for method & event name for event
		moduleName, eventName, err := m.FindEventNamesForEventID(id)
		// moduleName, eventName, err := "System", "ExtrinsicSuccess", nil
		if err != nil {
			return fmt.Errorf("unable to find event with EventID %v in metadata for event #%v: %s", id, i, err)
		}

		log.Debug(fmt.Sprintf("event #%v is in module %v with event name %v", i, moduleName, eventName))

		// check whether name for eventID exists in t
		field := val.FieldByName(fmt.Sprintf("%v_%v", moduleName, eventName))
		var holder reflect.Value
		if !field.IsValid() {
			log.Debug(fmt.Sprintf("unable to find field %v_%v for event #%v with EventID %v", moduleName, eventName, i, id))
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

			log.Debug(fmt.Sprintf("decoded event #%v", i))
		}
	}
	return nil
}

func getNewStructWithEventInfo(m *types.Metadata, moduleName types.Text, eventName types.Text) (reflect.Value, error) {
	switch m.Version {
	case 4:
		return getNewStructWithV13(m.AsMetadataV4, string(moduleName), string(eventName))
	case 7:
		return getNewStructWithV13(m.AsMetadataV7, string(moduleName), string(eventName))
	case 8:
		return getNewStructWithV13(m.AsMetadataV8, string(moduleName), string(eventName))
	case 9:
		return getNewStructWithV13(m.AsMetadataV9, string(moduleName), string(eventName))
	case 10:
		return getNewStructWithV13(m.AsMetadataV10, string(moduleName), string(eventName))
	case 11:
		return getNewStructWithV13(m.AsMetadataV11, string(moduleName), string(eventName))
	case 12:
		return getNewStructWithV13(m.AsMetadataV12, string(moduleName), string(eventName))
	case 13:
		return getNewStructWithV13(m.AsMetadataV13, string(moduleName), string(eventName))
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

func getNewStructWithV13(mv interface{}, moduleName string, eventName string) (reflect.Value, error) {
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
	if !mv.ExistsModuleMetadata(string(moduleName)) {
		return reflect.Value{}, fmt.Errorf("module %v does not exist in metadata", moduleName)
	}
	for _, mod := range mv.Pallets {
		if !mod.HasEvents {
			continue
		}
		if mod.Name != moduleName {
			continue
		}
		eventType := mod.Events.Type.Int64()
		if typ, ok := mv.EfficientLookup[eventType]; ok {
			if len(typ.Def.Variant.Variants) > 0 {
				for _, vars := range typ.Def.Variant.Variants {
					if vars.Name != eventName {
						continue
					}
					return getNewEventStructV14(vars.Fields)
				}
			}
		}
	}
	return reflect.Value{}, fmt.Errorf("unable to find event %s_%s", moduleName, eventName)
}

func getNewEventStructV14(fields []types.Si1Field) (reflect.Value, error) {
	rTypes := make([]reflect.Type, 0, len(fields)+2)
	rTypes = append(rTypes, reflect.TypeOf(types.Phase{}))
	if len(fields) > 0 {
		for _, f := range fields {
			if t, ok := TypeMap[string(f.TypeName)]; ok {
				rTypes = append(rTypes, t)
			} else {
				return reflect.Value{}, fmt.Errorf("unable to find type %v", f.TypeName)
			}
		}
	}
	rTypes = append(rTypes, reflect.TypeOf([]types.Hash{}))
	return makeStruct(rTypes...), nil
}

func getNewEventStructV13(rArgs reflect.Value) (reflect.Value, error) {
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
		if rModuleName.String() != moduleName {
			continue
		}
		rEvents := rModules.Index(i).FieldByName("Events")
		if !rEvents.IsValid() {
			return reflect.Value{}, errors.New("unable to find 'Events' field in module")
		}
		if rEvents.Kind() != reflect.Slice {
			return reflect.Value{}, errors.New("field 'Events' in module is not a slice")
		}
		return findEvent(rEvents, eventName)

	}
	return reflect.Value{}, fmt.Errorf("unable to find module '%s'", moduleName)
}

func findEvent(rEvents reflect.Value, eventName string) (reflect.Value, error) {
	for i := 0; i < rEvents.Len(); i++ {
		rEventName := rEvents.Index(i).FieldByName("Name")
		if !rEventName.IsValid() {
			return reflect.Value{}, errors.New("unable to find 'Name' field in event")
		}
		if rEventName.String() != eventName {
			continue
		}
		rArgs := rEvents.Index(i).FieldByName("Args")
		if !rArgs.IsValid() {
			return reflect.Value{}, errors.New("unable to find 'Args' field in event")
		}
		if rArgs.Kind() != reflect.Slice {
			return reflect.Value{}, errors.New("field 'Args' in event is not a slice")
		}
		return getNewEventStructV13(rArgs)
	}
	return reflect.Value{}, fmt.Errorf("unable to find event '%s'", eventName)
}
