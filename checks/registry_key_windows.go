package checks

import "errors"

func init() {
	availableChecks["registry"] = func(args map[string]interface{}) {
		return RegistryKeyChecker{}.FromArgs(args)
	}

	availableChecks["irp-stack-size"] = func(args map[string]interface{}) {
		args["root"] = "HKLM"
		args["path"] = "SYSTEM\\CurrentControlSet\\services\\LanmanServer\\Parameters"
		args["key"] = "IRPStackSize"
		args["value_type"] = "DWORD"
		return nil, nil
	}
}

// these are the valid windows root key names
var rootKeys = map[string]registry.Key{
	"HKEY_LOCAL_MACHINE": registry.LOCAL_MACHINE,
	"HKLM":               registry.LOCAL_MACHINE,

	"HKEY_CURRENT_CONFIG": registry.CURRENT_CONFIG,
	"HKCC":                registry.CURRENT_CONFIG,

	"HKEY_CLASSES_ROOT": registry.CLASSES_ROOT,
	"HKCR":              registry.CLASSES_ROOT,

	"HKEY_CURRENT_USER": registry.CURRENT_USER,
	"HKCU":              registry.CURRENT_USER,

	"HKEY_USERS": registry.USERS,
	"HKU":        registry.USERS,

	"HKEY_PERFORMANCE_DATA": registry.PERFORMANCE_DATA,
	"HKEY_DYN_DATA":         registry.PERFORMANCE_DATA,
}

// RegistryKeyChecker verifies the value of a given registry key is correct or
// exists.
type RegistryKeyChecker struct {
	Root      string
	Path      string
	Key       string
	ValueType string `mapstructure:"value_type"`
	Value     interface{}
}

// valueErr is a utility function for printing the key path value and what we expected + got
func (rkc RegistryKeyChecker) valueErr(value interface{}, expected interface{}) error {
	return fmt.Errorf("incorrect value for %s:%s\\%s, got: %v expected: %v", rkc.Root, rkc.Path, rkc.Key, val, intValue)
}

// Check implements Checker for RegistryKeyChecker
func (rkc RegistryKeyChecker) Check() error {
	reg, exists := rootKeys[rkc.Root]
	if !exists {
		return fmt.Errorf("%s is not a valid root key, valid root keys can be found here: https://docs.microsoft.com/en-us/windows/desktop/sysinfo/predefined-keys")
	}

	key, err := registry.OpenKey(reg, rkc.Path, registry.QUERY_VALUE)
	if err != nil {
		return fmt.Errorf("unable to open path: %s: %s", rkc.Path, err)
	}

	defer key.Close()

	// If no value provided we don't want to check the and opening the key
	// successfully was enough to verify a success.
	if rkc.Value == "" {
		return nil
	}

	if rkc.ValueType == "" {
		return errors.New("cannot check value without value_type argument")
	}

	if rkc.Key == "" {
		return errors.New("cannot check value without key argument")
	}

	switch {
	case rkc.ValueType == "DWORD" || rkc.ValueType == "QWORD":
		val, _, err := key.GetIntegerValue(rkc.Key)
		if err != nil {
			return err
		}

		intValue, ok := rkc.Value.(int)
		if !ok {
			return fmt.Errorf("can only compare %s with an integer, got %T", rkc.ValueType, rkc.Value)
		}

		if uint64(intValue) != val {
			return rkc.valueErr(val, intValue)
		}

		return nil
	case rkc.ValueType == "BINARY":
		val, _, err := key.GetBinaryValue(rkc.Key)
		if err != nil {
			return err
		}

		strValue, ok := rkc.Value.(string)
		if !ok {
			return fmt.Errorf("can only compare %s with a string of binary, got %T", rkc.ValueType, rkc.Value)
		}

		if []byte(strValue) != val {
			return rkc.valueErr(val, []byte(strValue))
		}

		return nil
	case rkc.ValueType == "SZ" || rkc.ValueType == "EXPAND_SZ":
		val, _, err := key.GetStringValue(rkc.Key)
		if err != nil {
			return err
		}

		strValue, ok := rkc.Value.(string)
		if !ok {
			return fmt.Errorf("can only compare %s with a string, got %T", rkc.ValueType, rkc.Value)
		}

		if strValue != val {
			return rkc.valueErr(val, strValue)
		}

		return nil
	case rkc.ValueType == "MULTI_SZ":
		val, _, err := key.GetStringsValue(rkc.Key)
		if err != nil {
			return err
		}

		strValues, ok := rkc.Value.([]string)
		if !ok {
			return fmt.Errorf("can only compare %s with a list of strings, got %T", rkc.ValueType, rkc.Value)
		}

		if len(strValues) != len(val) {
			return fmt.valueErr(val, strValues)
		}

		for i := range val {
			if val[i] != strValues[i] {
				return fmt.valueErr(val, strValues)
			}
		}

		return nil
	default:
		return fmt.Errorf("%s is not a known registry key value type", rck.ValueType)
	}
}

// FromArgs implements Argable for RegistryKeyChecker
func (rkc RegistryKeyChecker) FromArgs(args map[string]interface{}) (Checker, error) {
	if err := requiredArgs(args, "registry", "key"); err != nil {
		return nil, err
	}

	if err := decodeFromArgs(args, &rkc); err != nil {
		return nil, err
	}

	return rkc, nil
}